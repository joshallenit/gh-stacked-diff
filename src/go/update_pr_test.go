package stackeddiff

import (
	"log"
	"stackeddiff/testinginit"
	"testing"

	ex "stackeddiff/execute"

	"github.com/stretchr/testify/assert"
)

func Test_UpdatePr_OnRootCommit_UpdatesPr(t *testing.T) {
	assert := assert.New(t)
	testinginit.CdTestRepo()
	testinginit.AddCommit("first", "")

	testinginit.SetTestExecutor()

	CreateNewPr(true, "", ex.GetMainBranch(), GetBranchInfo("", IndicatorTypeGuess), log.Default())

	testinginit.AddCommit("second", "")

	commitsOnMain := GetAllCommits()

	UpdatePr(commitsOnMain[1].Commit, []string{}, log.Default())

	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "switch", GetBranchForCommit(commitsOnMain[1].Commit))

	commitsOnBranch := GetAllCommits()

	assert.Equal(2, len(commitsOnBranch))
}

func Test_UpdatePr_OnExistingRoot_UpdatesPr(t *testing.T) {
	assert := assert.New(t)
	testinginit.CdTestRepo()

	testinginit.AddCommit("first", "")
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "push", "origin", ex.GetMainBranch())

	testinginit.AddCommit("second", "")

	testinginit.SetTestExecutor()

	CreateNewPr(true, "", ex.GetMainBranch(), GetBranchInfo("", IndicatorTypeGuess), log.Default())

	testinginit.AddCommit("third", "")

	testinginit.AddCommit("fourth", "")

	commitsOnMain := GetAllCommits()

	UpdatePr(commitsOnMain[2].Commit, []string{}, log.Default())

	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "switch", GetBranchForCommit(commitsOnMain[2].Commit))

	allCommitsOnBranch := GetAllCommits()

	assert.Equal(3, len(allCommitsOnBranch))
	assert.Equal(allCommitsOnBranch[0].Subject, "fourth")
	assert.Equal(allCommitsOnBranch[1].Subject, "second")
	assert.Equal(allCommitsOnBranch[2].Subject, "first")

	newCommitsOnBranch := GetNewCommits(ex.GetMainBranch(), "HEAD")

	assert.Equal(2, len(newCommitsOnBranch))
	assert.Equal(newCommitsOnBranch[0].Subject, "fourth")
	assert.Equal(newCommitsOnBranch[1].Subject, "second")
}
