package commands

import (
	"log/slog"
	sd "stackeddiff"
	"testing"

	"github.com/stretchr/testify/assert"

	"slices"

	ex "github.com/joshallenit/stacked-diff/execute"
	"github.com/joshallenit/stacked-diff/util"
)

func TestSdUpdate_UpdatesPr(t *testing.T) {
	assert := assert.New(t)

	testutil.InitTest(slog.LevelInfo)

	testutil.AddCommit("first", "")

	testParseArguments("new")

	testutil.AddCommit("second", "")

	allCommits := sd.GetAllCommits()

	testParseArguments("update", allCommits[1].Commit)

	allCommits = sd.GetAllCommits()

	assert.Equal(2, len(allCommits))
	assert.Equal("first", allCommits[0].Subject)
	assert.Equal(testutil.InitialCommitSubject, allCommits[1].Subject)
}

func TestSdUpdate_WithListIndicators_UpdatesPr(t *testing.T) {
	assert := assert.New(t)

	testutil.InitTest(slog.LevelInfo)

	testutil.AddCommit("first", "")

	testParseArguments("new")

	testutil.AddCommit("second", "")
	testutil.AddCommit("third", "")

	testParseArguments("update", "3", "2", "1")

	allCommits := sd.GetAllCommits()

	assert.Equal(2, len(allCommits))
	assert.Equal("first", allCommits[0].Subject)
	assert.Equal(testutil.InitialCommitSubject, allCommits[1].Subject)

	ex.Execute(ex.ExecuteOptions{}, "sd", "checkout", "1")
	allCommits = sd.GetAllCommits()

	assert.Equal(4, len(allCommits))
	assert.Equal("third", allCommits[0].Subject)
	assert.Equal("second", allCommits[1].Subject)
	assert.Equal("first", allCommits[2].Subject)
	assert.Equal(testutil.InitialCommitSubject, allCommits[3].Subject)
}

func TestSdUpdate_WithReviewers_AddReviewers(t *testing.T) {
	assert := assert.New(t)

	testExecutor := testutil.InitTest(slog.LevelInfo)

	testutil.AddCommit("first", "")

	testParseArguments("new")

	testutil.AddCommit("second", "")

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
	assert.True(contains, util.FilterSlice(testExecutor.Responses, func(next ex.ExecutedResponse) bool {
		return next.ProgramName == "gh"
	}))
}
