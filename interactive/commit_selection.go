package interactive

import (
	"fmt"

	"github.com/joshallenit/gh-stacked-diff/v2/templates"

	"github.com/joshallenit/gh-stacked-diff/v2/util"
)

func GetCommitSelection(exit func(err any)) templates.GitLog {
	columns := []string{" ", "Commit", "Summary"}
	newCommits := templates.GetNewCommits("HEAD")
	index := 0
	rows := util.MapSlice(newCommits, func(commit templates.GitLog) []string {
		index++
		return []string{fmt.Sprint(index), commit.Commit, commit.Subject}
	})
	if len(rows) == 0 {
		panic("No new commits to select from")
	}
	selected := GetTableSelection(columns, rows)
	if selected == -1 {
		exit(nil)
	}
	return newCommits[selected]
}
