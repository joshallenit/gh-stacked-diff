package main

import (
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
	ParseArguments(outWriter, []string{"log"})
	out := outWriter.String()

	assert.Contains(out, "first")
}
