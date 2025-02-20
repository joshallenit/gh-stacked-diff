package commands

import (
	"log/slog"
	"slices"
	sd "stackeddiff"
	"testing"

	"github.com/stretchr/testify/assert"

	"errors"
	"stackeddiff/testinginit"

	ex "github.com/joshallenit/stacked-diff/execute"
	"github.com/joshallenit/stacked-diff/util"
)

func TestSdNew_CreatesPr(t *testing.T) {
	assert := assert.New(t)

	testinginit.InitTest(slog.LevelInfo)

	testinginit.AddCommit("first", "")

	testParseArguments("new")

	out := testParseArguments("log")

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

	testParseArguments("new", "--reviewers=mybestie")

	allCommits := sd.GetAllCommits()

	contains := slices.ContainsFunc(testExecutor.Responses, func(next ex.ExecutedResponse) bool {
		ghExpectedArgs := []string{"pr", "edit", allCommits[0].Branch, "--add-reviewer", "mybestie"}
		return next.ProgramName == "gh" && slices.Equal(next.Args, ghExpectedArgs)
	})
	assert.True(contains, util.FilterSlice(testExecutor.Responses, func(next ex.ExecutedResponse) bool {
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

	testParseArguments("new", "2")

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

	out := testParseArguments("new")

	assert.Contains(out, "Use \"--draft=false\" to avoid this warning")
	assert.Contains(out, "Created PR ")
}

func TestSdNew_WhenTwoPrsOnRoot_CreatesFromRoot(t *testing.T) {
	assert := assert.New(t)

	testinginit.InitTest(slog.LevelInfo)

	testinginit.AddCommit("first", "")
	testinginit.AddCommit("second", "")

	testParseArguments("new", "2")
	testParseArguments("new")

	mainCommits := sd.GetAllCommits()

	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "switch", mainCommits[1].Branch)
	firstCommits := sd.GetAllCommits()
	assert.Equal(2, len(firstCommits))
	assert.Equal(mainCommits[1].Subject, firstCommits[0].Subject)
	assert.Equal(testinginit.InitialCommitSubject, firstCommits[1].Subject)

	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "switch", mainCommits[0].Branch)
	secondCommits := sd.GetAllCommits()
	assert.Equal(2, len(secondCommits))
	assert.Equal(mainCommits[0].Subject, secondCommits[0].Subject)
	assert.Equal(testinginit.InitialCommitSubject, firstCommits[1].Subject)
}
