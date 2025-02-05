package main

import (
	"flag"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	sd "stackeddiff"
	ex "stackeddiff/execute"
	"stackeddiff/testinginit"
)

func TestSdCodeOwners_OutputsOwnersOfChangedFiles(t *testing.T) {
	assert := assert.New(t)

	testinginit.InitTest(slog.LevelInfo)

	testinginit.AddCommit("first", "first-not-changed")

	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "push", "origin", sd.GetMainBranch())

	testinginit.AddCommit("second", "second-changed")

	testinginit.AddCommit("third", "third-changed")

	codeOwners := "first-not-changed firstOwners\n" +
		"second-changed secondOwners\n" +
		"third-changed thirdOwners\n"

	ex.ExecuteOrDie(ex.ExecuteOptions{}, "mkdir", "-p", ".github")
	if writeErr := os.WriteFile(".github/CODEOWNERS", []byte(codeOwners), 0); writeErr != nil {
		panic(writeErr)
	}
	outWriter := testinginit.NewWriteRecorder()
	ParseArguments(outWriter, flag.NewFlagSet("sd", flag.ContinueOnError), []string{"code-owners"})
	out := outWriter.String()

	assert.Equal("Owner: secondOwners\n"+
		"second-changed\n"+
		"\n"+
		"Owner: thirdOwners\n"+
		"third-changed\n", out)
}
