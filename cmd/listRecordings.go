package cmd

import (
	"fmt"
	"time"

	"github.com/rokeller/zt-dl/zattoo"
	"github.com/spf13/cobra"
)

// listRecordingsCmd represents the list-recordings command
var listRecordingsCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"list-recordings"},
	Short:   "List recordings available for download",
	Long: `Lists recordings in your recording library that are currently available
for downloading`,

	SilenceErrors: false,
	RunE:          runlistRecordingsCmd,
}

func init() {
	rootCmd.AddCommand(listRecordingsCmd)
}

func runlistRecordingsCmd(cmd *cobra.Command, args []string) error {
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
	// fmt.Println("index,id,program_id,title,episode_title,level,start,end")
	for i, r := range rec {
		if r.End.After(now) {
			// Skip recordings which haven't finished recording yet.
			continue
		}
		if r.EpisodeTitle != "" {
			fmt.Printf("%4d: %s - %s (%s/%s) (ID %d)\n", i, r.Title, r.EpisodeTitle, r.ChannelId, r.Level, r.Id)
		} else {
			fmt.Printf("%4d: %s (%s/%s) (ID %d)\n", i, r.Title, r.ChannelId, r.Level, r.Id)
		}
		// fmt.Printf("%d,%d,%d,\"%s\",\"%s\",%s,%s,%s\n", i, r.Id, r.ProgramId, r.Title, r.EpisodeTitle, r.Level, r.Start, r.End)
	}

	return nil
}
