package streamspec

import (
	"fmt"
	"strconv"
	"unicode"
	"unicode/utf8"
)

type ParseError struct {
	Message string
	Span    Span
}

func (e *ParseError) Error() string {
	return e.Message
}

type parser struct {
	input    string
	pos      int
	seenType bool
}

// Parse parses a stream specifier string into an AST.
// The input should be the specifier portion only (e.g. "a:1", "m:language:eng", "u").
func Parse(input string) (*Specifier, error) {
	p := &parser{input: input}
	spec, err := p.parseSpecifier()
	if err != nil {
		return nil, err
	}
	p.skipWhitespace()
	if p.pos < len(p.input) {
		return nil, p.syntaxError("unexpected token after specifier")
	}
	return spec, nil
}

// IsSpecifier checks if the text is a valid stream specifier.
func IsSpecifier(text string) bool {
	p := &parser{input: text}
	_, err := p.parseSpecifier()
	return err == nil && p.pos == len(text)
}

func (p *parser) syntaxError(msg string) *ParseError {
	return &ParseError{
		Message: msg,
		Span:    Span{Start: p.pos, End: min(p.pos+1, len(p.input))},
	}
}

func (p *parser) peek() rune {
	if p.pos >= len(p.input) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(p.input[p.pos:])
	return r
}

func (p *parser) advance() rune {
	if p.pos >= len(p.input) {
		return 0
	}
	r, w := utf8.DecodeRuneInString(p.input[p.pos:])
	p.pos += w
	return r
}

func (p *parser) atEnd() bool {
	return p.pos >= len(p.input)
}

func (p *parser) skipWhitespace() {
	for p.pos < len(p.input) {
		r, w := utf8.DecodeRuneInString(p.input[p.pos:])
		if !unicode.IsSpace(r) {
			break
		}
		p.pos += w
	}
}

func (p *parser) hasPrefix(s string) bool {
	return p.pos+len(s) <= len(p.input) && p.input[p.pos:p.pos+len(s)] == s
}

// parseSpecifier parses a single stream specifier.
func (p *parser) parseSpecifier() (*Specifier, error) {
	start := p.pos

	if p.atEnd() {
		return nil, p.syntaxError("expected stream specifier")
	}

	// Check 'disp:' BEFORE stream type letter 'd' — this must come first
	if p.hasPrefix("disp:") {
		p.pos += 5
		disp, err := p.parseDispositionSpecifier()
		if err != nil {
			return nil, err
		}
		return p.withAdditional(&Specifier{
			Kind:    KindDisposition,
			Span:    Span{Start: start, End: p.pos},
			payload: disp,
		}, addlChain)
	}

	ch := p.peek()

	// 'u' - usable streams (terminates the specifier)
	if ch == 'u' {
		p.advance()
		return p.withAdditional(&Specifier{
			Kind:    KindUsable,
			Span:    Span{Start: start, End: p.pos},
			payload: struct{}{},
		}, addlTerminating)
	}

	// Stream type letter: v, V, a, s, d, t
	if isStreamTypeLetter(ch) {
		if p.seenType {
			return nil, p.syntaxError("stream type specified multiple times")
		}
		p.seenType = true
		st := p.parseStreamTypeLetter()
		return p.withAdditional(&Specifier{
			Kind:    KindStreamType,
			Span:    Span{Start: start, End: p.pos},
			payload: st,
		}, addlChain)
	}

	// 'g:' - group specifier
	if p.hasPrefix("g:") {
		p.pos += 2
		group, err := p.parseGroupSpecifier()
		if err != nil {
			return nil, err
		}
		return p.withAdditional(&Specifier{
			Kind:    KindGroup,
			Span:    Span{Start: start, End: p.pos},
			payload: group,
		}, addlChain)
	}

	// 'p:' - program specifier
	if p.hasPrefix("p:") {
		p.pos += 2
		prog, err := p.parseProgramSpecifier()
		if err != nil {
			return nil, err
		}
		return p.withAdditional(&Specifier{
			Kind:    KindProgram,
			Span:    Span{Start: start, End: p.pos},
			payload: prog,
		}, addlChain)
	}

	// '#' - stream ID (terminates the specifier)
	if ch == '#' {
		p.advance()
		id, err := p.scanStreamIDValue()
		if err != nil {
			return nil, err
		}
		return p.withAdditional(&Specifier{
			Kind:    KindStreamID,
			Span:    Span{Start: start, End: p.pos},
			payload: &StreamIDExpr{ID: id, Alt: false},
		}, addlTerminating)
	}

	// 'i:' - stream ID, alternate syntax (terminates the specifier)
	if p.hasPrefix("i:") {
		p.pos += 2
		id, err := p.scanStreamIDValue()
		if err != nil {
			return nil, err
		}
		return p.withAdditional(&Specifier{
			Kind:    KindStreamID,
			Span:    Span{Start: start, End: p.pos},
			payload: &StreamIDExpr{ID: id, Alt: true},
		}, addlTerminating)
	}

	// 'm:' - metadata specifier (terminates the specifier)
	if p.hasPrefix("m:") {
		p.pos += 2
		meta, err := p.parseMetadataSpecifier()
		if err != nil {
			return nil, err
		}
		return p.withAdditional(&Specifier{
			Kind:    KindMetadata,
			Span:    Span{Start: start, End: p.pos},
			payload: meta,
		}, addlMetadata)
	}

	// Numeric stream index (terminates the specifier)
	if ch >= '0' && ch <= '9' {
		idx, err := p.scanInteger()
		if err != nil {
			return nil, err
		}
		return p.withAdditional(&Specifier{
			Kind:    KindStreamIndex,
			Span:    Span{Start: start, End: p.pos},
			payload: idx,
		}, addlTerminating)
	}

	return nil, p.syntaxError(fmt.Sprintf("unexpected character %q in stream specifier", ch))
}

