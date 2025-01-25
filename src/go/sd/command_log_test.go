package main

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"

	"bytes"
	"stackeddiff/testinginit"
)

func TestSdLog_LogsOutput(t *testing.T) {
	assert := assert.New(t)

	testinginit.CdTestRepo()

	testinginit.AddCommit("first", "")

	outWriter := new(bytes.Buffer)
	ParseArguments(outWriter, flag.NewFlagSet("sd", flag.ExitOnError), []string{"log"})
	out := outWriter.String()

	assert.Contains(out, "first")
}
