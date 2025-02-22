package commands

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"

	ex "github.com/joshallenit/stacked-diff/execute"
	"github.com/joshallenit/stacked-diff/templates"
	"github.com/joshallenit/stacked-diff/testutil"
)

func TestSdWaitForMerge_WaitsForMerge(t *testing.T) {
	assert := assert.New(t)

	testExecutor := testutil.InitTest(slog.LevelInfo)

	testutil.AddCommit("first", "")
	allCommits := templates.GetAllCommits()
	testExecutor.SetResponse("2025-01-01", nil, "gh", "pr", "view", ex.MatchAnyRemainingArgs)

	out := testParseArguments("wait-for-merge", allCommits[0].Commit)

	assert.Contains(out, "Merged!")
}
