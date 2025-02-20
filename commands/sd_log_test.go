package commands

import (
	"log/slog"
	"stackeddiff/testinginit"
	"strings"
	"testing"

	"github.com/joshallenit/stacked-diff/util"
	"github.com/stretchr/testify/assert"
)

func TestPrintGitLog_OnEmptyRemote_PrintsLog(t *testing.T) {

	testinginit.InitTest(slog.LevelInfo)

	testinginit.AddCommit("first", "")

	outWriter := testinginit.NewWriteRecorder()
	PrintGitLog(outWriter)
	out := outWriter.String()

	if !strings.Contains(out, "first") {
		t.Errorf("'first' should be in %s", out)
	}
}

func Test_PrintGitLog_WhenRemoteHasSomeCommits_PrintsNewLogsOnly(t *testing.T) {
	testinginit.InitTest(slog.LevelInfo)

	testinginit.AddCommit("first", "")

	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "push", "origin", util.GetMainBranchOrDie())

	testinginit.AddCommit("second", "")

	outWriter := testinginit.NewWriteRecorder()
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
	testinginit.InitTest(slog.LevelInfo)

	testinginit.AddCommit("first", "")

	CreateNewPr(true, "", util.GetMainBranchOrDie(), GetBranchInfo("", IndicatorTypeGuess))

	outWriter := testinginit.NewWriteRecorder()
	PrintGitLog(outWriter)
	out := outWriter.String()

	if !strings.Contains(out, "✅") {
		t.Errorf("'✅' should be in %s", out)
	}
}

func TestPrintGitLog_WhenNotOnMain_OnlyShowsCommitsNotOnMain(t *testing.T) {
	assert := assert.New(t)

	testinginit.InitTest(slog.LevelInfo)

	testinginit.AddCommit("first", "")

	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "push", "origin", util.GetMainBranchOrDie())

	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "checkout", "-b", "my-branch")

	testinginit.AddCommit("second", "")

	outWriter := testinginit.NewWriteRecorder()
	PrintGitLog(outWriter)
	out := outWriter.String()

	assert.NotContains(out, "first")
	assert.Contains(out, "second")
}

func TestPrintGitLog_WhenCommitHasBranch_PrintsExtraBranchCommits(t *testing.T) {
	assert := assert.New(t)
	testinginit.InitTest(slog.LevelInfo)

	testinginit.AddCommit("first", "")

	CreateNewPr(true, "", util.GetMainBranchOrDie(), GetBranchInfo("", IndicatorTypeGuess))

	testinginit.AddCommit("second", "")

	allCommits := GetAllCommits()

	UpdatePr(GetBranchInfo(allCommits[1].Commit, IndicatorTypeCommit), []string{}, IndicatorTypeCommit)

	outWriter := testinginit.NewWriteRecorder()
	PrintGitLog(outWriter)
	out := outWriter.String()

	allCommits = GetAllCommits()
	assert.Equal("1. ✅ "+ex.Yellow+allCommits[0].Commit+ex.Reset+" first\n"+
		"      - second\n"+
		"      - first\n",
		out)
}
