package stackeddiff

import (
	"log/slog"
	"stackeddiff/testinginit"
	"testing"

	ex "stackeddiff/execute"

	"github.com/stretchr/testify/assert"
)

func Test_UpdatePr_OnRootCommit_UpdatesPr(t *testing.T) {
	assert := assert.New(t)
	testinginit.InitTest(slog.LevelInfo)
	testinginit.AddCommit("first", "")

	CreateNewPr(true, "", GetMainBranchOrDie(), GetBranchInfo("", IndicatorTypeGuess))

	testinginit.AddCommit("second", "")

	commitsOnMain := GetAllCommits()

	UpdatePr(GetBranchInfo(commitsOnMain[1].Commit, IndicatorTypeCommit), []string{}, IndicatorTypeCommit)

	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "switch", commitsOnMain[1].Branch)

	commitsOnBranch := GetAllCommits()

	assert.Equal(2, len(commitsOnBranch))
}

func Test_UpdatePr_OnExistingRoot_UpdatesPr(t *testing.T) {
	assert := assert.New(t)
	testinginit.InitTest(slog.LevelInfo)

	testinginit.AddCommit("first", "")
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "push", "origin", GetMainBranchOrDie())

	testinginit.AddCommit("second", "")

	CreateNewPr(true, "", GetMainBranchOrDie(), GetBranchInfo("", IndicatorTypeGuess))

	testinginit.AddCommit("third", "")

	testinginit.AddCommit("fourth", "")

	commitsOnMain := GetAllCommits()

	UpdatePr(GetBranchInfo(commitsOnMain[2].Commit, IndicatorTypeCommit), []string{}, IndicatorTypeCommit)

	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "switch", commitsOnMain[2].Branch)

	allCommitsOnBranch := GetAllCommits()

	assert.Equal(3, len(allCommitsOnBranch))
	assert.Equal(allCommitsOnBranch[0].Subject, "fourth")
	assert.Equal(allCommitsOnBranch[1].Subject, "second")
	assert.Equal(allCommitsOnBranch[2].Subject, "first")

	newCommitsOnBranch := getNewCommits(GetMainBranchOrDie(), "HEAD")

	assert.Equal(2, len(newCommitsOnBranch))
	assert.Equal(newCommitsOnBranch[0].Subject, "fourth")
	assert.Equal(newCommitsOnBranch[1].Subject, "second")
}
