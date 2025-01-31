package stackeddiff

import (
	"bytes"
	"log"
	"stackeddiff/testinginit"
	"strings"
	"testing"

	ex "stackeddiff/execute"

	"github.com/stretchr/testify/assert"
)

func Test_NewPr_OnNewRepo_CreatesPr(t *testing.T) {
	testinginit.CdTestRepo()

	testinginit.AddCommit("first", "")

	testinginit.SetTestExecutor()

	CreateNewPr(true, "", ex.GetMainBranch(), GetBranchInfo("", IndicatorTypeGuess), log.Default())

	// Check that the PR was created
	outWriter := new(bytes.Buffer)
	PrintGitLog(outWriter)
	out := outWriter.String()

	if !strings.Contains(out, "✅") {
		t.Errorf("'✅' should be in %s", out)
	}
}

func Test_NewPr_OnRepoWithPreviousCommit_CreatesPr(t *testing.T) {
	assert := assert.New(t)

	testinginit.CdTestRepo()

	testinginit.AddCommit("first", "")
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "push", "origin", ex.GetMainBranch())

	testinginit.AddCommit("second", "")
	allCommits := GetNewCommits(ex.GetMainBranch(), "HEAD")

	testinginit.SetTestExecutor()

	CreateNewPr(true, "", ex.GetMainBranch(), GetBranchInfo("", IndicatorTypeGuess), log.Default())

	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "switch", allCommits[0].Branch)
	commitsOnNewBranch := GetNewCommits(ex.GetMainBranch(), "HEAD")
	assert.Equal(1, len(commitsOnNewBranch))
	assert.Equal(allCommits[0].Subject, commitsOnNewBranch[0].Subject)
}

func Test_NewPr_WithMiddleCommit_CreatesPr(t *testing.T) {
	assert := assert.New(t)

	testinginit.CdTestRepo()

	testinginit.AddCommit("first", "")
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "push", "origin", ex.GetMainBranch())

	testinginit.AddCommit("second", "")

	testinginit.AddCommit("third", "")
	allCommits := GetNewCommits(ex.GetMainBranch(), "HEAD")

	testinginit.SetTestExecutor()

	CreateNewPr(true, "", ex.GetMainBranch(), GetBranchInfo("", IndicatorTypeGuess), log.Default())

	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "switch", allCommits[0].Branch)
	commitsOnNewBranch := GetNewCommits(ex.GetMainBranch(), "HEAD")
	assert.Equal(1, len(commitsOnNewBranch))
	assert.Equal(allCommits[0].Subject, commitsOnNewBranch[0].Subject)
}
