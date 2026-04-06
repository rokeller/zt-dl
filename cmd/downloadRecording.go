package cmd

import (
	"errors"
	"fmt"

	"github.com/rokeller/zt-dl/ffmpeg"
	"github.com/rokeller/zt-dl/zattoo"
	"github.com/spf13/cobra"
)

// downloadRecordingCmd represents the get-recording command
var downloadRecordingCmd = &cobra.Command{
	Use:     "download",
	Aliases: []string{"get-recording"},
	Short:   "Download a recording to a local file",
	Long: `Downloads audio and video streams of a recording to a local file.
This requires [1mffmpeg[0m and [1mffprobe[0m to be in the PATH.`,

	SilenceErrors: false,
	RunE:          runDownloadRecordingCmd,
}

func init() {
	addEmailAndDomainFlags(downloadRecordingCmd)
	addDownloadFlags(downloadRecordingCmd)
	rootCmd.AddCommand(downloadRecordingCmd)

	downloadRecordingCmd.Flags().StringP("out", "o", "", "Name of the output file")
	downloadRecordingCmd.MarkFlagRequired("out")

	downloadRecordingCmd.Flags().Int64P("rid", "r", -1, "ID of the recording to get")
	downloadRecordingCmd.MarkFlagRequired("rid")
}

func runDownloadRecordingCmd(cmd *cobra.Command, args []string) error {
	out, err := cmd.Flags().GetString("out")
	if nil != err {
		return err
	}
	recordingId, err := cmd.Flags().GetInt64("rid")
	if nil != err {
		return err
	}

	email := cmd.Flag(string(Email)).Value.String()
	domain := cmd.Flag(string(Domain)).Value.String()
	overwrite, _ := cmd.Flags().GetBool(string(Overwrite))
	selectStreams, _ := cmd.Flags().GetBool(string(SelectStreams))

	if selectStreams {
		return errors.New("manual stream selection not supported for this command yet - use the 'interactive' command instead")
	}

	acct := zattoo.NewAccount(email, domain)
	if err := acct.Login(); nil != err {
		return err
	}

	url, err := acct.GetRecordingStreamUrl(recordingId)
	if nil != err {
		return err
	}

	d := ffmpeg.NewDownloadable(url, out,
		ffmpeg.WithOverwrite(overwrite))

	fmt.Println("Detecting streams ...")
	if err := d.DetectStreams(cmd.Context()); nil != err {
		return err
	}

	fmt.Println("Starting download ...")
	return d.Download(cmd.Context(), ffmpeg.NewBestStreamsSelector(), nil)
}
