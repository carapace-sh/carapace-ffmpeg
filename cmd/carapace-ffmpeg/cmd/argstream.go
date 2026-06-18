package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/carapace-sh/carapace"
	"github.com/carapace-sh/carapace-ffmpeg/pkg/argstream"
	"github.com/spf13/cobra"
)

var argstreamCmd = &cobra.Command{
	Use:   "argstream <args...>",
	Short: "Parse ffmpeg argument stream",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		prog, err := argstream.Parse(args)
		if err != nil {
			return err
		}
		m, err := json.Marshal(prog)
		if err != nil {
			return err
		}
		fmt.Println(string(m))
		return nil
	},
	DisableFlagParsing: true,
}

var argstreamCompleteCmd = &cobra.Command{
	Use:   "argstream-complete <args...>",
	Short: "Get completion context for ffmpeg argument stream",
	Args:  cobra.MinimumNArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := argstream.ParseForCompletion(args)
		m, err := json.Marshal(ctx)
		if err != nil {
			return err
		}
		fmt.Println(string(m))
		return nil
	},
	DisableFlagParsing: true,
}

func init() {
	rootCmd.AddCommand(argstreamCmd)
	rootCmd.AddCommand(argstreamCompleteCmd)

	carapace.Gen(argstreamCmd)
	carapace.Gen(argstreamCompleteCmd)
}
