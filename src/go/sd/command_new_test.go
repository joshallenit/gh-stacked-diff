package main

import (
	"flag"
	"log/slog"
	"os"
	"slices"
	sd "stackeddiff"
	"testing"

	"github.com/stretchr/testify/assert"

	"errors"
	ex "stackeddiff/execute"
	"stackeddiff/sliceutil"
	"stackeddiff/testinginit"
)

func TestSdNew_CreatesPr(t *testing.T) {
	assert := assert.New(t)

	testinginit.InitTest(slog.LevelInfo)

	testinginit.AddCommit("first", "")

	parseArguments(os.Stdout, flag.NewFlagSet("sd", flag.ContinueOnError), []string{"new"})

	outWriter := testinginit.NewWriteRecorder()
	parseArguments(outWriter, flag.NewFlagSet("sd", flag.ContinueOnError), []string{"log"})
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

	parseArguments(os.Stdout, flag.NewFlagSet("sd", flag.ContinueOnError), []string{"new", "--reviewers=mybestie"})

	allCommits := sd.GetAllCommits()

	contains := slices.ContainsFunc(testExecutor.Responses, func(next ex.ExecutedResponse) bool {
		ghExpectedArgs := []string{"pr", "edit", allCommits[0].Branch, "--add-reviewer", "mybestie"}
		return next.ProgramName == "gh" && slices.Equal(next.Args, ghExpectedArgs)
	})
	assert.True(contains, sliceutil.FilterSlice(testExecutor.Responses, func(next ex.ExecutedResponse) bool {
		return next.ProgramName == "gh"
	}))
}

func TestSdNew_WhenUsingListIndex_UsesCorrectList(t *testing.T) {
	assert := assert.New(t)

	testinginit.InitTest(slog.LevelInfo)

	testinginit.AddCommit("first", "")
	testinginit.AddCommit("second", "")
	testinginit.AddCommit("third", "")
	testinginit.AddCommit("fourth", "")

	allCommits := sd.GetAllCommits()

	parseArguments(os.Stdout, flag.NewFlagSet("sd", flag.ContinueOnError), []string{"new", "2"})

	assert.Equal(true, sd.RemoteHasBranch(allCommits[1].Branch))
}

func TestSdNew_WhenDraftNotSupported_TriesAgainWithoutDraft(t *testing.T) {
	assert := assert.New(t)

	testExecutor := testinginit.InitTest(slog.LevelInfo)

	testinginit.AddCommit("first", "")

	draftNotSupported := "pull request create failed: GraphQL: Draft pull requests are not supported in this repository. (createPullRequest)"
	testExecutor.SetResponseFunc(draftNotSupported, errors.New("sss"), func(programName string, args ...string) bool {
		return programName == "gh" && args[0] == "pr" && args[1] == "create" && slices.Contains(args, "--draft")
	})

	out := testinginit.NewWriteRecorder()
	parseArguments(out, flag.NewFlagSet("sd", flag.ContinueOnError), []string{"new"})

	assert.Contains(out.String(), "Use \"--draft=false\" to avoid this warning")
	assert.Contains(out.String(), "Created PR ")
}