// additionalMode controls what may follow a specifier component.
//
// ffmpeg's stream_specifier_parse consumes a ':' separator after components
// that continue the specifier chain, so a trailing colon with nothing after
// it is accepted (empty additional specifier matches all streams).
// Components that terminate the specifier (index, stream ID, usable) reject
// everything after them as trailing garbage.
type additionalMode int

const (
	addlChain       additionalMode = iota // type, group, program, disposition: trailing ':' is a separator
	addlTerminating                       // index, stream ID, usable: nothing may follow
	addlMetadata                          // metadata: composition supported, but a dangling ':' is garbage
)

// withAdditional checks if the specifier is followed by ':' and parses an
// additional specifier according to the component's mode.
func (p *parser) withAdditional(spec *Specifier, mode additionalMode) (*Specifier, error) {
	if !p.atEnd() && p.peek() == ':' {
		if mode == addlTerminating {
			return nil, p.syntaxError("trailing garbage after stream specifier")
		}
		p.advance()
		if p.atEnd() {
			if mode == addlMetadata {
				return nil, p.syntaxError("trailing garbage after stream specifier")
			}
			// Empty additional specifier matches all remaining streams
			return spec, nil
		}
		additional, err := p.parseSpecifier()
		if err != nil {
			return nil, err
		}
		spec.Additional = additional
		spec.Span.End = p.pos
	}
	return spec, nil
}

func (p *parser) parseStreamTypeLetter() *StreamTypeExpr {
	var st StreamType
	switch p.peek() {
	case 'v':
		st = TypeVideo
	case 'V':
		st = TypeVideoNoAttached
	case 'a':
		st = TypeAudio
	case 's':
		st = TypeSubtitle
	case 'd':
		st = TypeData
	case 't':
		st = TypeAttachment
	default:
		return nil
	}
	p.advance()
	return &StreamTypeExpr{Type: st}
}

func (p *parser) parseGroupSpecifier() (*GroupExpr, error) {
	if p.atEnd() {
		return nil, p.syntaxError("expected group specifier after 'g:'")
	}
	ch := p.peek()
	if ch == '#' {
		p.advance()
		id := p.scanHexOrDecimalID()
		if id == "" {
			return nil, p.syntaxError("expected group ID after 'g:#'")
		}
		return &GroupExpr{Kind: GroupByID, ID: id}, nil
	}
	if p.hasPrefix("i:") {
		p.pos += 2
		id := p.scanHexOrDecimalID()
		if id == "" {
			return nil, p.syntaxError("expected group ID after 'g:i:'")
		}
		return &GroupExpr{Kind: GroupByID, ID: id}, nil
	}
	idx, err := p.scanInteger()
	if err != nil {
		return nil, err
	}
	return &GroupExpr{Kind: GroupByIndex, Index: idx}, nil
}

