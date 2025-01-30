package main

import (
	"flag"
	"os"
	sd "stackeddiff"
	"testing"

	"github.com/stretchr/testify/assert"

	"bytes"
	ex "stackeddiff/execute"
	"stackeddiff/testinginit"
)

func TestSdNew_CreatesPr(t *testing.T) {
	assert := assert.New(t)

	testinginit.CdTestRepo()

	testinginit.AddCommit("first", "")

	testinginit.SetTestExecutor()

	ParseArguments(os.Stdout, flag.NewFlagSet("sd", flag.ExitOnError), []string{"new"})

	outWriter := new(bytes.Buffer)
	ParseArguments(outWriter, flag.NewFlagSet("sd", flag.ExitOnError), []string{"log"})
	out := outWriter.String()

	assert.Contains(out, "✅")
}

func TestSdNew_WithReviewers_AddReviewers(t *testing.T) {
	assert := assert.New(t)

	testinginit.CdTestRepo()

	testinginit.AddCommit("first", "")

	testExecutor := testinginit.SetTestExecutor()

	testExecutor.SetResponse(
		// There has to be at least 4 checks, each with 3 values: status, conclusion, and state.
		"SUCCESS\nSUCCESS\nSUCCESS\n"+
			"SUCCESS\nSUCCESS\nSUCCESS\n"+
			"SUCCESS\nSUCCESS\nSUCCESS\n"+
			"SUCCESS\nSUCCESS\nSUCCESS\n",
		nil, "gh", "pr", "view", ex.MatchAnyRemainingArgs)

	ParseArguments(os.Stdout, flag.NewFlagSet("sd", flag.ExitOnError), []string{"new", "--reviewers=mybestie"})

	allCommits := sd.GetAllCommits()

	ghExpectedArgs := []string{"pr", "edit", sd.GetBranchForCommit(allCommits[0].Commit), "--add-reviewer", "mybestie"}
	expectedResponse := ex.ExecuteResponse{Out: "Ok", Err: nil, ProgramName: "gh", Args: ghExpectedArgs}
	assert.Contains(testExecutor.Responses, expectedResponse)
}

func TestSdNew_WhenUsingListIndex_UsesCorrectList(t *testing.T) {
	assert := assert.New(t)

	testinginit.CdTestRepo()

	testinginit.AddCommit("first", "")
	testinginit.AddCommit("second", "")
	testinginit.AddCommit("third", "")
	testinginit.AddCommit("fourth", "")

	allCommits := sd.GetAllCommits()

	testinginit.SetTestExecutor()

	ParseArguments(os.Stdout, flag.NewFlagSet("sd", flag.ExitOnError), []string{"new", "2"})

	assert.Equal(true, sd.RemoteHasBranch(sd.GetBranchForCommit(allCommits[1].Commit)))
}
