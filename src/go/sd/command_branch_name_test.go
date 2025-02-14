package main

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"

	sd "stackeddiff"
	"stackeddiff/testinginit"
)

func TestSdBranchName_OutputsBranchName(t *testing.T) {
	assert := assert.New(t)

	testinginit.InitTest(slog.LevelInfo)

	testinginit.AddCommit("first", "")

	allCommits := sd.GetAllCommits()

	out := testParseArguments("branch-name", allCommits[0].Commit)

	assert.Equal(allCommits[0].Branch, out)
}
