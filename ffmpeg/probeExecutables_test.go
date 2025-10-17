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
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "Path/WithShim",
			path:    dir,
			wantErr: false,
		},
		{
			name:    "Path/WithoutShim",
			path:    "",
			wantErr: true,
		},
		{
			name:    "Path/WithoutProbeShim",
			path:    path.Join(dir, "no-probe"),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldPath := os.Getenv("PATH")
			os.Setenv("PATH", tt.path)
			t.Cleanup(func() { os.Setenv("PATH", oldPath) })

			gotErr := ExecutablesPresent()
			if nil != gotErr && !tt.wantErr {
				t.Errorf("ExecutablesPresent() got error %v, want nil", gotErr)
			} else if nil == gotErr && tt.wantErr {
				t.Error("ExecutablesPresent() got nil, want error")
			}
		})
	}
}
