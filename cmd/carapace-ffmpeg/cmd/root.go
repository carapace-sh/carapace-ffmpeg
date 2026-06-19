package cmd

import (
	"fmt"
	"os"
	"slices"

	"github.com/carapace-sh/carapace"
	"github.com/carapace-sh/carapace-ffmpeg/pkg/argstream"
	ffmpeg "github.com/carapace-sh/carapace-ffmpeg/pkg/actions/tools/ffmpeg"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:                "ffmpeg",
	Short: "Hyper fast Audio and Video encoder",
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
			trailingSpace := c.Value == "" || (len(c.Value) > 0 && c.Value[0] == '-')
			ctx := argstream.ParseForCompletion(c.Args, trailingSpace)

			var actions []carapace.Action
			for _, expected := range ctx.ExpectedTokens {
				switch expected {
				case argstream.ExpectedGlobalOption,
					argstream.ExpectedInputOption,
					argstream.ExpectedOutputOption:
					actions = append(actions, actionOptions(ctx))
				case argstream.ExpectedInputURL:
					actions = append(actions, carapace.ActionFiles())
				case argstream.ExpectedOutputURL:
					actions = append(actions, carapace.ActionFiles())
				case argstream.ExpectedOptionValue:
					actions = append(actions, actionOptionValue(ctx))
				case argstream.ExpectedStreamSpecifier:
					actions = append(actions, actionStreamSpecifier(ctx))
				case argstream.ExpectedFilterValue:
					actions = append(actions, actionFilterValue())
				case argstream.ExpectedMapValue:
					actions = append(actions, actionMapValue(ctx))
				}
			}

			if len(actions) == 0 {
				return carapace.ActionFiles()
			}
			return carapace.Batch(actions...).ToA()
		}),
	)
}

func actionOptions(ctx *argstream.CompletionContext) carapace.Action {
	return carapace.ActionMultiPartsN(":", 2, func(c carapace.Context) carapace.Action {
		switch len(c.Parts) {
		case 0:
			return actionOptionNames(ctx).NoSpace(':')
		default:
			return actionStreamSpecifiers(ctx)
		}
	})
}

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
	default:
		return carapace.ActionValues()
	}
}

func actionCodec(ctx *argstream.CompletionContext) carapace.Action {
	switch ctx.Scope {
	case argstream.ScopeGlobal, argstream.ScopeInputFile:
		return carapace.Batch(
			ffmpeg.ActionCodecs(),
			ffmpeg.ActionDecoders(),
		).ToA()
	default:
		return carapace.Batch(
			ffmpeg.ActionCodecs(),
			ffmpeg.ActionEncoders(),
		).ToA()
	}
}

func actionStreamSpecifier(ctx *argstream.CompletionContext) carapace.Action {
	if ctx.CurrentOption == nil {
		return carapace.ActionValues()
	}
	return actionStreamSpecifiers(ctx)
}

func actionStreamSpecifiers(_ *argstream.CompletionContext) carapace.Action {
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
	).NoSpace(':')
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
