package cmd

import (
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
)

// go install stamps the module version into the build info.
// Builds from a local checkout report (devel).
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of gm",
	Run: func(cmd *cobra.Command, args []string) {
		version := "(devel)"
		if info, ok := debug.ReadBuildInfo(); ok {
			version = info.Main.Version
		}
		fmt.Println("gm", version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
