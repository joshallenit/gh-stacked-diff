package main

import (
	"log"
	"strings"
)

/*
replace the current commit during a rebase with the diff of a branch
*/
func main() {
	var commitWithConflicts = getCommitWithConflicts()
	branchInfo := GetBranchInfo(commitWithConflicts)
	Execute("git", "reset", "--hard", "HEAD")
	log.Println("Replacing HEAD for commit", commitWithConflicts, "with changes from branch", branchInfo.BranchName)
	diff := ExecuteWithOptions(ExecuteOptions{TrimSpace: false}, "git", "diff", "--binary", "origin/main", branchInfo.BranchName)
	ExecuteWithOptions(ExecuteOptions{Stdin: &diff}, "git", "apply")
	log.Println("Adding changes and continuing rebase")
	Execute("git", "add", ".")
	continueOptions := ExecuteOptions{EnvironmentVariables: make([]string, 1)}
	continueOptions.EnvironmentVariables[0] = "GIT_EDITOR=true"
	ExecuteWithOptions(continueOptions, "git", "rebase", "--continue")
}

func getCommitWithConflicts() string {
	statusLines := strings.Split(Execute("git", "status"), "\n")
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
