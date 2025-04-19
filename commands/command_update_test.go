package commands

import (
	"errors"
	"log/slog"

	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"

	"slices"

	"github.com/joshallenit/gh-stacked-diff/v2/interactive"
	"github.com/joshallenit/gh-stacked-diff/v2/templates"
	"github.com/joshallenit/gh-stacked-diff/v2/testutil"
	"github.com/joshallenit/gh-stacked-diff/v2/util"
)

func TestSdUpdate_OnRootCommit_UpdatesPr(t *testing.T) {
	assert := assert.New(t)
	testutil.InitTest(t, slog.LevelError)
	testutil.AddCommit("first", "")

	testParseArguments("new", "1")

	testutil.AddCommit("second", "")

	commitsOnMain := templates.GetAllCommits()

	testParseArguments("update", "2", "1")

	util.ExecuteOrDie(util.ExecuteOptions{}, "git", "switch", commitsOnMain[1].Branch)

	commitsOnBranch := templates.GetAllCommits()

	assert.Equal(3, len(commitsOnBranch))
}

func TestSdUpdate_OnExistingRoot_UpdatesPr(t *testing.T) {
	assert := assert.New(t)
	testutil.InitTest(t, slog.LevelError)

	testutil.AddCommit("first", "")
	util.ExecuteOrDie(util.ExecuteOptions{}, "git", "push", "origin", util.GetMainBranchOrDie())

	testutil.AddCommit("second", "")

	testParseArguments("new", "1")

	testutil.AddCommit("third", "")

	testutil.AddCommit("fourth", "")

	commitsOnMain := templates.GetAllCommits()

	testParseArguments("update", "3", "1")

	util.ExecuteOrDie(util.ExecuteOptions{}, "git", "switch", commitsOnMain[2].Branch)

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
	testutil.InitTest(t, slog.LevelError)

	testutil.AddCommit("first", "")

	testParseArguments("new", "1")

	testutil.AddCommit("second", "")

	allCommits := templates.GetAllCommits()

	testParseArguments("update", allCommits[1].Commit, "1")

	allCommits = templates.GetAllCommits()

	assert.Equal(2, len(allCommits))
	assert.Equal("first", allCommits[0].Subject)
	assert.Equal(testutil.InitialCommitSubject, allCommits[1].Subject)
}

func TestSdUpdate_WithListIndicators_UpdatesPr(t *testing.T) {
	assert := assert.New(t)
	testutil.InitTest(t, slog.LevelError)

	testutil.AddCommit("first", "")

	testParseArguments("new", "1")

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

	testExecutor := testutil.InitTest(t, slog.LevelError)

	testutil.AddCommit("first", "")

	testParseArguments("new", "1")

	testutil.AddCommit("second", "")

	allCommits := templates.GetAllCommits()

	testExecutor.SetResponse(
		// There has to be at least 4 checks, each with 3 values: status, conclusion, and state.
		"SUCCESS\nSUCCESS\nSUCCESS\n"+
			"SUCCESS\nSUCCESS\nSUCCESS\n"+
			"SUCCESS\nSUCCESS\nSUCCESS\n"+
			"SUCCESS\nSUCCESS\nSUCCESS\n",
		nil, "gh", "pr", "view", util.MatchAnyRemainingArgs)

	testParseArguments("update", "--reviewers=mybestie", "2", "1")

	contains := slices.ContainsFunc(testExecutor.Responses, func(next util.ExecutedResponse) bool {
		ghExpectedArgs := []string{"pr", "edit", allCommits[1].Branch, "--add-reviewer", "mybestie"}
		return next.ProgramName == "gh" && slices.Equal(next.Args, ghExpectedArgs)
	})
	assert.True(contains, util.FilterSlice(testExecutor.Responses, func(next util.ExecutedResponse) bool {
		return next.ProgramName == "gh"
	}))
}

func TestSdUpdate_WhenCherryPickFails_RestoresBranch(t *testing.T) {
	assert := assert.New(t)
	testutil.InitTest(t, slog.LevelError)

	testutil.AddCommit("first", "")
	testParseArguments("new", "1")
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

	testParseArguments("update", "3", "1")

	assert.Fail("did not panic on cherry-pick")
}

func TestSdUpdate_WhenPushFails_RestoresBranches(t *testing.T) {
	assert := assert.New(t)

	testExecutor := testutil.InitTest(t, slog.LevelError)

	testutil.AddCommit("first", "")
	firstBranch := templates.GetAllCommits()[0].Branch

	testParseArguments("new", "1")

	util.ExecuteOrDie(util.ExecuteOptions{}, "git", "switch", firstBranch)
	firstCommits := templates.GetAllCommits()
	util.ExecuteOrDie(util.ExecuteOptions{}, "git", "switch", util.GetMainBranchOrDie())

	testutil.AddCommit("second", "")

	allCommits := templates.GetAllCommits()
	testExecutor.SetResponse("", errors.New("Exit code 128"), "git", "push", util.MatchAnyRemainingArgs)
	defer func() {
		r := recover()
		if r != nil {
			assert.Equal(util.GetMainBranchOrDie(), util.GetCurrentBranchName())
			assert.Equal(allCommits, templates.GetAllCommits())

			util.ExecuteOrDie(util.ExecuteOptions{}, "git", "switch", firstBranch)
			assert.Equal(firstCommits, templates.GetAllCommits())
		}
	}()
	testParseArguments("update", "2", "1")

	assert.Fail("did not panic on cherry-pick")
}

