package filtergraph

// ParseForCompletion parses a partial filtergraph string and returns a
// CompletionContext describing what is expected at the end of the input.
func ParseForCompletion(input string) *CompletionContext {
	cursor := len(input)
	p := &compParser{
		input:  input,
		pos:    0,
		cursor: cursor,
		ctx:    &CompletionContext{},
	}
	p.parseFiltergraph()

	if len(p.ctx.ExpectedTokens) == 0 {
		p.ctx.ExpectedTokens = append(p.ctx.ExpectedTokens, ExpectedFilterName)
	}
	dedupTokens(&p.ctx.ExpectedTokens)
	return p.ctx
}

type compParser struct {
	input  string
	pos    int
	cursor int
	ctx    *CompletionContext

	chainIndex  int
	filterIndex int
	inFilter    bool
}

func (p *compParser) atCursorOrEnd() bool {
	return p.pos >= len(p.input) || p.pos >= p.cursor
}

func (p *compParser) peek() byte {
	if p.pos >= len(p.input) || p.pos >= p.cursor {
		return 0
	}
	return p.input[p.pos]
}

func (p *compParser) peekRaw() byte {
	if p.pos >= len(p.input) {
		return 0
	}
	return p.input[p.pos]
}

func (p *compParser) addExpected(t ExpectedToken) {
	p.ctx.ExpectedTokens = append(p.ctx.ExpectedTokens, t)
}

func (p *compParser) parseFiltergraph() {
	p.parseChain()
	for !p.atCursorOrEnd() && p.peekRaw() == ';' {
		p.pos++
		p.chainIndex++
		p.filterIndex = 0
		p.inFilter = false
		p.parseChain()
	}
}

func (p *compParser) parseChain() {
	// Optional input label
	if p.peek() == '[' {
		p.parseLabel()
		if p.atCursorOrEnd() {
			p.ctx.InLabel = true
			p.addExpected(ExpectedFilterName)
			return
		}
	}

	// First filter (required)
	p.parseFilter()

	// Additional filters separated by ','
	for !p.atCursorOrEnd() && p.peekRaw() == ',' {
		p.pos++
		p.filterIndex++
		p.parseFilter()
	}

	// Optional output label
	if p.peek() == '[' {
		p.parseLabel()
	}
}

func (p *compParser) parseFilter() {
	p.inFilter = true
	p.ctx.ChainIndex = p.chainIndex
	p.ctx.FilterIndex = p.filterIndex

	// Filter name
	name := p.scanFilterNameForCompletion()
	if p.atCursorOrEnd() {
		p.ctx.PartialIdent = name
		p.ctx.Filter = &FilterContext{
			Name:     name,
			ArgIndex: 0,
		}
		p.addExpected(ExpectedFilterName)
		if name != "" {
			p.addExpected(ExpectedFilterOption)
		}
		return
	}

	// Consume '='
	if p.peekRaw() == '=' {
		p.pos++
		argIndex := 0
		p.parseFilterOption(name, argIndex)
		argIndex++

		// Additional options separated by ':'
		for !p.atCursorOrEnd() && p.peekRaw() == ':' {
			p.pos++
			p.parseFilterOption(name, argIndex)
			argIndex++
		}
	}

	p.ctx.Filter = &FilterContext{
		Name: name,
	}
}

func (p *compParser) parseFilterOption(filterName string, argIndex int) {
	// Scan first segment
	first := p.scanOptionSegmentForCompletion()

	if p.atCursorOrEnd() {
		// Could be key or positional value
		p.ctx.PartialIdent = first
		p.ctx.Filter = &FilterContext{
			Name:     filterName,
			ArgIndex: argIndex,
		}
		p.addExpected(ExpectedFilterOptionValue)
		p.addExpected(ExpectedFilterOptionKey)
		return
	}

	// Check for '=' after the first segment
	if p.peekRaw() == '=' {
		p.pos++
		value := p.scanOptionSegmentForCompletion()
		p.ctx.Filter = &FilterContext{
			Name:     filterName,
			ArgIndex: argIndex,
			InKey:    false,
			InValue:  true,
		}
		p.ctx.PartialIdent = value
		p.addExpected(ExpectedFilterOptionValue)
		return
	}

	// Positional value
	p.ctx.Filter = &FilterContext{
		Name:     filterName,
		ArgIndex: argIndex,
	}
	p.ctx.PartialIdent = first
}

func (p *compParser) parseLabel() {
	if p.peek() != '[' {
		return
	}
	p.pos++ // consume '['
	start := p.pos
	for !p.atCursorOrEnd() && p.peek() != ']' {
		p.pos++
	}
	content := p.input[start:min(p.pos, p.cursor)]
	p.ctx.InLabel = true
	p.ctx.LabelContent = content
	p.ctx.PartialIdent = content
	p.addExpected(ExpectedLinkLabel)
	if !p.atCursorOrEnd() && p.peek() == ']' {
		p.pos++ // consume ']'
		p.ctx.InLabel = false
	}
}

func (p *compParser) scanFilterNameForCompletion() string {
	start := p.pos
	for !p.atCursorOrEnd() {
		ch := p.peek()
		if isFilterNameChar(ch) {
			p.pos++
		} else {
			break
		}
	}
	return p.input[start:p.pos]
}

func (p *compParser) scanOptionSegmentForCompletion() string {
	start := p.pos
	for !p.atCursorOrEnd() {
		ch := p.peek()
		if ch == ':' || ch == ',' || ch == ';' || ch == '[' || ch == ']' || ch == '=' {
			break
		}
		if ch == '\\' {
			p.pos++
			if !p.atCursorOrEnd() {
				p.pos++
			}
			continue
		}
		if ch == '\'' {
			p.pos++
			for !p.atCursorOrEnd() && p.peek() != '\'' {
				if p.peek() == '\\' {
					p.pos++
					if !p.atCursorOrEnd() {
						p.pos++
					}
					continue
				}
				p.pos++
			}
			if !p.atCursorOrEnd() {
				p.pos++ // closing quote
			}
			continue
		}
		p.pos++
	}
	return p.input[start:min(p.pos, p.cursor)]
}

func dedupTokens(tokens *[]ExpectedToken) {
	seen := make(map[ExpectedToken]bool)
	result := make([]ExpectedToken, 0, len(*tokens))
	for _, t := range *tokens {
		if !seen[t] {
			seen[t] = true
			result = append(result, t)
		}
	}
	*tokens = result
}
