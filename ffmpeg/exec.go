package ffmpeg

import (
	"context"
	"os/exec"
)

var cmdFactory func(ctx context.Context, name string, arg ...string) *exec.Cmd = exec.CommandContext
