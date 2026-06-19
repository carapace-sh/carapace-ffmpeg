package completer

import (
	"slices"
	"strings"

	"github.com/carapace-sh/carapace"
	ffmpeg "github.com/carapace-sh/carapace-ffmpeg/pkg/actions/tools/ffmpeg"
	"github.com/carapace-sh/carapace-ffmpeg/pkg/argstream"
	"github.com/carapace-sh/carapace-ffmpeg/pkg/filtergraph"
	"github.com/carapace-sh/carapace-ffmpeg/pkg/probe"
	"github.com/carapace-sh/carapace-ffmpeg/pkg/streamspec"
)

// ContextToArgs converts carapace.Context to the args and trailingSpace
// expected by argstream.ParseForCompletion.
func ContextToArgs(c carapace.Context) (args []string, trailingSpace bool) {
	n := len(c.Args)
	if n > 0 && c.Args[n-1] == "" {
		n--
	}
	args = c.Args[:n]
	if c.Value != "" {
		args = append(args, c.Value)
	}
	trailingSpace = c.Value == ""
	return
}

// IsMidTokenOptionWithSpec returns true when the current token is an option
// that contains a colon AND the option accepts stream specifiers.
func IsMidTokenOptionWithSpec(value string, profile *argstream.ToolProfile) bool {
	if !strings.HasPrefix(value, "-") || !strings.Contains(value, ":") {
		return false
	}
	optText := strings.TrimPrefix(value[1:], "-")
	baseName, _, _ := argstream.ParseOptionName(optText)
	optDef := profile.LookupOption(baseName)
	return optDef != nil && optDef.AcceptsSpec && optDef.ImplicitSpec == ""
}

// LookupOption looks up an option by name in the given profile's option index.
func LookupOption(profile *argstream.ToolProfile, name string) *argstream.OptionDef {
	return profile.LookupOption(name)
}

// ActionPartialOption handles completion when the cursor is mid-token within a
// partial option name (e.g. typing `-v` which might match `-vcodec`, `-vframes`,
// `-vn`, etc.). It returns option name completions so the shell can filter them
// against the partial prefix. For value options that have been recognized, it also
// includes the recognized option's value completions.
func ActionPartialOption(ctx *argstream.CompletionContext, profile *argstream.ToolProfile) carapace.Action {
	actions := []carapace.Action{ActionOptionNamesWithSpecSuffix(ctx, profile)}
	return carapace.Batch(actions...).ToA()
}

// ActionOptionNamesWithSpecSuffix returns option name completions for the given profile.
// Options that accept stream specifiers get Suffix(":") so the user
// can continue typing the specifier within the same token.
func ActionOptionNamesWithSpecSuffix(ctx *argstream.CompletionContext, profile *argstream.ToolProfile) carapace.Action {
	var specOptions, noSpecOptions []string
	for name, def := range profile.OptionIndex {
		switch {
		case def.Scope == argstream.ScopeGlobalOpt && !containsToken(ctx.ExpectedTokens, argstream.ExpectedGlobalOption):
			continue
		case def.Scope == argstream.ScopeInputOnlyOpt && !containsToken(ctx.ExpectedTokens, argstream.ExpectedInputOption):
			continue
		case def.Scope == argstream.ScopeOutputOnlyOpt && !containsToken(ctx.ExpectedTokens, argstream.ExpectedOutputOption):
			continue
		}

		if def.AcceptsSpec && def.ImplicitSpec == "" && def.Type == argstream.TypeValue {
			specOptions = append(specOptions, "-"+name, def.Description, def.Style())
		} else {
			noSpecOptions = append(noSpecOptions, "-"+name, def.Description, def.Style())
		}
	}

	specAction := carapace.ActionStyledValuesDescribed(specOptions...).Suffix(":").NoSpace(':')
	noSpecAction := carapace.ActionStyledValuesDescribed(noSpecOptions...)
	return carapace.Batch(specAction, noSpecAction).ToA()
}

// ActionOptionNames returns plain option name completions without
// any suffix or NoSpace (used inside ActionMultiParts).
func ActionOptionNames(ctx *argstream.CompletionContext, profile *argstream.ToolProfile) carapace.Action {
	var vals []string
	for name, def := range profile.OptionIndex {
		switch {
		case def.Scope == argstream.ScopeGlobalOpt && !containsToken(ctx.ExpectedTokens, argstream.ExpectedGlobalOption):
			continue
		case def.Scope == argstream.ScopeInputOnlyOpt && !containsToken(ctx.ExpectedTokens, argstream.ExpectedInputOption):
			continue
		case def.Scope == argstream.ScopeOutputOnlyOpt && !containsToken(ctx.ExpectedTokens, argstream.ExpectedOutputOption):
			continue
		}
		vals = append(vals, "-"+name, def.Description, def.Style())
	}
	return carapace.ActionStyledValuesDescribed(vals...)
}

