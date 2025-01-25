package main

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	sd "stackeddiff"
	"stackeddiff/testinginit"
)

func TestSdUpdate_UpdatesPr(t *testing.T) {
	assert := assert.New(t)

	testinginit.CdTestRepo()

	testinginit.AddCommit("first", "")

	testinginit.SetTestExecutor()

	ParseArguments(os.Stdout, flag.NewFlagSet("sd", flag.ExitOnError), []string{"new"})

	testinginit.AddCommit("second", "")

	allCommits := sd.GetAllCommits()

	ParseArguments(os.Stdout, flag.NewFlagSet("sd", flag.ExitOnError), []string{"update", allCommits[1].Commit})

	allCommits = sd.GetAllCommits()

	assert.Equal(1, len(allCommits))
	assert.Equal("first", allCommits[0].Subject)
}
