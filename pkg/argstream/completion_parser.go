package argstream

import "strings"

// ParseForCompletion parses a partial ffmpeg command argument list and returns
// a CompletionContext describing what is expected at the end.
// The args should be the raw arguments (e.g. from os.Args[1:]) up to the cursor.
// trailingSpace indicates whether the cursor is at a new position after the
// last argument (true) or mid-token within the last argument (false).
func ParseForCompletion(args []string, trailingSpace bool) *CompletionContext {
	return ParseForCompletionWithProfile(args, trailingSpace, DefaultFFmpegProfile)
}

// ParseForCompletionWithProfile parses a partial ff* tool argument list using the given profile.
func ParseForCompletionWithProfile(args []string, trailingSpace bool, profile *ToolProfile) *CompletionContext {
	ctx := &CompletionContext{
		Scope:       ScopeGlobal,
		InputCount:  0,
		OutputCount: 0,
	}

	i := 0
	var pendingOption *OptionContext
	var pendingSpecOption *OptionContext // option waiting for a stream specifier value

	for i < len(args) {
		arg := args[i]

		// If we have a pending stream specifier and the current arg is not an option,
		// consume it as the specifier then transition to pending value
		if pendingSpecOption != nil {
			if !isOption(arg) {
				if i == len(args)-1 && !trailingSpace {
					// Completing the stream specifier (mid-token)
					ctx.CurrentOption = pendingSpecOption
					ctx.PartialSpec = arg
					ctx.ExpectedTokens = append(ctx.ExpectedTokens, ExpectedStreamSpecifier)
					return ctx
				}
				// Consume the stream specifier and mark value as pending
				pendingSpecOption.StreamSpecifier = arg
				pendingOption = pendingSpecOption
				pendingSpecOption = nil
				i++
				continue
			}
			// Stream specifier was skipped — clear pending spec
			pendingSpecOption = nil
		}

		// If we have a pending option value and the current arg is not an option,
		// consume it as the value
		if pendingOption != nil {
			if !isOption(arg) {
				// Shells split "-c:v" at the colon (COMP_WORDBREAKS), producing
				// args [-c, :v] or [-c, v]. Detect this: when the pending option
				// accepts a stream specifier and the partial starts with colon,
				// it's actually a stream specifier continuation.
				if i == len(args)-1 && !trailingSpace {
					if pendingOption.AcceptsSpec && strings.HasPrefix(arg, ":") {
						ctx.CurrentOption = pendingOption
						ctx.PartialSpec = strings.TrimPrefix(arg, ":")
						ctx.ExpectedTokens = append(ctx.ExpectedTokens, ExpectedStreamSpecifier)
						return ctx
					}
					ctx.CurrentOption = pendingOption
					ctx.PartialValue = arg
					ctx.ExpectedTokens = append(ctx.ExpectedTokens, ExpectedOptionValue)
					return ctx
				}
				// When the shell splits "-c:" into [-c, :], the ":" arg
				// is consumed as the stream specifier prefix.
				if pendingOption.AcceptsSpec && strings.HasPrefix(arg, ":") {
					spec := strings.TrimPrefix(arg, ":")
					if spec == "" {
						// "-c:" split into [-c, :] — spec is empty, expect specifier at next position
						pendingSpecOption = pendingOption
						pendingOption = nil
						i++
						continue
					}
					pendingOption.StreamSpecifier = spec
					// Don't clear pendingOption — the option still needs its value.
					// The specifier was consumed from the shell-split, but the
					// value argument hasn't been provided yet.
					i++
					continue
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

			baseName, spec, hasColon := ParseOptionName(optName)
			optDef := profile.LookupOption(baseName)

			// Check if we're at the last argument and it's the one being completed
			if i == len(args)-1 && !trailingSpace {
				ctx.PartialOption = baseName
				ctx.PartialSpec = spec
				ctx.CurrentOption = buildOptionContext(baseName, spec, optDef)

				if optDef != nil && optDef.AcceptsSpec && spec == "" && optDef.ImplicitSpec == "" && hasColon {
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
						addScopeOptionsForProfile(ctx, ctx.Scope, profile)
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
					ctx.InputURLs = append(ctx.InputURLs, args[i])
					ctx.Scope = ScopeInputFile
					i++ // consume the URL
				} else {
					ctx.ExpectedTokens = append(ctx.ExpectedTokens, ExpectedInputURL)
					return ctx
				}
				continue
			}

			// If the option accepts a stream specifier and a colon was present
			// but no spec was provided, the next arg is the stream specifier.
			// (e.g. "-c:" "v" "libx264")
			// Without a colon (e.g. "-c" "libx264"), the value comes directly.
			if optDef != nil && optDef.AcceptsSpec && hasColon && spec == "" && optDef.ImplicitSpec == "" && optDef.Type == TypeValue {
				pendingSpecOption = buildOptionContext(baseName, spec, optDef)
				updateScope(ctx, optDef, profile)
				continue
			}

			// If the option takes a value, mark it as pending
			if optDef != nil && optDef.Type == TypeValue && (spec != "" || !optDef.AcceptsSpec || optDef.ImplicitSpec != "" || !hasColon) {
				pendingOption = buildOptionContext(baseName, spec, optDef)
			}

			// Update scope based on option
			updateScope(ctx, optDef, profile)
		} else {
			// Non-option: could be an output URL (ffmpeg) or input URL (ffplay/ffprobe),
			// or a partial option being typed (e.g. bare "-" at cursor position).
			if i == len(args)-1 && !trailingSpace && strings.HasPrefix(arg, "-") {
				// Partial option being typed — don't change scope,
				// return option completions based on current scope.
				ctx.PartialOption = strings.TrimLeft(arg, "-")
				addScopeOptionsForProfile(ctx, ctx.Scope, profile)
				return ctx
			}
			if profile.HasOutputSection {
				ctx.OutputCount++
				ctx.Scope = ScopeOutputFile
			} else {
				ctx.InputCount++
				ctx.Scope = ScopeInputFile
			}
			i++
		}
	}

	// If we have a pending stream specifier, that's what's expected next
	if pendingSpecOption != nil {
		ctx.CurrentOption = pendingSpecOption
		ctx.ExpectedTokens = append(ctx.ExpectedTokens, ExpectedStreamSpecifier)
		return ctx
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
		tokens := []ExpectedToken{ExpectedInputOption, ExpectedInputURL}
		if profile.HasOutputSection {
			tokens = append(tokens, ExpectedOutputOption, ExpectedOutputURL)
		}
		ctx.ExpectedTokens = append(ctx.ExpectedTokens, tokens...)
	case ScopeOutputFile:
		ctx.ExpectedTokens = append(ctx.ExpectedTokens, ExpectedOutputOption, ExpectedOutputURL)
	}

	return ctx
}

func updateScope(ctx *CompletionContext, optDef *OptionDef, profile *ToolProfile) {
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
		if profile.HasOutputSection {
			ctx.Scope = ScopeOutputFile
		}
	case ScopePerStreamOpt:
		// Per-stream stays in current scope
	}
}

func addScopeOptionsForProfile(ctx *CompletionContext, scope Scope, profile *ToolProfile) {
	switch scope {
	case ScopeGlobal:
		ctx.ExpectedTokens = append(ctx.ExpectedTokens, ExpectedGlobalOption, ExpectedInputOption)
	case ScopeInputFile:
		tokens := []ExpectedToken{ExpectedInputOption}
		if profile.HasOutputSection {
			tokens = append(tokens, ExpectedOutputOption)
		}
		ctx.ExpectedTokens = append(ctx.ExpectedTokens, tokens...)
	case ScopeOutputFile:
		ctx.ExpectedTokens = append(ctx.ExpectedTokens, ExpectedOutputOption)
	}
}

func buildOptionContext(baseName, spec string, optDef *OptionDef) *OptionContext {
	if optDef == nil {
		return &OptionContext{
			Name:        baseName,
			AcceptsSpec: false,
			IsBoolean:   false,
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
