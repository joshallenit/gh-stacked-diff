package stacked_diff

import (
	"bytes"
	"log"
	ex "stacked-diff-workflow/src/execute"
	testing_init "stacked-diff-workflow/src/testing-init"
	"strings"
	"testing"
)

func TestGitlog_OnEmptyRemote_PrintsLog(t *testing.T) {
	testing_init.CdTestRepo()

	testing_init.AddCommit("first", "")

	testExecutor := ex.TestExecutor{TestLogger: log.Default()}
	testExecutor.SetResponse("Ok", nil, "gh")
	ex.SetGlobalExecutor(testExecutor)

	outWriter := new(bytes.Buffer)
	PrintGitLog(outWriter)
	out := outWriter.String()

	if !strings.Contains(out, "first") {
		t.Errorf("'first' should be in %s", out)
	}
}

func Test_PrintGitLog_WhenRemoteHasSomeCommits_PrintsNewLogsOnly(t *testing.T) {
	testing_init.CdTestRepo()

	testing_init.AddCommit("first", "")

	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "push", "origin", ex.GetMainBranch())

	testing_init.AddCommit("second", "")

	testExecutor := ex.TestExecutor{TestLogger: log.Default()}
	testExecutor.SetResponse("Ok", nil, "gh")
	ex.SetGlobalExecutor(testExecutor)

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
	testing_init.CdTestRepo()

	testing_init.AddCommit("first", "")

	testExecutor := ex.TestExecutor{TestLogger: log.Default()}
	testExecutor.SetResponse("Ok", nil, "gh")
	ex.SetGlobalExecutor(testExecutor)

	CreateNewPr(true, "", ex.GetMainBranch(), 0, GetBranchInfo(""), log.Default())

	outWriter := new(bytes.Buffer)
	PrintGitLog(outWriter)
	out := outWriter.String()

	if !strings.Contains(out, "✅") {
		t.Errorf("'✅' should be in %s", out)
	}
}
