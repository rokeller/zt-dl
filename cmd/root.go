package cmd

import (
	"os"

	"github.com/rokeller/zt-dl/zattoo"
	"github.com/spf13/cobra"
)

var version string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use: "zt-dl",

	SilenceErrors: false,

	RunE:    runRootCmd,
	Version: version,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringP("email", "e", "", "Email address of your Zattoo account.")
	rootCmd.MarkPersistentFlagFilename("email")
}

func runRootCmd(cmd *cobra.Command, args []string) error {
	acct := zattoo.NewAccount(cmd.Flag("email").Value.String())
	if err := acct.Login(); nil != err {
		return err
	}
	if err := acct.GetAllRecordings(); nil != err {
		return err
	}

	return nil
}
