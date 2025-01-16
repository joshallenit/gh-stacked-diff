package testing_init

import (
	"os"
	"path"
	"runtime"
	ex "stacked-diff-workflow/src/execute"
)

var TestWorkingDir string

func init() {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("No caller information")
	}
	TestWorkingDir = path.Join(path.Dir(filename), "/../../../.test-stacked-diff-workflow")
}

func CdTestDir() {
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		panic("No caller information")
	}
	fn := runtime.FuncForPC(pc)

	individualTestDir := TestWorkingDir + "/" + fn.Name()
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "rm", "-rf", individualTestDir)
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "mkdir", "-p", individualTestDir)
	os.Chdir(individualTestDir)
	println("Changed to test directory: " + individualTestDir)
}
