package main

import (
	"log"
	"os"
	ex "stacked-diff-workflow/src/execute"
	testing_init "stacked-diff-workflow/src/testing-init"
	"testing"
)

func TestNewPr(t *testing.T) {

	testing_init.CdTestDir()
	// Create a git repository with a local remote
	remoteDir := "remote-repo"
	repositoryDir := "local-repo"

	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "init", "--bare", remoteDir)
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "clone", remoteDir, repositoryDir)

	os.Chdir(repositoryDir)

	ex.ExecuteOrDie(ex.ExecuteOptions{}, "touch", "README.md")
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "add", ".")
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "commit", "-m", "Add README")

	testExecutor := ex.TestExecutor{TestLogger: log.Default()}
	testExecutor.SetResponse("Ok", nil, "gh")
	ex.SetGlobalExecutor(testExecutor)

	PrintGitLog(os.Stdout)
}
