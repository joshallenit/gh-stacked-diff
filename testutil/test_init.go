/*
Utilities for unit testing this project.
*/
package testutil

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"runtime"
	"strings"
	"time"

	"testing"

	ex "github.com/joshallenit/gh-stacked-diff/v2/execute"
	"github.com/joshallenit/gh-stacked-diff/v2/util"
)

const InitialCommitSubject = "Initial empty commit"

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
	TestWorkingDir = path.Join(userCacheDir, "gh-stacked-diff-tests")
}

// CD into repository directory and set any global DI variables (slog, sleep, and executor).
func InitTest(t *testing.T, logLevel slog.Level) *ex.TestExecutor {
	startTime := time.Now()
	opts := util.PrettyHandlerOptions{
		SlogOpts: slog.HandlerOptions{
			Level: logLevel,
		},
	}
	handler := util.NewPrettyHandler(os.Stdout, opts)
	slog.SetDefault(slog.New(handler))
	testFunctionName := getTestFunctionName()
	slog.Info("Running test " + testFunctionName + "\n")
	t.Cleanup(func() {
		slog.Info(fmt.Sprint("Running test "+testFunctionName+" took ", time.Since(startTime), "\n"))
	})

	// Set new TestExecutor in case previous test has faked any of the git responses.
	testExecutor := setTestExecutor()

	cdTestRepo(testFunctionName)
	// Setup author config in case it is not set on machine.
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "config", "user.email", "unit-test@example.com")
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "config", "user.name", "Unit Test")
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "commit", "--allow-empty", "-m", InitialCommitSubject)
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "push", "origin", util.GetCurrentBranchName())

	util.SetDefaultSleep(func(d time.Duration) {
		slog.Debug(fmt.Sprint("Skipping sleep in tests ", d))
	})
	return testExecutor
}

func getTestFunctionName() string {
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
	// Reduce the length of the function name as otherwise on windows the OS max can be exceeded.
	functionParts := strings.Split(functionName, "/")
	return functionParts[len(functionParts)-1]
}

func cdTestRepo(testFunctionName string) {
	cdTestDir(testFunctionName)
	// Create a git repository with a local remote
	remoteDir := "remote-repo"
	repositoryDir := "local-repo"

	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "init", "--bare", remoteDir)
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "clone", remoteDir, repositoryDir)

	if err := os.Chdir(repositoryDir); err != nil {
		panic(err)
	}
	// os.Getwd() returns an unusable path ("c:\..."") in windows when running from Git Bash. Instead use "pwd"
	wd := strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "pwd"))
	slog.Info("Changed to test repository directory:\n" + wd)
}

func cdTestDir(testFunctionName string) {
	individualTestDir := path.Join(TestWorkingDir, testFunctionName)
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "rm", "-rf", individualTestDir)
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "mkdir", "-p", individualTestDir)
	if err := os.Chdir(individualTestDir); err != nil {
		panic(err)
	}
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
