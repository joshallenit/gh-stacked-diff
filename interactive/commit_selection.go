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

func GetCommitSelection(withPr bool) (templates.GitLog, error) {
	columns := []string{"Index", "Commit", "Summary"}
	newCommits := templates.GetNewCommits("HEAD")
	gitBranchArgs := make([]string, 0, len(newCommits)+2)
	gitBranchArgs = append(gitBranchArgs, "branch", "-l")
	for _, log := range newCommits {
		gitBranchArgs = append(gitBranchArgs, log.Branch)
	}
	prBranches := strings.Fields(ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", gitBranchArgs...))

	shownCommits := make([]templates.GitLog, 0, len(newCommits))
	rows := make([][]string, 0, len(newCommits))

	for i, commit := range newCommits {
		hasLocalBranch := slices.Contains(prBranches, commit.Branch)
		if (withPr && hasLocalBranch) || (!withPr && !hasLocalBranch) {
			shownCommits = append(shownCommits, commit)
			indexString := fmt.Sprint(i + 1)
			if withPr {
				indexString += " ✅"
			}
			rows = append(rows, []string{indexString, commit.Commit, commit.Subject})
		}
	}
	if len(rows) == 0 {
		if withPr {
			return templates.GitLog{}, errors.New("No new commits with PRs")
		} else {
			return templates.GitLog{}, errors.New("No new commits without PRs")
		}
	}
	selected := GetTableSelection(columns, rows)
	if selected == -1 {
		return templates.GitLog{}, UserCancelled
	}
	return shownCommits[selected], nil
}
