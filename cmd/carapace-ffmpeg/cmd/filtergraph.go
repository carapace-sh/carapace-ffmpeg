package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/carapace-sh/carapace"
	"github.com/carapace-sh/carapace-ffmpeg/pkg/filtergraph"
	"github.com/spf13/cobra"
)

var filtergraphCmd = &cobra.Command{
	Use:   "filtergraph <expression>",
	Short: "Parse a filtergraph expression",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fg, err := filtergraph.Parse(args[0])
		if err != nil {
			return err
		}
		m, err := json.Marshal(fg)
		if err != nil {
			return err
		}
		fmt.Println(string(m))
		return nil
	},
}

var filtergraphCompleteCmd = &cobra.Command{
	Use:   "filtergraph-complete <expression>",
	Short: "Get completion context for a filtergraph expression",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := filtergraph.ParseForCompletion(args[0])
		m, err := json.Marshal(ctx)
		if err != nil {
			return err
		}
		fmt.Println(string(m))
		return nil
	},
}

func init() {
	debugCmd.AddCommand(filtergraphCmd)
	debugCmd.AddCommand(filtergraphCompleteCmd)

	carapace.Gen(filtergraphCmd)
	carapace.Gen(filtergraphCompleteCmd)
}
