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
		"sample_rate": "44000"
	},
	{
		"index": 1,
		"codec_type": "audio",
		"sample_rate": "88000"
	},
	{
		"index": 2,
		"codec_type": "video",
		"width": 600,
		"height": 400,
		"avg_frame_rate": "50/1",
		"bit_rate": "1200"
	},
	{
		"index": 5,
		"codec_type": "video",
		"width": 1200,
		"height": 800,
		"avg_frame_rate": "30/1",
		"bit_rate": "3456000"
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

	expectedAudioStreams := []audioStream{
		{
			stream: stream{
				Index:     0,
				CodecType: "audio",
			},
			SampleRate: 44000,
		},
		{
			stream: stream{
				Index:     1,
				CodecType: "audio",
			},
			SampleRate: 88000,
		},
	}
	if len(d.audioStreams) != len(expectedAudioStreams) ||
		!reflect.DeepEqual(d.audioStreams, expectedAudioStreams) {
		t.Errorf("audioStreams mismatch: got %v, want %v", d.audioStreams, expectedAudioStreams)
	}

	expectedVideoStreams := []videoStream{
		{
			stream: stream{
				Index:     2,
				CodecType: "video",
			},
			Width:        600,
			Height:       400,
			AvgFrameRate: 50,
			BitRate:      1200,
		},
		{
			stream: stream{
				Index:     5,
				CodecType: "video",
			},
			Width:        1200,
			Height:       800,
			AvgFrameRate: 30,
			BitRate:      3456000,
		},
	}
	if len(d.videoStreams) != len(expectedVideoStreams) ||
		!reflect.DeepEqual(d.videoStreams, expectedVideoStreams) {
		t.Errorf("videoStreams mismatch: got %v, want %v", d.videoStreams, expectedVideoStreams)
	}
}

func Test_downloadable_getBestAudioStream(t *testing.T) {
	tests := []struct {
		name    string // description of this test case
		streams []audioStream
		want    *audioStream
	}{
		{
			name: "SingleSteram",
			streams: []audioStream{
				{
					stream:     stream{Index: 12, CodecType: "audio"},
					SampleRate: 3456,
				},
			},
			want: &audioStream{
				stream:     stream{Index: 12, CodecType: "audio"},
				SampleRate: 3456,
			},
		},
		{
			name: "MultipleStreams",
			streams: []audioStream{
				{
					stream:     stream{Index: 12, CodecType: "audio"},
					SampleRate: 123,
				},
				{
					stream:     stream{Index: 23, CodecType: "audio"},
					SampleRate: 234,
				},
			},
			want: &audioStream{
				stream:     stream{Index: 23, CodecType: "audio"},
				SampleRate: 234,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDownloadable("http://input", "./output")
			d.audioStreams = tt.streams
			got := d.getBestAudioStream()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getBestAudioStream() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_downloadable_getBestVideoStream(t *testing.T) {
	tests := []struct {
		name    string // description of this test case
		streams []videoStream
		want    *videoStream
	}{
		{
			name: "SingleSteram",
			streams: []videoStream{
				{
					stream:       stream{Index: 12, CodecType: "video"},
					Width:        1234,
					Height:       567,
					AvgFrameRate: 22,
					BitRate:      33333,
				},
			},
			want: &videoStream{
				stream:       stream{Index: 12, CodecType: "video"},
				Width:        1234,
				Height:       567,
				AvgFrameRate: 22,
				BitRate:      33333,
			},
		},
		{
			name: "MultipleStreams",
			streams: []videoStream{
				{
					stream:       stream{Index: 12, CodecType: "video"},
					Width:        123,
					Height:       56,
					AvgFrameRate: 22,
					BitRate:      33333,
				},
				{
					stream:       stream{Index: 34, CodecType: "video"},
					Width:        2345,
					Height:       678,
					AvgFrameRate: 33,
					BitRate:      444444,
				},
			},
			want: &videoStream{
				stream:       stream{Index: 34, CodecType: "video"},
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
			d.videoStreams = tt.streams
			got := d.getBestVideoStream()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getBestVideoStream() = %v, want %v", got, tt.want)
			}
		})
	}
}
