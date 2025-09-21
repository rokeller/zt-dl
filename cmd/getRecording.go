package cmd

import (
	"fmt"

	"github.com/rokeller/zt-dl/ffmpeg"
	"github.com/rokeller/zt-dl/zattoo"
	"github.com/spf13/cobra"
)

// getRecordingCmd represents the get-recording command
var getRecordingCmd = &cobra.Command{
	Use: "get-recording",

	SilenceErrors: false,

	RunE: runGetRecordingCmd,
}

func init() {
	rootCmd.AddCommand(getRecordingCmd)

	getRecordingCmd.Flags().StringP("out", "o", "", "Name of the output file")
	getRecordingCmd.MarkFlagRequired("out")

	getRecordingCmd.Flags().Int64P("rid", "r", -1, "ID of the recording to get")
	getRecordingCmd.MarkFlagRequired("rid")

	getRecordingCmd.Flags().Int64P("pid", "p", -1, "ID of the program to get")
}

func runGetRecordingCmd(cmd *cobra.Command, args []string) error {
	out, err := cmd.Flags().GetString("out")
	if nil != err {
		return err
	}
	recordingId, err := cmd.Flags().GetInt64("rid")
	if nil != err {
		return err
	}
	programId, err := cmd.Flags().GetInt64("pid")
	if nil != err {
		return err
	}

	acct := zattoo.NewAccount(cmd.Flag("email").Value.String())
	if err := acct.Login(); nil != err {
		return err
	}

	if -1 != programId {
		details, err := acct.GetProgramDetails(programId)
		if nil != err {
			return err
		}
		fmt.Printf("details: %#v\n", details)
	}

	url, err := acct.GetRecordingStreamUrl(recordingId)
	if nil != err {
		return err
	}

	d := ffmpeg.NewDownloadable(url, out)

	fmt.Println("Detecting streams ...")
	if err := d.DetectStreams(cmd.Context()); nil != err {
		return err
	}

	return d.Download(cmd.Context())
}
