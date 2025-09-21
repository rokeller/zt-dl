package ffmpeg

import (
	"reflect"
	"testing"
)

func TestNewDownloadable(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
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
