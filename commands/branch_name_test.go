package commands

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"

	sd "stackeddiff"
)

func TestSdBranchName_OutputsBranchName(t *testing.T) {
	assert := assert.New(t)

	testutil.InitTest(slog.LevelInfo)

	testutil.AddCommit("first", "")

	allCommits := sd.GetAllCommits()

	out := testParseArguments("branch-name", allCommits[0].Commit)

	assert.Equal(allCommits[0].Branch, out)
}
