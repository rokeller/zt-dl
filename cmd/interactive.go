package cmd

import (
	"fmt"
	"os"

	"github.com/rokeller/zt-dl/server"
	"github.com/rokeller/zt-dl/zattoo"
	"github.com/spf13/cobra"
)

const (
	OutDir    = Flag("outdir")
	Port      = Flag("port")
	OpenWebUI = Flag("open")
)

// interactiveCmd represents the interactive command
var interactiveCmd = &cobra.Command{
	Use:   "interactive",
	Short: "Run web interface for Zattoo recording download",
	Long: `Runs a local web server that lets you interact with zt-dl to examine and
download recordings from your recording library. Initiating downloads from the
web interface requires [1mffmpeg[0m and [1mffprobe[0m to be in the PATH.`,
	RunE: runInteractiveCmd,
}

func runInteractiveCmd(cmd *cobra.Command, args []string) error {
	email := cmd.Flag(string(Email)).Value.String()
	domain := cmd.Flag(string(Domain)).Value.String()
	acct := zattoo.NewAccount(email, domain)
	if err := acct.Login(); nil != err {
		return err
	}

	outdir := cmd.Flag(string(OutDir)).Value.String()
	port, _ := cmd.Flags().GetUint16(string(Port))
	open, _ := cmd.Flags().GetBool(string(OpenWebUI))
	return server.Serve(cmd.Context(), acct, outdir, port, open)
}

func init() {
	rootCmd.AddCommand(interactiveCmd)

	cwd, err := os.Getwd()
	if nil != err {
		fmt.Fprintf(os.Stderr, "failed to get current working directory: %v\n", err)
		os.Exit(1)
	}
	interactiveCmd.Flags().StringP(string(OutDir), "o", cwd,
		"The path to the directory where to download recordings to.")
	interactiveCmd.Flags().Uint16P(string(Port), "p", 8080,
		"The local port to run the web server on.")
	interactiveCmd.Flags().BoolP(string(OpenWebUI), "w", true,
		"If set, automatically opens the web UI in your default browser.")
}
