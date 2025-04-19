package commands

import (
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/joshallenit/gh-stacked-diff/v2/templates"
	"github.com/joshallenit/gh-stacked-diff/v2/testutil"
	"github.com/joshallenit/gh-stacked-diff/v2/util"
)

func TestSdReplaceCommit_WithMultipleCommits_ReplacesCommitWithBranch(t *testing.T) {
	assert := assert.New(t)
	testutil.InitTest(t, slog.LevelError)

	testutil.AddCommit("first", "1")
	util.ExecuteOrDie(util.ExecuteOptions{}, "git", "push", "origin", util.GetMainBranchOrDie())
	testutil.AddCommit("second", "will-be-replaced")
	testParseArguments("new", "1")
	testutil.AddCommit("fifth", "5")

	allCommits := templates.GetAllCommits()

	testParseArguments("checkout", allCommits[1].Commit)

	util.ExecuteOrDie(util.ExecuteOptions{}, "git", "reset", "--hard", allCommits[2].Commit)
	testutil.AddCommit("on-second-branch-only", "2")
	testutil.AddCommit("on-second-branch-only", "3")
	testutil.AddCommit("on-second-branch-only", "4")

	util.ExecuteOrDie(util.ExecuteOptions{}, "git", "checkout", util.GetMainBranchOrDie())

	testParseArguments("replace-commit", allCommits[1].Commit)

	allCommits = templates.GetAllCommits()

	assert.Equal(4, len(allCommits))
	assert.Equal("fifth", allCommits[0].Subject)
	assert.Equal("second", allCommits[1].Subject)
	assert.Equal("first", allCommits[2].Subject)
	assert.Equal(testutil.InitialCommitSubject, allCommits[3].Subject)

	dirEntries, err := os.ReadDir(".")
	if err != nil {
		panic("Could not read dir: " + err.Error())
	}
	assert.Equal(6, len(dirEntries))
	assert.Equal(".git", dirEntries[0].Name())
	assert.Equal("1", dirEntries[1].Name())
	assert.Equal("2", dirEntries[2].Name())
	assert.Equal("3", dirEntries[3].Name())
	assert.Equal("4", dirEntries[4].Name())
	assert.Equal("5", dirEntries[5].Name())
}

func TestSdReplaceCommit_WhenPrPushed_ReplacesCommitWithBranch(t *testing.T) {
	if true {
		return
	}
	assert := assert.New(t)
	testutil.InitTest(t, slog.LevelError)

	testutil.AddCommit("first", "1")
	testutil.AddCommit("second", "will-be-replaced")
	testParseArguments("new", "1")
	testutil.AddCommit("fifth", "5")

	allCommits := templates.GetAllCommits()

	testParseArguments("checkout", allCommits[1].Commit)

	util.ExecuteOrDie(util.ExecuteOptions{}, "git", "reset", "--hard", allCommits[2].Commit)
	testutil.AddCommit("on-second-branch-only", "2")
	testutil.AddCommit("on-second-branch-only", "3")
	testutil.AddCommit("on-second-branch-only", "4")

	util.ExecuteOrDie(util.ExecuteOptions{}, "git", "push", "origin", allCommits[1].Branch)

	util.ExecuteOrDie(util.ExecuteOptions{}, "git", "checkout", util.GetMainBranchOrDie())

	testParseArguments("replace-commit", allCommits[1].Commit)

	allCommits = templates.GetAllCommits()

	assert.Equal(3, len(allCommits))
	assert.Equal("fifth", allCommits[0].Subject)
	assert.Equal("second", allCommits[1].Subject)
	assert.Equal("first", allCommits[2].Subject)

	dirEntries, err := os.ReadDir(".")
	if err != nil {
		panic("Could not read dir: " + err.Error())
	}
	assert.Equal(6, len(dirEntries))
	assert.Equal(".git", dirEntries[0].Name())
	assert.Equal("1", dirEntries[1].Name())
	assert.Equal("2", dirEntries[2].Name())
	assert.Equal("3", dirEntries[3].Name())
	assert.Equal("4", dirEntries[4].Name())
	assert.Equal("5", dirEntries[5].Name())
}
