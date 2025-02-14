package stackeddiff

import (
	"log/slog"
	"os"
	"strings"

	"fmt"
	ex "stackeddiff/execute"
	"stackeddiff/sliceutil"
)

// Add commits from main to an existing PR.
func UpdatePr(destCommit BranchInfo, otherCommits []string, indicatorType IndicatorType) {
	requireMainBranch()
	requireCommitOnMain(destCommit.CommitHash)
	var commitsToCherryPick []string
	if len(otherCommits) > 0 {
		if indicatorType == IndicatorTypeGuess || indicatorType == IndicatorTypeList {
			commitsToCherryPick = sliceutil.MapSlice(otherCommits, func(commit string) string {
				return GetBranchInfo(commit, indicatorType).CommitHash
			})
		} else {
			commitsToCherryPick = otherCommits
		}
	} else {
		commitsToCherryPick = make([]string, 1)
		commitsToCherryPick[0] = strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "rev-parse", "--short", "HEAD"))
	}
	shouldPopStash := false
	stashResult := strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "stash", "save", "-u", "before update-pr "+destCommit.CommitHash))
	if strings.HasPrefix(stashResult, "Saved working") {
		slog.Info(stashResult)
		shouldPopStash = true
	}
	slog.Info(fmt.Sprint("Switching to branch ", destCommit.BranchName))
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "switch", destCommit.BranchName)
	slog.Info("Fast forwarding in case there were any commits made via github web interface")
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "fetch", "origin", destCommit.BranchName)
	forcePush := false
	if _, err := ex.Execute(ex.ExecuteOptions{}, "git", "merge", "--ff-only", "origin/"+destCommit.BranchName); err != nil {
		slog.Info(fmt.Sprint("Could not fast forward to match origin. Rebasing instead. ", err))
		ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "rebase", "origin", destCommit.BranchName)
		// As we rebased, a force push may be required.
		forcePush = true
	}

	slog.Info(fmt.Sprint("Cherry picking ", commitsToCherryPick))
	cherryPickArgs := make([]string, 1+len(commitsToCherryPick))
	cherryPickArgs[0] = "cherry-pick"
	for i, commit := range commitsToCherryPick {
		cherryPickArgs[i+1] = commit
	}
	_, cherryPickError := ex.Execute(ex.ExecuteOptions{}, "git", cherryPickArgs...)
	if cherryPickError != nil {
		slog.Info("First attempt at cherry-pick failed")
		ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "cherry-pick", "--abort")
		rebaseCommit := firstOriginMainCommit(GetMainBranchOrDie())
		slog.Info(fmt.Sprint("Rebasing with the base commit on "+GetMainBranchOrDie()+" branch, ", rebaseCommit,
			", in case the local "+GetMainBranchOrDie()+" was rebased with origin/"+GetMainBranchOrDie()))
		rebaseOutput, rebaseError := ex.Execute(ex.ExecuteOptions{}, "git", "rebase", rebaseCommit)
		if rebaseError != nil {
			slog.Info(fmt.Sprint(ex.Red+"Could not rebase, aborting... "+ex.Reset, rebaseOutput))
			ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "rebase", "--abort")
			ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "switch", GetMainBranchOrDie())
			popStash(shouldPopStash)
			os.Exit(1)
		}
		slog.Info(fmt.Sprint("Cherry picking again ", commitsToCherryPick))
		var cherryPickOutput string
		cherryPickOutput, cherryPickError = ex.Execute(ex.ExecuteOptions{}, "git", cherryPickArgs...)
		if cherryPickError != nil {
			slog.Info(fmt.Sprint(ex.Red+"Could not cherry-pick, aborting... "+ex.Reset, cherryPickArgs, " ", cherryPickOutput, " ", cherryPickError))
			ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "cherry-pick", "--abort")
			ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "switch", GetMainBranchOrDie())
			popStash(shouldPopStash)
			os.Exit(1)
		}
		forcePush = true
	}
	slog.Info("Pushing to remote")
	if forcePush {
		if _, err := ex.Execute(ex.ExecuteOptions{}, "git", "push", "origin", destCommit.BranchName); err != nil {
			slog.Info("Regular push failed, force pushing instead.")
			ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "push", "-f", "origin", destCommit.BranchName)
		}
	} else {
		ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "push", "origin", destCommit.BranchName)
	}
	slog.Info("Switching back to " + GetMainBranchOrDie())
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "switch", GetMainBranchOrDie())
	slog.Info(fmt.Sprint("Rebasing, marking as fixup ", commitsToCherryPick, " for target ", destCommit.CommitHash))
	environmentVariables := []string{
		"GIT_SEQUENCE_EDITOR=sequence_editor_mark_as_fixup " +
			destCommit.CommitHash + " " +
			strings.Join(commitsToCherryPick, " "),
	}
	slog.Debug(fmt.Sprint("Using sequence editor ", environmentVariables))
	options := ex.ExecuteOptions{EnvironmentVariables: environmentVariables, Output: ex.NewStandardOutput()}
	rootCommit := strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "log", "--max-parents=0", "--format=%h", "HEAD"))
	if rootCommit == destCommit.CommitHash {
		slog.Info("Rebasing root commit")
		ex.ExecuteOrDie(options, "git", "rebase", "-i", "--root")
	} else {
		ex.ExecuteOrDie(options, "git", "rebase", "-i", destCommit.CommitHash+"^")
	}
	popStash(shouldPopStash)
}
