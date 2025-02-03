package main

import (
	"flag"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"

	"stackeddiff/testinginit"
)

func TestSdLog_LogsOutput(t *testing.T) {
	assert := assert.New(t)

	testinginit.InitTest(slog.LevelInfo)

	testinginit.AddCommit("first", "")

	outWriter := testinginit.NewWriteRecorder()
	ParseArguments(outWriter, flag.NewFlagSet("sd", flag.ContinueOnError), []string{"log"})
	out := outWriter.String()

	assert.Contains(out, "first")
}

func TestGitlog_WhenManyCommits_PadsFirstCommits(t *testing.T) {
	assert := assert.New(t)

	testinginit.InitTest(slog.LevelInfo)

	testinginit.AddCommit("first", "")
	testinginit.AddCommit("second", "")
	testinginit.AddCommit("third", "")
	testinginit.AddCommit("fourth", "")
	testinginit.AddCommit("fifth", "")
	testinginit.AddCommit("sixth", "")
	testinginit.AddCommit("seventh", "")
	testinginit.AddCommit("eigth", "")
	testinginit.AddCommit("ninth", "")
	testinginit.AddCommit("tenth", "")

	outWriter := testinginit.NewWriteRecorder()
	ParseArguments(outWriter, flag.NewFlagSet("sd", flag.ContinueOnError), []string{"log"})
	out := outWriter.String()

	assert.Contains(out, "\n 2.    ")
	assert.Contains(out, "\n10.    ")
}
