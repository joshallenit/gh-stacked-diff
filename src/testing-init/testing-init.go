package testing_init

import (
	"os"
	"path"
	"runtime"
	ex "stacked-diff-workflow/src/execute"
)

var TestWorkingDir string

func init() {
	_, filename, _, _ := runtime.Caller(0)
	TestWorkingDir = path.Join(path.Dir(filename), "/../../../.test-stacked-diff-workflow")
}

func CdTestDir() {
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "rm", "-rf", TestWorkingDir)
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "mkdir", "-p", TestWorkingDir)
	os.Chdir(TestWorkingDir)
}
