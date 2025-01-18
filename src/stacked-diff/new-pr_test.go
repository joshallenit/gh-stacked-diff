package stacked_diff

import (
	"bytes"
	"log"
	ex "stacked-diff-workflow/src/execute"
	testing_init "stacked-diff-workflow/src/testing-init"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewPr_OnNewRepo_CreatesPr(t *testing.T) {
	testing_init.CdTestRepo()

	testing_init.AddCommit("first", "")

	testExecutor := ex.TestExecutor{TestLogger: log.Default()}
	testExecutor.SetResponse("Ok", nil, "gh")
	ex.SetGlobalExecutor(testExecutor)

	CreateNewPr(true, "", ex.GetMainBranch(), 0, GetBranchInfo(""), log.Default())

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

	testing_init.CdTestRepo()

	testing_init.AddCommit("first", "")
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "push", "origin", ex.GetMainBranch())

	testing_init.AddCommit("second", "")
	allCommits := GetNewCommits(ex.GetMainBranch())

	testExecutor := ex.TestExecutor{TestLogger: log.Default()}
	testExecutor.SetResponse("Ok", nil, "gh")
	ex.SetGlobalExecutor(testExecutor)

	CreateNewPr(true, "", ex.GetMainBranch(), 0, GetBranchInfo(""), log.Default())

	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "switch", GetBranchForCommit(allCommits[0].Commit))
	commitsOnNewBranch := GetNewCommits(ex.GetMainBranch())
	assert.Equal(1, len(commitsOnNewBranch))
	assert.Equal(allCommits[0].Subject, commitsOnNewBranch[0].Subject)
}

func Test_NewPr_WithMiddleCommit_CreatesPr(t *testing.T) {
	assert := assert.New(t)

	testing_init.CdTestRepo()

	testing_init.AddCommit("first", "")
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "push", "origin", ex.GetMainBranch())

	testing_init.AddCommit("second", "")

	testing_init.AddCommit("third", "")
	allCommits := GetNewCommits(ex.GetMainBranch())

	testExecutor := ex.TestExecutor{TestLogger: log.Default()}
	testExecutor.SetResponse("Ok", nil, "gh")
	ex.SetGlobalExecutor(testExecutor)

	CreateNewPr(true, "", ex.GetMainBranch(), 0, GetBranchInfo(""), log.Default())

	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "switch", GetBranchForCommit(allCommits[0].Commit))
	commitsOnNewBranch := GetNewCommits(ex.GetMainBranch())
	assert.Equal(1, len(commitsOnNewBranch))
	assert.Equal(allCommits[0].Subject, commitsOnNewBranch[0].Subject)
}
