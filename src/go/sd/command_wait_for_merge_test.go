package main

import (
	"flag"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"

	sd "stackeddiff"
	ex "stackeddiff/execute"
	"stackeddiff/testinginit"
)

func TestSdWaitForMerge_WaitsForMerge(t *testing.T) {
	assert := assert.New(t)

	testExecutor := testinginit.InitTest(slog.LevelInfo)

	testinginit.AddCommit("first", "")
	allCommits := sd.GetAllCommits()
	testExecutor.SetResponse("2025-01-01", nil, "gh", "pr", "view", ex.MatchAnyRemainingArgs)

	outWriter := testinginit.NewWriteRecorder()
	ParseArguments(
		outWriter,
		flag.NewFlagSet("sd", flag.ContinueOnError),
		[]string{"wait-for-merge", allCommits[0].Commit},
	)
	out := outWriter.String()

	assert.Contains(out, "Merged!")
}
