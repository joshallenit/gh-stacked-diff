package stackeddiff

import (
	"log"
	"os"
	"strings"

	ex "stackeddiff/execute"
)

func UpdatePr(commitOrPullRequest string, otherCommits []string, logger *log.Logger) {
	RequireMainBranch()
	branchInfo := GetBranchInfo(commitOrPullRequest)
	var commitsToCherryPick []string
	if len(otherCommits) > 0 {
		commitsToCherryPick = otherCommits
	} else {
		commitsToCherryPick = make([]string, 1)
		commitsToCherryPick[0] = strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "rev-parse", "--short", "HEAD"))
	}
	shouldPopStash := false
	stashResult := strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "stash", "save", "-u", "before update-pr "+commitOrPullRequest))
	if strings.HasPrefix(stashResult, "Saved working") {
		logger.Println(stashResult)
		shouldPopStash = true
	}
	logger.Println("Switching to branch", branchInfo.BranchName)
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "switch", branchInfo.BranchName)
	logger.Println("Fast forwarding in case there were any commits made via github web interface")
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "fetch", "origin", branchInfo.BranchName)
	forcePush := false
	if _, err := ex.Execute(ex.ExecuteOptions{}, "git", "merge", "--ff-only", "origin/"+branchInfo.BranchName); err != nil {
		logger.Println("Could not fast forward to match origin. Rebasing instead.", err)
		ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "rebase", "origin", branchInfo.BranchName)
		// As we rebased, a force push may be required.
		forcePush = true
	}

	logger.Println("Cherry picking", commitsToCherryPick)
	cherryPickArgs := make([]string, 1+len(commitsToCherryPick))
	cherryPickArgs[0] = "cherry-pick"
	for i, commit := range commitsToCherryPick {
		cherryPickArgs[i+1] = commit
	}
	cherryPickOutput, cherryPickError := ex.Execute(ex.ExecuteOptions{}, "git", cherryPickArgs...)
	if cherryPickError != nil {
		logger.Println("First attempt at cherry-pick failed")
		ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "cherry-pick", "--abort")
		rebaseBranch := FirstOriginMainCommit(ex.GetMainBranch())
		logger.Println("Rebasing with the base commit on "+ex.GetMainBranch()+" branch, ", rebaseBranch,
			", in case the local "+ex.GetMainBranch()+" was rebased with origin/"+ex.GetMainBranch())
		rebaseOutput, rebaseError := ex.Execute(ex.ExecuteOptions{}, "git", "rebase", rebaseBranch)
		if rebaseError != nil {
			logger.Println(ex.Red+"Could not rebase, aborting..."+ex.Reset, rebaseOutput)
			ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "rebase", "--abort")
			ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "switch", ex.GetMainBranch())
			PopStash(shouldPopStash)
			os.Exit(1)
		}
		logger.Println("Cherry picking again", commitsToCherryPick)
		cherryPickOutput, cherryPickError = ex.Execute(ex.ExecuteOptions{}, "git", cherryPickArgs...)
		if cherryPickError != nil {
			logger.Println(ex.Red+"Could not cherry-pick, aborting..."+ex.Reset, cherryPickArgs, cherryPickOutput, cherryPickError)
			ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "cherry-pick", "--abort")
			ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "switch", ex.GetMainBranch())
			PopStash(shouldPopStash)
			os.Exit(1)
		}
		forcePush = true
	}
	logger.Println("Pushing to remote")
	if forcePush {
		if _, err := ex.Execute(ex.ExecuteOptions{}, "git", "push", "origin", branchInfo.BranchName); err != nil {
			logger.Println("Regular push failed, force pushing instead.")
			ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "push", "-f", "origin", branchInfo.BranchName)
		}
	} else {
		ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "push", "origin", branchInfo.BranchName)
	}
	logger.Println("Switching back to " + ex.GetMainBranch())
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "switch", ex.GetMainBranch())
	logger.Println("Rebasing, marking as fixup", commitsToCherryPick, "for target", branchInfo.CommitHash)
	options := ex.ExecuteOptions{EnvironmentVariables: make([]string, 1), Output: ex.NewStandardOutput()}
	options.EnvironmentVariables[0] = "GIT_SEQUENCE_EDITOR=sequence_editor_mark_as_fixup " + branchInfo.CommitHash + " " + strings.Join(commitsToCherryPick, " ")
	rootCommit := strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "log", "--max-parents=0", "--format=%h", "HEAD"))
	if rootCommit == branchInfo.CommitHash {
		logger.Println("Rebasing root commit")
		ex.ExecuteOrDie(options, "git", "rebase", "-i", "--root")
	} else {
		ex.ExecuteOrDie(options, "git", "rebase", "-i", branchInfo.CommitHash+"^")
	}
	PopStash(shouldPopStash)
}
