package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/carapace-sh/carapace"
	"github.com/carapace-sh/carapace-ffmpeg/pkg/argstream"
	"github.com/carapace-sh/carapace-ffmpeg/pkg/completer"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:                "ffprobe",
	Short:              "FFprobe multimedia stream analyzer",
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
	profile := argstream.DefaultFFprobeProfile

	carapace.Gen(rootCmd).Standalone()

	carapace.Gen(rootCmd).PositionalAnyCompletion(
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
