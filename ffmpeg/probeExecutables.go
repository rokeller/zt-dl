package ffmpeg

import (
	"fmt"
	"os"
	"os/exec"
)

func ExecutablesPresent() bool {
	return probeExecutable("ffmpeg") && probeExecutable("ffprobe")
}

func probeExecutable(name string) bool {
	path, err := exec.LookPath(name)
	if nil != err {
		fmt.Fprintf(os.Stderr, "%s not found: %v\n", name, err)
		return false
	}

	fmt.Printf("%s found at %q.\n", name, path)
	return true
}
