package main

import (
	"flag"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"

	"os"
	sd "stackeddiff"
	"stackeddiff/testinginit"
)

func TestSdCheckout_ChecksOutBranch(t *testing.T) {
	assert := assert.New(t)

	testinginit.InitTest(slog.LevelInfo)

	testinginit.AddCommit("first", "")

	allCommits := sd.GetAllCommits()

	parseArguments(os.Stdout, flag.NewFlagSet("sd", flag.ContinueOnError), []string{"new"})

	parseArguments(os.Stdout, flag.NewFlagSet("sd", flag.ContinueOnError), []string{"checkout", allCommits[0].Commit})

	assert.Equal(allCommits[0].Branch, sd.GetCurrentBranchName())
}