// ActionOptions returns completions for option names appropriate to the current scope.
func ActionOptions(ctx *argstream.CompletionContext, profile *argstream.ToolProfile) carapace.Action {
	return ActionOptionNamesWithSpecSuffix(ctx, profile)
}

// ActionOptionValue returns completions for the value of the current option.
// filterValue is the partial filtergraph text when the option is ValueFilter (empty otherwise).
func ActionOptionValue(ctx *argstream.CompletionContext, codecAction func(*argstream.CompletionContext) carapace.Action, filterValue string) carapace.Action {
	if ctx.CurrentOption == nil {
		return carapace.ActionValues()
	}
	switch ctx.CurrentOption.ValueType {
	case argstream.ValueCodec:
		return codecAction(ctx)
	case argstream.ValueFormat:
		return ffmpeg.ActionFormats()
	case argstream.ValuePixelFormat:
		return ffmpeg.ActionPixelFormats()
	case argstream.ValueSampleFmt:
		return ffmpeg.ActionSampleFormats()
	case argstream.ValueChannelLayout:
		return ffmpeg.ActionChannelLayouts()
	case argstream.ValueFilter:
		isComplex := ctx.CurrentOption.CanonicalName == "filter_complex" || ctx.CurrentOption.CanonicalName == "lavfi"
		return ActionFilterValue(filterValue, isComplex)
	case argstream.ValueVideoSize:
		return ffmpeg.ActionVideoSizes()
	case argstream.ValueVideoRate:
		return ffmpeg.ActionFrameRates()
	case argstream.ValueBoolean:
		return ffmpeg.ActionBoolean()
	case argstream.ValueDisposition:
		return ffmpeg.ActionDispositions()
	case argstream.ValueBitrate:
		return ffmpeg.ActionBitrates()
	case argstream.ValueMapSpec:
		return carapace.ActionValues()
	case argstream.ValueMetadata:
		return carapace.ActionValues()
	case argstream.ValueFileURL:
		return carapace.ActionFiles()
	case argstream.ValueHWAccel:
		return ffmpeg.ActionHWAccels()
	case argstream.ValueLogLevel:
		return ffmpeg.ActionLogLevels()
	case argstream.ValueFPSMode:
		return ffmpeg.ActionFPSModes()
	case argstream.ValueCopyTB:
		return ffmpeg.ActionCopyTB()
	case argstream.ValueAbortOn:
		return ffmpeg.ActionAbortOn()
	case argstream.ValueDiscard:
		return ffmpeg.ActionDiscard()
	case argstream.ValueBSF:
		return ffmpeg.ActionBitstreamFilters()
	case argstream.ValuePrintGraphFmt:
		return ffmpeg.ActionPrintGraphsFormats()
	case argstream.ValueTarget:
		return ffmpeg.ActionTargets()
	case argstream.ValueTimestamp:
		return carapace.ActionValues("now")
	case argstream.ValueSwsFlags:
		return ffmpeg.ActionSwsFlags()
	case argstream.ValueDevice:
		if ctx.CurrentOption != nil && ctx.CurrentOption.CanonicalName == "sinks" {
			return ffmpeg.ActionDevices(ffmpeg.DeviceOpts{Demuxing: true})
		}
		return ffmpeg.ActionDevices(ffmpeg.DeviceOpts{Muxing: true})
	case argstream.ValueString:
		if ctx.CurrentOption != nil && ctx.CurrentOption.CanonicalName == "h" {
			return ffmpeg.ActionHelpTopics()
		}
		return carapace.ActionValues()
	case argstream.ValueShowMode:
		return ffmpeg.ActionShowModes()
	case argstream.ValueSyncType:
		return ffmpeg.ActionSyncTypes()
	case argstream.ValueStreamSpec:
		return ActionStreamSpecifiers()
	case argstream.ValueProbeOutputFmt:
		return ffmpeg.ActionProbeOutputFormats()
	case argstream.ValueDataDumpFmt:
		return ffmpeg.ActionDataDumpFormats()
	case argstream.ValueShowOptFields:
		return ffmpeg.ActionShowOptionalFields()
	case argstream.ValueVulkanParams:
		return carapace.ActionValues()
	default:
		return carapace.ActionValues()
	}
}

