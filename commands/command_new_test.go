package commands

import (
	"log/slog"
	"slices"

	"testing"

	"github.com/stretchr/testify/assert"

	"errors"

	"github.com/joshallenit/gh-stacked-diff/v2/interactive"
	"github.com/joshallenit/gh-stacked-diff/v2/templates"
	"github.com/joshallenit/gh-stacked-diff/v2/testutil"
	"github.com/joshallenit/gh-stacked-diff/v2/util"

	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func Test_NewPr_OnNewRepo_CreatesPr(t *testing.T) {
	testutil.InitTest(t, slog.LevelInfo)

	testutil.AddCommit("first", "")

	testParseArguments("new", "1")

	// Check that the PR was created
	outWriter := testutil.NewWriteRecorder()
	printGitLog(outWriter)
	out := outWriter.String()

	if !strings.Contains(out, "✅") {
		t.Errorf("'✅' should be in %s", out)
	}
}

func Test_NewPr_OnRepoWithPreviousCommit_CreatesPr(t *testing.T) {
	assert := assert.New(t)

	testutil.InitTest(t, slog.LevelInfo)

	testutil.AddCommit("first", "")
	util.ExecuteOrDie(util.ExecuteOptions{}, "git", "push", "origin", util.GetMainBranchOrDie())

	testutil.AddCommit("second", "")
	allCommits := templates.GetNewCommits("HEAD")

	testParseArguments("new", "1")

	util.ExecuteOrDie(util.ExecuteOptions{}, "git", "switch", allCommits[0].Branch)
	commitsOnNewBranch := templates.GetNewCommits("HEAD")
	assert.Equal(1, len(commitsOnNewBranch))
	assert.Equal(allCommits[0].Subject, commitsOnNewBranch[0].Subject)
}

func Test_NewPr_WithMiddleCommit_CreatesPr(t *testing.T) {
	assert := assert.New(t)

	testutil.InitTest(t, slog.LevelInfo)

	testutil.AddCommit("first", "")
	util.ExecuteOrDie(util.ExecuteOptions{}, "git", "push", "origin", util.GetMainBranchOrDie())

	testutil.AddCommit("second", "")

	testutil.AddCommit("third", "")
	allCommits := templates.GetNewCommits("HEAD")

	testParseArguments("new", "1")

	util.ExecuteOrDie(util.ExecuteOptions{}, "git", "switch", allCommits[0].Branch)
	commitsOnNewBranch := templates.GetNewCommits("HEAD")
	assert.Equal(1, len(commitsOnNewBranch))
	assert.Equal(allCommits[0].Subject, commitsOnNewBranch[0].Subject)
}

func TestSdNew_CreatesPr(t *testing.T) {
	assert := assert.New(t)

	testutil.InitTest(t, slog.LevelInfo)

	testutil.AddCommit("first", "")

	testParseArguments("new", "1")

	out := testParseArguments("log")

	assert.Contains(out, "✅")
}

func TestSdNew_WithReviewers_AddReviewers(t *testing.T) {
	assert := assert.New(t)

	testExecutor := testutil.InitTest(t, slog.LevelInfo)

	testutil.AddCommit("first", "")

	testExecutor.SetResponse(
		// There has to be at least 4 checks, each with 3 values: status, conclusion, and state.
		"SUCCESS\nSUCCESS\nSUCCESS\n"+
			"SUCCESS\nSUCCESS\nSUCCESS\n"+
			"SUCCESS\nSUCCESS\nSUCCESS\n"+
			"SUCCESS\nSUCCESS\nSUCCESS\n",
		nil, "gh", "pr", "view", util.MatchAnyRemainingArgs)

	testParseArguments("new", "--reviewers=mybestie", "1")

	allCommits := templates.GetAllCommits()

	contains := slices.ContainsFunc(testExecutor.Responses, func(next util.ExecutedResponse) bool {
		ghExpectedArgs := []string{"pr", "edit", allCommits[0].Branch, "--add-reviewer", "mybestie"}
		return next.ProgramName == "gh" && slices.Equal(next.Args, ghExpectedArgs)
	})
	assert.True(contains, util.FilterSlice(testExecutor.Responses, func(next util.ExecutedResponse) bool {
		return next.ProgramName == "gh"
	}))
}

func TestSdNew_WhenUsingListIndex_UsesCorrectList(t *testing.T) {
	assert := assert.New(t)

	testutil.InitTest(t, slog.LevelInfo)

	testutil.AddCommit("first", "")
	testutil.AddCommit("second", "")
	testutil.AddCommit("third", "")
	testutil.AddCommit("fourth", "")

	allCommits := templates.GetAllCommits()

	testParseArguments("new", "2")

	assert.Equal(true, util.RemoteHasBranch(allCommits[1].Branch))
}

func TestSdNew_WhenDraftNotSupported_TriesAgainWithoutDraft(t *testing.T) {
	assert := assert.New(t)

	testExecutor := testutil.InitTest(t, slog.LevelInfo)

	testutil.AddCommit("first", "")

	draftNotSupported := "pull request create failed: GraphQL: Draft pull requests are not supported in this repository. (createPullRequest)"
	testExecutor.SetResponseFunc(draftNotSupported, errors.New("Exit code 1"), func(programName string, args ...string) bool {
		return programName == "gh" && args[0] == "pr" && args[1] == "create" && slices.Contains(args, "--draft")
	})

	out := testParseArguments("new", "1")

	assert.Contains(out, "Use \"--draft=false\" to avoid this warning")
	assert.Contains(out, "Created PR ")
}

func TestSdNew_WhenTwoPrsOnRoot_CreatesFromRoot(t *testing.T) {
	assert := assert.New(t)

	testutil.InitTest(t, slog.LevelInfo)

	testutil.AddCommit("first", "")
	testutil.AddCommit("second", "")

	testParseArguments("new", "2")
	testParseArguments("new", "1")

	mainCommits := templates.GetAllCommits()

	util.ExecuteOrDie(util.ExecuteOptions{}, "git", "switch", mainCommits[1].Branch)
	firstCommits := templates.GetAllCommits()
	assert.Equal(2, len(firstCommits))
	assert.Equal(mainCommits[1].Subject, firstCommits[0].Subject)
	assert.Equal(testutil.InitialCommitSubject, firstCommits[1].Subject)

	util.ExecuteOrDie(util.ExecuteOptions{}, "git", "switch", mainCommits[0].Branch)
	secondCommits := templates.GetAllCommits()
	assert.Equal(2, len(secondCommits))
	assert.Equal(mainCommits[0].Subject, secondCommits[0].Subject)
	assert.Equal(testutil.InitialCommitSubject, firstCommits[1].Subject)
}

func TestSdNew_WhenCherryPickFails_RestoresBranch(t *testing.T) {
	assert := assert.New(t)

	testutil.InitTest(t, slog.LevelInfo)

	testutil.AddCommit("first", "")

	testutil.CommitFileChange("second", "first", "changes")

	allCommits := templates.GetAllCommits()

	restoreBranch := util.GetCurrentBranchName()
	defer func() {
		r := recover()
		if r != nil {
			assert.Equal(restoreBranch, util.GetCurrentBranchName())
			assert.Equal(allCommits, templates.GetAllCommits())
		}
	}()

	testParseArguments("new", "1")

	assert.Fail("did not panic on conflicts with cherry-pick")
}

func TestSdNew_WhenNewPrFails_RestoresBranch(t *testing.T) {
	assert := assert.New(t)

	testExecutor := testutil.InitTest(t, slog.LevelInfo)

	testutil.AddCommit("first", "")

	allCommits := templates.GetAllCommits()

	testExecutor.SetResponse("", errors.New("Exit Code 1"), "gh", "pr", "create", util.MatchAnyRemainingArgs)

	restoreBranch := util.GetCurrentBranchName()
	defer func() {
		r := recover()
		if r != nil {
			assert.Equal(restoreBranch, util.GetCurrentBranchName())
			assert.Equal(allCommits, templates.GetAllCommits())
		}
	}()

	testParseArguments("new", "1")

	assert.Fail("did not panic on PR create")
}

func TestSdNew_WhenDestinationCommitNotSpecified_CreatesPrWithSelectedCommit(t *testing.T) {
	assert := assert.New(t)
	testutil.InitTest(t, slog.LevelInfo)
	testutil.AddCommit("first", "")
	testutil.AddCommit("second", "")

	interactive.SendToProgram(t, 0,
		interactive.NewMessageKey(tea.KeyDown),
		interactive.NewMessageKey(tea.KeyEnter),
	)
	testParseArguments("new")

	allCommits := templates.GetAllCommits()

	assert.True(util.RemoteHasBranch(allCommits[1].Branch))
}

func TestSdNew_WhenDestinationCommitNotSpecified_WrapsCursorUp(t *testing.T) {
	assert := assert.New(t)
	testutil.InitTest(t, slog.LevelInfo)
	testutil.AddCommit("first", "")
	testutil.AddCommit("second", "")
	testutil.AddCommit("third", "")

	interactive.SendToProgram(t, 0,
		interactive.NewMessageKey(tea.KeyUp),
		interactive.NewMessageKey(tea.KeyEnter),
	)
	testParseArguments("new")

	allCommits := templates.GetAllCommits()

	assert.True(util.RemoteHasBranch(allCommits[2].Branch))
}

func TestSdNew_WhenDestinationCommitNotSpecified_WrapsCursorDown(t *testing.T) {
	assert := assert.New(t)
	testutil.InitTest(t, slog.LevelInfo)
	testutil.AddCommit("first", "")
	testutil.AddCommit("second", "")
	testutil.AddCommit("third", "")

	interactive.SendToProgram(t, 0,
		interactive.NewMessageKey(tea.KeyDown),
		interactive.NewMessageKey(tea.KeyDown),
		interactive.NewMessageKey(tea.KeyDown),
		interactive.NewMessageKey(tea.KeyEnter),
	)
	testParseArguments("new")

	allCommits := templates.GetAllCommits()

	assert.True(util.RemoteHasBranch(allCommits[0].Branch))
}

func TestSdNew_WhenDestinationCommitNotSpecifiedAndManyCommits_PadsIndex(t *testing.T) {
	assert := assert.New(t)
	testutil.InitTest(t, slog.LevelInfo)
	testutil.AddCommit("first", "")
	testutil.AddCommit("second", "")
	testutil.AddCommit("third", "")
	testutil.AddCommit("fourth", "")
	testutil.AddCommit("fifth", "")
	testutil.AddCommit("sixth", "")
	testutil.AddCommit("seventh", "")
	testutil.AddCommit("eighth", "")
	testutil.AddCommit("ninth", "")
	testutil.AddCommit("tenth", "")

	interactive.SendToProgram(t, 0,
		interactive.NewMessageRune('q'),
	)
	out := testutil.NewWriteRecorder()
	defer func() {
		r := recover()
		if r != nil {
			assert.Contains(out.String(), "│ 1│")
		}
	}()
	testParseArgumentsWithOut(out, "--log-level=error", "new")
	assert.Fail("did not panic on cancel")
}

func TestSdNew_WhenDestinationCommitNotSpecifiedAndManyCommitsAndExistingPr_PadsIndex(t *testing.T) {
	assert := assert.New(t)
	testutil.InitTest(t, slog.LevelInfo)
	testutil.AddCommit("first", "")
	testParseArguments("new", "1")
	testutil.AddCommit("second", "")
	testutil.AddCommit("third", "")
	testutil.AddCommit("fourth", "")
	testutil.AddCommit("fifth", "")
	testutil.AddCommit("sixth", "")
	testutil.AddCommit("seventh", "")
	testutil.AddCommit("eighth", "")
	testutil.AddCommit("ninth", "")
	testutil.AddCommit("tenth", "")

	interactive.SendToProgram(t, 0,
		interactive.NewMessageRune('q'),
	)
	out := testutil.NewWriteRecorder()
	defer func() {
		r := recover()
		if r != nil {
			assert.Contains(out.String(), "│ 1    │")
			assert.Contains(out.String(), "│10 ✅ │")
		}
	}()
	testParseArgumentsWithOut(out, "--log-level=error", "new")
	assert.Fail("did not panic on cancel")
}
