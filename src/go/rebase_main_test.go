package stackeddiff

import (
	"log"
	"os"
	"stackeddiff/testinginit"
	"testing"

	ex "stackeddiff/execute"

	"github.com/stretchr/testify/assert"
)

func Test_RebaseMain_WithDifferentCommits_DropsCommits(t *testing.T) {
	assert := assert.New(t)
	testinginit.CdTestRepo()

	testinginit.AddCommit("first", "")

	testinginit.AddCommit("second", "rebase-will-keep-this-file")

	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "push", "origin", ex.GetMainBranch())

	ex.ExecuteOrDie(ex.ExecuteOptions{}, "ls")
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "log")

	allOriginalCommits := GetAllCommits()

	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "reset", "--hard", allOriginalCommits[1].Commit)

	testinginit.AddCommit("second", "rebase-will-drop-this-file")

	ex.ExecuteOrDie(ex.ExecuteOptions{}, "ls")
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "log")

	testinginit.SetTestExecutor()

	RebaseMain(log.Default())

	ex.ExecuteOrDie(ex.ExecuteOptions{}, "ls")
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "log")

	dirEntries, err := os.ReadDir(".")
	if err != nil {
		panic("Could not read dir: " + err.Error())
	}
	assert.Equal(3, len(dirEntries))
	assert.Equal(".git", dirEntries[0].Name())
	assert.Equal("first", dirEntries[1].Name())
	assert.Equal("rebase-will-keep-this-file", dirEntries[2].Name())

	// so I could do something like this:
	// https://github.com/Nutlope/aicommits
	// to create an AI git commit message

	/*
					https://www.hatica.io/blog/ai-commit-tools/
				https://github.com/kamushadenes/chloe/blob/main/.github/scripts/release-notes.py
		the commit messages don't look that great, probably not the best idea
		but splitting commits into smaller ones based on compilability could be good. Maybe figuring out which dependencies are declared in which file and then which dependencies are used in each file? That could work


				I could actually try these commit messages to see how well they work or don't work
				Other ideas:
					- automatically add reviewers when creating a PR
					- Next `gh pr view 83824 --json latestReviews` and ensure developer is already not approved so that the review is not dismissed
					- show the commits in gitlog on the branches tabbed over so that it's easier to read
					- show the output of git commands in a tabbed window that uses ANSI escape codes to move around the screen
					- use Jira numbers to match commits, probably not the best as there are multiple PRs for the same Jira
					- use list numbers but how is this going to be distinguished from PR numbers?
					- merge all commands into one main and use "sd log" etc. <-- This is definitely the way to go, why do I have separate mains for eacch script?
					- better error handling so that it reverted on error rather then leaving in an indeterminite state... but wouldn't this mean that I have to save error codes so they can be reported upstream?
	*/
}