// ProbeAll probes all input URLs in the completion context and returns
// the merged stream info. Uses caching to avoid probing the same URL twice.
func ProbeAll(ctx *argstream.CompletionContext) []probe.StreamInfo {
	return probeURLs(ctx.InputURLs)
}

func probeURLs(urls []string) []probe.StreamInfo {
	var all []probe.StreamInfo
	seen := make(map[string]bool)
	for _, url := range urls {
		if seen[url] {
			continue
		}
		seen[url] = true
		streams, _ := probe.Probe(url)
		all = append(all, streams...)
	}
	return all
}

// streamTypeToCodecType maps a stream specifier type letter to the ffprobe codec_type value.
func streamTypeToCodecType(letter string) string {
	switch letter {
	case "v", "V":
		return "video"
	case "a":
		return "audio"
	case "s":
		return "subtitle"
	case "d":
		return "data"
	case "t":
		return "attachment"
	default:
		return ""
	}
}

// extractStreamTypeLetter extracts the stream type letter from a specifier string.
// For example "a:" returns "a", "a:1" returns "a", "v" returns "v", "m:language:eng" returns "".
// Returns empty for "disp:" since that's a disposition specifier, not a data stream.
func extractStreamTypeLetter(specifierPart string) string {
	if len(specifierPart) == 0 {
		return ""
	}
	// disp: is a disposition specifier, not a data stream type
	if len(specifierPart) >= 5 && specifierPart[:5] == "disp:" {
		return ""
	}
	ch := specifierPart[0]
	switch ch {
	case 'v', 'V', 'a', 's', 'd', 't':
		return string(ch)
	default:
		return ""
	}
}

// ActionStreamSpecifier handles stream specifier completion.
func ActionStreamSpecifier(ctx *argstream.CompletionContext, c carapace.Context) carapace.Action {
	return ActionStreamSpecifierWithStreams(ctx, c, nil)
}

// ActionStreamSpecifierWithStreams handles stream specifier completion with optional probed stream info.
func ActionStreamSpecifierWithStreams(ctx *argstream.CompletionContext, c carapace.Context, streams []probe.StreamInfo) carapace.Action {
	if ctx.CurrentOption == nil || !ctx.CurrentOption.AcceptsSpec {
		return carapace.ActionValues()
	}

	if colon, after, ok := strings.Cut(c.Value, ":"); ok {
		return actionStreamSpecifierAfter(after, streams).Invoke(
			carapace.Context{Value: after},
		).Prefix(colon + ":").ToA()
	}
	return ActionStreamSpecifiers()
}

// actionStreamSpecifierAfter returns context-aware completions for the
// specifier portion after the first colon (e.g. "l" in "-c:m:l").
func actionStreamSpecifierAfter(specifierPart string, streams []probe.StreamInfo) carapace.Action {
	return carapace.ActionCallback(func(c carapace.Context) carapace.Action {
		specCtx := streamspec.ParseForCompletion(specifierPart)

		prefixToReplace := specifierPart
		if specCtx.PartialIdent != "" {
			prefixToReplace = strings.TrimSuffix(specifierPart, specCtx.PartialIdent)
		}

		var actions []carapace.Action
		for _, token := range specCtx.ExpectedTokens {
			switch token {
			case streamspec.ExpectedSpecifierType, streamspec.ExpectedStreamTypeLetter:
				actions = append(actions, streamTypeActions(specCtx, prefixToReplace))
			case streamspec.ExpectedStreamIndex:
				actions = append(actions, actionStreamIndex(specCtx, specifierPart, streams, prefixToReplace))
			case streamspec.ExpectedMetadataKey:
				action := ffmpeg.ActionMetadataKeys()
				action = action.Invoke(carapace.Context{Value: specCtx.PartialIdent}).Prefix(prefixToReplace).ToA()
				actions = append(actions, action)
			case streamspec.ExpectedMetadataValue:
				actions = append(actions, actionMetadataValue(specCtx, streams, prefixToReplace))
			case streamspec.ExpectedDispositionName:
				action := actionDispositionName(specCtx, streams, prefixToReplace)
				actions = append(actions, action)
			case streamspec.ExpectedGroupSpecifier,
				streamspec.ExpectedGroupIndex,
				streamspec.ExpectedGroupID:
				actions = append(actions, carapace.ActionValues().Suffix(":").NoSpace(':'))
			case streamspec.ExpectedProgramID,
				streamspec.ExpectedStreamIDValue:
				actions = append(actions, carapace.ActionValues().Suffix(":").NoSpace(':'))
			}
		}

		if len(actions) == 0 {
			return streamTypeActions(specCtx, prefixToReplace)
		}
		return carapace.Batch(actions...).ToA()
	})
}

