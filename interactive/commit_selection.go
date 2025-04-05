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

type CommitType int

const (
	CommitTypePr CommitType = iota
	CommitTypeNoPr
	CommitTypeBoth
)

type CommitSelectionOptions struct {
	CommitType  CommitType
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

	rowEnabled := func(row int) bool {
		if options.CommitType == CommitTypeBoth {
			return true
		}
		hasLocalBranch := slices.Contains(prBranches, newCommits[row].Branch)
		return (options.CommitType == CommitTypePr && hasLocalBranch) || (options.CommitType == CommitTypeNoPr && !hasLocalBranch)
	}

	hasEnabledRow := false
	for i, commit := range newCommits {
		hasLocalBranch := slices.Contains(prBranches, commit.Branch)
		indexString := fmt.Sprint(i + 1)
		paddingLen := len(fmt.Sprint(len(newCommits))) - len(indexString)
		indexString = strings.Repeat(" ", paddingLen) + indexString
		if hasLocalBranch {
			indexString += " ✅"
		}
		row := []string{indexString, commit.Commit, commit.Subject}
		if rowEnabled(i) {
			hasEnabledRow = true
		}
		rows = append(rows, row)
	}

	if !hasEnabledRow {
		switch options.CommitType {
		case CommitTypePr:
			return []templates.GitLog{}, errors.New("no new commits with PRs")
		case CommitTypeNoPr:
			return []templates.GitLog{}, errors.New("no new commits without PRs")
		case CommitTypeBoth:
			return []templates.GitLog{}, errors.New("no new commits")
		default:
			panic("Unknown commit type " + fmt.Sprint(options.CommitType))
		}
	}

	selected := GetTableSelection(options.Prompt, columns, rows, options.MultiSelect, stdIo, rowEnabled)

	selectedCommits := make([]templates.GitLog, 0, len(selected))
	// reverse the selected indexes to do cherry-picks in order.
	for _, selectedRow := range slices.Backward(selected) {
		selectedCommits = append(selectedCommits, newCommits[selectedRow])
	}
	return selectedCommits, nil
}
