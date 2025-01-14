package main

import (
	"log"
	sd "stacked-diff-workflow/src/stacked-diff"
	testing_init "stacked-diff-workflow/src/testing-init"
	"testing"
)

func TestNewPr(t *testing.T) {
	testDir := testing_init.RootProjectDir + "/../.test/test-repo"
	sd.ExecuteOrDie(sd.ExecuteOptions{}, "rm", "-rf", testDir)
	sd.ExecuteOrDie(sd.ExecuteOptions{}, "mkdir", "-p", testDir)
	testExecutor := sd.TestExecutor{Dir: testDir, TestLogger: log.Default()}

	sd.SetGlobalExecutor(testExecutor)
	// Create the branches required and CD into them.

	// if I need to specify
	log.Println(sd.ExecuteOrDie(sd.ExecuteOptions{}, "pwd"))

	//CreateNewPr(true, "", "", 0, sd.GetBranchInfo(""), log.Default())
	t.Fatalf("Test not implemented")
}
