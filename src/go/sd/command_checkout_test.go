package main

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"

	"os"
	sd "stackeddiff"
	"stackeddiff/testinginit"
)

func TestSdCheckout_ChecksOutBranch(t *testing.T) {
	assert := assert.New(t)

	testinginit.CdTestRepo()

	testinginit.AddCommit("first", "")

	allCommits := sd.GetAllCommits()

	testinginit.SetTestExecutor()

	ParseArguments(os.Stdout, flag.NewFlagSet("sd", flag.ExitOnError), []string{"new"})

	ParseArguments(os.Stdout, flag.NewFlagSet("sd", flag.ExitOnError), []string{"checkout", allCommits[0].Commit})

	assert.Equal(allCommits[0].Branch, sd.GetCurrentBranchName())
}
