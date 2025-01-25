package main

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"bytes"
	sd "stackeddiff"
	ex "stackeddiff/execute"
	"stackeddiff/testinginit"
)

func TestSdLog_LogsOutput(t *testing.T) {
	assert := assert.New(t)

	testinginit.CdTestRepo()

	testinginit.AddCommit("first", "")

	outWriter := new(bytes.Buffer)
	ParseArguments(outWriter, flag.NewFlagSet("sd", flag.ExitOnError), []string{"log"})
	out := outWriter.String()

	assert.Contains(out, "first")
}

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

func TestSdUpdate_UpdatesPr(t *testing.T) {
	assert := assert.New(t)

	testinginit.CdTestRepo()

	testinginit.AddCommit("first", "")

	testinginit.SetTestExecutor()

	ParseArguments(os.Stdout, flag.NewFlagSet("sd", flag.ExitOnError), []string{"new"})

	testinginit.AddCommit("second", "")

	allCommits := sd.GetAllCommits()

	ParseArguments(os.Stdout, flag.NewFlagSet("sd", flag.ExitOnError), []string{"update", allCommits[1].Commit})

	allCommits = sd.GetAllCommits()

	assert.Equal(1, len(allCommits))
	assert.Equal("first", allCommits[0].Subject)
}

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

func TestSdBranchName_OutputsBranchName(t *testing.T) {
	assert := assert.New(t)

	testinginit.CdTestRepo()

	testinginit.AddCommit("first", "")

	allCommits := sd.GetAllCommits()

	outWriter := new(bytes.Buffer)
	ParseArguments(outWriter, flag.NewFlagSet("sd", flag.ExitOnError), []string{"branch-name", allCommits[0].Commit})
	out := outWriter.String()

	assert.Equal(sd.GetBranchInfo(allCommits[0].Commit).BranchName, out)
}

func TestSdNewWithReviewers_AddReviewers(t *testing.T) {
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
