package commands

import (
	"log/slog"
	"testing"

	"github.com/joshallenit/gh-testsd3/v2/templates"
	"github.com/joshallenit/gh-testsd3/v2/testutil"
	"github.com/stretchr/testify/assert"
)

func TestSdBranchName_OutputsBranchName(t *testing.T) {
	assert := assert.New(t)

	testutil.InitTest(slog.LevelInfo)

	testutil.AddCommit("first", "")

	allCommits := templates.GetAllCommits()

	out := testParseArguments("branch-name", allCommits[0].Commit)

	assert.Equal(allCommits[0].Branch, out)
}
