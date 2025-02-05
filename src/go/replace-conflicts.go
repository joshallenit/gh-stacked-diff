package stackeddiff

import (
	"fmt"
	"log/slog"
	ex "stackeddiff/execute"
	"strings"
)

func ReplaceConflicts() {
	commitWithConflicts := getCommitWithConflicts()
	branchInfo := GetBranchInfo(commitWithConflicts, IndicatorTypeCommit)
	Stash("replace-conflicts (you can discard this entry unless you called replace-conflicts by mistake")
	slog.Info(fmt.Sprint("Replacing changes (merge conflicts) for failed rebase of commit ", commitWithConflicts, ", with changes from associated branch, ", branchInfo.BranchName))
	diff := ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "diff", "--binary", "origin/"+ex.GetMainBranch(), branchInfo.BranchName)
	ex.ExecuteOrDie(ex.ExecuteOptions{Stdin: &diff}, "git", "apply")
	slog.Info("Adding changes and continuing rebase")
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "add", ".")
	continueOptions := ex.ExecuteOptions{EnvironmentVariables: make([]string, 1)}
	continueOptions.EnvironmentVariables[0] = "GIT_EDITOR=true"
	ex.ExecuteOrDie(continueOptions, "git", "rebase", "--continue")
}

func getCommitWithConflicts() string {
	statusLines := strings.Split(ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "status"), "\n")
	lastCommandDoneLine := -1
	inLast := false
	for i, line := range statusLines {
		if strings.HasPrefix(line, "Last ") {
			// find last pick line
			inLast = true
		} else if inLast {
			if strings.HasPrefix(line, "   ") {
				lastCommandDoneLine = i
			} else {
				break
			}
		}
	}
	if lastCommandDoneLine == -1 {
		panic("Cannot determine which commit is being rebased with because \"git status\" does not have a \"Last commands done\" section. To use this command you must be in the middle of a rebase")
	}
	// Return the 2nd field, from a string such as "pick f52e867 next1"
	return strings.Fields(statusLines[lastCommandDoneLine])[1]
}
