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

	testExecutor := testutil.InitTest(t, slog.LevelError)

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

	testParseArguments("add-reviewers", "--min-checks", "4", "--reviewers=mybestie", allCommits[0].Commit)

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

	testExecutor := testutil.InitTest(t, slog.LevelError)

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

	testParseArguments("add-reviewers", "--min-checks", "4", "--indicator=list", "--reviewers=mybestie", "1")

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

	testExecutor := testutil.InitTest(t, slog.LevelError)

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

	interactive.SendToProgram(0, interactive.NewMessageKey(tea.KeyEnter))
	testParseArguments("add-reviewers", "--min-checks", "4", "--indicator=list", "--reviewers=mybestie")

	contains := slices.ContainsFunc(testExecutor.Responses, func(next util.ExecutedResponse) bool {
		ghExpectedArgs := []string{"pr", "edit", allCommits[0].Branch, "--add-reviewer", "mybestie"}
		return next.ProgramName == "gh" && slices.Equal(next.Args, ghExpectedArgs)
	})
	assert.True(contains, util.FilterSlice(testExecutor.Responses, func(next util.ExecutedResponse) bool {
		return next.ProgramName == "gh"
	}))
}

func TestSdAddReviewers_WhenUserAlreadyApproved_DoesNotRequestReview(t *testing.T) {
	assert := assert.New(t)

	testExecutor := testutil.InitTest(t, slog.LevelError)

	testutil.AddCommit("first", "")

	testParseArguments("new", "1")

	allCommits := templates.GetAllCommits()
	checksSuccess := // There has to be at least 4 checks, each with 3 values: status, conclusion, and state.
		"SUCCESS\nSUCCESS\nSUCCESS\n" +
			"SUCCESS\nSUCCESS\nSUCCESS\n" +
			"SUCCESS\nSUCCESS\nSUCCESS\n" +
			"SUCCESS\nSUCCESS\nSUCCESS\n"
	testExecutor.SetResponseFunc(checksSuccess, nil, func(programName string, args ...string) bool {
		return programName == "gh" &&
			args[0] == "pr" &&
			args[1] == "view" &&
			slices.Contains(args, "statusCheckRollup")
	})

	approvedUsers := "alreadyapproved1\nalreadyapproved2"
	testExecutor.SetResponseFunc(approvedUsers, nil, func(programName string, args ...string) bool {
		return programName == "gh" &&
			args[0] == "pr" &&
			args[1] == "view" &&
			slices.Contains(args, "reviews")
	})

	out := testParseArguments("--log-level=info", "add-reviewers", "--min-checks", "4", "--reviewers=alreadyapproved2,mybestie,alreadyapproved1", "1")

	contains := slices.ContainsFunc(testExecutor.Responses, func(next util.ExecutedResponse) bool {
		ghExpectedArgs := []string{"pr", "edit", allCommits[0].Branch, "--add-reviewer", "mybestie"}
		return next.ProgramName == "gh" && slices.Equal(next.Args, ghExpectedArgs)
	})
	assert.True(contains, util.FilterSlice(testExecutor.Responses, func(next util.ExecutedResponse) bool {
		return next.ProgramName == "gh"
	}))

	assert.Contains(out, "alreadyapproved2,alreadyapproved1")
}

func TestSdAddReviewers_UserChoosesHistory_ChoosesSameReviewers(t *testing.T) {
	assert := assert.New(t)

	testExecutor := testutil.InitTest(t, slog.LevelError)

	testutil.AddCommit("first", "")

	testParseArguments("new", "1")

	allCommits := templates.GetAllCommits()
	checksSuccess := // There has to be at least 4 checks, each with 3 values: status, conclusion, and state.
		"SUCCESS\nSUCCESS\nSUCCESS\n" +
			"SUCCESS\nSUCCESS\nSUCCESS\n" +
			"SUCCESS\nSUCCESS\nSUCCESS\n" +
			"SUCCESS\nSUCCESS\nSUCCESS\n"
	testExecutor.SetResponseFunc(checksSuccess, nil, func(programName string, args ...string) bool {
		return programName == "gh" &&
			args[0] == "pr" &&
			args[1] == "view" &&
			slices.Contains(args, "statusCheckRollup")
	})

	testParseArguments("add-reviewers", "--min-checks", "4", "--reviewers=mybestie", "1")

	// Clear responses.
	testExecutor.Responses = []util.ExecutedResponse{}
	// What reviewers?
	interactive.SendToProgram(0,
		// History
		interactive.NewMessageKey(tea.KeyUp),
		interactive.NewMessageKey(tea.KeyEnter),
	)
	testParseArguments("add-reviewers", "--min-checks", "4", "1")

	contains := slices.ContainsFunc(testExecutor.Responses, func(next util.ExecutedResponse) bool {
		ghExpectedArgs := []string{"pr", "edit", allCommits[0].Branch, "--add-reviewer", "mybestie"}
		return next.ProgramName == "gh" && slices.Equal(next.Args, ghExpectedArgs)
	})
	assert.True(contains, util.FilterSlice(testExecutor.Responses, func(next util.ExecutedResponse) bool {
		return next.ProgramName == "gh"
	}))
}

func TestSdAddReviewers_UserChoosesHistoryFromTyped_ChoosesSameReviewers(t *testing.T) {
	assert := assert.New(t)

	testExecutor := testutil.InitTest(t, slog.LevelError)

	testutil.AddCommit("first", "")

	testParseArguments("new", "1")

	allCommits := templates.GetAllCommits()
	checksSuccess := // There has to be at least 4 checks, each with 3 values: status, conclusion, and state.
		"SUCCESS\nSUCCESS\nSUCCESS\n" +
			"SUCCESS\nSUCCESS\nSUCCESS\n" +
			"SUCCESS\nSUCCESS\nSUCCESS\n" +
			"SUCCESS\nSUCCESS\nSUCCESS\n"
	testExecutor.SetResponseFunc(checksSuccess, nil, func(programName string, args ...string) bool {
		return programName == "gh" &&
			args[0] == "pr" &&
			args[1] == "view" &&
			slices.Contains(args, "statusCheckRollup")
	})

	// What reviewers?
	interactive.SendToProgram(0,
		interactive.NewMessageRune('m'),
		interactive.NewMessageRune('y'),
		interactive.NewMessageKey(tea.KeyEnter),
	)
	testParseArguments("add-reviewers", "--min-checks", "4", "1")

	// Clear responses.
	testExecutor.Responses = []util.ExecutedResponse{}
	// What reviewers?
	interactive.SendToProgram(1,
		// History
		interactive.NewMessageKey(tea.KeyUp),
		interactive.NewMessageKey(tea.KeyEnter),
	)
	testParseArguments("add-reviewers", "--min-checks", "4", "1")

	contains := slices.ContainsFunc(testExecutor.Responses, func(next util.ExecutedResponse) bool {
		ghExpectedArgs := []string{"pr", "edit", allCommits[0].Branch, "--add-reviewer", "my"}
		return next.ProgramName == "gh" && slices.Equal(next.Args, ghExpectedArgs)
	})
	assert.True(contains, util.FilterSlice(testExecutor.Responses, func(next util.ExecutedResponse) bool {
		return next.ProgramName == "gh"
	}))
}
