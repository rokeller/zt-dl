package exec

import (
	"context"
	"fmt"
	"os/exec"
)

var CmdFactory func(ctx context.Context, name string, arg ...string) *exec.Cmd = exec.CommandContext

func ExecutablesPresent() error {
	if err := executablePresent("ffmpeg"); err != nil {
		return err
	}
	if err := executablePresent("ffprobe"); err != nil {
		return err
	}
	return nil
}

func executablePresent(name string) error {
	if path, err := exec.LookPath(name); err != nil {
		return err
	} else {
		fmt.Printf("Found %q: %s\n", name, path)
		return nil
	}
}
