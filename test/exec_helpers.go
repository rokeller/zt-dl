package test

import (
	"context"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"
)

const TestCallArgLen = 1 /* the executable itself */ + 3 /* -test.run, test-selector, double-dash */
const TestEnvVar = "GO_TEST_PROCESS"

func IsTestCall() bool {
	return os.Getenv(TestEnvVar) == "1"
}

func CallerFuncName(skip int) string {
	pc, _, _, ok := runtime.Caller(skip + 1 /* because calling CallerFuncName is another call */)
	if !ok {
		return "n/a"
	}
	f := runtime.FuncForPC(pc)
	if nil == f {
		return "n/a"
	}
	qualifiedName := f.Name()
	lastPeriod := strings.LastIndex(qualifiedName, ".")
	if lastPeriod >= 0 {
		return qualifiedName[lastPeriod+1:]
	}
	return "n/a"
}

func TestCommandContext(t *testing.T, testName string, ctx context.Context, name string, args ...string) (cmd *exec.Cmd) {
	t.Helper()

	testCallArgs := append(
		[]string{"-test.run", "^" + testName + "$", "--", name},
		args...)
	cmd = exec.CommandContext(ctx, os.Args[0], testCallArgs...)
	cmd.Env = []string{TestEnvVar + "=1"}

	return
}

func AssertArgs(expectedArgs ...string) {
	args := GetArgs()
	if len(args) != len(expectedArgs) {
		os.Exit(-1)
	}
	for i, a := range args {
		if a != expectedArgs[i] {
			os.Exit(1 + i)
		}
	}
}

func GetArgs() []string {
	return os.Args[TestCallArgLen:]
}
