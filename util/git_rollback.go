package util

import (
	"fmt"
	"log/slog"
	"slices"
)

type restoreBranchInfo struct {
	commit string
	branch string
}
type GitRollbackManager struct {
	restoreBranches []restoreBranchInfo
	deleteBranches  []string
}

func (rollbackManager *GitRollbackManager) SaveState() {
	restoreBranch := restoreBranchInfo{
		commit: ExecuteOrDie(ExecuteOptions{}, "git", "log", "-n", "1", "--pretty=format:%H"),
		branch: GetCurrentBranchName(),
	}
	rollbackManager.restoreBranches = append(rollbackManager.restoreBranches, restoreBranch)
}

func (rollbackManager *GitRollbackManager) Restore(err any) {
	if len(rollbackManager.restoreBranches) == 0 {
		// Nothing to restore.
		return
	}
	slog.Error(fmt.Sprint(err))
	tryAbort("cherry-pick")
	tryAbort("rebase")
	tryAbort("merge")
	for _, branchInfo := range slices.Backward(rollbackManager.restoreBranches) {
		slog.Info(fmt.Sprint("Restoring branch ", branchInfo.branch, " to ", branchInfo.commit))
		ExecuteOrDie(ExecuteOptions{}, "git", "switch", branchInfo.branch)
		ExecuteOrDie(ExecuteOptions{}, "git", "reset", "--hard", branchInfo.commit)
	}
	for _, branch := range rollbackManager.deleteBranches {
		slog.Info(fmt.Sprint("Deleting created branch ", branch))
		ExecuteOrDie(ExecuteOptions{}, "git", "branch", "-D", branch)
	}
}

// Abort the given git command if it is in progress.
func tryAbort(gitCommand string) {
	_, err := Execute(ExecuteOptions{}, "git", gitCommand, "--abort")
	if err == nil {
		slog.Info(fmt.Sprint("Aborted ", gitCommand))
	}
}

func (rollbackManager *GitRollbackManager) CreatedBranch(branchName string) {
	rollbackManager.deleteBranches = append(rollbackManager.deleteBranches, branchName)
}

func (rollbackManager *GitRollbackManager) Clear() {
	rollbackManager.restoreBranches = []restoreBranchInfo{}
	rollbackManager.deleteBranches = []string{}
}
