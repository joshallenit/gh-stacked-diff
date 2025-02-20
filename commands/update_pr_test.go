package commands

import (
	"log/slog"
	"testing"

	ex "github.com/joshallenit/stacked-diff/execute"
	"github.com/joshallenit/stacked-diff/templates"
	"github.com/joshallenit/stacked-diff/testutil"
	"github.com/joshallenit/stacked-diff/util"

	"github.com/stretchr/testify/assert"
)

func Test_UpdatePr_OnRootCommit_UpdatesPr(t *testing.T) {
	assert := assert.New(t)
	testutil.InitTest(slog.LevelInfo)
	testutil.AddCommit("first", "")

	CreateNewPr(true, "", util.GetMainBranchOrDie(), templates.GetBranchInfo("", IndicatorTypeGuess))

	testutil.AddCommit("second", "")

	commitsOnMain := GetAllCommits()

	UpdatePr(templates.GetBranchInfo(commitsOnMain[1].Commit, IndicatorTypeCommit), []string{}, IndicatorTypeCommit)

	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "switch", commitsOnMain[1].Branch)

	commitsOnBranch := GetAllCommits()

	assert.Equal(3, len(commitsOnBranch))
}

func Test_UpdatePr_OnExistingRoot_UpdatesPr(t *testing.T) {
	assert := assert.New(t)
	testutil.InitTest(slog.LevelInfo)

	testutil.AddCommit("first", "")
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "push", "origin", util.GetMainBranchOrDie())

	testutil.AddCommit("second", "")

	CreateNewPr(true, "", util.GetMainBranchOrDie(), templates.GetBranchInfo("", IndicatorTypeGuess))

	testutil.AddCommit("third", "")

	testutil.AddCommit("fourth", "")

	commitsOnMain := GetAllCommits()

	UpdatePr(templates.GetBranchInfo(commitsOnMain[2].Commit, IndicatorTypeCommit), []string{}, IndicatorTypeCommit)

	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "switch", commitsOnMain[2].Branch)

	allCommitsOnBranch := GetAllCommits()

	assert.Equal(4, len(allCommitsOnBranch))
	assert.Equal("fourth", allCommitsOnBranch[0].Subject)
	assert.Equal("second", allCommitsOnBranch[1].Subject)
	assert.Equal("first", allCommitsOnBranch[2].Subject)
	assert.Equal(testutil.InitialCommitSubject, allCommitsOnBranch[3].Subject)

	newCommitsOnBranch := templates.GetNewCommits("HEAD")

	assert.Equal(2, len(newCommitsOnBranch))
	assert.Equal(newCommitsOnBranch[0].Subject, "fourth")
	assert.Equal(newCommitsOnBranch[1].Subject, "second")
}
