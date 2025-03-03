package commands

import (
	"errors"
	"log/slog"

	"testing"

	"github.com/stretchr/testify/assert"

	"slices"

	ex "github.com/joshallenit/gh-testsd3/v2/execute"
	"github.com/joshallenit/gh-testsd3/v2/templates"
	"github.com/joshallenit/gh-testsd3/v2/testutil"
	"github.com/joshallenit/gh-testsd3/v2/util"
)

func Test_UpdatePr_OnRootCommit_UpdatesPr(t *testing.T) {
	assert := assert.New(t)
	testutil.InitTest(slog.LevelInfo)
	testutil.AddCommit("first", "")

	testParseArguments("new")

	testutil.AddCommit("second", "")

	commitsOnMain := templates.GetAllCommits()

	testParseArguments("update", "2")

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

	testParseArguments("new")

	testutil.AddCommit("third", "")

	testutil.AddCommit("fourth", "")

	commitsOnMain := templates.GetAllCommits()

	testParseArguments("update", "3")

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

func TestSdUpdate_WhenCherryPickFails_RestoresBranch(t *testing.T) {
	assert := assert.New(t)

	testutil.InitTest(slog.LevelInfo)

	testutil.AddCommit("first", "")
	testParseArguments("new")
	testutil.CommitFileChange("second", "first", "made change")
	testutil.CommitFileChange("third", "first", "another change")

	allCommits := templates.GetAllCommits()
	defer func() {
		r := recover()
		if r != nil {
			assert.Equal(util.GetMainBranchOrDie(), util.GetCurrentBranchName())
			assert.Equal(allCommits, templates.GetAllCommits())
		}
	}()

	testParseArguments("update", "3")

	assert.Fail("did not panic on cherry-pick")
}

func TestSdUpdate_WhenPushFails_RestoresBranches(t *testing.T) {
	assert := assert.New(t)

	testExecutor := testutil.InitTest(slog.LevelInfo)

	testutil.AddCommit("first", "")
	firstBranch := templates.GetAllCommits()[0].Branch

	testParseArguments("new")

	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "switch", firstBranch)
	firstCommits := templates.GetAllCommits()
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "switch", util.GetMainBranchOrDie())

	testutil.AddCommit("second", "")

	allCommits := templates.GetAllCommits()
	testExecutor.SetResponse("", errors.New("Exit code 128"), "git", "push", ex.MatchAnyRemainingArgs)
	defer func() {
		r := recover()
		if r != nil {
			assert.Equal(util.GetMainBranchOrDie(), util.GetCurrentBranchName())
			assert.Equal(allCommits, templates.GetAllCommits())

			ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "switch", firstBranch)
			assert.Equal(firstCommits, templates.GetAllCommits())
		}
	}()
	testParseArguments("update", "2")

	assert.Fail("did not panic on cherry-pick")
}