// streamTypeActions returns completions for stream specifier type forms
// filtered by the current partial ident. Forms that already end with ":"
// (e.g. "g:", "m:") get no additional suffix; others get Suffix(":")
// since all specifier types can be followed by additional specifiers.
func streamTypeActions(specCtx *streamspec.CompletionContext, prefixToReplace string) carapace.Action {
	var suffixed, unsuffixed []string
	for _, form := range specCtx.ValidForms {
		switch {
		case form.Prefix == "":
			continue
		case form.Suffix != "":
			suffixed = append(suffixed, form.Prefix, form.Description)
		default:
			unsuffixed = append(unsuffixed, form.Prefix, form.Description)
		}
	}
	var actions []carapace.Action
	if len(suffixed) > 0 {
		suffix := ":"
		for _, form := range specCtx.ValidForms {
			if form.Suffix != "" {
				suffix = form.Suffix
				break
			}
		}
		suffixedAction := carapace.ActionValuesDescribed(suffixed...).Uid("ffmpeg", "stream-specifier").Suffix(suffix).NoSpace(rune(suffix[0]))
		actions = append(actions, suffixedAction)
	}
	if len(unsuffixed) > 0 {
		actions = append(actions, carapace.ActionValuesDescribed(unsuffixed...).Uid("ffmpeg", "stream-specifier"))
	}
	combined := carapace.Batch(actions...).ToA()
	return combined.Invoke(carapace.Context{Value: specCtx.PartialIdent}).Prefix(prefixToReplace).ToA()
}

// ActionStreamSpecifiers returns completions for stream specifier types.
func ActionStreamSpecifiers() carapace.Action {
	return carapace.ActionValuesDescribed(
		"v", "video streams",
		"V", "video streams (excluding attached pictures)",
		"a", "audio streams",
		"s", "subtitle streams",
		"d", "data streams",
		"t", "attachment streams",
		"g", "stream group",
		"p", "program",
		"#", "stream by ID",
		"i", "stream by ID (alternate)",
		"m", "metadata",
		"disp", "disposition",
		"u", "usable configuration",
	).Uid("ffmpeg", "stream-specifier").Suffix(":").NoSpace(':')
}

// ActionStreamSpecifierParts returns context-aware stream specifier completions
// for the mid-token ActionMultiParts path. specifierPart is the full specifier
// text after the option name's colon, and partialValue is the currently typed
// part being completed (c.Value from ActionMultiParts callback).
func ActionStreamSpecifierParts(specifierPart string, partialValue string) carapace.Action {
	return ActionStreamSpecifierPartsWithStreams(specifierPart, partialValue, nil)
}

// ActionStreamSpecifierPartsWithStreams is like ActionStreamSpecifierParts but with
// optional probed stream info for stream-aware completion.
func ActionStreamSpecifierPartsWithStreams(specifierPart string, partialValue string, streams []probe.StreamInfo) carapace.Action {
	return carapace.ActionCallback(func(c carapace.Context) carapace.Action {
		specCtx := streamspec.ParseForCompletion(specifierPart)

		var actions []carapace.Action
		for _, token := range specCtx.ExpectedTokens {
			switch token {
			case streamspec.ExpectedSpecifierType, streamspec.ExpectedStreamTypeLetter:
				actions = append(actions, streamTypeActions(specCtx, ""))
			case streamspec.ExpectedStreamIndex:
				actions = append(actions, actionStreamIndex(specCtx, specifierPart, streams, ""))
			case streamspec.ExpectedMetadataKey:
				action := ffmpeg.ActionMetadataKeys()
				action = action.Invoke(carapace.Context{Value: partialValue}).ToA()
				actions = append(actions, action)
			case streamspec.ExpectedMetadataValue:
				actions = append(actions, actionMetadataValueParts(specCtx, streams, partialValue))
			case streamspec.ExpectedDispositionName:
				actions = append(actions, actionDispositionNameParts(specCtx, streams, partialValue))
			case streamspec.ExpectedGroupSpecifier,
				streamspec.ExpectedGroupIndex,
				streamspec.ExpectedGroupID:
				actions = append(actions, carapace.ActionValues().Suffix(":").NoSpace(':'))
			case streamspec.ExpectedProgramID,
				streamspec.ExpectedStreamIDValue:
				actions = append(actions, carapace.ActionValues().Suffix(":").NoSpace(':'))
			}
		}

		if len(actions) == 0 {
			return streamTypeActions(specCtx, "")
		}
		return carapace.Batch(actions...).ToA()
	})
}

