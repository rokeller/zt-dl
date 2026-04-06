package ffmpeg

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"testing"

	e "github.com/rokeller/zt-dl/exec"
	"github.com/rokeller/zt-dl/test"
)

func Test_downloadable_DetectStreams_ffprobe_Complete(t *testing.T) {
	if test.IsTestCall() {
		test.AssertArgs(
			"ffprobe",
			"-protocol_whitelist", "https,tls,tcp",
			"-print_format", "json",
			"-show_format",
			"-show_streams",
			"-i", "https://foo.bar.com/probe-source",
		)
		fmt.Println(`{
	"format": { "duration": "123456.789" },
	"streams": [
	{
		"index": 0,
		"codec_type": "audio",
		"codec_name": "aac",
		"sample_rate": "44000",
		"channels": 2,
		"channel_layout": "stereo",
		"tags": {
			"language": "en"
		}
	},
	{
		"index": 1,
		"codec_type": "audio",
		"codec_name": "test",
		"sample_rate": "88000",
		"channels": 5,
		"channel_layout": "test",
		"tags": {
			"language": "test"
		}
	},
	{
		"index": 2,
		"codec_type": "video",
		"codec_name": "h264",
		"width": 600,
		"height": 400,
		"avg_frame_rate": "50/1",
		"bit_rate": "1200"
	},
	{
		"index": 5,
		"codec_type": "video",
		"codec_name": "test",
		"width": 1200,
		"height": 800,
		"avg_frame_rate": "30/1",
		"bit_rate": "3456000"
	},
	{
		"index": 6,
		"codec_type": "subtitle",
		"codec_name": "srt"
	},
	{
		"index": 7,
		"codec_type": "subtitle",
		"codec_name": "srt",
		"tags": {}
	},
	{
		"index": 8,
		"codec_type": "subtitle",
		"codec_name": "srt",
		"tags": {
			"language": "en"
		}
	}
]}`)
		os.Exit(0)
		return
	}

	me := test.CallerFuncName(0)
	e.CmdFactory = func(ctx context.Context, name string, arg ...string) *exec.Cmd {
		return test.TestCommandContext(t, me, ctx, name, arg...)
	}

	d := NewDownloadable("https://foo.bar.com/probe-source", "target.mp4")
	err := d.DetectStreams(t.Context())
	if nil != err {
		t.Errorf("downloadable.DetectStreams() got error %v, want nil", err)
	}

	expectedStreams := []SourceStream{
		&AudioStream{
			Stream: Stream{
				Index:     0,
				CodecType: "audio",
				CodecName: "aac",
			},
			SampleRate:    44000,
			Channels:      2,
			ChannelLayout: "stereo",
			Language:      "en",
		},
		&AudioStream{
			Stream: Stream{
				Index:     1,
				CodecType: "audio",
				CodecName: "test",
			},
			SampleRate:    88000,
			Channels:      5,
			ChannelLayout: "test",
			Language:      "test",
		},
		&VideoStream{
			Stream: Stream{
				Index:     2,
				CodecType: "video",
				CodecName: "h264",
			},
			Width:        600,
			Height:       400,
			AvgFrameRate: 50,
			BitRate:      1200,
		},
		&VideoStream{
			Stream: Stream{
				Index:     5,
				CodecType: "video",
				CodecName: "test",
			},
			Width:        1200,
			Height:       800,
			AvgFrameRate: 30,
			BitRate:      3456000,
		},
		&SubtitleStream{
			Stream: Stream{
				Index:     8,
				CodecType: "subtitle",
				CodecName: "srt",
			},
			Language: "en",
		},
	}
	if len(d.streams) != len(expectedStreams) ||
		!reflect.DeepEqual(d.streams, expectedStreams) {
		t.Errorf("streams mismatch: got %v, want %v", d.streams, expectedStreams)
	}
}

func Test_downloadable_getBestAudioStream(t *testing.T) {
	tests := []struct {
		name    string // description of this test case
		streams []SourceStream
		want    *AudioStream
	}{
		{
			name: "SingleSteram",
			streams: []SourceStream{
				&AudioStream{
					Stream:     Stream{Index: 12, CodecType: "audio"},
					SampleRate: 3456,
				},
			},
			want: &AudioStream{
				Stream:     Stream{Index: 12, CodecType: "audio"},
				SampleRate: 3456,
			},
		},
		{
			name: "MultipleStreams",
			streams: []SourceStream{
				&AudioStream{
					Stream:     Stream{Index: 12, CodecType: "audio"},
					SampleRate: 123,
				},
				&AudioStream{
					Stream:     Stream{Index: 23, CodecType: "audio"},
					SampleRate: 234,
				},
			},
			want: &AudioStream{
				Stream:     Stream{Index: 23, CodecType: "audio"},
				SampleRate: 234,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDownloadable("http://input", "./output")
			d.streams = tt.streams
			got := bestAudioStream(d.streams)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getBestAudioStream() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_downloadable_getBestVideoStream(t *testing.T) {
	tests := []struct {
		name    string // description of this test case
		streams []SourceStream
		want    *VideoStream
	}{
		{
			name: "SingleSteram",
			streams: []SourceStream{
				&VideoStream{
					Stream:       Stream{Index: 12, CodecType: "video"},
					Width:        1234,
					Height:       567,
					AvgFrameRate: 22,
					BitRate:      33333,
				},
			},
			want: &VideoStream{
				Stream:       Stream{Index: 12, CodecType: "video"},
				Width:        1234,
				Height:       567,
				AvgFrameRate: 22,
				BitRate:      33333,
			},
		},
		{
			name: "MultipleStreams",
			streams: []SourceStream{
				&VideoStream{
					Stream:       Stream{Index: 12, CodecType: "video"},
					Width:        123,
					Height:       56,
					AvgFrameRate: 22,
					BitRate:      33333,
				},
				&VideoStream{
					Stream:       Stream{Index: 34, CodecType: "video"},
					Width:        2345,
					Height:       678,
					AvgFrameRate: 33,
					BitRate:      444444,
				},
			},
			want: &VideoStream{
				Stream:       Stream{Index: 34, CodecType: "video"},
				Width:        2345,
				Height:       678,
				AvgFrameRate: 33,
				BitRate:      444444,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDownloadable("http://input", "./output")
			d.streams = tt.streams
			got := bestVideoStream(d.streams)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getBestVideoStream() = %v, want %v", got, tt.want)
			}
		})
	}
}
