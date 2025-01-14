package testing_init

import (
	"os"
	"path"
	"runtime"
	sd "stacked-diff-workflow/src/stacked-diff"
)

var TestWorkingDir string

func init() {
	_, filename, _, _ := runtime.Caller(0)
	TestWorkingDir = path.Join(path.Dir(filename), "/../../../.test-stacked-diff-workflow")
}

func CdTestDir() {
	sd.ExecuteOrDie(sd.ExecuteOptions{}, "rm", "-rf", TestWorkingDir)
	sd.ExecuteOrDie(sd.ExecuteOptions{}, "mkdir", "-p", TestWorkingDir)
	os.Chdir(TestWorkingDir)
}
