package commands

import (
	"log/slog"
	"testing"

	"strings"

	"github.com/joshallenit/stacked-diff/templates"
	testutil "github.com/joshallenit/stacked-diff/test_util"

	ex "github.com/joshallenit/stacked-diff/execute"

	"github.com/joshallenit/stacked-diff/util"
	"github.com/stretchr/testify/assert"
)

func TestPrintGitLog_OnEmptyRemote_PrintsLog(t *testing.T) {

	testutil.InitTest(slog.LevelInfo)

	testutil.AddCommit("first", "")

	outWriter := testutil.NewWriteRecorder()
	PrintGitLog(outWriter)
	out := outWriter.String()

	if !strings.Contains(out, "first") {
		t.Errorf("'first' should be in %s", out)
	}
}

func Test_PrintGitLog_WhenRemoteHasSomeCommits_PrintsNewLogsOnly(t *testing.T) {
	testutil.InitTest(slog.LevelInfo)

	testutil.AddCommit("first", "")

	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "push", "origin", util.GetMainBranchOrDie())

	testutil.AddCommit("second", "")

	outWriter := testutil.NewWriteRecorder()
	PrintGitLog(outWriter)
	out := outWriter.String()

	if strings.Contains(out, "first") {
		t.Errorf("'first' should not be in %s", out)
	}
	if !strings.Contains(out, "second") {
		t.Errorf("'second' should be in %s", out)
	}
}

func TestPrintGitLog_WhenPrCreatedForSomeCommits_PrintsCheckForCommitsWithPrs(t *testing.T) {
	testutil.InitTest(slog.LevelInfo)

	testutil.AddCommit("first", "")

	CreateNewPr(true, "", util.GetMainBranchOrDie(), templates.GetBranchInfo("", IndicatorTypeGuess))

	outWriter := testutil.NewWriteRecorder()
	PrintGitLog(outWriter)
	out := outWriter.String()

	if !strings.Contains(out, "✅") {
		t.Errorf("'✅' should be in %s", out)
	}
}

func TestPrintGitLog_WhenNotOnMain_OnlyShowsCommitsNotOnMain(t *testing.T) {
	assert := assert.New(t)

	testutil.InitTest(slog.LevelInfo)

	testutil.AddCommit("first", "")

	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "push", "origin", util.GetMainBranchOrDie())

	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "checkout", "-b", "my-branch")

	testutil.AddCommit("second", "")

	outWriter := testutil.NewWriteRecorder()
	PrintGitLog(outWriter)
	out := outWriter.String()

	assert.NotContains(out, "first")
	assert.Contains(out, "second")
}

func TestPrintGitLog_WhenCommitHasBranch_PrintsExtraBranchCommits(t *testing.T) {
	assert := assert.New(t)
	testutil.InitTest(slog.LevelInfo)

	testutil.AddCommit("first", "")

	CreateNewPr(true, "", util.GetMainBranchOrDie(), templates.GetBranchInfo("", IndicatorTypeGuess))

	testutil.AddCommit("second", "")

	allCommits := GetAllCommits()

	UpdatePr(templates.GetBranchInfo(allCommits[1].Commit, IndicatorTypeCommit), []string{}, IndicatorTypeCommit)

	outWriter := testutil.NewWriteRecorder()
	PrintGitLog(outWriter)
	out := outWriter.String()

	allCommits = GetAllCommits()
	assert.Equal("1. ✅ "+ex.Yellow+allCommits[0].Commit+ex.Reset+" first\n"+
		"      - second\n"+
		"      - first\n",
		out)
}

func TestSdLog_LogsOutput(t *testing.T) {
	assert := assert.New(t)

	testutil.InitTest(slog.LevelInfo)

	testutil.AddCommit("first", "")

	out := testParseArguments("log")

	assert.Contains(out, "first")
}

func TestGitlog_WhenManyCommits_PadsFirstCommits(t *testing.T) {
	assert := assert.New(t)

	testutil.InitTest(slog.LevelInfo)

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

	testutil.InitTest(slog.LevelInfo)

	testutil.AddCommit("first", "")
	testutil.AddCommit("second", "")

	testParseArguments("new", "2")
	testParseArguments("new", "1")

	out := testParseArguments("log")

	assert.Regexp("✅.*first", out)
	assert.Regexp("✅.*second", out)
}