// actionStreamIndex returns completions for stream indices, using probed stream info
// when available to list only actual stream indices for the resolved type.
func actionStreamIndex(specCtx *streamspec.CompletionContext, specifierPart string, streams []probe.StreamInfo, prefixToReplace string) carapace.Action {
	if len(streams) > 0 {
		var codecType string
		if specCtx.CurrentKind == streamspec.KindStreamType {
			codecType = streamTypeToCodecType(extractStreamTypeLetter(specifierPart))
		}
		// When CurrentKind is KindStreamIndex (bare numeric), codecType stays ""
		// which means all stream indices are returned.
		indices := probe.StreamIndices(streams, codecType)
		if len(indices) > 0 {
			action := carapace.ActionValues(indices...)
			if specCtx.PartialIdent != "" {
				action = action.Invoke(carapace.Context{Value: specCtx.PartialIdent}).Prefix(prefixToReplace).ToA()
			}
			return action.Suffix(":").NoSpace(':')
		}
	}
	return carapace.ActionValues()
}

// actionMetadataValue returns completions for metadata values, using probed stream info
// when available to list actual tag values from input streams.
func actionMetadataValue(specCtx *streamspec.CompletionContext, streams []probe.StreamInfo, prefixToReplace string) carapace.Action {
	if len(streams) > 0 && specCtx.MetadataKey != "" {
		vals := probe.MetadataValues(streams, specCtx.MetadataKey)
		if len(vals) > 0 {
			action := carapace.ActionValues(vals...)
			if specCtx.PartialIdent != "" {
				action = action.Invoke(carapace.Context{Value: specCtx.PartialIdent}).Prefix(prefixToReplace).ToA()
			}
			return action.Suffix(":").NoSpace(':')
		}
	}
	return carapace.ActionValues().Suffix(":").NoSpace(':')
}

// actionMetadataValueParts returns metadata value completions for the ActionMultiParts path.
func actionMetadataValueParts(specCtx *streamspec.CompletionContext, streams []probe.StreamInfo, partialValue string) carapace.Action {
	if len(streams) > 0 && specCtx.MetadataKey != "" {
		vals := probe.MetadataValues(streams, specCtx.MetadataKey)
		if len(vals) > 0 {
			action := carapace.ActionValues(vals...)
			action = action.Invoke(carapace.Context{Value: partialValue}).ToA()
			return action.Suffix(":").NoSpace(':')
		}
	}
	return carapace.ActionValues().Suffix(":").NoSpace(':')
}

// actionDispositionName returns disposition completions, preferring probed stream info
// when available to show only dispositions that are actually set in the input.
func actionDispositionName(specCtx *streamspec.CompletionContext, streams []probe.StreamInfo, prefixToReplace string) carapace.Action {
	if len(streams) > 0 {
		active := probe.ActiveDispositions(streams)
		if len(active) > 0 {
			action := carapace.ActionValues(active...)
			if specCtx.PartialIdent != "" {
				action = action.Invoke(carapace.Context{Value: specCtx.PartialIdent}).Prefix(prefixToReplace).ToA()
			}
			return action.Suffix(":").NoSpace(':')
		}
	}
	action := ffmpeg.ActionDispositions()
	action = action.Invoke(carapace.Context{Value: specCtx.PartialIdent}).Prefix(prefixToReplace).ToA()
	return action.Suffix(":").NoSpace(':')
}

// actionDispositionNameParts returns disposition completions for the ActionMultiParts path.
func actionDispositionNameParts(_ *streamspec.CompletionContext, streams []probe.StreamInfo, partialValue string) carapace.Action {
	if len(streams) > 0 {
		active := probe.ActiveDispositions(streams)
		if len(active) > 0 {
			action := carapace.ActionValues(active...)
			action = action.Invoke(carapace.Context{Value: partialValue}).ToA()
			return action.Suffix(":").NoSpace(':')
		}
	}
	action := ffmpeg.ActionDispositions()
	action = action.Invoke(carapace.Context{Value: partialValue}).ToA()
	return action.Suffix(":").NoSpace(':')
}

