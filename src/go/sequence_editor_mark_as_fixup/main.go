/*
For use as a sequence editor for an interactive git rebase.
Marks commits as fixup commits.

usage: sequence_editor_mark_as_fixup targetCommit fixupCommit1 [fixupCommit2...] rebaseFilename
*/
package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

/*
 */
func main() {
	slog.Debug(fmt.Sprint("Got args ", os.Args))
	if len(os.Args) < 3 {
		fmt.Printf("usage: sequence_editor_mark_as_fixup targetCommit fixupCommit1 [fixupCommit2...] rebaseFilename")
		os.Exit(1)
	}
	targetCommit := os.Args[1]
	fixupCommits := os.Args[2 : len(os.Args)-1]
	rebaseFilename := os.Args[len(os.Args)-1]

	data, err := os.ReadFile(rebaseFilename)

	if err != nil {
		panic(fmt.Sprint("Could not open ", rebaseFilename, err))
	}

	originalText := string(data)
	var newText strings.Builder

	fixupLines := make([]string, len(fixupCommits))
	i := 0

	lines := strings.Split(strings.TrimSuffix(originalText, "\n"), "\n")
	for _, line := range lines {
		if isFixupLine(line, fixupCommits) {
			fixupLines[i] = strings.Replace(line, "pick", "fixup", 1)
			i++
		}
	}
	if i != len(fixupCommits) {
		panic(fmt.Sprint("Could only find ", i, " of ", len(fixupCommits), " fixup commits ", fixupCommits, " in ", lines))
	}
	for _, line := range lines {
		if strings.HasPrefix(line, "pick ") && strings.Index(line, targetCommit) != -1 {
			newText.WriteString(line)
			newText.WriteString("\n")
			for _, fixupLine := range fixupLines {
				newText.WriteString(fixupLine)
				newText.WriteString("\n")
			}
		} else if !isFixupLine(line, fixupCommits) {
			newText.WriteString(line)
			newText.WriteString("\n")
		}
	}

	err = os.WriteFile(rebaseFilename, []byte(newText.String()), 0)
	if err != nil {
		panic(err)
	}
}

func isFixupLine(line string, fixupCommits []string) bool {
	if !strings.HasPrefix(line, "pick ") {
		return false
	}
	for _, commit := range fixupCommits {
		if strings.Index(line, commit) != -1 {
			return true
		}
	}
	return false
}
