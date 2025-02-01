package main

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"

	sd "stackeddiff"
	ex "stackeddiff/execute"
	"stackeddiff/testinginit"
)

func TestSdWaitForMerge_WaitsForMerge(t *testing.T) {
	assert := assert.New(t)

	testinginit.CdTestRepo()

	testinginit.AddCommit("first", "")
	allCommits := sd.GetAllCommits()
	testingExecutor := testinginit.SetTestExecutor()
	testingExecutor.SetResponse("2025-01-01", nil, "gh", "pr", "view", ex.MatchAnyRemainingArgs)

	outWriter := testinginit.NewWriteRecorder()
	ParseArguments(
		outWriter,
		flag.NewFlagSet("sd", flag.ContinueOnError),
		[]string{"--log-level=debug", "wait-for-merge", allCommits[0].Commit},
	)
	out := outWriter.String()

	assert.Contains(out, "Merged!")
}