// ActionFilterValue returns completions for filter values using the filtergraph parser.
// isComplex indicates whether link labels are allowed (-filter_complex, -lavfi).
func ActionFilterValue(value string, isComplex bool) carapace.Action {
	return carapace.ActionCallback(func(c carapace.Context) carapace.Action {
		fgCtx := filtergraph.ParseForCompletion(value)
		fgCtx.IsComplex = isComplex

		// prefixToReplace is the part of the value before the partial ident
		// that needs to be prefixed back onto completion results.
		prefixToReplace := value
		if fgCtx.PartialIdent != "" {
			prefixToReplace = strings.TrimSuffix(value, fgCtx.PartialIdent)
		}

		var actions []carapace.Action
		for _, token := range fgCtx.ExpectedTokens {
			switch token {
			case filtergraph.ExpectedFilterName:
				action := ffmpeg.ActionFilters()
				// Invoke with PartialIdent (or empty) as the Value so carapace
				// filters filter names correctly, then prefix back the
				// preceding filtergraph text.
				action = action.Invoke(carapace.Context{Value: fgCtx.PartialIdent}).Prefix(prefixToReplace).ToA()
				actions = append(actions, action.NoSpace())
			case filtergraph.ExpectedFilterOptionKey:
				if fgCtx.Filter != nil && fgCtx.Filter.Name != "" {
					action := ffmpeg.ActionFilterOptions(fgCtx.Filter.Name, fgCtx.Filter.OptionKeys)
					action = action.Invoke(carapace.Context{Value: fgCtx.PartialIdent}).Prefix(prefixToReplace).ToA()
					actions = append(actions, action.NoSpace())
				}
			case filtergraph.ExpectedFilterOptionValue:
				if fgCtx.Filter != nil && fgCtx.Filter.Name != "" && fgCtx.Filter.OptionKey != "" {
					action := ffmpeg.ActionFilterOptionValue(fgCtx.Filter.Name, fgCtx.Filter.OptionKey)
					action = action.Invoke(carapace.Context{Value: fgCtx.PartialIdent}).Prefix(prefixToReplace).ToA()
					actions = append(actions, action.NoSpace())
				}
			case filtergraph.ExpectedFilterOption:
				// Ambiguous position — could be key or positional value.
				// Offer key=value completions with Suffix("=") for known filters.
				if fgCtx.Filter != nil && fgCtx.Filter.Name != "" {
					action := ffmpeg.ActionFilterOptions(fgCtx.Filter.Name, fgCtx.Filter.OptionKeys)
					action = action.Invoke(carapace.Context{Value: fgCtx.PartialIdent}).Prefix(prefixToReplace).ToA()
					actions = append(actions, action.NoSpace('='))
				}
			case filtergraph.ExpectedLinkLabel:
				if fgCtx.IsComplex {
					actions = append(actions, carapace.ActionValues().NoSpace())
				}
			case filtergraph.ExpectedChainSeparator:
				actions = append(actions, carapace.ActionValues(";").NoSpace())
			}
		}

		if len(actions) == 0 {
			return ffmpeg.ActionFilters().NoSpace()
		}
		return carapace.Batch(actions...).ToA()
	})
}

// ActionDecoderOnlyCodec returns codec completions restricted to decoders.
// Used by ffplay and ffprobe which only decode (no encoding).
func ActionDecoderOnlyCodec(ctx *argstream.CompletionContext) carapace.Action {
	audio := true
	subtitle := true
	video := true
	if ctx.CurrentOption != nil {
		spec := ctx.CurrentOption.StreamSpecifier
		if spec != "" {
			if strings.HasPrefix(spec, "a") {
				audio = true
				subtitle = false
				video = false
			} else if strings.HasPrefix(spec, "v") || strings.HasPrefix(spec, "V") {
				audio = false
				subtitle = false
				video = true
			} else if strings.HasPrefix(spec, "s") {
				audio = false
				subtitle = true
				video = false
			} else if strings.HasPrefix(spec, "d") {
				audio = false
				subtitle = false
				video = false
			}
		}
	}
	return carapace.Batch(
		ffmpeg.ActionDecodableCodecs(ffmpeg.CodecOpts{Audio: audio, Subtitle: subtitle, Video: video}),
		ffmpeg.ActionDecoders(ffmpeg.DecoderOpts{Audio: audio, Subtitle: subtitle, Video: video}),
	).ToA()
}

func containsToken(tokens []argstream.ExpectedToken, t argstream.ExpectedToken) bool {
	return slices.Contains(tokens, t)
}
