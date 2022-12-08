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
	for i, _ := range logsNoColor {
		commit := logsNoColor[i][0:strings.Index(logsNoColor[i], " ")]
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
