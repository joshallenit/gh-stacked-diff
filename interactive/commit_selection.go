package interactive

import (
	"fmt"
	"slices"

	"github.com/joshallenit/gh-stacked-diff/v2/templates"

	"errors"
	"strings"

	ex "github.com/joshallenit/gh-stacked-diff/v2/execute"
)

var UserCancelled error = errors.New("User cancelled")

func GetCommitSelection(withPr bool, prompt string) (templates.GitLog, error) {
	columns := []string{"Index", "Commit", "Summary"}
	newCommits := templates.GetNewCommits("HEAD")
	gitBranchArgs := make([]string, 0, len(newCommits)+2)
	gitBranchArgs = append(gitBranchArgs, "branch", "-l")
	for _, log := range newCommits {
		gitBranchArgs = append(gitBranchArgs, log.Branch)
	}
	prBranches := strings.Fields(ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", gitBranchArgs...))

	rows := make([][]string, 0, len(newCommits))

	for i, commit := range newCommits {
		hasLocalBranch := slices.Contains(prBranches, commit.Branch)
		indexString := fmt.Sprint(i + 1)
		if hasLocalBranch {
			indexString += " ✅"
		}
		rows = append(rows, []string{indexString, commit.Commit, commit.Subject})
	}
	// so I need multi-select which is going to need a different style
	// and I'm going to need a disabled selection too... so how should that behave?
	if len(rows) == 0 {
		if withPr {
			return templates.GitLog{}, errors.New("No new commits with PRs")
		} else {
			return templates.GitLog{}, errors.New("No new commits without PRs")
		}
	}
	selected := GetTableSelection(prompt, columns, rows, false, func(row int) bool {
		hasLocalBranch := slices.Contains(prBranches, newCommits[row].Branch)
		return (withPr && hasLocalBranch) || (!withPr && !hasLocalBranch)
	})
	if selected == -1 {
		return templates.GitLog{}, UserCancelled
	}
	return newCommits[selected], nil
}
