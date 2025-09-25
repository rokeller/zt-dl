package ffmpeg

import (
	"context"
	"os"
	"os/exec"
	"reflect"
	"testing"

	"github.com/rokeller/zt-dl/test"
)

func TestNewDownloadable(t *testing.T) {
	tests := []struct {
		name       string // description of this test case
		inputUrl   string
		outputPath string
		want       downloadable
	}{
		{
			name:       "Passthrough of parameters",
			inputUrl:   "abc",
			outputPath: "def",
			want: downloadable{
				inputUrl:   "abc",
				outputPath: "def",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewDownloadable(tt.inputUrl, tt.outputPath)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewDownloadable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_downloadable_Download(t *testing.T) {
	tests := []struct {
		name    string // description of this test case
		d       downloadable
		wantErr bool
	}{
		{
			name:    "Missing Audio",
			d:       downloadable{},
			wantErr: true,
		},
		{
			name: "Missing Video",
			d: downloadable{
				audioStreams: []audioStream{
					{},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := tt.d.Download(context.Background())
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Download() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Download() succeeded unexpectedly")
			}
		})
	}
}

func Test_downloadable_Download_ffmpeg(t *testing.T) {
	if test.IsTestCall() {
		test.AssertArgs(
			"ffmpeg",
			"-protocol_whitelist", "https,tls,tcp",
			"-i", "https://foo.bar.com/source",
			"-map", "0:0",
			"-map", "0:2",
			"-c", "copy",
			"target.mp4",
		)
		os.Exit(0)
		return
	}

	me := test.CallerFuncName(0)
	cmdFactory = func(ctx context.Context, name string, arg ...string) *exec.Cmd {
		return test.TestCommandContext(t, me, ctx, name, arg...)
	}

	d := NewDownloadable("https://foo.bar.com/source", "target.mp4")
	d.audioStreams = []audioStream{
		{
			stream: stream{
				Index: 0,
			},
			SampleRate: 1234,
		},
	}
	d.videoStreams = []videoStream{
		{
			stream: stream{
				Index: 2,
			},
			Width:        987,
			Height:       876,
			AvgFrameRate: 12,
			BitRate:      12345,
		},
	}
	err := d.Download(t.Context())
	if nil != err {
		t.Errorf("downloadable.Download() got error %v, want nil", err)
	}
}
