package commands

import (
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	ex "github.com/joshallenit/gh-testsd3/v2/execute"
	"github.com/joshallenit/gh-testsd3/v2/testutil"
	"github.com/joshallenit/gh-testsd3/v2/util"
)

func TestSdCodeOwners_OutputsOwnersOfChangedFiles(t *testing.T) {
	assert := assert.New(t)

	testutil.InitTest(slog.LevelInfo)

	testutil.AddCommit("first", "first-not-changed")

	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "push", "origin", util.GetMainBranchOrDie())

	testutil.AddCommit("second", "second-changed")

	testutil.AddCommit("third", "third-changed")

	codeOwners := "first-not-changed firstOwners\n" +
		"second-changed secondOwners\n" +
		"third-changed thirdOwners\n"

	ex.ExecuteOrDie(ex.ExecuteOptions{}, "mkdir", "-p", ".github")
	if writeErr := os.WriteFile(".github/CODEOWNERS", []byte(codeOwners), os.ModePerm); writeErr != nil {
		panic(writeErr)
	}
	out := testParseArguments("code-owners")

	assert.Equal("Owner: secondOwners\n"+
		"second-changed\n"+
		"\n"+
		"Owner: thirdOwners\n"+
		"third-changed\n", out)
}
