package mapvalue

// ParseForCompletion parses a partial -map value and returns a
// CompletionContext describing what is expected at the end of the input.
func ParseForCompletion(input string) *CompletionContext {
	cursor := len(input)
	p := &compParser{
		input:  input,
		pos:    0,
		cursor: cursor,
		ctx:    &CompletionContext{},
	}
	p.parseMapValue()

	if len(p.ctx.ExpectedTokens) == 0 {
		p.ctx.ExpectedTokens = append(p.ctx.ExpectedTokens, ExpectedFileIndex)
	}
	dedupTokens(&p.ctx.ExpectedTokens)
	return p.ctx
}

type compParser struct {
	input  string
	pos    int
	cursor int
	ctx    *CompletionContext
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

func (p *compParser) parseMapValue() {
	// Check for link label form: [label]
	if p.peek() == '[' {
		p.pos++ // consume '['
		p.ctx.IsLinkLabel = true
		start := p.pos
		for !p.atCursorOrEnd() && p.peekRaw() != ']' {
			p.pos++
		}
		p.ctx.PartialIdent = p.input[start:min(p.pos, p.cursor)]
		p.addExpected(ExpectedLinkLabel)
		return
	}

	// Check for negative map prefix
	if p.peek() == '-' {
		p.pos++
		p.ctx.Negate = true
	}

	// Parse file index
	p.parseFileIndex()

	if p.atCursorOrEnd() {
		p.addExpected(ExpectedSpecifier)
		p.addExpected(ExpectedOptional)
		return
	}

	// Parse stream specifier after ':'
	if p.peekRaw() == ':' {
		p.pos++
		p.ctx.HasSpecifier = true
		p.addExpected(ExpectedSpecifier)
		start := p.pos
		for !p.atCursorOrEnd() {
			ch := p.peekRaw()
			if ch == ':' || ch == '?' {
				break
			}
			p.pos++
		}
		p.ctx.Specifier = p.input[start:min(p.pos, p.cursor)]
		if p.atCursorOrEnd() {
			p.addExpected(ExpectedViewSpecifier)
			p.addExpected(ExpectedOptional)
			return
		}
	}

	// View specifier
	if !p.atCursorOrEnd() && p.peekRaw() == ':' {
		p.pos++
		p.addExpected(ExpectedViewSpecifier)
		for !p.atCursorOrEnd() && p.peekRaw() != '?' && p.peekRaw() != ':' {
			p.pos++
		}
		if p.atCursorOrEnd() {
			p.addExpected(ExpectedOptional)
			return
		}
	}

	// Optional '?'
	if !p.atCursorOrEnd() && p.peekRaw() == '?' {
		p.pos++
		p.addExpected(ExpectedOptional)
	}
}

func (p *compParser) parseFileIndex() {
	start := p.pos
	for !p.atCursorOrEnd() && p.peek() >= '0' && p.peek() <= '9' {
		p.pos++
	}
	if p.pos > start {
		p.ctx.PartialIdent = p.input[start:p.pos]
		p.ctx.FileIndex = mustAtoi(p.input[start:p.pos])
	}
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

func mustAtoi(s string) int {
	n := 0
	for _, ch := range s {
		n = n*10 + int(ch-'0')
	}
	return n
}
