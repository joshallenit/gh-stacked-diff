package main

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"

	"stackeddiff/testinginit"
)

func TestSdLog_LogsOutput(t *testing.T) {
	assert := assert.New(t)

	testinginit.InitTest(slog.LevelInfo)

	testinginit.AddCommit("first", "")

	out := testParseArguments("log")

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

	out := testParseArguments("log")

	assert.Contains(out, "\n 2.    ")
	assert.Contains(out, "\n10.    ")
}

func TestSdLog_WhenMultiplePrs_MatchesAllPrs(t *testing.T) {
	assert := assert.New(t)

	testinginit.InitTest(slog.LevelInfo)

	testinginit.AddCommit("first", "")
	testinginit.AddCommit("second", "")

	testParseArguments("new", "2")
	testParseArguments("new", "1")

	out := testParseArguments("log")

	assert.Regexp("✅.*first", out)
	assert.Regexp("✅.*second", out)
}
