package main

import (
	"log"
	"regexp"
	"strconv"
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
}

func getCommitWithConflicts() string {
	statusLines := strings.Split(Execute("git", "status"), "\n")
	var commandsDoneLine = -1
	for i, line := range statusLines {
		if strings.HasPrefix(line, "Last ") {
			commandsDoneLine = i
		}
	}
	if commandsDoneLine == -1 {
		log.Fatal("Cannot determine which commit is being rebased with because \"git status\" does not have a \"Last commands done\" line. To use this command you must be in the middle of a rebase")
	}
	expression := regexp.MustCompile(".*\\(([[:digit:]]+).*")
	summaryMatches := expression.FindStringSubmatch(statusLines[commandsDoneLine])
	totalCommandsDone, err := strconv.Atoi(summaryMatches[1])
	if err != nil {
		log.Fatal("Cannot parse number of done tasks from", summaryMatches)
	}
	// Return the 2nd field, from a string such as "pick f52e867 next1"
	return strings.Fields(statusLines[commandsDoneLine+totalCommandsDone])[1]
}
