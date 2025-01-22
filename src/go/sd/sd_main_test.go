package main

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"bytes"
	"stackeddiff/testinginit"
)

func Test_SdLog_LogsOutput(t *testing.T) {
	assert := assert.New(t)

	testinginit.CdTestRepo()

	testinginit.AddCommit("first", "")

	outWriter := new(bytes.Buffer)
	ParseArguments(outWriter, flag.NewFlagSet("sd", flag.ExitOnError), []string{"log"})
	out := outWriter.String()

	assert.Contains(out, "first")
}

func Test_SdNew_CreatesPr(t *testing.T) {
	assert := assert.New(t)

	testinginit.CdTestRepo()

	testinginit.AddCommit("first", "")

	testinginit.IgnoreGithubCli()

	ParseArguments(os.Stdout, flag.NewFlagSet("sd", flag.ExitOnError), []string{"new"})

	outWriter := new(bytes.Buffer)
	ParseArguments(outWriter, flag.NewFlagSet("sd", flag.ExitOnError), []string{"log"})
	out := outWriter.String()

	assert.Contains(out, "✅")
}
