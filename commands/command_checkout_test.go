package commands

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"

	sd "stackeddiff"
	"stackeddiff/testinginit"
)

func TestSdCheckout_ChecksOutBranch(t *testing.T) {
	assert := assert.New(t)

	testinginit.InitTest(slog.LevelInfo)

	testinginit.AddCommit("first", "")

	allCommits := sd.GetAllCommits()

	testParseArguments("new")

	testParseArguments("checkout", allCommits[0].Commit)

	assert.Equal(allCommits[0].Branch, sd.GetCurrentBranchName())
}
