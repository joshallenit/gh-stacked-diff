package stackeddiff

import (
	"fmt"
	"log/slog"
	"slices"
	"strings"

	ex "stackeddiff/execute"
)

// Bring local main branch up to date with remote
func RebaseMain() {
	requireMainBranch()
	shouldPopStash := stash("rebase-main")

	slog.Info("Fetching...")
	ex.ExecuteOrDie(ex.ExecuteOptions{Output: ex.NewStandardOutput()}, "git", "fetch")
	slog.Info("Getting merged branches from Github...")
	mergedBranches := getMergedBranches()
	localLogs := getNewCommits("HEAD")
	dropCommits := getDropCommits(localLogs, mergedBranches)
	slog.Info("Rebasing...")
	var rebaseError error
	if len(dropCommits) > 0 {
		environmentVariables := []string{
			"GIT_SEQUENCE_EDITOR=sequence_editor_drop_already_merged " +
				strings.Join(mapSlice(dropCommits, func(gitLog GitLog) string {
					return gitLog.Commit
				}), " ")}
		options := ex.ExecuteOptions{
			EnvironmentVariables: environmentVariables,
			Output:               ex.NewStandardOutput(),
		}
		_, rebaseError = ex.Execute(options, "git", "rebase", "-i", "origin/"+GetMainBranchOrDie())
		slog.Info("Deleting merged branches...")
		dropBranches(dropCommits)
	} else {
		options := ex.ExecuteOptions{
			Output: ex.NewStandardOutput(),
		}
		_, rebaseError = ex.Execute(options, "git", "rebase", "origin/"+GetMainBranchOrDie())
	}
	if rebaseError != nil {
		popStash(shouldPopStash)
	} else {
		slog.Warn("Rebase failed, check output ^^ for details. Continue rebase manually.")
	}
}

func getMergedBranches() []string {
	mergedBranchesRaw := ex.ExecuteOrDie(ex.ExecuteOptions{},
		"gh", "pr", "list", "--author", "@me", "--state", "merged", "--base", GetMainBranchOrDie(),
		"--json", "headRefName", "--jq", ".[ ] | .headRefName")
	return strings.Split(strings.TrimSpace(mergedBranchesRaw), "\n")
}

func getDropCommits(localLogs []GitLog, mergedBranches []string) []GitLog {
	// Look for matching summaries between localLogs and mergedBranches
	var dropCommits []GitLog
	for _, localLog := range localLogs {
		if slices.ContainsFunc(mergedBranches, func(mergedBranch string) bool {
			return mergedBranch == localLog.Branch
		}) {
			slog.Info(fmt.Sprint("Force dropping as it was already merged: ", localLog.Commit, " ", localLog.Subject))
			dropCommits = append(dropCommits, localLog)
		}
	}
	// Verify that there is only one local commit with that hash
	checkUniqueBranches(dropCommits)
	return dropCommits
}

// panics if there are duplicate branches in dropCommits.
func checkUniqueBranches(dropCommits []GitLog) {
	branchToCommit := make(map[string]string)
	for _, dropCommit := range dropCommits {
		if otherCommit, ok := branchToCommit[dropCommit.Branch]; ok {
			panic(fmt.Sprint("Multiple commits, (", dropCommit.Commit, ", ", otherCommit, "), have the same branch:\n",
				dropCommit.Branch, "\n",
				"Ensure that all the commits in the diff stack have unique commit summaries."))
		}
		branchToCommit[dropCommit.Branch] = dropCommit.Commit
	}
}

func mapSlice[V, R any](slice []V, f func(V) R) []R {
	var mapped []R
	for _, s := range slice {
		mapped = append(mapped, f(s))
	}
	return mapped
}

func dropBranches(dropCommits []GitLog) {
	for _, dropCommit := range dropCommits {
		// Ignore any error.
		ex.Execute(ex.ExecuteOptions{Output: ex.NewStandardOutput()}, "git", "branch", "-D", dropCommit.Branch)
	}
}
