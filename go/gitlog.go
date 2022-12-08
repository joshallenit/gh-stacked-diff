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
	logsColor := strings.Split(Execute("git", "--no-pager", "log", "origin/main..HEAD", "--pretty=oneline", "--abbrev-commit", "--color=always"), "\n")
	logsNoColor := strings.Split(Execute("git", "--no-pager", "log", "origin/main..HEAD", "--pretty=oneline", "--abbrev-commit"), "\n")
	if len(logsNoColor) == 0 {
		return
	}
	for i, _ := range logsNoColor {
		index := strings.Index(logsNoColor[i], " ")
		if index == -1 {
			continue
		}
		commit := logsNoColor[i][0:index]
		branchInfo := GetBranchInfo(commit)
		_, err := ExecuteFailable("git", "rev-parse", "--verify", branchInfo.BranchName)
		if err != nil {
			fmt.Print("   ")
		} else {
			fmt.Print("✅ ")
		}
		fmt.Println(logsColor[i])
	}
}
