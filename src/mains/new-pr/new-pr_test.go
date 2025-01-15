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
	//testExecutor := sd.TestExecutor{Dir: testing_init.TestWorkingDir, TestLogger: log.Default()}

	//defaultBranch := strings.TrimSpace(sd.ExecuteOrDie(sd.ExecuteOptions{}, "git", "config", "init.defaultBranch"))
	sd.ExecuteOrDie(sd.ExecuteOptions{}, "git", "init", "--bare", remoteDir)
	sd.ExecuteOrDie(sd.ExecuteOptions{}, "git", "clone", remoteDir, repositoryDir)

	os.Chdir(repositoryDir)
	sd.ExecuteOrDie(sd.ExecuteOptions{}, "git", "branch", "--unset-upstream")

	log.Println(sd.ExecuteOrDie(sd.ExecuteOptions{}, "pwd"))

	sd.ExecuteOrDie(sd.ExecuteOptions{}, "touch", "README.md")
	sd.ExecuteOrDie(sd.ExecuteOptions{}, "git", "add", ".")
	sd.ExecuteOrDie(sd.ExecuteOptions{}, "git", "commit", "-m", "Add README")

	//sd.SetGlobalExecutor(testExecutor)
	// Create the branches required and CD into them.

	CreateNewPr(true, "", sd.GetMainBranch(), 0, sd.GetBranchInfo(""), log.Default())
	t.Fatalf("Test not implemented")
}
