package ffmpeg

import (
	"os"
	"path"
	"runtime"
	"testing"
)

func TestExecutablesPresent(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "..", "shims")

	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "Path/WithShim",
			path: dir,
			want: true,
		},
		{
			name: "Path/WithoutShim",
			path: "",
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldPath := os.Getenv("PATH")
			os.Setenv("PATH", tt.path)
			t.Cleanup(func() { os.Setenv("PATH", oldPath) })

			if got := ExecutablesPresent(); got != tt.want {
				t.Errorf("ExecutablesPresent() = %v, want %v", got, tt.want)
			}
		})
	}
}
