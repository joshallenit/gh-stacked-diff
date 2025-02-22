package commands

import (
	"log/slog"

	"testing"

	"github.com/stretchr/testify/assert"

	"slices"

	ex "github.com/joshallenit/stacked-diff/v2/execute"
	"github.com/joshallenit/stacked-diff/v2/templates"
	"github.com/joshallenit/stacked-diff/v2/testutil"
	"github.com/joshallenit/stacked-diff/v2/util"
)

func Test_UpdatePr_OnRootCommit_UpdatesPr(t *testing.T) {
	assert := assert.New(t)
	testutil.InitTest(slog.LevelInfo)
	testutil.AddCommit("first", "")

	createNewPr(true, "", util.GetMainBranchOrDie(), templates.GetBranchInfo("", templates.IndicatorTypeGuess))

	testutil.AddCommit("second", "")

	commitsOnMain := templates.GetAllCommits()

	updatePr(templates.GetBranchInfo(commitsOnMain[1].Commit, templates.IndicatorTypeCommit), []string{}, templates.IndicatorTypeCommit)

	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "switch", commitsOnMain[1].Branch)

	commitsOnBranch := templates.GetAllCommits()

	assert.Equal(3, len(commitsOnBranch))
}

func Test_UpdatePr_OnExistingRoot_UpdatesPr(t *testing.T) {
	assert := assert.New(t)
	testutil.InitTest(slog.LevelInfo)

	testutil.AddCommit("first", "")
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "push", "origin", util.GetMainBranchOrDie())

	testutil.AddCommit("second", "")

	createNewPr(true, "", util.GetMainBranchOrDie(), templates.GetBranchInfo("", templates.IndicatorTypeGuess))

	testutil.AddCommit("third", "")

	testutil.AddCommit("fourth", "")

	commitsOnMain := templates.GetAllCommits()

	updatePr(templates.GetBranchInfo(commitsOnMain[2].Commit, templates.IndicatorTypeCommit), []string{}, templates.IndicatorTypeCommit)

	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "switch", commitsOnMain[2].Branch)

	allCommitsOnBranch := templates.GetAllCommits()

	assert.Equal(4, len(allCommitsOnBranch))
	assert.Equal("fourth", allCommitsOnBranch[0].Subject)
	assert.Equal("second", allCommitsOnBranch[1].Subject)
	assert.Equal("first", allCommitsOnBranch[2].Subject)
	assert.Equal(testutil.InitialCommitSubject, allCommitsOnBranch[3].Subject)

	newCommitsOnBranch := templates.GetNewCommits("HEAD")

	assert.Equal(2, len(newCommitsOnBranch))
	assert.Equal(newCommitsOnBranch[0].Subject, "fourth")
	assert.Equal(newCommitsOnBranch[1].Subject, "second")
}

func TestSdUpdate_UpdatesPr(t *testing.T) {
	assert := assert.New(t)

	testutil.InitTest(slog.LevelInfo)

	testutil.AddCommit("first", "")

	testParseArguments("new")

	testutil.AddCommit("second", "")

	allCommits := templates.GetAllCommits()

	testParseArguments("update", allCommits[1].Commit)

	allCommits = templates.GetAllCommits()

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

	allCommits := templates.GetAllCommits()

	assert.Equal(2, len(allCommits))
	assert.Equal("first", allCommits[0].Subject)
	assert.Equal(testutil.InitialCommitSubject, allCommits[1].Subject)

	testParseArguments("checkout", "1")
	allCommits = templates.GetAllCommits()

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

	allCommits := templates.GetAllCommits()

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
