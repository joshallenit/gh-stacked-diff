package stackeddiff

import (
	"log/slog"
	"os"
	"stackeddiff/testinginit"
	"testing"

	ex "stackeddiff/execute"

	"github.com/stretchr/testify/assert"
)

func Test_RebaseMain_WithDifferentCommits_DropsCommits(t *testing.T) {
	assert := assert.New(t)
	testExecutor := testinginit.InitTest(slog.LevelInfo)

	testinginit.AddCommit("first", "")

	testinginit.AddCommit("second", "rebase-will-keep-this-file")

	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "push", "origin", GetMainBranchOrDie())

	allOriginalCommits := GetAllCommits()

	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "reset", "--hard", allOriginalCommits[1].Commit)

	testinginit.AddCommit("second", "rebase-will-drop-this-file")

	testExecutor.SetResponse(allOriginalCommits[0].Branch+" fakeMergeCommit",
		nil, "gh", "pr", "list", ex.MatchAnyRemainingArgs)

	RebaseMain()

	dirEntries, err := os.ReadDir(".")
	if err != nil {
		panic("Could not read dir: " + err.Error())
	}
	assert.Equal(3, len(dirEntries))
	assert.Equal(".git", dirEntries[0].Name())
	assert.Equal("first", dirEntries[1].Name())
	assert.Equal("rebase-will-keep-this-file", dirEntries[2].Name())
}
