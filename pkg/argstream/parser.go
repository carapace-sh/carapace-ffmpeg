package argstream

import (
	"strings"
)

type ParseError struct {
	Message string
	Span    Span
}

func (e *ParseError) Error() string {
	return e.Message
}

// Parse parses a ffmpeg command argument list into a Program AST.
// The input should be the raw arguments (e.g. from os.Args[1:]).
func Parse(args []string) (*Program, error) {
	return ParseWithProfile(args, DefaultFFmpegProfile)
}

// ParseWithProfile parses an ff* tool argument list into a Program AST using the given profile.
func ParseWithProfile(args []string, profile *ToolProfile) (*Program, error) {
	p := &parser{args: args, profile: profile}
	return p.parseProgram()
}

type parser struct {
	args    []string
	pos     int
	profile *ToolProfile
}

func (p *parser) syntaxError(msg string) *ParseError {
	span := Span{Start: p.pos, End: p.pos + 1}
	if p.pos < len(p.args) {
		span.End = p.pos + 1
	}
	return &ParseError{Message: msg, Span: span}
}

func (p *parser) atEnd() bool {
	return p.pos >= len(p.args)
}

func (p *parser) peek() string {
	if p.pos >= len(p.args) {
		return ""
	}
	return p.args[p.pos]
}

func (p *parser) advance() string {
	if p.pos >= len(p.args) {
		return ""
	}
	arg := p.args[p.pos]
	p.pos++
	return arg
}

