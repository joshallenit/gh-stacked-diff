package stacked_diff

import (
	"log"
	ex "stacked-diff-workflow/src/execute"
	testing_init "stacked-diff-workflow/src/testing-init"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_UpdatePr_OnRootCommit_UpdatesPr(t *testing.T) {
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

func Test_UpdatePr_OnExistingRoot_UpdatesPr(t *testing.T) {
	assert := assert.New(t)
	testing_init.CdTestRepo()

	testing_init.AddCommit("first", "")
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "push", "origin", ex.GetMainBranch())

	testing_init.AddCommit("second", "")

	testExecutor := ex.TestExecutor{TestLogger: log.Default()}
	testExecutor.SetResponse("Ok", nil, "gh")
	ex.SetGlobalExecutor(testExecutor)

	CreateNewPr(true, "", ex.GetMainBranch(), 0, GetBranchInfo(""), log.Default())

	testing_init.AddCommit("third", "")

	testing_init.AddCommit("fourth", "")

	commitsOnMain := GetAllCommits()

	UpdatePr(commitsOnMain[2].Commit, []string{}, log.Default())

	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "switch", GetBranchForCommit(commitsOnMain[2].Commit))

	allCommitsOnBranch := GetAllCommits()

	assert.Equal(3, len(allCommitsOnBranch))
	assert.Equal(allCommitsOnBranch[0].Subject, "fourth")
	assert.Equal(allCommitsOnBranch[1].Subject, "second")
	assert.Equal(allCommitsOnBranch[2].Subject, "first")

	newCommitsOnBranch := GetNewCommits(ex.GetMainBranch())

	assert.Equal(2, len(newCommitsOnBranch))
	assert.Equal(newCommitsOnBranch[0].Subject, "fourth")
	assert.Equal(newCommitsOnBranch[1].Subject, "second")
}
