package main

import (
	"fmt"
	sd "stacked-diff-workflow/cmd/stacked-diff"
	"strings"
)

/*
Outputs abbreviated git log that only shows what has changed, useful for copying commit hashes.
Adds a checkmark beside commits that have an associated branch.
*/
func main() {
	logsColorRaw := sd.ExecuteOrDie(sd.ExecuteOptions{}, "git", "--no-pager", "log", "origin/"+sd.GetMainBranch()+"..HEAD", "--pretty=oneline", "--abbrev-commit", "--color=always")
	logsColor := strings.Split(strings.TrimSpace(logsColorRaw), "\n")
	logsNoColorRaw := sd.ExecuteOrDie(sd.ExecuteOptions{}, "git", "--no-pager", "log", "origin/"+sd.GetMainBranch()+"..HEAD", "--pretty=oneline", "--abbrev-commit")
	logsNoColor := strings.Split(strings.TrimSpace(logsNoColorRaw), "\n")
	if len(logsNoColor) == 0 {
		return
	}
	if sd.GetCurrentBranchName() == sd.GetMainBranch() {
		for i, _ := range logsNoColor {
			index := strings.Index(logsNoColor[i], " ")
			if index == -1 {
				continue
			}
			commit := logsNoColor[i][0:index]

			branchName := sd.GetBranchForCommit(commit)
			checkedBranch := strings.TrimSpace(sd.ExecuteOrDie(sd.ExecuteOptions{}, "git", "branch", "-l", branchName))
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
