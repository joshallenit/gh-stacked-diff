package main

import (
	"fmt"
	"io"
	"os"
	ex "stacked-diff-workflow/src/execute"
	sd "stacked-diff-workflow/src/stacked-diff"
	"strings"
)

/*
Outputs abbreviated git log that only shows what has changed, useful for copying commit hashes.
Adds a checkmark beside commits that have an associated branch.
*/
func main() {
	PrintGitLog(os.Stdout)
}

func PrintGitLog(out io.Writer) {
	// Check that remote has main branch
	var logsColorRaw string
	var logsNoColorRaw string
	if sd.RemoteHasBranch(ex.GetMainBranch()) {
		println("here1")
		logsColorRaw = ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "--no-pager", "log", "origin/"+ex.GetMainBranch()+"..HEAD", "--pretty=oneline", "--abbrev-commit", "--color=always")
		logsNoColorRaw = ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "--no-pager", "log", "origin/"+ex.GetMainBranch()+"..HEAD", "--pretty=oneline", "--abbrev-commit")
	} else {
		println("here2")
		logsColorRaw = ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "--no-pager", "log", "--pretty=oneline", "--abbrev-commit", "--color=always")
		logsNoColorRaw = ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "--no-pager", "log", "--pretty=oneline", "--abbrev-commit")
	}
	logsColor := strings.Split(strings.TrimSpace(logsColorRaw), "\n")
	logsNoColor := strings.Split(strings.TrimSpace(logsNoColorRaw), "\n")
	if len(logsNoColor) == 0 {
		return
	}
	if sd.GetCurrentBranchName() == ex.GetMainBranch() {
		for i, _ := range logsNoColor {
			index := strings.Index(logsNoColor[i], " ")
			if index == -1 {
				continue
			}
			commit := logsNoColor[i][0:index]

			branchName := sd.GetBranchForCommit(commit)
			checkedBranch := strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "branch", "-l", branchName))
			if checkedBranch == "" {
				fmt.Fprint(out, "   ")
			} else {
				fmt.Fprint(out, "✅ ")
			}
			fmt.Fprintln(out, logsColor[i])
		}
	} else {
		fmt.Fprintln(out, logsColorRaw)
	}
}
