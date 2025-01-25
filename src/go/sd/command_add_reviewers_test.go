package main

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	sd "stackeddiff"
	ex "stackeddiff/execute"
	"stackeddiff/testinginit"
)

func TestSdAddReviewers_AddReviewers(t *testing.T) {
	assert := assert.New(t)

	testinginit.CdTestRepo()

	testinginit.AddCommit("first", "")

	testExecutor := testinginit.SetTestExecutor()

	ParseArguments(os.Stdout, flag.NewFlagSet("sd", flag.ExitOnError), []string{"new"})

	allCommits := sd.GetAllCommits()
	testExecutor.SetResponse(
		// There has to be at least 4 checks, each with 3 values: status, conclusion, and state.
		"SUCCESS\nSUCCESS\nSUCCESS\n"+
			"SUCCESS\nSUCCESS\nSUCCESS\n"+
			"SUCCESS\nSUCCESS\nSUCCESS\n"+
			"SUCCESS\nSUCCESS\nSUCCESS\n",
		nil, "gh", "pr", "view", ex.MatchAnyRemainingArgs)

	ParseArguments(os.Stdout, flag.NewFlagSet("sd", flag.ExitOnError), []string{"add-reviewers", "--reviewers=mybestie", allCommits[0].Commit})

	ghExpectedArgs := []string{"pr", "edit", sd.GetBranchForCommit(allCommits[0].Commit), "--add-reviewer", "mybestie"}
	expectedResponse := ex.ExecuteResponse{Out: "Ok", Err: nil, ProgramName: "gh", Args: ghExpectedArgs}
	assert.Contains(testExecutor.Responses, expectedResponse)
}
