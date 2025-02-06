package main

import (
	"flag"
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

	outWriter := testinginit.NewWriteRecorder()
	parseArguments(outWriter, flag.NewFlagSet("sd", flag.ContinueOnError), []string{"branch-name", allCommits[0].Commit})
	out := outWriter.String()

	assert.Equal(allCommits[0].Branch, out)
}
