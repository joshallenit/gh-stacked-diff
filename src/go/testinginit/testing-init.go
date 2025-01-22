package testinginit

import (
	"log"
	"os"
	"path"
	"runtime"

	"log/slog"
	ex "stackeddiff/execute"
)

var TestWorkingDir string
var thisFile string

func init() {
	_, file, _, ok := runtime.Caller(0)
	thisFile = file
	if !ok {
		panic("No caller information")
	}
	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		panic("Cannot find UserCacheDir: " + err.Error())
	}
	TestWorkingDir = path.Join(userCacheDir, "stacked-diff-workflow-unit-tests")
}

func CdTestDir() {
	var functionName string
	for i := 0; i < 10; i++ {
		pc, file, _, ok := runtime.Caller(i)
		if !ok {
			panic("No caller information")
		}
		if file != thisFile {
			functionName = runtime.FuncForPC(pc).Name()
			break
		}
	}
	if functionName == "" {
		panic("Could not find caller outside of " + thisFile)
	}
	individualTestDir := path.Join(TestWorkingDir, functionName)
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "rm", "-rf", individualTestDir)
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "mkdir", "-p", individualTestDir)
	os.Chdir(individualTestDir)
	// Use pwd instead of individualTestDir to avoid having a mix of \ and / path separators on Windows.
	pwd := ex.ExecuteOrDie(ex.ExecuteOptions{}, "pwd")
	log.Println("Changed to test directory: " + pwd)
}

func CdTestRepo() {
	CdTestDir()
	// Create a git repository with a local remote
	remoteDir := "remote-repo"
	repositoryDir := "local-repo"

	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "init", "--bare", remoteDir)
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "clone", remoteDir, repositoryDir)

	os.Chdir(repositoryDir)
	log.Println("Changed to repository directory: " + repositoryDir)
}

func AddCommit(commitMessage string, fileName string) {
	if fileName == "" {
		fileName = commitMessage
	}
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "touch", fileName)
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "add", ".")
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "commit", "-m", commitMessage)
}

func IgnoreGithubCli() {
	testExecutor := ex.TestExecutor{TestLogger: slog.Default()}
	testExecutor.SetResponse("Ok", nil, "gh")
	ex.SetGlobalExecutor(testExecutor)
}
