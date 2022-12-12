package main

import (
	"fmt"
	"strings"
)

/*
Outputs abbreviated git log that only shows what has changed, useful for copying commit hashes.
Adds a checkmark beside commits that have an associated branch.
*/
func main() {
	logsColorRaw := Execute("git", "--no-pager", "log", "origin/main..HEAD", "--pretty=oneline", "--abbrev-commit", "--color=always")
	logsColor := strings.Split(logsColorRaw, "\n")
	logsNoColor := strings.Split(Execute("git", "--no-pager", "log", "origin/main..HEAD", "--pretty=oneline", "--abbrev-commit"), "\n")
	if len(logsNoColor) == 0 {
		return
	}
	if GetCurrentBranchName() == "main" {
		for i, _ := range logsNoColor {
			index := strings.Index(logsNoColor[i], " ")
			if index == -1 {
				continue
			}
			commit := logsNoColor[i][0:index]

			branchName := GetBranchForCommit(commit)
			checkedBranch := Execute("git", "branch", "-l", branchName)
			if checkedBranch == "" {
				fmt.Print("   ")
			} else {
				fmt.Print("✅ ")
			}
			fmt.Println(logsColor[i])
		}
	} else {
		fmt.Println(logsColorRaw)
	}
}