func (p *parser) parseProgram() (*Program, error) {
	start := p.pos
	prog := &Program{}

	scope := ScopeGlobal
	inputCount := 0
	outputCount := 0

	for !p.atEnd() {
		arg := p.peek()

		// Option token: starts with '-' and is not just '-'
		if isOption(arg) {
			optName := arg[1:] // strip leading '-'
			baseName, spec, _ := ParseOptionName(optName)
			optDef := p.profile.lookupOption(baseName)

			// Determine the option's effective scope based on position
			switch {
			case optDef != nil && optDef.Scope == ScopeGlobalOpt:
				// Global option - OK anywhere but semantically global
				p.advance()
				tok := Token{
					Kind:            KindGlobalOption,
					OptionName:      baseName,
					StreamSpecifier: spec,
					Span:            Span{Start: p.pos - 1, End: p.pos},
				}
				if optDef.Type == TypeValue {
					if p.atEnd() {
						// Missing value — create token anyway, value is empty
					} else {
						tok.Value = p.advance()
						tok.Span.End = p.pos
					}
				}
				prog.Tokens = append(prog.Tokens, tok)

			case baseName == "i":
				// Input file marker
				p.advance()
				scope = ScopeInputFile
				if p.atEnd() {
					return nil, p.syntaxError("expected input URL after -i")
				}
				url := p.advance()
				inp := InputFile{Index: inputCount, URL: url, Span: Span{Start: p.pos - 1, End: p.pos}}
				prog.InputFiles = append(prog.InputFiles, inp)
				prog.Tokens = append(prog.Tokens, Token{
					Kind: KindInputURL,
					URL:  url,
					Span: Span{Start: p.pos - 1, End: p.pos},
				})
				inputCount++
				scope = ScopeInputFile

			case optDef != nil && (optDef.Scope == ScopeInputOnlyOpt || optDef.Scope == ScopePerFileOpt):
				// Input option
				p.advance()
				var kind TokenKind
				if scope == ScopeOutputFile {
					kind = KindOutputOption
				} else {
					kind = KindInputOption
				}
				tok := Token{
					Kind:            kind,
					OptionName:      baseName,
					StreamSpecifier: spec,
					Span:            Span{Start: p.pos - 1, End: p.pos},
				}
				if optDef.Type == TypeValue {
					if !p.atEnd() {
						tok.Value = p.advance()
						tok.Span.End = p.pos
					}
				}
				prog.Tokens = append(prog.Tokens, tok)

			case optDef != nil && optDef.Scope == ScopeOutputOnlyOpt:
				if !p.profile.HasOutputSection {
					break
				}
				// Output option (valid in output section)
				p.advance()
				tok := Token{
					Kind:            KindOutputOption,
					OptionName:      baseName,
					StreamSpecifier: spec,
					Span:            Span{Start: p.pos - 1, End: p.pos},
				}
				if optDef.Type == TypeValue {
					if !p.atEnd() {
						tok.Value = p.advance()
						tok.Span.End = p.pos
					}
				}
				prog.Tokens = append(prog.Tokens, tok)

			case optDef != nil && optDef.Scope == ScopePerStreamOpt:
				// Per-stream option — scope depends on position
				p.advance()
				var kind TokenKind
				if scope == ScopeOutputFile {
					kind = KindOutputOption
				} else {
					kind = KindInputOption
				}
				effectiveSpec := spec
				if spec == "" && optDef.ImplicitSpec != "" {
					effectiveSpec = optDef.ImplicitSpec
				}
				tok := Token{
					Kind:            kind,
					OptionName:      baseName,
					StreamSpecifier: effectiveSpec,
					Span:            Span{Start: p.pos - 1, End: p.pos},
				}
				if optDef.Type == TypeValue {
					if !p.atEnd() {
						tok.Value = p.advance()
						tok.Span.End = p.pos
					}
				}
				prog.Tokens = append(prog.Tokens, tok)

			default:
				// Unknown option — treat as output option or output URL
				// depending on whether it looks like an option
				p.advance()
				tok := Token{
					Kind:            KindOutputOption,
					OptionName:      baseName,
					StreamSpecifier: spec,
					Span:            Span{Start: p.pos - 1, End: p.pos},
				}
				// Heuristic: if the option name is >1 char and known options
				// take values, assume this one does too
				if len(baseName) > 1 && !p.isKnownBoolean(baseName) {
					if !p.atEnd() && !isOption(p.peek()) {
						tok.Value = p.advance()
						tok.Span.End = p.pos
					}
				}
				prog.Tokens = append(prog.Tokens, tok)
				scope = ScopeOutputFile
			}

		} else {
			// Non-option: treat as output URL (ffmpeg) or input URL (ffplay/ffprobe)
			p.advance()
			url := p.args[p.pos-1]
			if p.profile.HasOutputSection {
				prog.Tokens = append(prog.Tokens, Token{
					Kind: KindOutputURL,
					URL:  url,
					Span: Span{Start: p.pos - 1, End: p.pos},
				})
				prog.OutputFiles = append(prog.OutputFiles, OutputFile{
					Index: outputCount,
					URL:   url,
					Span:  Span{Start: p.pos - 1, End: p.pos},
				})
				outputCount++
				scope = ScopeOutputFile
			} else {
				// For ffplay/ffprobe, positional arg is the input URL
				prog.Tokens = append(prog.Tokens, Token{
					Kind: KindInputURL,
					URL:  url,
					Span: Span{Start: p.pos - 1, End: p.pos},
				})
				prog.InputFiles = append(prog.InputFiles, InputFile{
					Index: inputCount,
					URL:   url,
					Span:  Span{Start: p.pos - 1, End: p.pos},
				})
				inputCount++
				scope = ScopeInputFile
			}
		}
	}

	prog.Span = Span{Start: start, End: p.pos}
	return prog, nil
}

// isOption checks if an argument looks like an ffmpeg option.
func isOption(arg string) bool {
	if len(arg) == 0 || arg[0] != '-' {
		return false
	}
	// '-' alone is not an option (stdin/stdout)
	if arg == "-" {
		return false
	}
	// '--' is end-of-options marker
	if arg == "--" {
		return false
	}
	// '--help' is accepted as an alias for '-h'
	if strings.HasPrefix(arg, "--") {
		return true
	}
	return true
}

// isKnownBoolean checks if a name is a known boolean option.
func (p *parser) isKnownBoolean(name string) bool {
	if opt, ok := p.profile.OptionIndex[name]; ok {
		return opt.Type == TypeBoolean
	}
	// Fallback for options not yet in the index
	return false
}