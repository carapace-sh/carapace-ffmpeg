package cmd

import (
	"fmt"
	"os"

	"github.com/carapace-sh/carapace"
	ffmpeg "github.com/carapace-sh/carapace-ffmpeg/pkg/actions/tools/ffmpeg"
	"github.com/carapace-sh/carapace-ffmpeg/pkg/argstream"
	"github.com/carapace-sh/carapace-ffmpeg/pkg/completer"
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
	profile := argstream.DefaultFFmpegProfile

	carapace.Gen(rootCmd).Standalone()

	carapace.Gen(rootCmd).PositionalAnyCompletion(
		carapace.ActionCallback(func(c carapace.Context) carapace.Action {
			if completer.IsMidTokenOptionWithSpec(c.Value, profile) {
				args, _ := completer.ContextToArgs(c)
				ctx := argstream.ParseForCompletionWithProfile(args, false, profile)
				return carapace.ActionMultiParts(":", func(c carapace.Context) carapace.Action {
					switch len(c.Parts) {
					case 0:
						return completer.ActionOptionNames(ctx, profile).NoSpace(':')
					default:
						return completer.ActionStreamSpecifiers()
					}
				})
			}

			args, trailingSpace := completer.ContextToArgs(c)
			ctx := argstream.ParseForCompletionWithProfile(args, trailingSpace, profile)

			// When completing a partial option name (e.g. `-v` matching `-vcodec`,
			// `-vframes`, `-vn`), include option names so the shell can filter them.
			if ctx.PartialOption != "" && !trailingSpace {
				return carapace.Batch(
					completer.ActionPartialOption(ctx, profile),
					actionOptionValueIfExpected(ctx, c),
				).ToA()
			}

			var actions []carapace.Action
			for _, token := range ctx.ExpectedTokens {
				switch token {
				case argstream.ExpectedGlobalOption, argstream.ExpectedInputOption, argstream.ExpectedOutputOption:
					actions = append(actions, completer.ActionOptions(ctx, profile))
				case argstream.ExpectedInputURL:
					actions = append(actions, carapace.ActionFiles())
				case argstream.ExpectedOutputURL:
					actions = append(actions, carapace.ActionFiles())
				case argstream.ExpectedOptionValue:
					actions = append(actions, actionOptionValue(ctx, c))
				case argstream.ExpectedStreamSpecifier:
					actions = append(actions, completer.ActionStreamSpecifier(ctx, c))
				case argstream.ExpectedFilterValue:
					actions = append(actions, actionFilterValue(ctx, c))
				case argstream.ExpectedMapValue:
					actions = append(actions, carapace.ActionValues())
				}
			}

			if len(actions) == 0 {
				return carapace.ActionValues()
			}
			return carapace.Batch(actions...).ToA()
		}),
	)
}

// actionOptionValueIfExpected returns option value completions only when
// the completion context expects an option value (e.g. when a partial option
// like `-v` resolved to a known value option like `-vloglevel`).
func actionOptionValueIfExpected(ctx *argstream.CompletionContext, c carapace.Context) carapace.Action {
	for _, token := range ctx.ExpectedTokens {
		if token == argstream.ExpectedOptionValue {
			return actionOptionValue(ctx, c)
		}
	}
	return carapace.ActionValues()
}

// actionOptionValue returns completions for option values, with ffmpeg-specific
// codec handling (encoder vs. decoder depending on scope).
func actionOptionValue(ctx *argstream.CompletionContext, c carapace.Context) carapace.Action {
	return completer.ActionOptionValue(ctx, actionCodec, c.Value)
}

func actionFilterValue(ctx *argstream.CompletionContext, c carapace.Context) carapace.Action {
	isComplex := ctx.CurrentOption != nil &&
		(ctx.CurrentOption.CanonicalName == "filter_complex" || ctx.CurrentOption.CanonicalName == "lavfi")
	return completer.ActionFilterValue(c.Value, isComplex)
}

// actionCodec returns codec completions scoped to the current position.
// In global/input scope: codecs + decoders (decoding context).
// In output scope: codecs + encoders (encoding context).
func actionCodec(ctx *argstream.CompletionContext) carapace.Action {
	audio := true
	subtitle := true
	video := true
	if ctx.CurrentOption != nil {
		spec := ctx.CurrentOption.StreamSpecifier
		if spec != "" {
			if len(spec) > 0 && spec[0] == 'a' {
				audio = true
				subtitle = false
				video = false
			} else if len(spec) > 0 && (spec[0] == 'v' || spec[0] == 'V') {
				audio = false
				subtitle = false
				video = true
			} else if len(spec) > 0 && spec[0] == 's' {
				audio = false
				subtitle = true
				video = false
			} else if len(spec) > 0 && spec[0] == 'd' {
				audio = false
				subtitle = false
				video = false
			}
		} else if optDef := argstream.LookupOption(ctx.CurrentOption.Name); optDef != nil && optDef.ImplicitSpec != "" {
			audio = optDef.ImplicitSpec == "a"
			subtitle = optDef.ImplicitSpec == "s"
			video = optDef.ImplicitSpec == "v"
		}
	}

	switch ctx.Scope {
	case argstream.ScopeGlobal, argstream.ScopeInputFile:
		return carapace.Batch(
			ffmpeg.ActionDecodableCodecs(ffmpeg.CodecOpts{Audio: audio, Subtitle: subtitle, Video: video}),
			ffmpeg.ActionDecoders(ffmpeg.DecoderOpts{Audio: audio, Subtitle: subtitle, Video: video}),
		).ToA()
	default:
		return carapace.Batch(
			ffmpeg.ActionEncodableCodecs(ffmpeg.CodecOpts{Audio: audio, Subtitle: subtitle, Video: video}),
			ffmpeg.ActionEncoders(ffmpeg.EncoderOpts{Audio: audio, Subtitle: subtitle, Video: video}),
		).ToA()
	}
}
