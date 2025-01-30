package stackeddiff

import (
	"fmt"
	"io"
	"strings"

	ex "stackeddiff/execute"
)

func PrintGitLog(out io.Writer) {
	gitArgs := []string{"--no-pager", "log", "--pretty=oneline", "--abbrev-commit"}
	if RemoteHasBranch(ex.GetMainBranch()) {
		gitArgs = append(gitArgs, "origin/"+ex.GetMainBranch()+"..HEAD")
	}
	if GetCurrentBranchName() != ex.GetMainBranch() {
		gitArgs = append(gitArgs, "--color=always")
		ex.ExecuteOrDie(ex.ExecuteOptions{Output: &ex.ExecutionOutput{Stdout: out, Stderr: out}}, "git", gitArgs...)
		return
	}
	logsRaw := ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", gitArgs...)
	logs := strings.Split(strings.TrimSpace(logsRaw), "\n")
	if len(logs) == 0 {
		return
	}
	for i, log := range logs {
		index := strings.Index(log, " ")
		if index == -1 {
			continue
		}
		commit := log[0:index]

		branchName := GetBranchForCommit(commit)
		checkedBranch := strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "branch", "-l", branchName))
		numberPrefix := getNumberPrefix(i, logs)
		if checkedBranch == "" {
			fmt.Fprint(out, numberPrefix+"   ")
		} else {
			fmt.Fprint(out, numberPrefix+"✅ ")
		}

		fmt.Fprintln(out, ex.Yellow+commit+ex.Reset+" "+log[index+1:])
		// find first commit that is not in main branch
		if checkedBranch != "" {
			branchCommits := GetNewCommits(ex.GetMainBranch(), branchName)
			if len(branchCommits) > 1 {
				for _, branchCommit := range branchCommits {
					fmt.Fprintln(out, "   - "+branchCommit.Subject)
				}
			}
		}
	}
}

func getNumberPrefix(i int, logs []string) string {
	maxIndex := fmt.Sprint(len(logs))
	currentIndex := fmt.Sprint(i + 1)
	padding := strings.Repeat(" ", len(maxIndex)-len(currentIndex))
	return padding + currentIndex + ". "
}
