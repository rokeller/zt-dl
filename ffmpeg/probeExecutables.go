package ffmpeg

import (
	"os/exec"
)

func ExecutablesPresent() error {
	if err := probeExecutable("ffmpeg"); nil != err {
		return err
	}
	if err := probeExecutable("ffprobe"); nil != err {
		return err
	}
	return nil
}

func probeExecutable(name string) error {
	_, err := exec.LookPath(name)
	if nil != err {
		return err
	}
	return nil
}
