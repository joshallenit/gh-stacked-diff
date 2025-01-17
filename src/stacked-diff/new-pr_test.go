package stacked_diff

import (
	"bytes"
	"log"
	ex "stacked-diff-workflow/src/execute"
	testing_init "stacked-diff-workflow/src/testing-init"
	"strings"
	"testing"
)

func Test_NewPr_OnNewRepo_CreatesPr(t *testing.T) {
	testing_init.CdTestRepo()

	testing_init.AddCommit("first", "")

	testExecutor := ex.TestExecutor{TestLogger: log.Default()}
	testExecutor.SetResponse("Ok", nil, "gh")
	ex.SetGlobalExecutor(testExecutor)

	CreateNewPr(true, "", ex.GetMainBranch(), 0, GetBranchInfo(""), log.Default())

	// Check that the PR was created
	outWriter := new(bytes.Buffer)
	PrintGitLog(outWriter)
	out := outWriter.String()

	if !strings.Contains(out, "✅") {
		t.Errorf("'✅' should be in %s", out)
	}
}
