package main

import (
	"flag"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	sd "stackeddiff"
	ex "stackeddiff/execute"
	"stackeddiff/testinginit"
)

func TestSdRebaseMain_WithDifferentCommits_DropsCommits(t *testing.T) {
	assert := assert.New(t)
	testExecutor := testinginit.InitTest(slog.LevelInfo)

	testinginit.AddCommit("first", "")
	testinginit.AddCommit("second", "rebase-will-keep-this-file")

	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "push", "origin", sd.GetMainBranchOrDie())

	allOriginalCommits := sd.GetAllCommits()

	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "reset", "--hard", allOriginalCommits[1].Commit)

	testinginit.AddCommit("second", "rebase-will-drop-this-file")

	testExecutor.SetResponse(allOriginalCommits[0].Branch, nil, "gh", "pr", "list", ex.MatchAnyRemainingArgs)

	parseArguments(os.Stdout, flag.NewFlagSet("sd", flag.ContinueOnError), []string{"rebase-main"})

	dirEntries, err := os.ReadDir(".")
	if err != nil {
		panic("Could not read dir: " + err.Error())
	}
	assert.Equal(3, len(dirEntries))
	assert.Equal(".git", dirEntries[0].Name())
	assert.Equal("first", dirEntries[1].Name())
	assert.Equal("rebase-will-keep-this-file", dirEntries[2].Name())
}

func TestSdRebaseMain_WithMulitpleMergedBranches_DropsCommitss(t *testing.T) {
	assert := assert.New(t)
	testExecutor := testinginit.InitTest(slog.LevelInfo)

	testinginit.AddCommit("first", "1")
	testinginit.AddCommit("second", "2")
	testinginit.AddCommit("third", "3")

	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "push", "origin", sd.GetMainBranchOrDie())

	allOriginalCommits := sd.GetAllCommits()

	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "reset", "--hard", allOriginalCommits[2].Commit)

	testinginit.AddCommit("second", "2-rebase-will-drop-this-file")
	testinginit.AddCommit("third", "3-rebase-will-drop-this-file")
	testinginit.AddCommit("fourth", "4")

	testExecutor.SetResponse(allOriginalCommits[0].Branch+"\n"+allOriginalCommits[1].Branch, nil, "gh", "pr", "list", ex.MatchAnyRemainingArgs)

	parseArguments(os.Stdout, flag.NewFlagSet("sd", flag.ContinueOnError), []string{"rebase-main"})

	dirEntries, err := os.ReadDir(".")
	if err != nil {
		panic("Could not read dir: " + err.Error())
	}
	assert.Equal(5, len(dirEntries))
	assert.Equal(".git", dirEntries[0].Name())
	assert.Equal("1", dirEntries[1].Name())
	assert.Equal("2", dirEntries[2].Name())
	assert.Equal("3", dirEntries[3].Name())
	assert.Equal("4", dirEntries[4].Name())
}

func TestSdRebaseMain_WithDuplicateBranches_Panics(t *testing.T) {
	assert := assert.New(t)
	testExecutor := testinginit.InitTest(slog.LevelInfo)

	testinginit.AddCommit("first", "1")
	testinginit.AddCommit("second", "2.1")
	testinginit.AddCommit("second", "2.2")

	allOriginalCommits := sd.GetAllCommits()

	testExecutor.SetResponse(allOriginalCommits[0].Branch, nil, "gh", "pr", "list", ex.MatchAnyRemainingArgs)

	// Return on panic
	defer func() { _ = recover() }()

	parseArguments(os.Stdout, flag.NewFlagSet("sd", flag.ContinueOnError), []string{"rebase-main"})

	// Never reaches here if `OtherFunctionThatPanics` panics.
	assert.Fail("did not panic with duplicate branches")
}

func TestSdRebaseMain_WhenRebaseFails_DropsBranches(t *testing.T) {
	assert := assert.New(t)
	testExecutor := testinginit.InitTest(slog.LevelInfo)

	testinginit.AddCommit("first", "file-with-conflicts")
	testinginit.CommitFileChange("second", "file-with-conflicts", "1")

	parseArguments(os.Stdout, flag.NewFlagSet("sd", flag.ContinueOnError), []string{"new"})

	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "push", "origin", sd.GetMainBranchOrDie())
	allCommits := sd.GetAllCommits()
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "reset", "--hard", allCommits[1].Commit)
	testinginit.CommitFileChange("second", "file-with-conflicts", "2")

	testExecutor.SetResponse(allCommits[0].Branch, nil, "gh", "pr", "list", ex.MatchAnyRemainingArgs)

	parseArguments(os.Stdout, flag.NewFlagSet("sd", flag.ContinueOnError), []string{"--log-level=debug", "rebase-main"})

	branches := ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "branch")
	assert.NotContains(branches, "second")
}
