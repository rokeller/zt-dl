package server

import (
	"context"
	"os"
	"os/exec"
	"runtime"
	"testing"
	"time"

	e "github.com/rokeller/zt-dl/exec"
	"github.com/rokeller/zt-dl/test"
	"github.com/rokeller/zt-dl/zattoo"
)

func TestServe(t *testing.T) {
	type args struct {
		port int
	}
	tests := []struct {
		name    string
		port    int
		wantErr bool
	}{
		{
			name:    "Success",
			port:    8001,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), time.Millisecond*100)
			defer cancel()

			a := &zattoo.Account{}
			outdir, err := os.MkdirTemp(os.TempDir(), "*")
			if nil != err {
				t.Fatalf("failed to get temporary dir: %v", err)
			}
			if err := Serve(ctx, a, outdir, tt.port); (nil != err) != tt.wantErr {
				t.Errorf("Serve() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_open(t *testing.T) {
	if test.IsTestCall() {
		var args []string
		switch runtime.GOOS {
		case "windows":
			args = []string{"cmd", "/c", "start"}
		case "darwin":
			args = []string{"open"}
		default: // "linux", "freebsd", "openbsd", "netbsd"
			args = []string{"xdg-open"}
		}
		args = append(args, "https://localhost:8080/test-open")
		test.AssertArgs(args...)
		os.Exit(0)
		return
	}

	me := test.CallerFuncName(0)
	e.CmdFactory = func(ctx context.Context, name string, arg ...string) *exec.Cmd {
		return test.TestCommandContext(t, me, ctx, name, arg...)
	}

	if err := open(t.Context(), "http://localhost:8080/test-open"); nil != err {
		t.Errorf("open() error = %v, wantErr %v", err, nil)
	}
}
