package ffmpeg

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"
)

type downloadable struct {
	inputUrl   string
	outputPath string

	format       format
	audioStreams []audioStream
	videoStreams []videoStream
}

func NewDownloadable(inputUrl, outputPath string) downloadable {
	return downloadable{
		inputUrl:   inputUrl,
		outputPath: outputPath,
	}
}

func (d *downloadable) Download(ctx context.Context, progress DownloadProgressHandler) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	durationMsec := d.format.Duration.Milliseconds()
	audio := d.getBestAudioStream()
	video := d.getBestVideoStream()
	if nil == audio {
		return errors.New("failed to get best audio stream")
	}
	if nil == video {
		return errors.New("failed to get best video stream")
	}

	fmt.Printf("Selected audio stream: sample rate %dHz (stream #%d)\n",
		audio.SampleRate, audio.Index)
	fmt.Printf("Selected video stream: width/height %d/%d, bit rate %dbps, avg frame rate %dfps (stream #%d)\n",
		video.Width, video.Height, video.BitRate, video.AvgFrameRate, video.Index)
	fmt.Printf("Duration: %s\n", d.format.Duration)

	ffmpegCmd := cmdFactory(ctx, "ffmpeg",
		"-protocol_whitelist", protocolWhiteList,
		"-i", d.inputUrl,
		"-map", fmt.Sprintf("0:%d", audio.Index),
		"-map", fmt.Sprintf("0:%d", video.Index),
		"-c", "copy",
		d.outputPath)

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
		return fmt.Errorf("failed to wait for ffmpeg to finish: %w", err)
	}
	progress.Finished()
	return nil
}
