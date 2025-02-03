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

func TestSdNew_CreatesPr(t *testing.T) {
	assert := assert.New(t)

	testinginit.InitTest(slog.LevelInfo)

	testinginit.AddCommit("first", "")

	ParseArguments(os.Stdout, flag.NewFlagSet("sd", flag.ContinueOnError), []string{"new"})

	outWriter := testinginit.NewWriteRecorder()
	ParseArguments(outWriter, flag.NewFlagSet("sd", flag.ContinueOnError), []string{"log"})
	out := outWriter.String()

	assert.Contains(out, "✅")
}

func TestSdNew_WithReviewers_AddReviewers(t *testing.T) {
	assert := assert.New(t)

	testExecutor := testinginit.InitTest(slog.LevelInfo)

	testinginit.AddCommit("first", "")

	testExecutor.SetResponse(
		// There has to be at least 4 checks, each with 3 values: status, conclusion, and state.
		"SUCCESS\nSUCCESS\nSUCCESS\n"+
			"SUCCESS\nSUCCESS\nSUCCESS\n"+
			"SUCCESS\nSUCCESS\nSUCCESS\n"+
			"SUCCESS\nSUCCESS\nSUCCESS\n",
		nil, "gh", "pr", "view", ex.MatchAnyRemainingArgs)

	ParseArguments(os.Stdout, flag.NewFlagSet("sd", flag.ContinueOnError), []string{"new", "--reviewers=mybestie"})

	allCommits := sd.GetAllCommits()

	ghExpectedArgs := []string{"pr", "edit", allCommits[0].Branch, "--add-reviewer", "mybestie"}
	expectedResponse := ex.ExecuteResponse{Out: "Ok", Err: nil, ProgramName: "gh", Args: ghExpectedArgs}
	assert.Contains(testExecutor.Responses, expectedResponse)
}

func TestSdNew_WhenUsingListIndex_UsesCorrectList(t *testing.T) {
	assert := assert.New(t)

	testinginit.InitTest(slog.LevelInfo)

	testinginit.AddCommit("first", "")
	testinginit.AddCommit("second", "")
	testinginit.AddCommit("third", "")
	testinginit.AddCommit("fourth", "")

	allCommits := sd.GetAllCommits()

	ParseArguments(os.Stdout, flag.NewFlagSet("sd", flag.ContinueOnError), []string{"new", "2"})

	assert.Equal(true, sd.RemoteHasBranch(allCommits[1].Branch))
}
