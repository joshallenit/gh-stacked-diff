package commands

import (
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/joshallenit/gh-stacked-diff/v2/templates"
	"github.com/joshallenit/gh-stacked-diff/v2/testutil"
	"github.com/joshallenit/gh-stacked-diff/v2/util"
)

func TestSdRebaseMain_WithDifferentCommits_DropsCommits(t *testing.T) {
	assert := assert.New(t)
	testExecutor := testutil.InitTest(t, slog.LevelInfo)

	testutil.AddCommit("first", "")
	testutil.AddCommit("second", "rebase-will-keep-this-file")

	util.ExecuteOrDie(util.ExecuteOptions{}, "git", "push", "origin", util.GetMainBranchOrDie())

	allOriginalCommits := templates.GetAllCommits()

	util.ExecuteOrDie(util.ExecuteOptions{}, "git", "reset", "--hard", allOriginalCommits[1].Commit)

	testutil.AddCommit("second", "rebase-will-drop-this-file")

	testExecutor.SetResponse(allOriginalCommits[0].Branch+" fakeMergeCommit",
		nil, "gh", "pr", "list", util.MatchAnyRemainingArgs)

	testParseArguments("rebase-main")

	dirEntries, err := os.ReadDir(".")
	if err != nil {
		panic("Could not read dir: " + err.Error())
	}
	assert.Equal(3, len(dirEntries))
	assert.Equal(".git", dirEntries[0].Name())
	assert.Equal("first", dirEntries[1].Name())
	assert.Equal("rebase-will-keep-this-file", dirEntries[2].Name())
}

func TestSdRebaseMain_WithMulitpleMergedBranches_DropsCommits(t *testing.T) {
	assert := assert.New(t)
	testExecutor := testutil.InitTest(t, slog.LevelInfo)

	testutil.AddCommit("first", "1")
	testutil.AddCommit("second", "2")
	testutil.AddCommit("third", "3")

	util.ExecuteOrDie(util.ExecuteOptions{}, "git", "push", "origin", util.GetMainBranchOrDie())

	allOriginalCommits := templates.GetAllCommits()

	util.ExecuteOrDie(util.ExecuteOptions{}, "git", "reset", "--hard", allOriginalCommits[2].Commit)

	testutil.AddCommit("second", "2-rebase-will-drop-this-file")
	testutil.AddCommit("third", "3-rebase-will-drop-this-file")
	testutil.AddCommit("fourth", "4")

	testExecutor.SetResponse(
		allOriginalCommits[0].Branch+" fakeMergeCommit\n"+
			allOriginalCommits[1].Branch+" fakeMergeCommit",
		nil, "gh", "pr", "list", util.MatchAnyRemainingArgs)

	testParseArguments("rebase-main")

	dirEntries, err := os.ReadDir(".")
	if err != nil {
		panic("Could not read dir: " + err.Error())
	}
	assert.Equal(5, len(dirEntries))
	assert.Equal(".git", dirEntries[0].Name())
	assert.Equal("1", dirEntries[1].Name())
	assert.Equal("2", dirEntries[2].Name())
	assert.Equal("3", dirEntries[3].Name())
	assert.Equal("4", dirEntries[4].Name())
}

func TestSdRebaseMain_WithDuplicateBranches_Panics(t *testing.T) {
	assert := assert.New(t)
	testExecutor := testutil.InitTest(t, slog.LevelInfo)

	testutil.AddCommit("first", "1")
	testutil.AddCommit("second", "2.1")
	testutil.AddCommit("second", "2.2")

	allOriginalCommits := templates.GetAllCommits()

	testExecutor.SetResponse(allOriginalCommits[0].Branch+" fakeMergeCommit",
		nil, "gh", "pr", "list", util.MatchAnyRemainingArgs)

	// Return on panic
	defer func() { _ = recover() }()

	testParseArguments("rebase-main")

	assert.Fail("did not panic with duplicate branches")
}

func TestSdRebaseMain_WhenRebaseFails_DropsBranches(t *testing.T) {
	assert := assert.New(t)
	testExecutor := testutil.InitTest(t, slog.LevelDebug)

	testutil.AddCommit("first", "file-with-conflicts")
	testutil.CommitFileChange("second", "change-value-to-avoid-same-hash", "1")
	testutil.CommitFileChange("third", "file-with-conflicts", "1")
	testParseArguments("new", "2")
	util.ExecuteOrDie(util.ExecuteOptions{}, "git", "push", "origin", util.GetMainBranchOrDie())

	allCommits := templates.GetAllCommits()
	util.ExecuteOrDie(util.ExecuteOptions{}, "git", "reset", "--hard", allCommits[2].Commit)
	// If this runs in the same second then it will generate the same commit hash, so change value.
	testutil.CommitFileChange("second", "change-value-to-avoid-same-hash", "2")
	testutil.CommitFileChange("fourth", "file-with-conflicts", "2")

	testExecutor.SetResponse(allCommits[1].Branch+" fakeMergeCommit",
		nil, "gh", "pr", "list", util.MatchAnyRemainingArgs)

	branches := util.ExecuteOrDie(util.ExecuteOptions{}, "git", "branch")
	assert.Contains(branches, "second")

	out := testParseArguments("rebase-main")

	assert.Contains(out, "Rebase failed")
	branches = util.ExecuteOrDie(util.ExecuteOptions{}, "git", "branch")
	assert.NotContains(branches, "second")
}

