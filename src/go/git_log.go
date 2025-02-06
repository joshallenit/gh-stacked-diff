package stackeddiff

import (
	"fmt"
	"io"
	"slices"
	"strings"

	ex "stackeddiff/execute"
)

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
	logs := GetNewCommits(GetMainBranchOrDie(), "HEAD")
	gitBranchArgs := make([]string, 0, len(logs)+2)
	gitBranchArgs = append(gitBranchArgs, "branch", "-l")
	for _, log := range logs {
		gitBranchArgs = append(gitBranchArgs, log.Branch)
	}
	checkedBranchesRaw := ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", gitBranchArgs...)
	checkedBranches := strings.Split(strings.TrimSpace(checkedBranchesRaw), "\n")
	for i, log := range logs {
		numberPrefix := getNumberPrefix(i, len(logs))
		if slices.Contains(checkedBranches, log.Branch) {
			fmt.Fprint(out, numberPrefix+"✅ ")
		} else {
			fmt.Fprint(out, numberPrefix+"   ")
		}
		fmt.Fprintln(out, ex.Yellow+log.Commit+ex.Reset+" "+log.Subject)
		// find first commit that is not in main branch
		if slices.Contains(checkedBranches, log.Branch) {
			branchCommits := GetNewCommits(GetMainBranchOrDie(), log.Branch)
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
