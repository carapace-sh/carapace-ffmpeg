package streamspec

import (
	"fmt"
	"slices"
	"unicode/utf8"
)

func allForms() []SpecifierForm {
	return []SpecifierForm{
		{Prefix: "", Description: "stream index (0, 1, 2...)"},
		{Prefix: "v", Description: "all video streams", Suffix: ":"},
		{Prefix: "V", Description: "video streams excluding attached pictures", Suffix: ":"},
		{Prefix: "a", Description: "all audio streams", Suffix: ":"},
		{Prefix: "s", Description: "all subtitle streams", Suffix: ":"},
		{Prefix: "d", Description: "all data streams", Suffix: ":"},
		{Prefix: "t", Description: "all attachment streams", Suffix: ":"},
		{Prefix: "g", Description: "stream group by index or ID", Suffix: ":"},
		{Prefix: "p", Description: "program by ID", Suffix: ":"},
		{Prefix: "#", Description: "stream by ID"},
		{Prefix: "i", Description: "stream by ID (alternate)", Suffix: ":"},
		{Prefix: "m", Description: "stream by metadata key", Suffix: ":"},
		{Prefix: "disp", Description: "stream by disposition", Suffix: ":"},
		{Prefix: "u", Description: "streams with usable configuration", Suffix: ":"},
	}
}

