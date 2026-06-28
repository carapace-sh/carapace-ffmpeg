package cmd

import (
	"slices"
	"strings"

	"github.com/carapace-sh/carapace"
	ffmpeg "github.com/carapace-sh/carapace-ffmpeg/pkg/actions/tools/ffmpeg"
	"github.com/carapace-sh/carapace-ffmpeg/pkg/argstream"
	"github.com/carapace-sh/carapace-ffmpeg/pkg/completer"
	spec "github.com/carapace-sh/carapace-spec"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:               "carapace-ffmpeg",
	Short:             "FFmpeg completion provider",
	CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
}

var gen *carapace.Carapace

func Execute() {
	gen.Execute()
}

func init() {
	rootCmd.AddCommand(
		ffmpegCmd,
		ffplayCmd,
		ffprobeCmd,
		debugCmd,
	)

	gen = carapace.Gen(rootCmd,
		carapace.WithSubcommands(ffmpegCmd, ffplayCmd, ffprobeCmd),
		carapace.WithDefault("ffmpeg"),
	)

	gen.Standalone()

	gen.PositionalAnyCompletion(
		carapace.ActionValues("ffmpeg", "ffplay", "ffprobe", "debug"),
	)
}

var ffmpegCmd = &cobra.Command{
	Use:                "ffmpeg",
	Short:              "Hyper fast Audio and Video encoder",
	Run:                func(cmd *cobra.Command, args []string) {},
	DisableFlagParsing: true,
}

var ffplayCmd = &cobra.Command{
	Use:                "ffplay",
	Short:              "FFplay media player",
	Run:                func(cmd *cobra.Command, args []string) {},
	DisableFlagParsing: true,
}

var ffprobeCmd = &cobra.Command{
	Use:                "ffprobe",
	Short:              "FFprobe multimedia stream analyzer",
	Run:                func(cmd *cobra.Command, args []string) {},
	DisableFlagParsing: true,
}

var debugCmd = &cobra.Command{
	Use:   "debug",
	Short: "Parse ffmpeg stream specifiers, filtergraphs, and argument streams",
}

func init() {
	initFFmpeg()
	initFFplay()
	initFFprobe()
	initDebug()
}

