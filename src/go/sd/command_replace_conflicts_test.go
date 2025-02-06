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

func TestSdReplaceConflicts_WhenConflictOnLastCommit_ReplacesCommit(t *testing.T) {
	assert := assert.New(t)
	testinginit.InitTest(slog.LevelInfo)

	testinginit.AddCommit("first", "file-with-conflicts")
	testinginit.CommitFileChange("second", "file-with-conflicts", "1")
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "push", "origin", sd.GetMainBranchOrDie())
	allCommits := sd.GetAllCommits()
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "reset", "--hard", allCommits[1].Commit)
	testinginit.CommitFileChange("third", "file-with-conflicts", "2")

	ParseArguments(os.Stdout, flag.NewFlagSet("sd", flag.ContinueOnError), []string{"new"})

	ParseArguments(
		os.Stdout,
		flag.NewFlagSet("sd", flag.ContinueOnError),
		[]string{"checkout", "1"},
	)

	_, mergeErr := ex.Execute(ex.ExecuteOptions{}, "git", "merge", "origin/"+sd.GetMainBranchOrDie())
	assert.NotNil(mergeErr)

	if writeErr := os.WriteFile("file-with-conflicts", []byte("1\n2"), 0); writeErr != nil {
		panic(writeErr)
	}
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "add", ".")

	continueOptions := ex.ExecuteOptions{EnvironmentVariables: []string{"GIT_EDITOR=true"}}
	ex.ExecuteOrDie(continueOptions, "git", "merge", "--continue")
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "switch", sd.GetMainBranchOrDie())

	_, rebaseErr := ex.Execute(ex.ExecuteOptions{}, "git", "rebase", "origin/"+sd.GetMainBranchOrDie())
	assert.NotNil(rebaseErr)

	ParseArguments(
		os.Stdout,
		flag.NewFlagSet("sd", flag.ContinueOnError),
		[]string{"replace-conflicts", "--confirm=true"},
	)

	allCommits = sd.GetAllCommits()

	assert.Equal(3, len(allCommits))
	assert.Equal("third", allCommits[0].Subject)
	assert.Equal("second", allCommits[1].Subject)
	assert.Equal("first", allCommits[2].Subject)

	dirEntries, err := os.ReadDir(".")
	if err != nil {
		panic("Could not read dir: " + err.Error())
	}
	assert.Equal(2, len(dirEntries))
	assert.Equal(".git", dirEntries[0].Name())
	assert.Equal("file-with-conflicts", dirEntries[1].Name())

	contents, readErr := os.ReadFile("file-with-conflicts")
	assert.Nil(readErr)
	// Add a .? to account for eol on windows.
	assert.Regexp("1.?\n2", string(contents))
}
