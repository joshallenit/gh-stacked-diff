package commands

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"

	ex "github.com/joshallenit/gh-stacked-diff/v2/execute"
	"github.com/joshallenit/gh-stacked-diff/v2/templates"
	"github.com/joshallenit/gh-stacked-diff/v2/testutil"
)

func TestSdWaitForMerge_WaitsForMerge(t *testing.T) {
	assert := assert.New(t)

	testExecutor := testutil.InitTest(t, slog.LevelInfo)

	testutil.AddCommit("first", "")
	allCommits := templates.GetAllCommits()
	testExecutor.SetResponse("2025-01-01", nil, "gh", "pr", "view", ex.MatchAnyRemainingArgs)

	out := testParseArguments("wait-for-merge", allCommits[0].Commit)

	assert.Contains(out, "Merged!")
}
