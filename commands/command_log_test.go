package commands

import (
	"log/slog"
	"testing"

	"strings"

	"github.com/fatih/color"
	"github.com/joshallenit/gh-stacked-diff/v2/templates"
	"github.com/joshallenit/gh-stacked-diff/v2/testutil"

	"github.com/joshallenit/gh-stacked-diff/v2/util"
	"github.com/stretchr/testify/assert"
)

func TestPrintGitLog_OnEmptyRemote_PrintsLog(t *testing.T) {

	testutil.InitTest(t, slog.LevelInfo)

	testutil.AddCommit("first", "")

	out := testParseArguments("log")

	if !strings.Contains(out, "first") {
		t.Errorf("'first' should be in %s", out)
	}
}

func TestPrintGitLog_WhenRemoteHasSomeCommits_PrintsNewLogsOnly(t *testing.T) {
	testutil.InitTest(t, slog.LevelInfo)

	testutil.AddCommit("first", "")

	util.ExecuteOrDie(util.ExecuteOptions{}, "git", "push", "origin", util.GetMainBranchOrDie())

	testutil.AddCommit("second", "")

	out := testParseArguments("log")

	if strings.Contains(out, "first") {
		t.Errorf("'first' should not be in %s", out)
	}
	if !strings.Contains(out, "second") {
		t.Errorf("'second' should be in %s", out)
	}
}

func TestPrintGitLog_WhenPrCreatedForSomeCommits_PrintsCheckForCommitsWithPrs(t *testing.T) {
	testutil.InitTest(t, slog.LevelInfo)

	testutil.AddCommit("first", "")

	testParseArguments("new", "1")

	out := testParseArguments("log")

	if !strings.Contains(out, "✅") {
		t.Errorf("'✅' should be in %s", out)
	}
}

func TestPrintGitLog_WhenNotOnMain_OnlyShowsCommitsNotOnMain(t *testing.T) {
	assert := assert.New(t)

	testutil.InitTest(t, slog.LevelInfo)

	testutil.AddCommit("first", "")

	util.ExecuteOrDie(util.ExecuteOptions{}, "git", "push", "origin", util.GetMainBranchOrDie())

	util.ExecuteOrDie(util.ExecuteOptions{}, "git", "checkout", "-b", "my-branch")

	testutil.AddCommit("second", "")

	out := testParseArguments("log")

	assert.NotContains(out, "first")
	assert.Contains(out, "second")
}

func TestPrintGitLog_WhenCommitHasBranch_PrintsExtraBranchCommits(t *testing.T) {
	assert := assert.New(t)
	testutil.InitTest(t, slog.LevelInfo)

	testutil.AddCommit("first", "")

	testParseArguments("new", "1")

	testutil.AddCommit("second", "")

	testParseArguments("update", "2", "1")

	out := testParseArguments("log")

	allCommits := templates.GetAllCommits()
	assert.Equal("1. ✅ "+color.YellowString(allCommits[0].Commit)+" first\n"+
		"      - second\n"+
		"      - first\n",
		out)
}

func TestSdLog_LogsOutput(t *testing.T) {
	assert := assert.New(t)

	testutil.InitTest(t, slog.LevelInfo)

	testutil.AddCommit("first", "")

	out := testParseArguments("log")

	assert.Contains(out, "first")
}

func TestSdLog_WhenManyCommits_PadsFirstCommits(t *testing.T) {
	assert := assert.New(t)

	testutil.InitTest(t, slog.LevelInfo)

	testutil.AddCommit("first", "")
	testutil.AddCommit("second", "")
	testutil.AddCommit("third", "")
	testutil.AddCommit("fourth", "")
	testutil.AddCommit("fifth", "")
	testutil.AddCommit("sixth", "")
	testutil.AddCommit("seventh", "")
	testutil.AddCommit("eigth", "")
	testutil.AddCommit("ninth", "")
	testutil.AddCommit("tenth", "")

	out := testParseArguments("log")

	assert.Contains(out, "\n 2.    ")
	assert.Contains(out, "\n10.    ")
}

func TestSdLog_WhenMultiplePrs_MatchesAllPrs(t *testing.T) {
	assert := assert.New(t)

	testutil.InitTest(t, slog.LevelInfo)

	testutil.AddCommit("first", "")
	testutil.AddCommit("second", "")

	testParseArguments("new", "2")
	testParseArguments("new", "1")

	out := testParseArguments("log")

	assert.Regexp("✅.*first", out)
	assert.Regexp("✅.*second", out)
}
