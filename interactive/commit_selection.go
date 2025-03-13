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

func GetPrSelection(prompt string) (templates.GitLog, error) {
	prSelected, err := getCommitSelection(true, false, prompt)
	if err == nil {
		return prSelected[0], nil
	} else {
		return templates.GitLog{}, err
	}
}

func GetCommitSelection(prompt string) ([]templates.GitLog, error) {
	return getCommitSelection(false, true, prompt)
}

func getCommitSelection(withPr bool, multiSelect bool, prompt string) ([]templates.GitLog, error) {
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
			return []templates.GitLog{}, errors.New("No new commits with PRs")
		} else {
			return []templates.GitLog{}, errors.New("No new commits without PRs")
		}
	}
	selected := GetTableSelection(prompt, columns, rows, multiSelect, func(row int) bool {
		hasLocalBranch := slices.Contains(prBranches, newCommits[row].Branch)
		return (withPr && hasLocalBranch) || (!withPr && !hasLocalBranch)
	})
	if len(selected) == 0 {
		return []templates.GitLog{}, UserCancelled
	}
	selectedCommits := make([]templates.GitLog, len(selected))
	for i, selectedRow := range selected {
		selectedCommits[i] = newCommits[selectedRow]
	}
	return selectedCommits, nil
}
