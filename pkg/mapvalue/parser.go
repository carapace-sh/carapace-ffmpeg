package mapvalue

import (
	"fmt"
	"strconv"
	"strings"
)

// MapValue represents a parsed -map value.
//
// Syntax: [-]input_file_id[:stream_specifier][:view_specifier][:?] | [linklabel]
type MapValue struct {
	Negate       bool   // true if prefixed with '-'
	FileIndex    int    // input file index (0-based)
	HasSpecifier bool   // true if stream specifier is present
	Specifier    string // raw stream specifier string (e.g. "a:1", "v")
	HasView      bool   // true if view specifier is present
	ViewSpec     string // raw view specifier
	Optional     bool   // true if trailing '?' is present
	IsLinkLabel  bool   // true when this is [linklabel] from filtergraph
	LinkLabel    string // the label name (without brackets)
	Span         Span
}

type ParseError struct {
	Message string
	Span    Span
}

func (e *ParseError) Error() string {
	return e.Message
}

type parser struct {
	input string
	pos   int
}

// Parse parses a -map value string.
func Parse(input string) (*MapValue, error) {
	p := &parser{input: input}
	val, err := p.parseMapValue()
	if err != nil {
		return nil, err
	}
	if p.pos < len(p.input) {
		return nil, p.syntaxError("unexpected token after map value")
	}
	return val, nil
}

func (p *parser) syntaxError(msg string) *ParseError {
	return &ParseError{
		Message: msg,
		Span:    Span{Start: p.pos, End: min(p.pos+1, len(p.input))},
	}
}

func (p *parser) peek() byte {
	if p.pos >= len(p.input) {
		return 0
	}
	return p.input[p.pos]
}

func (p *parser) advance() byte {
	if p.pos >= len(p.input) {
		return 0
	}
	ch := p.input[p.pos]
	p.pos++
	return ch
}

func (p *parser) atEnd() bool {
	return p.pos >= len(p.input)
}

func (p *parser) parseMapValue() (*MapValue, error) {
	start := p.pos

	// Check for link label form: [label]
	if p.peek() == '[' {
		p.advance()
		labelStart := p.pos
		for !p.atEnd() && p.peek() != ']' {
			p.advance()
		}
		label := p.input[labelStart:p.pos]
		if p.atEnd() {
			return nil, p.syntaxError("expected ']'")
		}
		p.advance() // consume ']'
		return &MapValue{
			IsLinkLabel: true,
			LinkLabel:   label,
			Span:        Span{Start: start, End: p.pos},
		}, nil
	}

	mv := &MapValue{Span: Span{Start: start, End: p.pos}}

	// Check for negative map prefix
	if p.peek() == '-' {
		p.advance()
		mv.Negate = true
	}

	// Parse file index
	fileIdx, err := p.parseInteger()
	if err != nil {
		return nil, err
	}
	mv.FileIndex = fileIdx

	// Parse optional stream specifier (after ':')
	if !p.atEnd() && p.peek() == ':' {
		p.advance()
		mv.HasSpecifier = true
		spec, remainder, err := p.scanStreamSpec()
		if err != nil {
			return nil, err
		}
		mv.Specifier = spec
		if err := p.applyRemainder(mv, remainder); err != nil {
			return nil, err
		}
	}

	mv.Span.End = p.pos
	return mv, nil
}

// applyRemainder handles what follows the stream specifier: a view specifier
// keyword, an optional '?' marker, or an error for anything else. This mirrors
// ffmpeg's -map parsing, which splits the view specifier off the remainder
// returned by stream_specifier_parse.
func (p *parser) applyRemainder(mv *MapValue, remainder string) error {
	if remainder == "" {
		return nil
	}
	// p.pos still points at the raw rest (including the skipped separator);
	// the remainder text starts at the end of the input minus its length.
	restStart := len(p.input) - len(remainder)
	if remainder == "?" {
		mv.Optional = true
		p.pos = len(p.input)
		return nil
	}
	if !hasPrefixView(remainder) {
		p.pos = restStart
		return p.syntaxError("trailing garbage after stream specifier")
	}
	view := remainder
	optional := false
	if i := strings.IndexByte(view, '?'); i >= 0 {
		view = view[:i]
		optional = true
	}
	if err := validateViewSpec(view); err != nil {
		p.pos = restStart
		return p.syntaxError(err.Error())
	}
	mv.HasView = true
	mv.ViewSpec = view
	mv.Optional = optional
	p.pos = restStart + len(view)
	if optional {
		p.pos++
	}
	if p.pos < len(p.input) {
		return p.syntaxError("trailing garbage after view specifier")
	}
	return nil
}

