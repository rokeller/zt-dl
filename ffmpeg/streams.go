package ffmpeg

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type probeResult struct {
	Format  formatJson   `json:"format"`
	Streams []streamJson `json:"streams"`
}

type format struct {
	Duration time.Duration
}
type stream struct {
	Index     int
	CodecType string
}

type audioStream struct {
	stream
	SampleRate int
}

type videoStream struct {
	stream
	Width        int
	Height       int
	AvgFrameRate int
	BitRate      int
}

type formatJson struct {
	Duration string `json:"duration"` // seconds as string
}

type streamJson struct {
	Index        int    `json:"index"`
	CodecType    string `json:"codec_type"`
	SampleRate   string `json:"sample_rate"`    // audio streams only
	Width        int    `json:"width"`          // video streams only
	Height       int    `json:"height"`         // video streams only
	AvgFrameRate string `json:"avg_frame_rate"` // video streams only
	BitRate      string `json:"bit_rate"`       // video streams only

}

func (d *downloadable) DetectStreams(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ffprobeCmd := cmdFactory(ctx, "ffprobe",
		"-protocol_whitelist", protocolWhiteList,
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		"-i", d.inputUrl)

	output, err := ffprobeCmd.Output()
	if nil != err {
		return fmt.Errorf("failed to run ffprobe: %w", err)
	}

	var res probeResult
	r := bytes.NewReader(output)
	if err := json.NewDecoder(r).Decode(&res); nil != err {
		return fmt.Errorf("failed to JSON decode ffprobe output: %w", err)
	}

	duration, err := strconv.ParseFloat(res.Format.Duration, 32)
	if nil != err {
		return fmt.Errorf("failed to parse duration %q from ffprobe output: %w", res.Format.Duration, err)
	}
	f := format{
		Duration: time.Second * time.Duration(duration),
	}

	as := make([]audioStream, 0)
	vs := make([]videoStream, 0)

	for _, s := range res.Streams {
		switch s.CodecType {
		case "audio":
			sampleRate, err := strconv.ParseInt(s.SampleRate, 10, 0)
			if nil != err {
				return fmt.Errorf("failed to parse sample rate from %q: %w", s.SampleRate, err)
			}
			audio := audioStream{
				stream: stream{
					Index:     s.Index,
					CodecType: s.CodecType,
				},
				SampleRate: int(sampleRate),
			}
			as = append(as, audio)
		case "video":
			avgFrameRateStr := s.AvgFrameRate
			slashPos := strings.Index(avgFrameRateStr, "/")
			if slashPos >= 0 {
				avgFrameRateStr = avgFrameRateStr[0:slashPos]
			}
			avgFrameRate, err := strconv.ParseInt(avgFrameRateStr, 10, 0)
			if nil != err {
				return fmt.Errorf("failed to parse average frame rate from %q: %w", s.AvgFrameRate, err)
			}
			bitRate, err := strconv.ParseInt(s.BitRate, 10, 0)
			if nil != err {
				return fmt.Errorf("failed to parse bit rate from %q: %w", s.BitRate, err)
			}
			video := videoStream{
				stream: stream{
					Index:     s.Index,
					CodecType: s.CodecType,
				},
				Width:        s.Width,
				Height:       s.Height,
				AvgFrameRate: int(avgFrameRate),
				BitRate:      int(bitRate),
			}
			vs = append(vs, video)
		}
	}

	d.format = f
	d.audioStreams = as
	d.videoStreams = vs

	return nil
}

func (d *downloadable) getBestAudioStream() *audioStream {
	if len(d.audioStreams) == 1 {
		return &d.audioStreams[0]
	}

	var best *audioStream
	for _, as := range d.audioStreams {
		if nil == best || as.SampleRate > best.SampleRate {
			best = &as
		}
	}

	return best
}

func (d *downloadable) getBestVideoStream() *videoStream {
	if len(d.videoStreams) == 1 {
		return &d.videoStreams[0]
	}

	var best *videoStream
	for _, vs := range d.videoStreams {
		if nil == best ||
			(vs.Width > best.Width && vs.Height > best.Height && vs.AvgFrameRate > best.AvgFrameRate) {
			best = &vs
		}
	}

	return best
}
