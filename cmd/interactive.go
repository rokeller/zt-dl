package cmd

import (
	"fmt"
	"os"

	"github.com/rokeller/zt-dl/server"
	"github.com/rokeller/zt-dl/zattoo"
	"github.com/spf13/cobra"
)

// interactiveCmd represents the interactive command
var interactiveCmd = &cobra.Command{
	Use: "interactive",
	// 	Short: "A brief description of your command",
	// 	Long: `A longer description that spans multiple lines and likely contains examples
	// and usage of using your command. For example:

	// Cobra is a CLI library for Go that empowers applications.
	// This application is a tool to generate the needed files
	// to quickly create a Cobra application.`,
	RunE: runInteractiveCmd,
}

func runInteractiveCmd(cmd *cobra.Command, args []string) error {
	email := cmd.Flag(string(Email)).Value.String()
	domain := cmd.Flag(string(Domain)).Value.String()
	outdir := cmd.Flag("outdir").Value.String()
	acct := zattoo.NewAccount(email, domain)
	if err := acct.Login(); nil != err {
		return err
	}

	return server.Serve(cmd.Context(), acct, outdir, 8080)
}

func init() {
	rootCmd.AddCommand(interactiveCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// interactiveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// interactiveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	cwd, err := os.Getwd()
	if nil != err {
		fmt.Fprintf(os.Stderr, "failed to get current working directory: %v\n", err)
		os.Exit(1)
	}
	interactiveCmd.Flags().StringP("outdir", "o", cwd, "The path to the directory where to download recordings to.")
}
