package argstream

import "strings"

// ParseForCompletion parses a partial ffmpeg command argument list and returns
// a CompletionContext describing what is expected at the end.
// The args should be the raw arguments (e.g. from os.Args[1:]) up to the cursor.
func ParseForCompletion(args []string) *CompletionContext {
	ctx := &CompletionContext{
		Scope:       ScopeGlobal,
		InputCount:  0,
		OutputCount: 0,
	}

	i := 0
	for i < len(args) {
		arg := args[i]

		// Check if this is an option
		if isOption(arg) {
			optName := arg[1:] // strip '-'
			optName = strings.TrimPrefix(optName, "-")

			baseName, spec := ParseOptionName(optName)
			optDef := LookupOption(baseName)

			// Check if we're at the last argument (the one being completed)
			if i == len(args)-1 {
				// We're completing the option name or its specifier
				ctx.PartialOption = baseName
				ctx.PartialSpec = spec
				ctx.CurrentOption = buildOptionContext(baseName, spec, optDef)

				// If the option is complete (has a space after it), the cursor
				// is at a new position — but in our model, the last arg IS the option,
				// so we're completing the option name itself.
				if optDef != nil && optDef.AcceptsSpec && spec == "" {
					ctx.ExpectedTokens = append(ctx.ExpectedTokens, ExpectedStreamSpecifier)
				}

				switch {
				case optDef != nil && optDef.Scope == ScopeGlobalOpt:
					ctx.ExpectedTokens = append(ctx.ExpectedTokens, ExpectedGlobalOption)
					ctx.Scope = ScopeGlobal
				case baseName == "i":
					ctx.ExpectedTokens = append(ctx.ExpectedTokens, ExpectedInputURL)
				default:
					addScopeOption(ctx, optDef, baseName)
				}
				if optDef != nil && optDef.Type == TypeValue {
					if spec != "" || !optDef.AcceptsSpec {
						ctx.ExpectedTokens = append(ctx.ExpectedTokens, ExpectedOptionValue)
						ctx.PartialValue = ""
					}
				}
				return ctx
			}

			// Not the last arg — consume it and possibly its value
			i++

			// Handle -i (input file marker)
			if baseName == "i" {
				if i < len(args) {
					ctx.InputCount++
					ctx.Scope = ScopeInputFile
					i++ // consume the URL
				}
				continue
			}

			// Consume option value if it takes one
			if optDef != nil && optDef.Type == TypeValue {
				if i < len(args) {
					if i == len(args)-1 && !isOption(args[i]) {
						// The value position is the cursor — complete it
						ctx.CurrentOption = buildOptionContext(baseName, spec, optDef)
						ctx.PartialValue = args[i]
						ctx.ExpectedTokens = append(ctx.ExpectedTokens, ExpectedOptionValue)
						return ctx
					}
					if !isOption(args[i]) {
						i++ // consume the value
					}
				}
			}

			// Update scope based on option
			updateScope(ctx, optDef)
		} else {
			// Non-option: output URL
			ctx.OutputCount++
			ctx.Scope = ScopeOutputFile
			i++
		}
	}

	// If we've consumed all args, we're at a new completion position
	switch ctx.Scope {
	case ScopeGlobal:
		ctx.ExpectedTokens = append(ctx.ExpectedTokens, ExpectedGlobalOption, ExpectedInputOption, ExpectedInputURL)
	case ScopeInputFile:
		ctx.ExpectedTokens = append(ctx.ExpectedTokens, ExpectedInputOption, ExpectedInputURL, ExpectedOutputOption, ExpectedOutputURL)
	case ScopeOutputFile:
		ctx.ExpectedTokens = append(ctx.ExpectedTokens, ExpectedOutputOption, ExpectedOutputURL)
	}

	return ctx
}

func updateScope(ctx *CompletionContext, optDef *OptionDef) {
	if optDef == nil {
		return
	}
	switch optDef.Scope {
	case ScopeGlobalOpt:
		// Global stays in current scope
	case ScopePerFileOpt:
		// Don't change scope
	case ScopeInputOnlyOpt:
		ctx.Scope = ScopeInputFile
	case ScopeOutputOnlyOpt:
		ctx.Scope = ScopeOutputFile
	case ScopePerStreamOpt:
		// Per-stream stays in current scope
	}
}

func addScopeOption(ctx *CompletionContext, _ *OptionDef, _ string) {
	switch ctx.Scope {
	case ScopeGlobal, ScopeInputFile:
		ctx.ExpectedTokens = append(ctx.ExpectedTokens, ExpectedInputOption)
	case ScopeOutputFile:
		ctx.ExpectedTokens = append(ctx.ExpectedTokens, ExpectedOutputOption)
	}
}

func buildOptionContext(baseName, spec string, optDef *OptionDef) *OptionContext {
	if optDef == nil {
		return &OptionContext{
			Name:        baseName,
			AcceptsSpec: false,
			IsBoolean:  false,
		}
	}
	return &OptionContext{
		Name:            baseName,
		CanonicalName:   optDef.CanonicalName,
		StreamSpecifier: spec,
		ValueType:       optDef.ValueType,
		AcceptsSpec:     optDef.AcceptsSpec,
		IsBoolean:       optDef.Type == TypeBoolean,
		Style:           optDef.Style(),
	}
}