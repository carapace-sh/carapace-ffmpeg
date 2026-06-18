package cmd

import (
	"fmt"
	"os"

	"github.com/carapace-sh/carapace"
	spec "github.com/carapace-sh/carapace-spec"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "carapace-ffmpeg-debug",
	Short: "Parse ffmpeg stream specifiers, filtergraphs, and argument streams",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	carapace.Gen(rootCmd)
	spec.Register(rootCmd)
}
