package main

import (
	"log/slog"
	sd "stackeddiff"
	"testing"

	"github.com/stretchr/testify/assert"

	"slices"
	ex "stackeddiff/execute"
	"stackeddiff/sliceutil"
	"stackeddiff/testinginit"
)

func TestSdUpdate_UpdatesPr(t *testing.T) {
	assert := assert.New(t)

	testinginit.InitTest(slog.LevelInfo)

	testinginit.AddCommit("first", "")

	testParseArguments("new")

	testinginit.AddCommit("second", "")

	allCommits := sd.GetAllCommits()

	testParseArguments("update", allCommits[1].Commit)

	allCommits = sd.GetAllCommits()

	assert.Equal(2, len(allCommits))
	assert.Equal("first", allCommits[0].Subject)
	assert.Equal(testinginit.InitialCommitSubject, allCommits[1].Subject)
}

func TestSdUpdate_WithReviewers_AddReviewers(t *testing.T) {
	assert := assert.New(t)

	testExecutor := testinginit.InitTest(slog.LevelInfo)

	testinginit.AddCommit("first", "")

	testParseArguments("new")

	testinginit.AddCommit("second", "")

	allCommits := sd.GetAllCommits()

	testExecutor.SetResponse(
		// There has to be at least 4 checks, each with 3 values: status, conclusion, and state.
		"SUCCESS\nSUCCESS\nSUCCESS\n"+
			"SUCCESS\nSUCCESS\nSUCCESS\n"+
			"SUCCESS\nSUCCESS\nSUCCESS\n"+
			"SUCCESS\nSUCCESS\nSUCCESS\n",
		nil, "gh", "pr", "view", ex.MatchAnyRemainingArgs)

	testParseArguments("update", "--reviewers=mybestie", "2")

	contains := slices.ContainsFunc(testExecutor.Responses, func(next ex.ExecutedResponse) bool {
		ghExpectedArgs := []string{"pr", "edit", allCommits[1].Branch, "--add-reviewer", "mybestie"}
		return next.ProgramName == "gh" && slices.Equal(next.Args, ghExpectedArgs)
	})
	assert.True(contains, sliceutil.FilterSlice(testExecutor.Responses, func(next ex.ExecutedResponse) bool {
		return next.ProgramName == "gh"
	}))
}
