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

func (d *downloadable) Download(ctx context.Context) error {
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
		return err
	}

	tracker := downloadProgressTracker{
		outType: "stderr",
		source:  stderr,
		target:  os.Stdout,

		start:        time.Now().UTC(),
		durationMsec: durationMsec,
	}
	go tracker.showDownloadProgress()

	// Now start the ffmpeg process ...
	fmt.Println("Starting download ...")
	if err := ffmpegCmd.Start(); nil != err {
		return err
	}

	if err := ffmpegCmd.Wait(); err != nil {
		return err
	}

	fmt.Println("Finished download.")
	fmt.Printf("Recording written to '%s'.\n", d.outputPath)

	return nil
}