func (p *parser) parseProgramSpecifier() (*ProgramExpr, error) {
	if p.atEnd() {
		return nil, p.syntaxError("expected program ID after 'p:'")
	}
	id := p.scanHexOrDecimalID()
	if id == "" {
		return nil, p.syntaxError("expected program ID after 'p:'")
	}
	return &ProgramExpr{ID: id}, nil
}

func (p *parser) scanStreamIDValue() (string, error) {
	if p.atEnd() {
		return "", p.syntaxError("expected stream ID")
	}
	id := p.scanHexOrDecimalID()
	if id == "" {
		return "", p.syntaxError("expected stream ID")
	}
	return id, nil
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

func (p *parser) parseMetadataSpecifier() (*MetadataExpr, error) {
	if p.atEnd() {
		return nil, p.syntaxError("expected metadata key after 'm:'")
	}
	key := p.scanMetadataKey()
	if key == "" {
		return nil, p.syntaxError("expected metadata key after 'm:'")
	}
	if !p.atEnd() && p.peek() == ':' {
		p.advance()
		value := p.scanMetadataValue()
		return &MetadataExpr{Key: key, Value: value}, nil
	}
	return &MetadataExpr{Key: key}, nil
}

func (p *parser) scanMetadataKey() string {
	start := p.pos
	for p.pos < len(p.input) {
		ch := p.input[p.pos]
		if ch == ':' {
			break
		}
		if ch == '\\' && p.pos+1 < len(p.input) {
			p.pos += 2
			continue
		}
		p.pos++
	}
	return p.input[start:p.pos]
}

func (p *parser) scanMetadataValue() string {
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

func (p *parser) parseDispositionSpecifier() (*DispositionExpr, error) {
	if p.atEnd() {
		return nil, p.syntaxError("expected disposition after 'disp:'")
	}
	disp := p.scanDispositions()
	if len(disp) == 0 {
		return nil, p.syntaxError("expected disposition name after 'disp:'")
	}
	return &DispositionExpr{Dispositions: disp}, nil
}

func (p *parser) scanDispositions() []string {
	var result []string
	start := p.pos
	for p.pos < len(p.input) {
		ch := p.input[p.pos]
		if ch == '+' {
			if p.pos > start {
				result = append(result, p.input[start:p.pos])
			}
			p.pos++
			start = p.pos
			continue
		}
		if !isDispositionChar(ch) {
			break
		}
		p.pos++
	}
	if p.pos > start {
		result = append(result, p.input[start:p.pos])
	}
	return result
}

func isDispositionChar(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_'
}

// scanInteger scans an integer using strtol base-0 semantics, as ffmpeg
// does: decimal, 0x-prefixed hexadecimal, or leading-zero octal.
func (p *parser) scanInteger() (int, error) {
	start := p.pos
	if p.pos < len(p.input) && (p.input[p.pos] == '+' || p.input[p.pos] == '-') {
		p.pos++
	}
	if p.hasPrefix("0x") || p.hasPrefix("0X") {
		p.pos += 2
		for p.pos < len(p.input) && isHexDigit(p.input[p.pos]) {
			p.pos++
		}
	} else if p.pos < len(p.input) && p.input[p.pos] == '0' {
		p.pos++
		for p.pos < len(p.input) && p.input[p.pos] >= '0' && p.input[p.pos] <= '7' {
			p.pos++
		}
	} else {
		for p.pos < len(p.input) && p.input[p.pos] >= '0' && p.input[p.pos] <= '9' {
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

func isHexDigit(ch byte) bool {
	return (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')
}

func isStreamTypeLetter(ch rune) bool {
	return ch == 'v' || ch == 'V' || ch == 'a' || ch == 's' || ch == 'd' || ch == 't'
}
