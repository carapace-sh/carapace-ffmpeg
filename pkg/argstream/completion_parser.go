package argstream

import "strings"

// ParseForCompletion parses a partial ffmpeg command argument list and returns
// a CompletionContext describing what is expected at the end.
// The args should be the raw arguments (e.g. from os.Args[1:]) up to the cursor.
// trailingSpace indicates whether the cursor is at a new position after the
// last argument (true) or mid-token within the last argument (false).
func ParseForCompletion(args []string, trailingSpace bool) *CompletionContext {
	ctx := &CompletionContext{
		Scope:       ScopeGlobal,
		InputCount:  0,
		OutputCount: 0,
	}

	i := 0
	var pendingOption *OptionContext

	for i < len(args) {
		arg := args[i]

		// If we have a pending option value and the current arg is not an option,
		// consume it as the value
		if pendingOption != nil {
			if !isOption(arg) {
				if i == len(args)-1 && !trailingSpace {
					// Completing the option value (mid-token)
					ctx.CurrentOption = pendingOption
					ctx.PartialValue = arg
					ctx.ExpectedTokens = append(ctx.ExpectedTokens, ExpectedOptionValue)
					return ctx
				}
				// Consume the value
				pendingOption = nil
				i++
				continue
			}
			// The pending value was skipped (next arg is an option) — clear pending
			pendingOption = nil
		}

		// Check if this is an option
		if isOption(arg) {
			optName := arg[1:] // strip '-'
			optName = strings.TrimPrefix(optName, "-")

			baseName, spec := ParseOptionName(optName)
			optDef := LookupOption(baseName)

			// Check if we're at the last argument and it's the one being completed
			if i == len(args)-1 && !trailingSpace {
				ctx.PartialOption = baseName
				ctx.PartialSpec = spec
				ctx.CurrentOption = buildOptionContext(baseName, spec, optDef)

				if optDef != nil && optDef.AcceptsSpec && spec == "" && optDef.ImplicitSpec == "" {
					ctx.ExpectedTokens = append(ctx.ExpectedTokens, ExpectedStreamSpecifier)
				}

				optionComplete := optDef != nil && optDef.Type == TypeValue && (spec != "" || !optDef.AcceptsSpec || optDef.ImplicitSpec != "")
				if optionComplete {
					ctx.ExpectedTokens = append(ctx.ExpectedTokens, ExpectedOptionValue)
					ctx.PartialValue = ""
				} else {
					switch {
					case optDef != nil && optDef.Scope == ScopeGlobalOpt:
						ctx.ExpectedTokens = append(ctx.ExpectedTokens, ExpectedGlobalOption)
						ctx.Scope = ScopeGlobal
					case baseName == "i":
						ctx.ExpectedTokens = append(ctx.ExpectedTokens, ExpectedInputURL)
					default:
						addScopeOption(ctx, optDef, baseName)
					}
				}
				return ctx
			}

			// Not completing the last arg — consume it
			i++

			// Handle -i (input file marker)
			if baseName == "i" {
				if i < len(args) && !isOption(args[i]) {
					ctx.InputCount++
					ctx.Scope = ScopeInputFile
					i++ // consume the URL
				}
				continue
			}

			// If the option takes a value, mark it as pending
			if optDef != nil && optDef.Type == TypeValue && (spec != "" || !optDef.AcceptsSpec || optDef.ImplicitSpec != "") {
				pendingOption = buildOptionContext(baseName, spec, optDef)
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

	// If we have a pending option value, that's what's expected next
	if pendingOption != nil {
		ctx.CurrentOption = pendingOption
		ctx.ExpectedTokens = append(ctx.ExpectedTokens, ExpectedOptionValue)
		return ctx
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
	effectiveSpec := spec
	if spec == "" && optDef.ImplicitSpec != "" {
		effectiveSpec = optDef.ImplicitSpec
	}
	return &OptionContext{
		Name:            baseName,
		CanonicalName:   optDef.CanonicalName,
		StreamSpecifier: effectiveSpec,
		ValueType:       optDef.ValueType,
		AcceptsSpec:     optDef.AcceptsSpec && optDef.ImplicitSpec == "",
		IsBoolean:       optDef.Type == TypeBoolean,
		Style:           optDef.Style(),
	}
}
