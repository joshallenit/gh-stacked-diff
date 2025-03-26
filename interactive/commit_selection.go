package interactive

import (
	"fmt"
	"slices"

	"github.com/joshallenit/gh-stacked-diff/v2/templates"
	"github.com/joshallenit/gh-stacked-diff/v2/util"

	"errors"
	"strings"

	ex "github.com/joshallenit/gh-stacked-diff/v2/execute"
)

type CommitSelectionOptions struct {
	WithPr      bool
	MultiSelect bool
	Prompt      string
}

// Returns an empty array if user cancelled.
func GetCommitSelection(stdIo util.StdIo, options CommitSelectionOptions) ([]templates.GitLog, error) {
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
		if options.WithPr {
			return []templates.GitLog{}, errors.New("No new commits with PRs")
		} else {
			return []templates.GitLog{}, errors.New("No new commits without PRs")
		}
	}
	selected := GetTableSelection(options.Prompt, columns, rows, options.MultiSelect, stdIo.In, func(row int) bool {
		hasLocalBranch := slices.Contains(prBranches, newCommits[row].Branch)
		return (options.WithPr && hasLocalBranch) || (!options.WithPr && !hasLocalBranch)
	})
	selectedCommits := make([]templates.GitLog, 0, len(selected))
	// reverse the selected indexes to do cherry-picks in order.
	for _, selectedRow := range slices.Backward(selected) {
		selectedCommits = append(selectedCommits, newCommits[selectedRow])
	}
	return selectedCommits, nil
}
