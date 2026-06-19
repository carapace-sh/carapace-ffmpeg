package cmd

import (
	"fmt"
	"os"

	"github.com/carapace-sh/carapace"
	"github.com/carapace-sh/carapace-ffmpeg/pkg/argstream"
	"github.com/carapace-sh/carapace-ffmpeg/pkg/completer"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:                "ffplay",
	Short:              "FFplay media player",
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
	profile := argstream.DefaultFFplayProfile

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
					actions = append(actions, completer.ActionOptionValue(ctx, completer.ActionDecoderOnlyCodec))
				case argstream.ExpectedStreamSpecifier:
					actions = append(actions, completer.ActionStreamSpecifier(ctx, c))
				case argstream.ExpectedFilterValue:
					actions = append(actions, completer.ActionFilterValue())
				}
			}

			if len(actions) == 0 {
				return carapace.ActionValues()
			}
			return carapace.Batch(actions...).ToA()
		}),
	)
}
