package mapvalue

import (
	"fmt"
	"strconv"
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
		specStart := p.pos
		// Stream specifier extends until a ':' followed by a view keyword, '?' (optional), or end
		for !p.atEnd() {
			ch := p.peek()
			if ch == '?' {
				break
			}
			if ch == ':' {
				// Check if the next segment is a view specifier keyword
				remaining := p.input[p.pos+1:]
				if hasPrefixView(remaining) {
					break
				}
				p.advance() // consume the colon, continue building specifier
				continue
			}
			p.advance()
		}
		mv.Specifier = p.input[specStart:p.pos]
	}

	// Check for view specifier (after ':')
	if !p.atEnd() && p.peek() == ':' {
		p.advance()
		mv.HasView = true
		viewStart := p.pos
		// View specifier extends until next ':' or '?' or end
		// But view specifiers can also contain ':' (e.g. view:all)
		for !p.atEnd() {
			ch := p.peek()
			if ch == '?' {
				break
			}
			p.advance()
		}
		mv.ViewSpec = p.input[viewStart:p.pos]
	}

	// Check for optional '?' suffix
	if !p.atEnd() && p.peek() == '?' {
		p.advance()
		mv.Optional = true
	}

	mv.Span.End = p.pos
	return mv, nil
}

func (p *parser) parseInteger() (int, error) {
	start := p.pos
	for !p.atEnd() && p.peek() >= '0' && p.peek() <= '9' {
		p.advance()
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
