package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/carapace-sh/carapace"
	"github.com/carapace-sh/carapace-ffmpeg/pkg/streamspec"
	"github.com/spf13/cobra"
)

var streamspecCmd = &cobra.Command{
	Use:   "streamspec <specifier>",
	Short: "Parse a stream specifier",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		spec, err := streamspec.Parse(args[0])
		if err != nil {
			return err
		}
		m, err := json.Marshal(spec)
		if err != nil {
			return err
		}
		fmt.Println(string(m))
		return nil
	},
}

var streamspecCompleteCmd = &cobra.Command{
	Use:   "streamspec-complete <specifier>",
	Short: "Get completion context for a stream specifier",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := streamspec.ParseForCompletion(args[0])
		m, err := json.Marshal(ctx)
		if err != nil {
			return err
		}
		fmt.Println(string(m))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(streamspecCmd)
	rootCmd.AddCommand(streamspecCompleteCmd)

	carapace.Gen(streamspecCmd)
	carapace.Gen(streamspecCompleteCmd)
}
