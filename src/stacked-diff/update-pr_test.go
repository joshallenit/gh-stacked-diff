package stacked_diff

import (
	"log"
	ex "stacked-diff-workflow/src/execute"
	testing_init "stacked-diff-workflow/src/testing-init"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_UpdatePr_(t *testing.T) {
	assert := assert.New(t)
	testing_init.CdTestRepo()

	testing_init.AddCommit("first", "")

	testExecutor := ex.TestExecutor{TestLogger: log.Default()}
	testExecutor.SetResponse("Ok", nil, "gh")
	ex.SetGlobalExecutor(testExecutor)

	CreateNewPr(true, "", ex.GetMainBranch(), 0, GetBranchInfo(""), log.Default())

	testing_init.AddCommit("second", "")

	commitsOnMain := GetAllCommits()

	UpdatePr(commitsOnMain[1].Commit, []string{}, log.Default())

	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "switch", GetBranchForCommit(commitsOnMain[1].Commit))

	commitsOnBranch := GetAllCommits()

	assert.Equal(2, len(commitsOnBranch))
}
