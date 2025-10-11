package exec

import (
	"context"
	"os/exec"
)

var CmdFactory func(ctx context.Context, name string, arg ...string) *exec.Cmd = exec.CommandContext
