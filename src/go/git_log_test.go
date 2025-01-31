package stackeddiff

import (
	"bytes"
	"log"
	"stackeddiff/testinginit"
	"strings"
	"testing"

	ex "stackeddiff/execute"

	"github.com/stretchr/testify/assert"
)

func TestGitlog_OnEmptyRemote_PrintsLog(t *testing.T) {

	testinginit.CdTestRepo()

	testinginit.AddCommit("first", "")

	testinginit.SetTestExecutor()

	outWriter := new(bytes.Buffer)
	PrintGitLog(outWriter)
	out := outWriter.String()

	if !strings.Contains(out, "first") {
		t.Errorf("'first' should be in %s", out)
	}
}

func Test_PrintGitLog_WhenRemoteHasSomeCommits_PrintsNewLogsOnly(t *testing.T) {
	testinginit.CdTestRepo()

	testinginit.AddCommit("first", "")

	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "push", "origin", ex.GetMainBranch())

	testinginit.AddCommit("second", "")

	testinginit.SetTestExecutor()

	outWriter := new(bytes.Buffer)
	PrintGitLog(outWriter)
	out := outWriter.String()

	if strings.Contains(out, "first") {
		t.Errorf("'first' should not be in %s", out)
	}
	if !strings.Contains(out, "second") {
		t.Errorf("'second' should be in %s", out)
	}
}

func TestGitlog_WhenPrCreatedForSomeCommits_PrintsCheckForCommitsWithPrs(t *testing.T) {
	testinginit.CdTestRepo()

	testinginit.AddCommit("first", "")

	testinginit.SetTestExecutor()

	CreateNewPr(true, "", ex.GetMainBranch(), GetBranchInfo("", IndicatorTypeGuess), log.Default())

	outWriter := new(bytes.Buffer)
	PrintGitLog(outWriter)
	out := outWriter.String()

	if !strings.Contains(out, "✅") {
		t.Errorf("'✅' should be in %s", out)
	}
}

func TestGitlog_WhenNotOnMain_OnlyShowsCommitsNotOnMain(t *testing.T) {
	assert := assert.New(t)

	testinginit.CdTestRepo()

	testinginit.AddCommit("first", "")

	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "push", "origin", ex.GetMainBranch())

	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "checkout", "-b", "my-branch")

	testinginit.AddCommit("second", "")

	outWriter := new(bytes.Buffer)
	PrintGitLog(outWriter)
	out := outWriter.String()

	assert.NotContains(out, "first")
	assert.Contains(out, "second")
}

func TestGitlog_WhenCommitHasBranch_PrintsExtraBranchCommits(t *testing.T) {
	assert := assert.New(t)
	testinginit.CdTestRepo()

	testinginit.AddCommit("first", "")

	testinginit.SetTestExecutor()

	CreateNewPr(true, "", ex.GetMainBranch(), GetBranchInfo("", IndicatorTypeGuess), log.Default())

	testinginit.AddCommit("second", "")

	allCommits := GetAllCommits()

	UpdatePr(GetBranchInfo(allCommits[1].Commit, IndicatorTypeCommit), []string{}, IndicatorTypeCommit, log.Default())

	outWriter := new(bytes.Buffer)
	PrintGitLog(outWriter)
	out := outWriter.String()

	allCommits = GetAllCommits()
	assert.Equal("1. ✅ "+ex.Yellow+allCommits[0].Commit+ex.Reset+" first\n"+
		"      - second\n"+
		"      - first\n",
		out)
}
