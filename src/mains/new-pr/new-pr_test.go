package main

import (
	"log"
	"os"
	sd "stacked-diff-workflow/src/stacked-diff"
	testing_init "stacked-diff-workflow/src/testing-init"
	"testing"
)

func TestNewPr(t *testing.T) {
	testing_init.CdTestDir()
	// Create a git repository with a local remote
	remoteDir := "remote-repo"
	repositoryDir := "local-repo"

	sd.ExecuteOrDie(sd.ExecuteOptions{}, "git", "init", "--bare", remoteDir)
	sd.ExecuteOrDie(sd.ExecuteOptions{}, "git", "clone", remoteDir, repositoryDir)

	os.Chdir(repositoryDir)

	log.Println(sd.ExecuteOrDie(sd.ExecuteOptions{}, "pwd"))

	sd.ExecuteOrDie(sd.ExecuteOptions{}, "touch", "README.md")
	sd.ExecuteOrDie(sd.ExecuteOptions{}, "git", "add", ".")
	sd.ExecuteOrDie(sd.ExecuteOptions{}, "git", "commit", "-m", "Add README")

	testExecutor := sd.TestExecutor{TestLogger: log.Default()}
	testExecutor.SetResponse("Ok", nil, "gh")
	sd.SetGlobalExecutor(testExecutor)

	CreateNewPr(true, "", sd.GetMainBranch(), 0, sd.GetBranchInfo(""), log.Default())
}
