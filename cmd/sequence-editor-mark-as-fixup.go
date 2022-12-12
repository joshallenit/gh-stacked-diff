package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

/*
 */
func main() {
	if len(os.Args) < 3 {
		fmt.Printf("Usage: sequence-editor-mark-as-fixup targetCommit fixupCommit1 [fixupCommit2...] rebaseFilename")
		os.Exit(1)
	}
	targetCommit := os.Args[1]
	fixupCommits := os.Args[2 : len(os.Args)-1]
	rebaseFilename := os.Args[len(os.Args)-1]

	log.Println("Got args", os.Args)

	data, err := os.ReadFile(rebaseFilename)

	if err != nil {
		log.Fatal("Could not open ", rebaseFilename, err)
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
		log.Fatal("Could only find ", i, " of ", len(fixupCommits), " fixup commits ", fixupCommits, " in ", lines)
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
		log.Fatal(err)
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
