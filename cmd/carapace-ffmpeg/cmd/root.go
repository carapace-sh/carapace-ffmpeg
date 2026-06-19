package cmd

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/carapace-sh/carapace"
	"github.com/carapace-sh/carapace-ffmpeg/pkg/argstream"
	ffmpeg "github.com/carapace-sh/carapace-ffmpeg/pkg/actions/tools/ffmpeg"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:                "ffmpeg",
	Short:              "Hyper fast Audio and Video encoder",
	Run:                func(cmd *cobra.Command, args []string) {},
	DisableFlagParsing: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	carapace.Gen(rootCmd).Standalone()

	carapace.Gen(rootCmd).PositionalAnyCompletion(
		carapace.ActionCallback(func(c carapace.Context) carapace.Action {
			// When the current token is a mid-token option with a colon
			// (e.g. "-c:" or "-c:v"), use ActionMultiParts to handle
			// the colon-separated completion within the single token.
			// This takes priority over the argstream dispatch since the
			// user is actively typing an option+specifier as one word.
			if isMidTokenOptionWithSpec(c.Value) {
				args, _ := contextToArgs(c)
				ctx := argstream.ParseForCompletion(args, false)
				return carapace.ActionMultiParts(":", func(c carapace.Context) carapace.Action {
					switch len(c.Parts) {
					case 0:
						return actionOptionNames(ctx).NoSpace(':')
					default:
						return actionStreamSpecifiers()
					}
				})
			}

			args, trailingSpace := contextToArgs(c)
			ctx := argstream.ParseForCompletion(args, trailingSpace)

			var actions []carapace.Action
			for _, token := range ctx.ExpectedTokens {
				switch token {
				case argstream.ExpectedGlobalOption, argstream.ExpectedInputOption, argstream.ExpectedOutputOption:
					actions = append(actions, actionOptions(ctx, c))
				case argstream.ExpectedInputURL:
					actions = append(actions, carapace.ActionFiles())
				case argstream.ExpectedOutputURL:
					actions = append(actions, carapace.ActionFiles())
				case argstream.ExpectedOptionValue:
					actions = append(actions, actionOptionValue(ctx))
				case argstream.ExpectedStreamSpecifier:
					actions = append(actions, actionStreamSpecifier(ctx, c))
				case argstream.ExpectedFilterValue:
					actions = append(actions, actionFilterValue())
				case argstream.ExpectedMapValue:
					actions = append(actions, actionMapValue(ctx))
				}
			}

			if len(actions) == 0 {
				return carapace.ActionValues()
			}
			return carapace.Batch(actions...).ToA()
		}),
	)
}

// isMidTokenOptionWithSpec returns true when the current token is an option
// that contains a colon AND the option accepts stream specifiers.
// This detects mid-token option+specifier like "-c:" or "-c:v" where
// the user is typing both the option name and specifier as one word.
func isMidTokenOptionWithSpec(value string) bool {
	if !strings.HasPrefix(value, "-") || !strings.Contains(value, ":") {
		return false
	}
	optText := strings.TrimPrefix(value[1:], "-")
	baseName, _, _ := argstream.ParseOptionName(optText)
	optDef := argstream.LookupOption(baseName)
	return optDef != nil && optDef.AcceptsSpec && optDef.ImplicitSpec == ""
}

