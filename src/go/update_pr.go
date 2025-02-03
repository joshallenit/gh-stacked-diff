package stackeddiff

import (
	"log/slog"
	"os"
	"strings"

	"fmt"
	ex "stackeddiff/execute"
)

func UpdatePr(destCommit BranchInfo, otherCommits []string, indicatorType IndicatorType) {
	RequireMainBranch()
	RequireCommitOnMain(destCommit.CommitHash)
	var commitsToCherryPick []string
	if len(otherCommits) > 0 {
		commitsToCherryPick = otherCommits
	} else {
		commitsToCherryPick = make([]string, 1)
		commitsToCherryPick[0] = strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "rev-parse", "--short", "HEAD"))
	}
	shouldPopStash := false
	stashResult := strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "stash", "save", "-u", "before update-pr "+destCommit.CommitHash))
	if strings.HasPrefix(stashResult, "Saved working") {
		slog.Info(fmt.Sprint(stashResult))
		shouldPopStash = true
	}
	slog.Info(fmt.Sprint("Switching to branch", destCommit.BranchName))
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "switch", destCommit.BranchName)
	slog.Info(fmt.Sprint("Fast forwarding in case there were any commits made via github web interface"))
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "fetch", "origin", destCommit.BranchName)
	forcePush := false
	if _, err := ex.Execute(ex.ExecuteOptions{}, "git", "merge", "--ff-only", "origin/"+destCommit.BranchName); err != nil {
		slog.Info(fmt.Sprint("Could not fast forward to match origin. Rebasing instead.", err))
		ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "rebase", "origin", destCommit.BranchName)
		// As we rebased, a force push may be required.
		forcePush = true
	}

	slog.Info(fmt.Sprint("Cherry picking", commitsToCherryPick))
	cherryPickArgs := make([]string, 1+len(commitsToCherryPick))
	cherryPickArgs[0] = "cherry-pick"
	for i, commit := range commitsToCherryPick {
		cherryPickArgs[i+1] = commit
	}
	_, cherryPickError := ex.Execute(ex.ExecuteOptions{}, "git", cherryPickArgs...)
	if cherryPickError != nil {
		slog.Info("First attempt at cherry-pick failed")
		ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "cherry-pick", "--abort")
		rebaseCommit := FirstOriginMainCommit(ex.GetMainBranch())
		if rebaseCommit == "" {
			panic("There is no " + ex.GetMainBranch() + " branch on origin, nothing to rebase")
		}
		slog.Info(fmt.Sprint("Rebasing with the base commit on "+ex.GetMainBranch()+" branch, ", rebaseCommit,
			", in case the local "+ex.GetMainBranch()+" was rebased with origin/"+ex.GetMainBranch()))
		rebaseOutput, rebaseError := ex.Execute(ex.ExecuteOptions{}, "git", "rebase", rebaseCommit)
		if rebaseError != nil {
			slog.Info(fmt.Sprint(ex.Red+"Could not rebase, aborting..."+ex.Reset, rebaseOutput))
			ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "rebase", "--abort")
			ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "switch", ex.GetMainBranch())
			PopStash(shouldPopStash)
			os.Exit(1)
		}
		slog.Info(fmt.Sprint("Cherry picking again", commitsToCherryPick))
		var cherryPickOutput string
		cherryPickOutput, cherryPickError = ex.Execute(ex.ExecuteOptions{}, "git", cherryPickArgs...)
		if cherryPickError != nil {
			slog.Info(fmt.Sprint(ex.Red+"Could not cherry-pick, aborting..."+ex.Reset, cherryPickArgs, cherryPickOutput, cherryPickError))
			ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "cherry-pick", "--abort")
			ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "switch", ex.GetMainBranch())
			PopStash(shouldPopStash)
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
	slog.Info("Switching back to " + ex.GetMainBranch())
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "switch", ex.GetMainBranch())
	slog.Info(fmt.Sprint("Rebasing, marking as fixup", commitsToCherryPick, "for target", destCommit.CommitHash))
	environmentVariables := []string{
		"GIT_SEQUENCE_EDITOR=" +
			sequenceEditorPath("sequence_editor_mark_as_fixup") + " " +
			destCommit.CommitHash + " " +
			strings.Join(commitsToCherryPick, " "),
	}
	options := ex.ExecuteOptions{EnvironmentVariables: environmentVariables, Output: ex.NewStandardOutput()}
	rootCommit := strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "log", "--max-parents=0", "--format=%h", "HEAD"))
	if rootCommit == destCommit.CommitHash {
		slog.Info("Rebasing root commit")
		ex.ExecuteOrDie(options, "git", "rebase", "-i", "--root")
	} else {
		ex.ExecuteOrDie(options, "git", "rebase", "-i", destCommit.CommitHash+"^")
	}
	PopStash(shouldPopStash)
}