// scanStreamSpec consumes a stream specifier following ffmpeg's
// stream_specifier_parse semantics and returns the raw specifier text along
// with the remainder that follows it (empty when the specifier consumed the
// whole input). Components that terminate the specifier (numeric index,
// stream ID, metadata, usable) put everything after them into the remainder.
func (p *parser) scanStreamSpec() (string, string, error) {
	specStart := p.pos
	seenType := false

	for p.pos < len(p.input) {
		ch := p.input[p.pos]
		switch {
		case ch >= '0' && ch <= '9':
			// Stream index terminates the specifier
			if _, err := p.scanBase0(); err != nil {
				return "", "", err
			}
			return p.finishSpec(specStart)

		case isStreamTypeLetter(ch) && (p.pos+1 >= len(p.input) || !isAlnum(p.input[p.pos+1])):
			if seenType {
				return "", "", p.syntaxError("stream type specified multiple times")
			}
			seenType = true
			p.pos++

		case ch == 'g' && p.hasPrefixAt(p.pos, "g:"):
			p.pos += 2
			if p.atEnd() {
				return "", "", p.syntaxError("expected group specifier after 'g:'")
			}
			if p.peek() == '#' {
				p.pos++
				if err := p.scanRequiredID("expected group ID after 'g:#'"); err != nil {
					return "", "", err
				}
			} else if p.hasPrefixAt(p.pos, "i:") {
				p.pos += 2
				if err := p.scanRequiredID("expected group ID after 'g:i:'"); err != nil {
					return "", "", err
				}
			} else if _, err := p.scanBase0(); err != nil {
				return "", "", err
			}

		case ch == 'p' && p.hasPrefixAt(p.pos, "p:"):
			p.pos += 2
			if err := p.scanRequiredID("expected program ID after 'p:'"); err != nil {
				return "", "", err
			}

		case p.hasPrefixAt(p.pos, "disp:"):
			p.pos += 5
			p.scanDispChars()

		case ch == '#' || (ch == 'i' && p.hasPrefixAt(p.pos, "i:")):
			// Stream ID terminates the specifier
			p.pos += 1 + boolToInt(ch == 'i')
			if err := p.scanRequiredID("expected stream ID"); err != nil {
				return "", "", err
			}
			return p.finishSpec(specStart)

		case ch == 'm' && p.hasPrefixAt(p.pos, "m:"):
			// Metadata specifier terminates the specifier
			p.pos += 2
			key := p.scanMetadataToken()
			if key == "" {
				return "", "", p.syntaxError("expected metadata key after 'm:'")
			}
			if !p.atEnd() && p.peek() == ':' {
				p.pos++
				p.scanMetadataToken()
			}
			return p.finishSpec(specStart)

		case ch == 'u' && (p.pos+1 >= len(p.input) || p.input[p.pos+1] == ':'):
			// Usable-only terminates the specifier
			p.pos++
			return p.finishSpec(specStart)

		default:
			// Unknown character ends the specifier
			return p.finishSpec(specStart)
		}

		if !p.atEnd() && p.peek() == ':' {
			p.pos++
		}
	}

	return p.input[specStart:p.pos], "", nil
}

// finishSpec returns the raw specifier consumed so far and the remainder
// that follows it. Separator colons are excluded from both: ffmpeg consumes
// a ':' separator after each component and skips one more when splitting
// off the remainder, so at most two colons separate the specifier from the
// remainder text.
func (p *parser) finishSpec(specStart int) (string, string, error) {
	spec := p.input[specStart:p.pos]
	rest := p.input[p.pos:]
	if rest != "" {
		spec = strings.TrimSuffix(spec, ":")
		if rest[0] == ':' {
			rest = rest[1:]
		}
	}
	return spec, rest, nil
}

func (p *parser) scanRequiredID(msg string) error {
	id := p.scanHexOrDecimalID()
	if id == "" {
		return p.syntaxError(msg)
	}
	return nil
}

