package commands

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	ex "stackeddiff/execute"
	"strings"
)

// Creates a new pull request via Github CLI.
func CreateNewPr(draft bool, featureFlag string, baseBranch string, branchInfo BranchInfo) {
	requireMainBranch()
	requireCommitOnMain(branchInfo.CommitHash)

	var commitToBranchFrom string
	shouldPopStash := stash("sd new " + flag.Arg(0))
	if baseBranch == GetMainBranchOrDie() {
		commitToBranchFrom = firstOriginMainCommit(GetMainBranchOrDie())
		slog.Info(fmt.Sprint("Switching to branch ", branchInfo.BranchName, " based off commit ", commitToBranchFrom))
	} else {
		commitToBranchFrom = baseBranch
		slog.Info(fmt.Sprint("Switching to branch ", branchInfo.BranchName, " based off branch ", baseBranch))
	}
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "branch", "--no-track", branchInfo.BranchName, commitToBranchFrom)
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "switch", branchInfo.BranchName)
	slog.Info(fmt.Sprint("Cherry picking ", branchInfo.CommitHash))
	cherryPickOutput, cherryPickError := ex.Execute(ex.ExecuteOptions{}, "git", "cherry-pick", branchInfo.CommitHash)
	if cherryPickError != nil {
		slog.Info(fmt.Sprint(ex.Red+"Could not cherry-pick, aborting... "+ex.Reset, cherryPickOutput, " ", cherryPickError))
		ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "cherry-pick", "--abort")
		ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "switch", GetMainBranchOrDie())
		slog.Info(fmt.Sprint("Deleting created branch ", branchInfo.BranchName))
		ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "branch", "-D", branchInfo.BranchName)
		popStash(shouldPopStash)
		os.Exit(1)
	}
	slog.Info("Pushing to remote")
	// -u is required because in newer versions of Github CLI the upstream must be set.
	pushOutput, pushErr := ex.Execute(ex.ExecuteOptions{}, "git", "-c", "push.default=current", "push", "-f", "-u")
	if pushErr != nil {
		slog.Info(fmt.Sprint(ex.Red+"Could not push: "+ex.Reset, " ", pushOutput))
		ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "switch", GetMainBranchOrDie())
		slog.Info(fmt.Sprint("Deleting created branch ", branchInfo.BranchName))
		ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "branch", "-D", branchInfo.BranchName)
		popStash(shouldPopStash)
		os.Exit(1)
	}
	prText := getPullRequestText(branchInfo.CommitHash, featureFlag)
	slog.Info("Creating PR via gh")
	createPrOutput, createPrErr := createPr(prText, baseBranch, draft)
	if createPrErr != nil {
		slog.Info(fmt.Sprint(ex.Red+"Could not create PR: "+ex.Reset, createPrOutput, " ", createPrErr))
		ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "switch", GetMainBranchOrDie())
		slog.Info(fmt.Sprint("Deleting created branch ", branchInfo.BranchName))
		ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "branch", "-D", branchInfo.BranchName)
		popStash(shouldPopStash)
		os.Exit(1)
	} else {
		slog.Info(fmt.Sprint("Created PR ", createPrOutput))
	}
	if prViewOutput, prViewErr := ex.Execute(ex.ExecuteOptions{}, "gh", "pr", "view", "--web"); prViewErr != nil {
		slog.Info(fmt.Sprint(ex.Red+"Could not open browser to PR: "+ex.Reset, prViewOutput, " ", prViewErr))
	}
	slog.Info(fmt.Sprint("Switching back to " + GetMainBranchOrDie()))
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "switch", GetMainBranchOrDie())
	popStash(shouldPopStash)
	/*
	   This avoids this hint when using `git fetch && git-rebase origin/main` which is not appropriate for stacked diff workflow:
	   > hint: use --reapply-cherry-picks to include skipped commits
	   > hint: Disable this message with "git config advice.skippedCherryPicks false",
	*/
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "config", "advice.skippedCherryPicks", "false")
}

func createPr(prText pullRequestText, baseBranch string, draft bool) (string, error) {
	createPrArgsNoDraft := []string{"pr", "create", "--title", prText.Title, "--body", prText.Description, "--fill", "--base", baseBranch}
	createPrArgs := createPrArgsNoDraft
	if draft {
		createPrArgs = append(createPrArgs, "--draft")
	}
	createPrOutput, createPrErr := ex.Execute(ex.ExecuteOptions{}, "gh", createPrArgs...)
	if createPrErr != nil && draft && strings.Contains(createPrOutput, "Draft pull requests are not supported") {
		slog.Warn("Draft PRs not supported, trying again without draft.\nUse \"--draft=false\" to avoid this warning.")
		createPrOutput, createPrErr = ex.Execute(ex.ExecuteOptions{}, "gh", createPrArgsNoDraft...)
	}
	return createPrOutput, createPrErr
}
