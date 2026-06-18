package filtergraph

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

// Parse parses a filtergraph string into an AST.
func Parse(input string) (*Filtergraph, error) {
	p := &parser{input: input}
	fg, err := p.parseFiltergraph()
	if err != nil {
		return nil, err
	}
	if p.pos < len(p.input) {
		return nil, p.syntaxError("unexpected token after filtergraph")
	}
	return fg, nil
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

// parseFiltergraph parses the top-level: chain (';' chain)*
func (p *parser) parseFiltergraph() (*Filtergraph, error) {
	start := p.pos
	var chains []*Chain

	chain, err := p.parseChain()
	if err != nil {
		return nil, err
	}
	chains = append(chains, chain)

	for !p.atEnd() && p.peek() == ';' {
		p.advance()
		chain, err := p.parseChain()
		if err != nil {
			return nil, err
		}
		chains = append(chains, chain)
	}

	return &Filtergraph{
		Chains: chains,
		Span:   Span{Start: start, End: p.pos},
	}, nil
}

// parseChain parses: [input_label]* filter (',' filter)* [output_label]*
func (p *parser) parseChain() (*Chain, error) {
	start := p.pos

	var inputLabels []string
	for !p.atEnd() && p.peek() == '[' {
		label, err := p.parseLabel()
		if err != nil {
			return nil, err
		}
		inputLabels = append(inputLabels, label)
	}

	filter, err := p.parseFilter()
	if err != nil {
		return nil, err
	}
	filters := []*Filter{filter}

	for !p.atEnd() && p.peek() == ',' {
		p.advance()
		f, err := p.parseFilter()
		if err != nil {
			return nil, err
		}
		filters = append(filters, f)
	}

	var outputLabels []string
	for !p.atEnd() && p.peek() == '[' {
		label, err := p.parseLabel()
		if err != nil {
			return nil, err
		}
		outputLabels = append(outputLabels, label)
	}

	return &Chain{
		InputLabels:  inputLabels,
		Filters:      filters,
		OutputLabels: outputLabels,
		Span:         Span{Start: start, End: p.pos},
	}, nil
}

// parseFilter parses: name ('=' option (':' option)*)?
func (p *parser) parseFilter() (*Filter, error) {
	start := p.pos

	name := p.scanFilterName()
	if name == "" {
		return nil, p.syntaxError("expected filter name")
	}

	var options []*FilterOption
	if !p.atEnd() && p.peek() == '=' {
		p.advance()
		opt, err := p.parseFilterOption()
		if err != nil {
			return nil, err
		}
		options = append(options, opt)

		for !p.atEnd() && p.peek() == ':' {
			p.advance()
			opt, err := p.parseFilterOption()
			if err != nil {
				return nil, err
			}
			options = append(options, opt)
		}
	}

	return &Filter{
		Name:    name,
		Options: options,
		Span:    Span{Start: start, End: p.pos},
	}, nil
}

// parseFilterOption parses a single option: either key=value or positional value.
func (p *parser) parseFilterOption() (*FilterOption, error) {
	start := p.pos

	// Scan the first segment (could be key or positional value)
	first := p.scanOptionSegment()

	// Check if followed by '=' — if so, it's key=value
	if !p.atEnd() && p.peek() == '=' {
		p.advance()
		value := p.scanOptionSegment()
		return &FilterOption{
			Key:   first,
			Value: value,
			Span:  Span{Start: start, End: p.pos},
		}, nil
	}

	// Positional value
	return &FilterOption{
		Key:   "",
		Value: first,
		Span:  Span{Start: start, End: p.pos},
	}, nil
}

// parseLabel parses [label_text]
func (p *parser) parseLabel() (string, error) {
	if p.peek() != '[' {
		return "", p.syntaxError("expected '['")
	}
	p.advance() // consume '['
	start := p.pos
	for !p.atEnd() && p.peek() != ']' {
		p.advance()
	}
	label := p.input[start:p.pos]
	if p.atEnd() {
		return "", p.syntaxError("expected ']'")
	}
	p.advance() // consume ']'
	return label, nil
}

// scanFilterName scans a filter name: alphanumeric, underscore, hyphen
func (p *parser) scanFilterName() string {
	start := p.pos
	for !p.atEnd() {
		ch := p.peek()
		if isFilterNameChar(ch) {
			p.advance()
		} else {
			break
		}
	}
	return p.input[start:p.pos]
}

// scanOptionSegment scans a filter option value segment (between '=' or ':' delimiters).
// Handles quoting and escaping.
func (p *parser) scanOptionSegment() string {
	var sb []byte
	for !p.atEnd() {
		ch := p.peek()
		if ch == ':' || ch == ',' || ch == ';' || ch == '[' || ch == ']' {
			break
		}
		if ch == '\\' {
			p.advance()
			if !p.atEnd() {
				sb = append(sb, p.advance())
			}
			continue
		}
		if ch == '\'' {
			p.advance()
			for !p.atEnd() && p.peek() != '\'' {
				if p.peek() == '\\' {
					p.advance()
					if !p.atEnd() {
						sb = append(sb, p.advance())
					}
					continue
				}
				sb = append(sb, p.advance())
			}
			if !p.atEnd() {
				p.advance() // closing quote
			}
			continue
		}
		if ch == '=' {
			break
		}
		sb = append(sb, p.advance())
	}
	return string(sb)
}

func isFilterNameChar(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_' || ch == '-'
}