func TestSdRebaseMain_WithMergedPrAlreadyRebased_KeepsCommits(t *testing.T) {
	assert := assert.New(t)
	testExecutor := testutil.InitTest(t, slog.LevelInfo)

	testutil.AddCommit("first", "")
	testutil.AddCommit("second", "second-1")
	testutil.AddCommit("third", "")

	util.ExecuteOrDie(util.ExecuteOptions{}, "git", "push", "origin", util.GetMainBranchOrDie())
	allCommits := templates.GetAllCommits()
	util.ExecuteOrDie(util.ExecuteOptions{}, "git", "reset", "--hard", allCommits[1].Commit)

	testutil.AddCommit("second", "second-2")
	testParseArguments("new", "1")

	// Use the commit of the first "second" commit as the branch
	// that was merged so that the second "second" commit is not dropped.
	testExecutor.SetResponse(allCommits[1].Branch+" "+allCommits[1].Commit,
		nil, "gh", "pr", "list", util.MatchAnyRemainingArgs)

	testParseArguments("rebase-main")

	branches := util.ExecuteOrDie(util.ExecuteOptions{}, "git", "branch")
	assert.Contains(branches, "second")
}

func TestSdRebaseMain_WithDroppedCommits_DropsBranches(t *testing.T) {
	assert := assert.New(t)
	testExecutor := testutil.InitTest(t, slog.LevelInfo)

	testutil.AddCommit("first", "")
	testutil.AddCommit("second", "rebase-will-keep-this-file")

	testParseArguments("new", "1")

	util.ExecuteOrDie(util.ExecuteOptions{}, "git", "push", "origin", util.GetMainBranchOrDie())

	allOriginalCommits := templates.GetAllCommits()

	util.ExecuteOrDie(util.ExecuteOptions{}, "git", "reset", "--hard", allOriginalCommits[1].Commit)

	testutil.AddCommit("second", "rebase-will-drop-this-file")

	testExecutor.SetResponse(allOriginalCommits[0].Branch+" "+getBranchLatestCommit(allOriginalCommits[0].Branch),
		nil, "gh", "pr", "list", util.MatchAnyRemainingArgs)

	testParseArguments("rebase-main")

	assert.False(util.RemoteHasBranch(allOriginalCommits[0].Branch))
	assert.False(util.GetLocalHasBranchOrDie(allOriginalCommits[0].Branch))
}

func TestSdRebaseMain_WithSquashedMerge_DropsBranches(t *testing.T) {
	assert := assert.New(t)
	testExecutor := testutil.InitTest(t, slog.LevelInfo)

	testutil.AddCommit("first", "")
	testutil.AddCommit("second", "rebase-will-keep-this-file")

	testParseArguments("new", "1")

	util.ExecuteOrDie(util.ExecuteOptions{}, "git", "push", "origin", util.GetMainBranchOrDie())

	allOriginalCommits := templates.GetAllCommits()
	util.ExecuteOrDie(util.ExecuteOptions{}, "git", "switch", allOriginalCommits[0].Branch)

	beforeSquash := templates.GetAllCommits()
	testutil.AddCommit("fake squash commit", "")
	squashedCommit := getBranchLatestCommit("HEAD")

	util.ExecuteOrDie(util.ExecuteOptions{}, "git", "push", "origin", allOriginalCommits[0].Branch)
	util.ExecuteOrDie(util.ExecuteOptions{}, "git", "reset", "--hard", beforeSquash[0].Commit)
	util.ExecuteOrDie(util.ExecuteOptions{}, "git", "switch", util.GetMainBranchOrDie())
	util.ExecuteOrDie(util.ExecuteOptions{}, "git", "reset", "--hard", allOriginalCommits[1].Commit)

	testutil.AddCommit("second", "rebase-will-drop-this-file")

	testExecutor.SetResponse(allOriginalCommits[0].Branch+" "+squashedCommit,
		nil, "gh", "pr", "list", util.MatchAnyRemainingArgs)

	testParseArguments("rebase-main")

	assert.False(util.RemoteHasBranch(allOriginalCommits[0].Branch))
	assert.False(util.GetLocalHasBranchOrDie(allOriginalCommits[0].Branch))
}
