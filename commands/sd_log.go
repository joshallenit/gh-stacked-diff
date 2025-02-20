package commands

import (
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/fatih/color"
)

// Prints changes in the current branch compared to the main branch to out.
func PrintGitLog(out io.Writer) {
	if GetCurrentBranchName() != GetMainBranchOrDie() {
		gitArgs := []string{"--no-pager", "log", "--pretty=oneline", "--abbrev-commit"}
		if RemoteHasBranch(GetMainBranchOrDie()) {
			gitArgs = append(gitArgs, "origin/"+GetMainBranchOrDie()+"..HEAD")
		}
		gitArgs = append(gitArgs, "--color=always")
		ex.ExecuteOrDie(ex.ExecuteOptions{Output: &ex.ExecutionOutput{Stdout: out, Stderr: out}}, "git", gitArgs...)
		return
	}
	logs := getNewCommits("HEAD")
	gitBranchArgs := make([]string, 0, len(logs)+2)
	gitBranchArgs = append(gitBranchArgs, "branch", "-l")
	for _, log := range logs {
		gitBranchArgs = append(gitBranchArgs, log.Branch)
	}
	checkedBranches := strings.Fields(ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", gitBranchArgs...))
	for i, log := range logs {
		numberPrefix := getNumberPrefix(i, len(logs))
		if slices.Contains(checkedBranches, log.Branch) {
			// Use color for ✅ otherwise in Git Bash on Windows it will appear as black and white.
			fmt.Fprint(out, numberPrefix+color.GreenString("✅ "))
		} else {
			fmt.Fprint(out, numberPrefix+"   ")
		}
		fmt.Fprintln(out, ex.Yellow+log.Commit+ex.Reset+" "+log.Subject)
		// find first commit that is not in main branch
		if slices.Contains(checkedBranches, log.Branch) {
			branchCommits := getNewCommits(log.Branch)
			if len(branchCommits) > 1 {
				for _, branchCommit := range branchCommits {
					padding := strings.Repeat(" ", len(numberPrefix))
					fmt.Fprintln(out, padding+"   - "+branchCommit.Subject)
				}
			}
		}
	}
}

func getNumberPrefix(i int, numLogs int) string {
	maxIndex := fmt.Sprint(numLogs)
	currentIndex := fmt.Sprint(i + 1)
	padding := strings.Repeat(" ", len(maxIndex)-len(currentIndex))
	return padding + currentIndex + ". "
}
