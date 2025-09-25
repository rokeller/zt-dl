package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/rokeller/zt-dl/zattoo"
	"github.com/spf13/cobra"
)

var version string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "zt-dl",
	Short: "Get a textual list of recordings from Zattoo",
	Long: `Lists all currently available (i.e. completed) recordings from your Zattoo subscription.

The information provided in the output includes the recording ID (the number in
the final parenthesis per line), which is needed when downloading a specific recording.`,

	SilenceErrors: false,
	RunE:          runRootCmd,
	Version:       version,
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
	rootCmd.PersistentFlags().StringP(string(Email), "e", "", "Email address of your Zattoo account.")
	rootCmd.MarkPersistentFlagRequired(string(Email))

	rootCmd.PersistentFlags().StringP(string(Domain), "d", "zattoo.com", "Domain of your Zattoo subscription.")
}

func runRootCmd(cmd *cobra.Command, args []string) error {
	email := cmd.Flag(string(Email)).Value.String()
	domain := cmd.Flag(string(Domain)).Value.String()
	acct := zattoo.NewAccount(email, domain)
	if err := acct.Login(); nil != err {
		return err
	}
	rec, err := acct.GetAllRecordings()
	if nil != err {
		return err
	}

	now := time.Now()
	fmt.Println("Ready recordings:")
	for i, r := range rec {
		if r.End.After(now) {
			// Skip recordings which haven't finished recording yet.
			continue
		}
		if r.EpisodeTitle != "" {
			fmt.Printf("%4d: %s - %s (%s/%s) (#%d)\n", i, r.Title, r.EpisodeTitle, r.ChannelId, r.Level, r.Id)
		} else {
			fmt.Printf("%4d: %s (%s/%s) (#%d)\n", i, r.Title, r.ChannelId, r.Level, r.Id)
		}
	}

	return nil
}
