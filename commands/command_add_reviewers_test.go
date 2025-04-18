package commands

import (
	"log/slog"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/joshallenit/gh-stacked-diff/v2/interactive"
	"github.com/joshallenit/gh-stacked-diff/v2/templates"
	"github.com/joshallenit/gh-stacked-diff/v2/testutil"
	"github.com/joshallenit/gh-stacked-diff/v2/util"
)

func TestSdAddReviewers_AddReviewers(t *testing.T) {
	assert := assert.New(t)

	testExecutor := testutil.InitTest(t, slog.LevelInfo)

	testutil.AddCommit("first", "")

	testParseArguments("new", "1")

	allCommits := templates.GetAllCommits()
	testExecutor.SetResponse(
		// There has to be at least 4 checks, each with 3 values: status, conclusion, and state.
		"SUCCESS\nSUCCESS\nSUCCESS\n"+
			"SUCCESS\nSUCCESS\nSUCCESS\n"+
			"SUCCESS\nSUCCESS\nSUCCESS\n"+
			"SUCCESS\nSUCCESS\nSUCCESS\n",
		nil, "gh", "pr", "view", util.MatchAnyRemainingArgs)

	testParseArguments("add-reviewers", "--reviewers=mybestie", allCommits[0].Commit)

	contains := slices.ContainsFunc(testExecutor.Responses, func(next util.ExecutedResponse) bool {
		ghExpectedArgs := []string{"pr", "edit", allCommits[0].Branch, "--add-reviewer", "mybestie"}
		return next.ProgramName == "gh" && slices.Equal(next.Args, ghExpectedArgs)
	})
	assert.True(contains, util.FilterSlice(testExecutor.Responses, func(next util.ExecutedResponse) bool {
		return next.ProgramName == "gh"
	}))
}

func TestSdAddReviewers_WhenUsingListIndicator_AddReviewers(t *testing.T) {
	assert := assert.New(t)

	testExecutor := testutil.InitTest(t, slog.LevelInfo)

	testutil.AddCommit("first", "")

	testParseArguments("new", "1")

	allCommits := templates.GetAllCommits()
	testExecutor.SetResponse(
		// There has to be at least 4 checks, each with 3 values: status, conclusion, and state.
		"SUCCESS\nSUCCESS\nSUCCESS\n"+
			"SUCCESS\nSUCCESS\nSUCCESS\n"+
			"SUCCESS\nSUCCESS\nSUCCESS\n"+
			"SUCCESS\nSUCCESS\nSUCCESS\n",
		nil, "gh", "pr", "view", util.MatchAnyRemainingArgs)

	testParseArguments("add-reviewers", "--indicator=list", "--reviewers=mybestie", "1")

	contains := slices.ContainsFunc(testExecutor.Responses, func(next util.ExecutedResponse) bool {
		ghExpectedArgs := []string{"pr", "edit", allCommits[0].Branch, "--add-reviewer", "mybestie"}
		return next.ProgramName == "gh" && slices.Equal(next.Args, ghExpectedArgs)
	})
	assert.True(contains, util.FilterSlice(testExecutor.Responses, func(next util.ExecutedResponse) bool {
		return next.ProgramName == "gh"
	}))
}

func TestSdAddReviewers_WhenOmittingCommitIndicator_AsksForSelection(t *testing.T) {
	assert := assert.New(t)

	testExecutor := testutil.InitTest(t, slog.LevelInfo)

	testutil.AddCommit("first", "")

	testParseArguments("new", "1")

	allCommits := templates.GetAllCommits()
	testExecutor.SetResponse(
		// There has to be at least 4 checks, each with 3 values: status, conclusion, and state.
		"SUCCESS\nSUCCESS\nSUCCESS\n"+
			"SUCCESS\nSUCCESS\nSUCCESS\n"+
			"SUCCESS\nSUCCESS\nSUCCESS\n"+
			"SUCCESS\nSUCCESS\nSUCCESS\n",
		nil, "gh", "pr", "view", util.MatchAnyRemainingArgs)

	interactive.SendToProgram(t, 0,
		interactive.NewMessageKey(tea.KeyEnter),
	)
	testParseArguments("add-reviewers", "--indicator=list", "--reviewers=mybestie")

	contains := slices.ContainsFunc(testExecutor.Responses, func(next util.ExecutedResponse) bool {
		ghExpectedArgs := []string{"pr", "edit", allCommits[0].Branch, "--add-reviewer", "mybestie"}
		return next.ProgramName == "gh" && slices.Equal(next.Args, ghExpectedArgs)
	})
	assert.True(contains, util.FilterSlice(testExecutor.Responses, func(next util.ExecutedResponse) bool {
		return next.ProgramName == "gh"
	}))
}
