package stackeddiff

import (
	"log/slog"
	"stackeddiff/testinginit"
	"strings"
	"testing"

	ex "stackeddiff/execute"

	"github.com/stretchr/testify/assert"
)

func Test_NewPr_OnNewRepo_CreatesPr(t *testing.T) {
	testinginit.InitTest(slog.LevelInfo)

	testinginit.AddCommit("first", "")

	CreateNewPr(true, "", GetMainBranchOrDie(), GetBranchInfo("", IndicatorTypeGuess))

	// Check that the PR was created
	outWriter := testinginit.NewWriteRecorder()
	PrintGitLog(outWriter)
	out := outWriter.String()

	if !strings.Contains(out, "✅") {
		t.Errorf("'✅' should be in %s", out)
	}
}

func Test_NewPr_OnRepoWithPreviousCommit_CreatesPr(t *testing.T) {
	assert := assert.New(t)

	testinginit.InitTest(slog.LevelInfo)

	testinginit.AddCommit("first", "")
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "push", "origin", GetMainBranchOrDie())

	testinginit.AddCommit("second", "")
	allCommits := GetNewCommits(GetMainBranchOrDie(), "HEAD")

	CreateNewPr(true, "", GetMainBranchOrDie(), GetBranchInfo("", IndicatorTypeGuess))

	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "switch", allCommits[0].Branch)
	commitsOnNewBranch := GetNewCommits(GetMainBranchOrDie(), "HEAD")
	assert.Equal(1, len(commitsOnNewBranch))
	assert.Equal(allCommits[0].Subject, commitsOnNewBranch[0].Subject)
}

func Test_NewPr_WithMiddleCommit_CreatesPr(t *testing.T) {
	assert := assert.New(t)

	testinginit.InitTest(slog.LevelInfo)

	testinginit.AddCommit("first", "")
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "push", "origin", GetMainBranchOrDie())

	testinginit.AddCommit("second", "")

	testinginit.AddCommit("third", "")
	allCommits := GetNewCommits(GetMainBranchOrDie(), "HEAD")

	CreateNewPr(true, "", GetMainBranchOrDie(), GetBranchInfo("", IndicatorTypeGuess))

	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "switch", allCommits[0].Branch)
	commitsOnNewBranch := GetNewCommits(GetMainBranchOrDie(), "HEAD")
	assert.Equal(1, len(commitsOnNewBranch))
	assert.Equal(allCommits[0].Subject, commitsOnNewBranch[0].Subject)
}