// contextToArgs converts carapace.Context to the args and trailingSpace
// expected by argstream.ParseForCompletion.
//
// c.Args contains the positional arguments up to (but not including) the
// current token being completed. c.Value contains the current token.
//
// Some shell protocols include a trailing empty string in c.Args to mark
// the word-break position. We strip it. When c.Value is non-empty, it is
// the current token being completed and must be appended to args so the
// argstream parser can see it. trailingSpace is true when the cursor is
// at a new blank position after the last token (c.Value == "").
func contextToArgs(c carapace.Context) (args []string, trailingSpace bool) {
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

// actionOptions returns completions for ffmpeg option names.
// When called inside ActionMultiParts (for mid-token colon options),
// the plain option names are returned without suffix.
// Otherwise, options that accept stream specifiers get Suffix(":")
// so the user can continue typing the specifier.
func actionOptions(ctx *argstream.CompletionContext, _ carapace.Context) carapace.Action {
	return actionOptionNamesWithSpecSuffix(ctx)
}

// actionOptionNamesWithSpecSuffix returns option name completions.
// Options that accept stream specifiers get Suffix(":") so the user
// can continue typing the specifier within the same token.
func actionOptionNamesWithSpecSuffix(ctx *argstream.CompletionContext) carapace.Action {
	var specOptions, noSpecOptions []string
	for name, def := range argstream.OptionIndex {
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

// actionOptionNames returns plain option name completions without
// any suffix or NoSpace (used inside ActionMultiParts where the
// colon handling is already managed).
func actionOptionNames(ctx *argstream.CompletionContext) carapace.Action {
	var vals []string
	for name, def := range argstream.OptionIndex {
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

func actionOptionValue(ctx *argstream.CompletionContext) carapace.Action {
	if ctx.CurrentOption == nil {
		return carapace.ActionValues()
	}
	switch ctx.CurrentOption.ValueType {
	case argstream.ValueCodec:
		return actionCodec(ctx)
	case argstream.ValueFormat:
		return ffmpeg.ActionFormats()
	case argstream.ValuePixelFormat:
		return ffmpeg.ActionPixelFormats()
	case argstream.ValueSampleFmt:
		return ffmpeg.ActionSampleFormats()
	case argstream.ValueChannelLayout:
		return ffmpeg.ActionChannelLayouts()
	case argstream.ValueFilter:
		return actionFilterValue()
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
		return actionMapValue(ctx)
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
		return ffmpeg.ActionDevices()
	case argstream.ValueString:
		if ctx.CurrentOption.CanonicalName == "h" {
			return ffmpeg.ActionHelpTopics()
		}
		return carapace.ActionValues()
	default:
		return carapace.ActionValues()
	}
}

// actionStreamSpecifier handles stream specifier completion.
// When the cursor is mid-token inside an option with a colon
// (handled by ActionMultiParts in actionOptions), this is not reached.
// This is called when the argstream parser reports ExpectedStreamSpecifier
// for the separate-arg form (e.g. "-c:" followed by a new arg after
// shell splits the colon into a separate word).
func actionStreamSpecifier(ctx *argstream.CompletionContext, c carapace.Context) carapace.Action {
	if ctx.CurrentOption == nil || !ctx.CurrentOption.AcceptsSpec {
		return carapace.ActionValues()
	}

	if colon, after, ok := strings.Cut(c.Value, ":"); ok {
		return actionStreamSpecifiers().Invoke(
			carapace.Context{Value: after},
		).Prefix(colon + ":").ToA()
	}
	return actionStreamSpecifiers()
}

func actionStreamSpecifiers() carapace.Action {
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
	).NoSpace(':').Uid("ffmpeg", "stream-specifier")
}

func actionCodec(ctx *argstream.CompletionContext) carapace.Action {
	switch ctx.Scope {
	case argstream.ScopeGlobal, argstream.ScopeInputFile:
		return carapace.Batch(
			ffmpeg.ActionCodecs(ffmpeg.CodecOpts{}.Default()),
			ffmpeg.ActionDecoders(ffmpeg.DecoderOpts{}.Default()),
		).ToA()
	default:
		return carapace.Batch(
			ffmpeg.ActionCodecs(ffmpeg.CodecOpts{}.Default()),
			ffmpeg.ActionEncoders(ffmpeg.EncoderOpts{}.Default()),
		).ToA()
	}
}

func actionFilterValue() carapace.Action {
	return ffmpeg.ActionFilters().NoSpace()
}

func actionMapValue(_ *argstream.CompletionContext) carapace.Action {
	return carapace.ActionCallback(func(c carapace.Context) carapace.Action {
		return carapace.ActionValues()
	})
}

func containsToken(tokens []argstream.ExpectedToken, t argstream.ExpectedToken) bool {
	return slices.Contains(tokens, t)
}