func (p *parser) scanHexOrDecimalID() string {
	start := p.pos
	for p.pos < len(p.input) {
		ch := p.input[p.pos]
		if isHexDigit(ch) || ch == 'x' || ch == 'X' {
			p.pos++
		} else {
			break
		}
	}
	if p.pos == start {
		return ""
	}
	return p.input[start:p.pos]
}

// scanMetadataToken scans a metadata key or value up to an unescaped ':'.
// Colons in the key or value must be backslash-escaped.
func (p *parser) scanMetadataToken() string {
	start := p.pos
	for p.pos < len(p.input) {
		ch := p.input[p.pos]
		if ch == '\\' && p.pos+1 < len(p.input) {
			p.pos += 2
			continue
		}
		if ch == ':' {
			break
		}
		p.pos++
	}
	return p.input[start:p.pos]
}

func (p *parser) scanDispChars() {
	for p.pos < len(p.input) && isDispositionChar(p.input[p.pos]) {
		p.pos++
	}
}

// scanBase0 scans an integer with strtol base-0 semantics, as ffmpeg does:
// decimal, 0x-prefixed hexadecimal, or leading-zero octal.
func (p *parser) scanBase0() (int, error) {
	start := p.pos
	if !p.atEnd() && (p.peek() == '+' || p.peek() == '-') {
		p.pos++
	}
	if p.hasPrefixAt(p.pos, "0x") || p.hasPrefixAt(p.pos, "0X") {
		p.pos += 2
		for !p.atEnd() && isHexDigit(p.peek()) {
			p.pos++
		}
	} else if !p.atEnd() && p.peek() == '0' {
		p.pos++
		for !p.atEnd() && p.peek() >= '0' && p.peek() <= '7' {
			p.pos++
		}
	} else {
		for !p.atEnd() && p.peek() >= '0' && p.peek() <= '9' {
			p.pos++
		}
	}
	text := p.input[start:p.pos]
	if text == "" || text == "+" || text == "-" {
		p.pos = start
		return 0, p.syntaxError("expected integer")
	}
	n, err := strconv.ParseInt(text, 0, 64)
	if err != nil {
		p.pos = start
		return 0, p.syntaxError("invalid integer")
	}
	return int(n), nil
}

func (p *parser) parseInteger() (int, error) {
	return p.scanBase0()
}

// validateViewSpec checks a view specifier keyword and its value, mirroring
// ffmpeg's view_specifier_parse: view:<id|all>, vidx:<index>, vpos:<left|right>.
func validateViewSpec(vs string) error {
	keyword, value, _ := strings.Cut(vs, ":")
	switch keyword {
	case "view":
		if value == "all" {
			return nil
		}
		if _, err := strconv.ParseInt(value, 0, 64); err != nil || value == "" {
			return fmt.Errorf("invalid view ID: %s", value)
		}
	case "vidx":
		if _, err := strconv.ParseInt(value, 0, 64); err != nil || value == "" {
			return fmt.Errorf("invalid view index: %s", value)
		}
	case "vpos":
		if value != "left" && value != "right" {
			return fmt.Errorf("invalid view position: %s", value)
		}
	}
	return nil
}

// Format returns the string representation of a MapValue.
func Format(mv *MapValue) string {

	if mv.IsLinkLabel {
		return fmt.Sprintf("[%s]", mv.LinkLabel)
	}
	var s string
	if mv.Negate {
		s += "-"
	}
	s += fmt.Sprintf("%d", mv.FileIndex)
	if mv.HasSpecifier {
		s += ":" + mv.Specifier
	}
	if mv.HasView {
		s += ":" + mv.ViewSpec
	}
	if mv.Optional {
		s += "?"
	}
	return s
}

// hasPrefixView checks if the remaining string starts with a view specifier keyword.
// View specifiers: view:, vidx:, vpos:
func hasPrefixView(s string) bool {
	return hasPrefixWord(s, "view:") || hasPrefixWord(s, "vidx:") || hasPrefixWord(s, "vpos:")
}

func hasPrefixWord(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func (p *parser) hasPrefixAt(pos int, s string) bool {
	return pos+len(s) <= len(p.input) && p.input[pos:pos+len(s)] == s
}

func isStreamTypeLetter(ch byte) bool {
	switch ch {
	case 'v', 'V', 'a', 's', 'd', 't':
		return true
	}
	return false
}

func isAlnum(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9')
}

func isHexDigit(ch byte) bool {
	return (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')
}

func isDispositionChar(ch byte) bool {
	return isAlnum(ch) || ch == '_'
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
