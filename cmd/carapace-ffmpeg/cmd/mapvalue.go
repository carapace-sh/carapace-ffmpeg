package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/carapace-sh/carapace"
	"github.com/carapace-sh/carapace-ffmpeg/pkg/mapvalue"
	"github.com/spf13/cobra"
)

var mapvalueCmd = &cobra.Command{
	Use:   "mapvalue <value>",
	Short: "Parse a -map value",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mv, err := mapvalue.Parse(args[0])
		if err != nil {
			return err
		}
		m, err := json.Marshal(mv)
		if err != nil {
			return err
		}
		fmt.Println(string(m))
		return nil
	},
}

var mapvalueCompleteCmd = &cobra.Command{
	Use:   "mapvalue-complete <value>",
	Short: "Get completion context for a -map value",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := mapvalue.ParseForCompletion(args[0])
		m, err := json.Marshal(ctx)
		if err != nil {
			return err
		}
		fmt.Println(string(m))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(mapvalueCmd)
	rootCmd.AddCommand(mapvalueCompleteCmd)

	carapace.Gen(mapvalueCmd)
	carapace.Gen(mapvalueCompleteCmd)
}
