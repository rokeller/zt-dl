package cmd

import (
	"context"
	"os"

	"github.com/spf13/cobra"
)

var version string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "zt-dl",
	Short: "Tool to help downloading recordings from Zattoo",

	SilenceErrors: false,
	Version:       version,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(ctx context.Context) {
	err := rootCmd.ExecuteContext(ctx)
	if nil != err {
		os.Exit(1)
	}
}

func init() {
	addEmailAndDomainFlags(rootCmd)
}
