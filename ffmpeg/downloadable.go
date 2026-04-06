package ffmpeg

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	e "github.com/rokeller/zt-dl/exec"
)

type downloadable struct {
	inputUrl   string
	outputPath string

	overwrite bool

	format  format
	streams []SourceStream
}

func NewDownloadable(
	inputUrl, outputPath string,
	options ...DownloadableOption,
) *downloadable {
	d := &downloadable{
		inputUrl:   inputUrl,
		outputPath: outputPath,
	}
	for _, option := range options {
		option(d)
	}
	return d
}

func (d *downloadable) Download(
	ctx context.Context,
	selector StreamsSelector,
	progress DownloadProgressHandler,
) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	durationMsec := d.format.Duration.Milliseconds()
	if nil == d.streams || len(d.streams) <= 0 {
		return errors.New("no streams available for download")
	}
	streams, err := selector.SelectStreams(d.streams)
	if nil != err {
		return fmt.Errorf("failed to select streams to download: %w", err)
	} else if len(streams) <= 0 {
		return errors.New("no streams selected for download")
	}

	args := []string{
		"-protocol_whitelist", protocolWhiteList,
		"-i", d.inputUrl,
	}

	if d.overwrite {
		args = append(args, "-y")
	} else {
		args = append(args, "-n")
	}

	fmt.Printf("Duration: %s\n", d.format.Duration)
	fmt.Println("Selected stream(s) for download:")
	for i, s := range streams {
		t := "Unknown"
		switch s.(type) {
		case *AudioStream:
			t = "Audio"
		case *SubtitleStream:
			t = "Subtitle"
		case *VideoStream:
			t = "Video"
		}
		fmt.Printf("    [%0d] %s: %s\n", i+1, t, s.String())
		args = append(args, "-map", fmt.Sprintf("0:%d", s.Index()))
	}

	args = append(args, "-c", "copy", d.outputPath)
	ffmpegCmd := e.CmdFactory(ctx, "ffmpeg", args...)
	stderr, err := ffmpegCmd.StderrPipe()
	if nil != err {
		return fmt.Errorf("failed to redirect stderr to pipe: %w", err)
	}

	if nil == progress {
		progress = consoleProgressHandler{
			target:     os.Stdout,
			outputPath: d.outputPath,
		}
	}

	tracker := downloadProgressTracker{
		handler: progress,
		source:  stderr,

		start:        time.Now().UTC(),
		durationMsec: durationMsec,
	}
	go tracker.trackProgress()

	// Now start the ffmpeg process ...
	progress.Start()
	if err := ffmpegCmd.Start(); nil != err {
		return fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	if err := ffmpegCmd.Wait(); nil != err {
		return fmt.Errorf("ffmpeg failed: %w", err)
	}
	progress.Finished()
	return nil
}
