package main

import (
	"os"
	sd "stacked-diff-workflow/src/stacked-diff"
)

/*
Outputs abbreviated git log that only shows what has changed, useful for copying commit hashes.
Adds a checkmark beside commits that have an associated branch.
*/
func main() {
	sd.PrintGitLog(os.Stdout)
}