func TestSdUpdate_WhenCherryPickCommitsNotSpecifiedAndUserQuits_NoOp(t *testing.T) {
	assert := assert.New(t)
	testutil.InitTest(t, slog.LevelError)
	testutil.AddCommit("first", "")

	testParseArguments("new", "1")

	testutil.AddCommit("second", "")

	commitsOnMain := templates.GetAllCommits()

	defer func() {
		r := recover()
		if r != nil {
			assert.Equal(commitsOnMain, templates.GetAllCommits())
		}
	}()

	interactive.SendToProgram(t, 0, interactive.NewMessageRune('q'))
	testParseArguments("update", "2")

	assert.Fail("did not panic on quit")
}

func TestSdUpdate_WhenCherryPickCommitsNotSpecified_CherryPicsUserSelection(t *testing.T) {
	assert := assert.New(t)
	testutil.InitTest(t, slog.LevelError)
	testutil.AddCommit("first", "")

	testParseArguments("new", "1")

	testutil.AddCommit("second", "")

	interactive.SendToProgram(t, 0, interactive.NewMessageKey(tea.KeyEnter))
	testParseArguments("update", "2")

	allCommits := templates.GetAllCommits()

	assert.Equal(2, len(allCommits))
	assert.Equal("first", allCommits[0].Subject)
	assert.Equal(testutil.InitialCommitSubject, allCommits[1].Subject)
}

func TestSdUpdate_WhenDestinationCommitNotSpecified_UpdatesSelectedPr(t *testing.T) {
	assert := assert.New(t)
	testutil.InitTest(t, slog.LevelError)
	testutil.AddCommit("first", "")

	testParseArguments("new", "1")

	testutil.AddCommit("second", "")

	interactive.SendToProgram(t, 0, interactive.NewMessageKey(tea.KeyEnter))
	interactive.SendToProgram(t, 1, interactive.NewMessageKey(tea.KeyEnter))
	testParseArguments("update")

	allCommits := templates.GetAllCommits()

	assert.Equal(2, len(allCommits))
	assert.Equal("first", allCommits[0].Subject)
	assert.Equal(testutil.InitialCommitSubject, allCommits[1].Subject)
}

func TestSdUpdate_WhenDestinationCommitNotSpecifiedAndMultiplePossibleValues_UpdatesSelectedPr(t *testing.T) {
	assert := assert.New(t)
	testutil.InitTest(t, slog.LevelError)
	testutil.AddCommit("first", "destination")
	testParseArguments("new", "1")
	testutil.AddCommit("second", "")
	testParseArguments("new", "1")
	testutil.AddCommit("third", "")
	testutil.AddCommit("fourth", "added1")
	testutil.AddCommit("fifth", "added2")
	testutil.AddCommit("sixth", "")

	interactive.SendToProgram(t, 0,
		interactive.NewMessageKey(tea.KeyDown),
		interactive.NewMessageKey(tea.KeyEnter),
	)
	interactive.SendToProgram(t, 1,
		interactive.NewMessageKey(tea.KeyDown),
		interactive.NewMessageRune(' '),
		interactive.NewMessageKey(tea.KeyDown),
		interactive.NewMessageKey(tea.KeyEnter),
	)
	testParseArguments("update")

	allCommits := templates.GetAllCommits()

	assert.Equal(5, len(allCommits))
	assert.Equal("sixth", allCommits[0].Subject)
	assert.Equal("third", allCommits[1].Subject)
	assert.Equal("second", allCommits[2].Subject)
	assert.Equal("first", allCommits[3].Subject)
	assert.Equal(testutil.InitialCommitSubject, allCommits[4].Subject)
	assert.True(util.RemoteHasBranch(allCommits[3].Branch))
}

func TestSdUpdate_WhenBranchAlreadyMergedAndUserDoesNotConfirm_Cancels(t *testing.T) {
	assert := assert.New(t)
	testExecutor := testutil.InitTest(t, slog.LevelError)
	testutil.AddCommit("first", "")

	testParseArguments("new", "1")

	testutil.AddCommit("second", "")

	allCommits := templates.GetAllCommits()

	interactive.SendToProgram(t, 0, interactive.NewMessageRune('n'))
	testExecutor.SetResponse(allCommits[1].Branch+" fakeMergeCommit",
		nil, "gh", "pr", "list", util.MatchAnyRemainingArgs)

	defer func() {
		r := recover()
		if r != nil {
			assert.Equal(allCommits, templates.GetAllCommits())
		}
	}()

	testParseArguments("update", "2", "1")

	assert.Fail("did not cancel")
}

func TestSdUpdate_WhenBranchAlreadyMergedAndUserConfirms_Updates(t *testing.T) {
	assert := assert.New(t)
	testExecutor := testutil.InitTest(t, slog.LevelError)
	testutil.AddCommit("first", "")

	testParseArguments("new", "1")

	testutil.AddCommit("second", "")

	allCommits := templates.GetAllCommits()
	testExecutor.SetResponse(allCommits[1].Branch+" fakeMergeCommit",
		nil, "gh", "pr", "list", util.MatchAnyRemainingArgs)

	interactive.SendToProgram(t, 0, interactive.NewMessageRune('y'))
	testParseArguments("update", "2", "1")

	assert.Equal(2, len(templates.GetAllCommits()))
}
