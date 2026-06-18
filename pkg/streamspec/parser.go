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
	input string
	pos   int
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
			Kind: KindDisposition,
			Span: Span{Start: start, End: p.pos},
			payload: disp,
		}, start)
	}

	ch := p.peek()

	// 'u' - usable streams
	if ch == 'u' {
		p.advance()
		return p.withAdditional(&Specifier{
			Kind: KindUsable,
			Span: Span{Start: start, End: p.pos},
			payload: struct{}{},
		}, start)
	}

	// Stream type letter: v, V, a, s, d, t
	if isStreamTypeLetter(ch) {
		st := p.parseStreamTypeLetter()
		return p.withAdditional(&Specifier{
			Kind: KindStreamType,
			Span: Span{Start: start, End: p.pos},
			payload: st,
		}, start)
	}

	// 'g:' - group specifier
	if p.hasPrefix("g:") {
		p.pos += 2
		group, err := p.parseGroupSpecifier()
		if err != nil {
			return nil, err
		}
		return p.withAdditional(&Specifier{
			Kind: KindGroup,
			Span: Span{Start: start, End: p.pos},
			payload: group,
		}, start)
	}

	// 'p:' - program specifier
	if p.hasPrefix("p:") {
		p.pos += 2
		prog, err := p.parseProgramSpecifier()
		if err != nil {
			return nil, err
		}
		return p.withAdditional(&Specifier{
			Kind: KindProgram,
			Span: Span{Start: start, End: p.pos},
			payload: prog,
		}, start)
	}

	// '#' - stream ID
	if ch == '#' {
		p.advance()
		id, err := p.scanStreamIDValue()
		if err != nil {
			return nil, err
		}
		return p.withAdditional(&Specifier{
			Kind: KindStreamID,
			Span: Span{Start: start, End: p.pos},
			payload: &StreamIDExpr{ID: id, Alt: false},
		}, start)
	}

	// 'i:' - stream ID (alternate)
	if p.hasPrefix("i:") {
		p.pos += 2
		id, err := p.scanStreamIDValue()
		if err != nil {
			return nil, err
		}
		return p.withAdditional(&Specifier{
			Kind: KindStreamID,
			Span: Span{Start: start, End: p.pos},
			payload: &StreamIDExpr{ID: id, Alt: true},
		}, start)
	}

	// 'm:' - metadata specifier
	if p.hasPrefix("m:") {
		p.pos += 2
		meta, err := p.parseMetadataSpecifier()
		if err != nil {
			return nil, err
		}
		return p.withAdditional(&Specifier{
			Kind: KindMetadata,
			Span: Span{Start: start, End: p.pos},
			payload: meta,
		}, start)
	}

	// Numeric stream index
	if ch >= '0' && ch <= '9' {
		idx, err := p.scanInteger()
		if err != nil {
			return nil, err
		}
		return p.withAdditional(&Specifier{
			Kind: KindStreamIndex,
			Span: Span{Start: start, End: p.pos},
			payload: idx,
		}, start)
	}

	return nil, p.syntaxError(fmt.Sprintf("unexpected character %q in stream specifier", ch))
}

// withAdditional checks if the specifier is followed by ':' and parses an additional specifier.
func (p *parser) withAdditional(spec *Specifier, _ int) (*Specifier, error) {
	if !p.atEnd() && p.peek() == ':' {
		p.advance()
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
		return &GroupExpr{Kind: GroupByID, ID: id}, nil
	}
	if p.hasPrefix("i:") {
		p.pos += 2
		id := p.scanHexOrDecimalID()
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
	return &ProgramExpr{ID: id}, nil
}

func (p *parser) scanStreamIDValue() (string, error) {
	if p.atEnd() {
		return "", p.syntaxError("expected stream ID")
	}
	return p.scanHexOrDecimalID(), nil
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
		if ch == ':' || ch == '\\' {
			break
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
		if ch == ':' || ch == '\\' {
			break
		}
		p.pos++
	}
	if p.pos > start {
		result = append(result, p.input[start:p.pos])
	}
	return result
}

func (p *parser) scanInteger() (int, error) {
	start := p.pos
	for p.pos < len(p.input) && p.input[p.pos] >= '0' && p.input[p.pos] <= '9' {
		p.pos++
	}
	if p.pos == start {
		return 0, p.syntaxError("expected integer")
	}
	n, err := strconv.Atoi(p.input[start:p.pos])
	if err != nil {
		return 0, p.syntaxError("invalid integer")
	}
	return n, nil
}

func isHexDigit(ch byte) bool {
	return (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')
}

func isStreamTypeLetter(ch rune) bool {
	return ch == 'v' || ch == 'V' || ch == 'a' || ch == 's' || ch == 'd' || ch == 't'
}
