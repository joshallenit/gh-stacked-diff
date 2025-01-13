package main

import (
	"log"
	sd "stacked-diff-workflow/cmd/stacked-diff"
	"testing"
)

func TestNewPr(t *testing.T) {
	// Create the branches required and CD into them.
	sd.Execute(sd.ExecuteOptions{}, "mkdir", "test-repo")
	sd.Execute(sd.ExecuteOptions{}, "cd", "test-repo")
	CreateNewPr(true, "", "", 0, sd.GetBranchInfo(""), log.Default())
	t.Fatalf("Test not implemented")
}
