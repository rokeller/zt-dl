package ffmpeg

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	e "github.com/rokeller/zt-dl/exec"
)

// probeResult matches the JSON output from ffprobe.
type probeResult struct {
	Format  formatJson   `json:"format"`
	Streams []streamJson `json:"streams"`
}
type formatJson struct {
	Duration string `json:"duration"` // seconds as string
}
type streamJson struct {
	Index         int            `json:"index"`
	CodecType     string         `json:"codec_type"`
	CodecName     string         `json:"codec_name"`
	SampleRate    string         `json:"sample_rate"`    // audio streams only
	Channels      int            `json:"channels"`       // audio streams only
	ChannelLayout string         `json:"channel_layout"` // audio streams only
	Tags          map[string]any `json:"tags"`           // audio and subtitle streams
	Width         int            `json:"width"`          // video streams only
	Height        int            `json:"height"`         // video streams only
	AvgFrameRate  string         `json:"avg_frame_rate"` // video streams only
	BitRate       string         `json:"bit_rate"`       // video streams only
}

type format struct {
	Duration time.Duration
}

type SourceStream interface {
	fmt.Stringer
	Index() int
}

type Stream struct {
	Index     int
	CodecType string
	CodecName string
}

type AudioStream struct {
	Stream
	SampleRate    int
	Channels      int
	ChannelLayout string
	Language      string
}

var _ SourceStream = &AudioStream{}

// String implements [SourceStream]
func (s *AudioStream) String() string {
	return fmt.Sprintf(
		"%s, sample rate %dHz, %d channels (%s), language %q (stream #%d)",
		s.CodecName, s.SampleRate, s.Channels, s.ChannelLayout, s.Language, s.Stream.Index)
}

// Index implements [SourceStream]
func (s *AudioStream) Index() int {
	return s.Stream.Index
}

type SubtitleStream struct {
	Stream
	Language string
}

var _ SourceStream = &SubtitleStream{}

// String implements [SourceStream]
func (s *SubtitleStream) String() string {
	return fmt.Sprintf("%s, language %q (stream #%d)",
		s.CodecName, s.Language, s.Stream.Index)
}

// Index implements [SourceStream]
func (s *SubtitleStream) Index() int {
	return s.Stream.Index
}

type VideoStream struct {
	Stream
	Width        int
	Height       int
	AvgFrameRate int
	BitRate      int
}

var _ SourceStream = &VideoStream{}

// String implements [SourceStream]
func (s *VideoStream) String() string {
	return fmt.Sprintf("%s, dimensions %dx%d, bit rate %dbps, avg frame rate %dfps (stream #%d)",
		s.CodecName, s.Width, s.Height, s.BitRate, s.AvgFrameRate, s.Stream.Index)
}

// Index implements [SourceStream]
func (s *VideoStream) Index() int {
	return s.Stream.Index
}

func (d *downloadable) DetectStreams(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ffprobeCmd := e.CmdFactory(ctx, "ffprobe",
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

	streams := make([]SourceStream, 0)

	for _, s := range res.Streams {
		switch s.CodecType {
		case "audio":
			if audio, err := s.audioStream(); nil != err {
				return err
			} else if nil != audio {
				streams = append(streams, audio)
			}
		case "subtitle":
			if subtitle, err := s.subtitleStream(); nil != err {
				return err
			} else if nil != subtitle {
				streams = append(streams, subtitle)
			}
		case "video":
			if video, err := s.videoStream(); nil != err {
				return err
			} else if nil != video {
				streams = append(streams, video)
			}
		}
	}

	d.format = f
	d.streams = streams
	return nil
}

func (s streamJson) audioStream() (*AudioStream, error) {
	sampleRate, err := strconv.ParseInt(s.SampleRate, 10, 0)
	if nil != err {
		return nil, fmt.Errorf("failed to parse sample rate from %q: %w", s.SampleRate, err)
	}
	lang := "<not-specified>"
	if nil != s.Tags {
		if l, found := s.Tags["language"]; found {
			lang = l.(string)
		}
	}
	audio := AudioStream{
		Stream: Stream{
			Index:     s.Index,
			CodecType: s.CodecType,
			CodecName: s.CodecName,
		},
		SampleRate:    int(sampleRate),
		Channels:      s.Channels,
		ChannelLayout: s.ChannelLayout,
		Language:      lang,
	}
	return &audio, nil
}

func (s streamJson) subtitleStream() (*SubtitleStream, error) {
	if nil == s.Tags {
		fmt.Printf("WARN: failed to find tags for subtitle stream %d.\n", s.Index)
		return nil, nil
	}
	lang, found := s.Tags["language"]
	if !found {
		fmt.Printf("WARN: failed to find language tag for subtitle stream %d.\n", s.Index)
		return nil, nil
	}
	subtitle := SubtitleStream{
		Stream: Stream{
			Index:     s.Index,
			CodecType: s.CodecType,
			CodecName: s.CodecName,
		},
		Language: lang.(string),
	}
	return &subtitle, nil
}

func (s streamJson) videoStream() (*VideoStream, error) {
	avgFrameRateStr := s.AvgFrameRate
	slashPos := strings.Index(avgFrameRateStr, "/")
	if slashPos >= 0 {
		avgFrameRateStr = avgFrameRateStr[0:slashPos]
	}
	avgFrameRate, err := strconv.ParseInt(avgFrameRateStr, 10, 0)
	if nil != err {
		return nil, fmt.Errorf("failed to parse average frame rate from %q: %w", s.AvgFrameRate, err)
	}
	bitRate, err := strconv.ParseInt(s.BitRate, 10, 0)
	if nil != err {
		return nil, fmt.Errorf("failed to parse bit rate from %q: %w", s.BitRate, err)
	}
	video := VideoStream{
		Stream: Stream{
			Index:     s.Index,
			CodecType: s.CodecType,
			CodecName: s.CodecName,
		},
		Width:        s.Width,
		Height:       s.Height,
		AvgFrameRate: int(avgFrameRate),
		BitRate:      int(bitRate),
	}
	return &video, nil
}
