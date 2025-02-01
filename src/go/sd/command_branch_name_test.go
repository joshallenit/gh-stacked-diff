package main

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"

	"bytes"
	sd "stackeddiff"
	"stackeddiff/testinginit"
)

func TestSdBranchName_OutputsBranchName(t *testing.T) {
	assert := assert.New(t)

	testinginit.CdTestRepo()

	testinginit.AddCommit("first", "")

	allCommits := sd.GetAllCommits()

	outWriter := new(bytes.Buffer)
	ParseArguments(outWriter, flag.NewFlagSet("sd", flag.ContinueOnError), []string{"branch-name", allCommits[0].Commit})
	out := outWriter.String()

	assert.Equal(allCommits[0].Branch, out)
}
