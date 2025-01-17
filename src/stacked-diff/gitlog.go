package stacked_diff

import (
	"fmt"
	"io"
	ex "stacked-diff-workflow/src/execute"
	"strings"
)

func PrintGitLog(out io.Writer) {
	// Check that remote has main branch

	gitArgs := []string{"--no-pager", "log", "--pretty=oneline", "--abbrev-commit", "--color=always"}
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
	for _, log := range logs {
		index := strings.Index(log, " ")
		if index == -1 {
			continue
		}
		commit := log[0:index]

		branchName := GetBranchForCommit(commit)
		checkedBranch := strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "branch", "-l", branchName))
		if checkedBranch == "" {
			fmt.Fprint(out, "   ")
		} else {
			fmt.Fprint(out, "✅ ")
		}
		fmt.Fprintln(out, ex.Yellow+log+ex.Reset)
	}
}