func initFFmpeg() {
	profile := argstream.DefaultFFmpegProfile

	carapace.Gen(ffmpegCmd).Standalone()

	carapace.Gen(ffmpegCmd).PositionalAnyCompletion(
		carapace.ActionCallback(func(c carapace.Context) carapace.Action {
			if completer.IsMidTokenOptionWithSpec(c.Value, profile) {
				args, _ := completer.ContextToArgs(c)
				ctx := argstream.ParseForCompletionWithProfile(args, false, profile)
				streams := completer.ProbeAll(ctx)
				originalValue := c.Value
				return carapace.ActionMultiParts(":", func(c carapace.Context) carapace.Action {
					switch len(c.Parts) {
					case 0:
						return completer.ActionOptionNames(ctx, profile).NoSpace(':')
					default:
						specifierPart := ""
						if _, after, ok := strings.Cut(originalValue, ":"); ok {
							specifierPart = after
						}
						return completer.ActionStreamSpecifierPartsWithStreams(specifierPart, c.Value, streams)
					}
				})
			}

			args, trailingSpace := completer.ContextToArgs(c)
			ctx := argstream.ParseForCompletionWithProfile(args, trailingSpace, profile)
			streams := completer.ProbeAll(ctx)

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
					actions = append(actions, completer.ActionStreamSpecifierWithStreams(ctx, c, streams))
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

func initFFplay() {
	profile := argstream.DefaultFFplayProfile

	carapace.Gen(ffplayCmd).Standalone()

	carapace.Gen(ffplayCmd).PositionalAnyCompletion(
		carapace.ActionCallback(func(c carapace.Context) carapace.Action {
			if completer.IsMidTokenOptionWithSpec(c.Value, profile) {
				args, _ := completer.ContextToArgs(c)
				ctx := argstream.ParseForCompletionWithProfile(args, false, profile)
				streams := completer.ProbeAll(ctx)
				originalValue := c.Value
				return carapace.ActionMultiParts(":", func(c carapace.Context) carapace.Action {
					switch len(c.Parts) {
					case 0:
						return completer.ActionOptionNames(ctx, profile).NoSpace(':')
					default:
						specifierPart := ""
						if _, after, ok := strings.Cut(originalValue, ":"); ok {
							specifierPart = after
						}
						return completer.ActionStreamSpecifierPartsWithStreams(specifierPart, c.Value, streams)
					}
				})
			}

			args, trailingSpace := completer.ContextToArgs(c)
			ctx := argstream.ParseForCompletionWithProfile(args, trailingSpace, profile)
			streams := completer.ProbeAll(ctx)

			if ctx.PartialOption != "" && !trailingSpace {
				return completer.ActionPartialOption(ctx, profile)
			}

			var actions []carapace.Action
			for _, token := range ctx.ExpectedTokens {
				switch token {
				case argstream.ExpectedGlobalOption, argstream.ExpectedInputOption:
					actions = append(actions, completer.ActionOptions(ctx, profile))
				case argstream.ExpectedInputURL:
					actions = append(actions, carapace.ActionFiles())
				case argstream.ExpectedOptionValue:
					actions = append(actions, completer.ActionOptionValue(ctx, completer.ActionDecoderOnlyCodec, c.Value))
				case argstream.ExpectedStreamSpecifier:
					actions = append(actions, completer.ActionStreamSpecifierWithStreams(ctx, c, streams))
				case argstream.ExpectedFilterValue:
					isComplex := ctx.CurrentOption != nil &&
						(ctx.CurrentOption.CanonicalName == "filter_complex" || ctx.CurrentOption.CanonicalName == "lavfi")
					actions = append(actions, completer.ActionFilterValue(c.Value, isComplex, completer.FilterOptsFromContext(ctx)))
				}
			}

			if len(actions) == 0 {
				return carapace.ActionValues()
			}
			return carapace.Batch(actions...).ToA()
		}),
	)
}

func initFFprobe() {
	profile := argstream.DefaultFFprobeProfile

	carapace.Gen(ffprobeCmd).Standalone()

	carapace.Gen(ffprobeCmd).PositionalAnyCompletion(
		carapace.ActionCallback(func(c carapace.Context) carapace.Action {
			if completer.IsMidTokenOptionWithSpec(c.Value, profile) {
				args, _ := completer.ContextToArgs(c)
				ctx := argstream.ParseForCompletionWithProfile(args, false, profile)
				streams := completer.ProbeAll(ctx)
				originalValue := c.Value
				return carapace.ActionMultiParts(":", func(c carapace.Context) carapace.Action {
					switch len(c.Parts) {
					case 0:
						return completer.ActionOptionNames(ctx, profile).NoSpace(':')
					default:
						specifierPart := ""
						if _, after, ok := strings.Cut(originalValue, ":"); ok {
							specifierPart = after
						}
						return completer.ActionStreamSpecifierPartsWithStreams(specifierPart, c.Value, streams)
					}
				})
			}

			args, trailingSpace := completer.ContextToArgs(c)
			ctx := argstream.ParseForCompletionWithProfile(args, trailingSpace, profile)
			streams := completer.ProbeAll(ctx)

			if ctx.PartialOption != "" && !trailingSpace {
				return completer.ActionPartialOption(ctx, profile)
			}

			var actions []carapace.Action
			for _, token := range ctx.ExpectedTokens {
				switch token {
				case argstream.ExpectedGlobalOption, argstream.ExpectedInputOption:
					actions = append(actions, completer.ActionOptions(ctx, profile))
				case argstream.ExpectedInputURL:
					actions = append(actions, carapace.ActionFiles())
				case argstream.ExpectedOptionValue:
					actions = append(actions, completer.ActionOptionValue(ctx, completer.ActionDecoderOnlyCodec, c.Value))
				case argstream.ExpectedStreamSpecifier:
					actions = append(actions, completer.ActionStreamSpecifierWithStreams(ctx, c, streams))
				}
			}

			if len(actions) == 0 {
				return carapace.ActionValues()
			}
			return carapace.Batch(actions...).ToA()
		}),
	)
}

func initDebug() {
	carapace.Gen(debugCmd)
	spec.Register(debugCmd)
}

func actionOptionValueIfExpected(ctx *argstream.CompletionContext, c carapace.Context) carapace.Action {
	if slices.Contains(ctx.ExpectedTokens, argstream.ExpectedOptionValue) {
		return actionOptionValue(ctx, c)
	}
	return carapace.ActionValues()
}

func actionOptionValue(ctx *argstream.CompletionContext, c carapace.Context) carapace.Action {
	return completer.ActionOptionValue(ctx, actionCodec, c.Value)
}

func actionFilterValue(ctx *argstream.CompletionContext, c carapace.Context) carapace.Action {
	isComplex := ctx.CurrentOption != nil &&
		(ctx.CurrentOption.CanonicalName == "filter_complex" || ctx.CurrentOption.CanonicalName == "lavfi")
	return completer.ActionFilterValue(c.Value, isComplex, completer.FilterOptsFromContext(ctx))
}

func actionCodec(ctx *argstream.CompletionContext) carapace.Action {
	audio := true
	subtitle := true
	video := true
	if ctx.CurrentOption != nil {
		spec := ctx.CurrentOption.StreamSpecifier
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
		if optDef := argstream.LookupOption(ctx.CurrentOption.Name); optDef != nil && optDef.ImplicitSpec != "" {
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
