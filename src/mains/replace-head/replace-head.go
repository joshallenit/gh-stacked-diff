package main

import (
	"log"
	sd "stacked-diff-workflow/src/stacked-diff"
	"strings"
)

/*
replace the current commit during a rebase with the diff of a branch
*/
func main() {
	var commitWithConflicts = getCommitWithConflicts()
	branchInfo := sd.GetBranchInfo(commitWithConflicts)
	sd.ExecuteOrDie(sd.ExecuteOptions{}, "git", "reset", "--hard", "HEAD")
	log.Println("Replacing HEAD for commit", commitWithConflicts, "with changes from branch", branchInfo.BranchName)
	diff := sd.ExecuteOrDie(sd.ExecuteOptions{}, "git", "diff", "--binary", "origin/"+sd.GetMainBranch(), branchInfo.BranchName)
	sd.ExecuteOrDie(sd.ExecuteOptions{Stdin: &diff, PipeToStdout: true}, "git", "apply")
	log.Println("Adding changes and continuing rebase")
	sd.ExecuteOrDie(sd.ExecuteOptions{}, "git", "add", ".")
	continueOptions := sd.ExecuteOptions{EnvironmentVariables: make([]string, 1), PipeToStdout: true}
	continueOptions.EnvironmentVariables[0] = "GIT_EDITOR=true"
	sd.ExecuteOrDie(continueOptions, "git", "rebase", "--continue")
}

func getCommitWithConflicts() string {
	statusLines := strings.Split(strings.TrimSpace(sd.ExecuteOrDie(sd.ExecuteOptions{}, "git", "status")), "\n")
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
		log.Fatal("Cannot determine which commit is being rebased with because \"git status\" does not have a \"Last commands done\" section. To use this command you must be in the middle of a rebase")
	}
	// Return the 2nd field, from a string such as "pick f52e867 next1"
	return strings.Fields(statusLines[lastCommandDoneLine])[1]
}
