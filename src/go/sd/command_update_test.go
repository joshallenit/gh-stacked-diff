package main

import (
	"flag"
	"log/slog"
	"os"
	sd "stackeddiff"
	"testing"

	"github.com/stretchr/testify/assert"

	ex "stackeddiff/execute"
	"stackeddiff/testinginit"
)

func TestSdUpdate_UpdatesPr(t *testing.T) {
	assert := assert.New(t)

	testinginit.InitTest(slog.LevelInfo)

	testinginit.AddCommit("first", "")

	ParseArguments(os.Stdout, flag.NewFlagSet("sd", flag.ContinueOnError), []string{"new"})

	testinginit.AddCommit("second", "")

	allCommits := sd.GetAllCommits()

	ParseArguments(os.Stdout, flag.NewFlagSet("sd", flag.ContinueOnError), []string{"update", allCommits[1].Commit})

	allCommits = sd.GetAllCommits()

	assert.Equal(1, len(allCommits))
	assert.Equal("first", allCommits[0].Subject)
}

func TestSdUpdate_WithReviewers_AddReviewers(t *testing.T) {
	assert := assert.New(t)

	testExecutor := testinginit.InitTest(slog.LevelInfo)

	testinginit.AddCommit("first", "")

	ParseArguments(os.Stdout, flag.NewFlagSet("sd", flag.ContinueOnError), []string{"new"})

	testinginit.AddCommit("second", "")

	allCommits := sd.GetAllCommits()

	testExecutor.SetResponse(
		// There has to be at least 4 checks, each with 3 values: status, conclusion, and state.
		"SUCCESS\nSUCCESS\nSUCCESS\n"+
			"SUCCESS\nSUCCESS\nSUCCESS\n"+
			"SUCCESS\nSUCCESS\nSUCCESS\n"+
			"SUCCESS\nSUCCESS\nSUCCESS\n",
		nil, "gh", "pr", "view", ex.MatchAnyRemainingArgs)

	ParseArguments(os.Stdout, flag.NewFlagSet("sd", flag.ContinueOnError), []string{"update", "--reviewers=mybestie", "2"})

	ghExpectedArgs := []string{"pr", "edit", allCommits[1].Branch, "--add-reviewer", "mybestie"}
	expectedResponse := ex.ExecuteResponse{Out: "Ok", Err: nil, ProgramName: "gh", Args: ghExpectedArgs}
	assert.Contains(testExecutor.Responses, expectedResponse)
}