// ParseForCompletion parses a partial stream specifier and returns a
// CompletionContext describing what is expected at the end of the input.
func ParseForCompletion(input string) *CompletionContext {
	cursor := len(input)
	p := &compParser{
		input:  input,
		pos:    0,
		cursor: cursor,
		ctx:    &CompletionContext{},
	}
	p.parseSpecifier()

	if len(p.ctx.ExpectedTokens) == 0 {
		p.ctx.ExpectedTokens = append(p.ctx.ExpectedTokens, ExpectedSpecifierType)
	}
	p.ctx.ExpectedTokens = dedupTokens(p.ctx.ExpectedTokens)
	p.ctx.ValidForms = dedupForms(p.ctx.ValidForms)
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

func (p *compParser) peek() rune {
	if p.pos >= len(p.input) || p.pos >= p.cursor {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(p.input[p.pos:])
	return r
}

func (p *compParser) advance() rune {
	if p.pos >= len(p.input) || p.pos >= p.cursor {
		return 0
	}
	r, w := utf8.DecodeRuneInString(p.input[p.pos:])
	p.pos += w
	return r
}

// peekRaw returns the next rune from input without cursor constraint.
// Used for prefix matching (e.g. checking if 'g' is followed by ':')
// where the prefix has already been partially consumed.
func (p *compParser) peekRaw(offset int) rune {
	pos := p.pos + offset
	if pos >= len(p.input) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(p.input[pos:])
	return r
}

func (p *compParser) addExpected(t ExpectedToken) {
	p.ctx.ExpectedTokens = append(p.ctx.ExpectedTokens, t)
}

func (p *compParser) addForm(prefix, desc string, suffixes ...string) {
	suffix := ""
	if len(suffixes) > 0 {
		suffix = suffixes[0]
	}
	p.ctx.ValidForms = append(p.ctx.ValidForms, SpecifierForm{Prefix: prefix, Description: desc, Suffix: suffix})
}

func (p *compParser) addTopLevelForms() {
	for _, f := range allForms() {
		p.ctx.ValidForms = append(p.ctx.ValidForms, f)
	}
}

func (p *compParser) parseSpecifier() {
	if p.atCursorOrEnd() {
		p.addTopLevelForms()
		p.addExpected(ExpectedSpecifierType)
		p.addExpected(ExpectedStreamIndex)
		p.addExpected(ExpectedStreamTypeLetter)
		return
	}

	ch := p.peek()

	// Check 'disp:' BEFORE stream type letter 'd' — this must come first
	if ch == 'd' {
		remaining := p.input[p.pos:]
		if len(remaining) >= 5 && remaining[:5] == "disp:" {
			p.pos += 5
			if p.pos > p.cursor {
				p.pos = p.cursor
			}
			p.ctx.CurrentKind = KindDisposition
			p.parseDispositionSpecifier()
			return
		}
	}

	// 'u' - usable
	if ch == 'u' {
		p.advance()
		p.ctx.CurrentKind = KindUsable
		if p.atCursorOrEnd() {
			return
		}
		if p.peek() == ':' {
			p.advance()
			p.parseSpecifier()
		}
		return
	}

	// Stream type letter
	if isStreamTypeLetter(ch) {
		p.advance()
		p.ctx.CurrentKind = KindStreamType
		if p.atCursorOrEnd() {
			p.addExpected(ExpectedStreamIndex)
			return
		}
		if p.peek() == ':' {
			p.advance()
			p.parseSpecifier()
		}
		return
	}

	// 'g:' - group
	if ch == 'g' {
		if p.pos+1 < len(p.input) && p.input[p.pos+1] == ':' {
			p.pos += 2
			if p.pos > p.cursor {
				p.pos = p.cursor
			}
			p.ctx.CurrentKind = KindGroup
			p.parseGroupSpecifier()
			return
		}
		// 'g' alone without colon - might be partial
		p.advance()
		p.ctx.CurrentKind = KindGroup
		p.addForm(":", "colon after 'g'")
		return
	}

	// 'p:' - program
	if ch == 'p' {
		if p.pos+1 < len(p.input) && p.input[p.pos+1] == ':' {
			p.pos += 2
			if p.pos > p.cursor {
				p.pos = p.cursor
			}
			p.ctx.CurrentKind = KindProgram
			p.parseProgramSpecifier()
			return
		}
		p.advance()
		p.ctx.CurrentKind = KindProgram
		p.addForm(":", "colon after 'p'")
		return
	}

	// '#' - stream ID
	if ch == '#' {
		p.advance()
		p.ctx.CurrentKind = KindStreamID
		if p.atCursorOrEnd() {
			p.addExpected(ExpectedStreamIDValue)
			return
		}
		p.scanHexOrDecimalID()
		if p.atCursorOrEnd() {
			return
		}
		if p.peek() == ':' {
			p.advance()
			p.parseSpecifier()
		}
		return
	}

	// 'i:' - stream ID (alternate)
	if ch == 'i' {
		if p.pos+1 < len(p.input) && p.input[p.pos+1] == ':' {
			p.pos += 2
			if p.pos > p.cursor {
				p.pos = p.cursor
			}
			p.ctx.CurrentKind = KindStreamID
			if p.atCursorOrEnd() {
				p.addExpected(ExpectedStreamIDValue)
				return
			}
			p.scanHexOrDecimalID()
			if p.atCursorOrEnd() {
				return
			}
			if p.peek() == ':' {
				p.advance()
				p.parseSpecifier()
			}
			return
		}
	}

	// 'm:' - metadata
	if ch == 'm' {
		if p.pos+1 < len(p.input) && p.input[p.pos+1] == ':' {
			p.pos += 2
			if p.pos > p.cursor {
				p.pos = p.cursor
			}
			p.ctx.CurrentKind = KindMetadata
			p.parseMetadataSpecifier()
			return
		}
	}

	// Numeric stream index
	if ch >= '0' && ch <= '9' {
		p.scanIntegerForCompletion()
		p.ctx.CurrentKind = KindStreamIndex
		if p.atCursorOrEnd() {
			return
		}
		if p.peek() == ':' {
			p.advance()
			p.parseSpecifier()
		}
		return
	}

	p.addTopLevelForms()
	p.addExpected(ExpectedSpecifierType)
}

func (p *compParser) parseGroupSpecifier() {
	if p.atCursorOrEnd() {
		p.addExpected(ExpectedGroupIndex)
		p.addExpected(ExpectedGroupID)
		p.addForm("#", "group by ID")
		p.addForm("i", "group by ID (alternate)", ":")
		return
	}
	ch := p.peek()
	if ch == '#' {
		p.advance()
		if p.atCursorOrEnd() {
			p.addExpected(ExpectedGroupID)
			return
		}
		p.scanHexOrDecimalID()
	} else if ch == 'i' && p.peekRaw(1) == ':' {
		p.pos += 2
		if p.pos > p.cursor {
			p.pos = p.cursor
		}
		if p.atCursorOrEnd() {
			p.addExpected(ExpectedGroupID)
			return
		}
		p.scanHexOrDecimalID()
	} else if ch >= '0' && ch <= '9' {
		p.scanIntegerForCompletion()
	}
	if p.atCursorOrEnd() {
		return
	}
	if p.peek() == ':' {
		p.advance()
		p.parseSpecifier()
	}
}

func (p *compParser) parseProgramSpecifier() {
	if p.atCursorOrEnd() {
		p.addExpected(ExpectedProgramID)
		return
	}
	p.scanHexOrDecimalID()
	if p.atCursorOrEnd() {
		return
	}
	if p.peek() == ':' {
		p.advance()
		p.parseSpecifier()
	}
}

func (p *compParser) parseMetadataSpecifier() {
	if p.atCursorOrEnd() {
		p.addExpected(ExpectedMetadataKey)
		return
	}
	start := p.pos
	for !p.atCursorOrEnd() {
		ch := p.peek()
		if ch == ':' {
			break
		}
		p.advance()
	}
	keyPart := p.input[start:p.pos]
	p.ctx.PartialIdent = keyPart
	p.ctx.MetadataKey = keyPart

	if p.atCursorOrEnd() {
		p.addExpected(ExpectedMetadataKey)
		if keyPart != "" {
			p.addExpected(ExpectedMetadataValue)
		}
		return
	}

	// Consume colon after key
	p.advance()
	p.ctx.InMetadataValue = true
	start = p.pos
	for !p.atCursorOrEnd() {
		ch := p.peek()
		if ch == ':' {
			break
		}
		p.advance()
	}
	p.ctx.PartialIdent = p.input[start:p.pos]
	p.addExpected(ExpectedMetadataValue)

	// If cursor is at a colon after the value, transition to additional specifier
	if !p.atCursorOrEnd() && p.peek() == ':' {
		p.ctx.InMetadataValue = false
		p.advance()
		p.parseSpecifier()
	}
}

func (p *compParser) parseDispositionSpecifier() {
	if p.atCursorOrEnd() {
		p.addExpected(ExpectedDispositionName)
		return
	}
	var dispositions []string
	start := p.pos
	for !p.atCursorOrEnd() {
		ch := p.peek()
		if ch == '+' {
			if p.pos > start {
				dispositions = append(dispositions, p.input[start:p.pos])
			}
			p.advance()
			start = p.pos
			continue
		}
		if ch == ':' {
			break
		}
		p.advance()
	}
	if p.pos > start {
		dispositions = append(dispositions, p.input[start:p.pos])
	}
	p.ctx.Dispositions = dispositions
	if len(dispositions) > 0 {
		p.ctx.PartialIdent = dispositions[len(dispositions)-1]
	}
	p.addExpected(ExpectedDispositionName)

	// If cursor is at a colon after dispositions, transition to additional specifier
	if !p.atCursorOrEnd() && p.peek() == ':' {
		p.advance()
		p.parseSpecifier()
	}
}

func (p *compParser) scanIntegerForCompletion() {
	start := p.pos
	for !p.atCursorOrEnd() && p.peek() >= '0' && p.peek() <= '9' {
		p.advance()
	}
	if p.pos > start {
		p.ctx.PartialIdent = p.input[start:p.pos]
	}
}

func (p *compParser) scanHexOrDecimalID() {
	start := p.pos
	for !p.atCursorOrEnd() {
		ch := p.input[p.pos]
		if isHexDigit(ch) || ch == 'x' || ch == 'X' {
			p.pos++
		} else {
			break
		}
	}
	if p.pos > start {
		p.ctx.PartialIdent = p.input[start:p.pos]
	}
}

func dedupTokens(tokens []ExpectedToken) []ExpectedToken {
	seen := make(map[ExpectedToken]bool)
	result := make([]ExpectedToken, 0, len(tokens))
	for _, t := range tokens {
		if !seen[t] {
			seen[t] = true
			result = append(result, t)
		}
	}
	return result
}

func dedupForms(forms []SpecifierForm) []SpecifierForm {
	seen := make(map[string]bool)
	result := make([]SpecifierForm, 0, len(forms))
	for _, f := range forms {
		key := fmt.Sprintf("%s:%s", f.Prefix, f.Description)
		if !seen[key] {
			seen[key] = true
			result = append(result, f)
		}
	}
	slices.SortFunc(result, func(a, b SpecifierForm) int {
		if a.Prefix < b.Prefix {
			return -1
		}
		if a.Prefix > b.Prefix {
			return 1
		}
		return 0
	})
	return result
}
