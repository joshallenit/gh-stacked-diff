package testinginit

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"runtime"
	ex "stackeddiff/execute"
	"strings"
	"time"
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

func InitTest(logLevel slog.Level) *ex.TestExecutor {
	opts := ex.PrettyHandlerOptions{
		SlogOpts: slog.HandlerOptions{
			Level: logLevel,
		},
	}
	handler := ex.NewPrettyHandler(os.Stdout, opts)
	slog.SetDefault(slog.New(handler))
	cdTestRepo()
	ex.SetDefaultSleep(func(d time.Duration) {
		slog.Debug(fmt.Sprint("Skipping sleep in tests ", d))
	})
	return setTestExecutor()
}

func cdTestRepo() {
	cdTestDir()
	// Create a git repository with a local remote
	remoteDir := "remote-repo"
	repositoryDir := "local-repo"

	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "init", "--bare", remoteDir)
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "clone", remoteDir, repositoryDir)

	os.Chdir(repositoryDir)
	// os.Getwd() returns an unusable path ("c:\..."") in windows when running from Git Bash. Instead use "pwd"
	wd := strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "pwd"))
	slog.Info("Changed to test repository directory:\n" + wd)
}

func cdTestDir() {
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
}

func setTestExecutor() *ex.TestExecutor {
	testExecutor := ex.TestExecutor{}
	testExecutor.SetResponse("Ok", nil, "gh", ex.MatchAnyRemainingArgs)
	testExecutor.SetResponse("Ok", nil, "say", ex.MatchAnyRemainingArgs)
	ex.SetGlobalExecutor(&testExecutor)
	return &testExecutor
}

func AddCommit(commitMessage string, filename string) {
	if filename == "" {
		filename = commitMessage
	}
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "touch", filename)
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "add", ".")
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "commit", "-m", commitMessage)
}

func CommitFileChange(commitMessage string, filename string, fileContents string) {
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "touch", filename)
	if writeErr := os.WriteFile(filename, []byte(fileContents), 0); writeErr != nil {
		panic(writeErr)
	}
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "add", ".")
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "commit", "-m", commitMessage)
}
