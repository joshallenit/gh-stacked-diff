package stackeddiff

import (
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strings"

	ex "stackeddiff/execute"

	"github.com/hairyhenderson/go-codeowners"
)

// Returns changed files and their owners.
func ChangedFilesOwnersString() string {
	var ownerString strings.Builder
	ownedFiles := changedFilesOwners(getChangedFiles())
	keys := make([]string, 0, len(ownedFiles))
	for k := range ownedFiles {
		keys = append(keys, k)
	}
	slices.Sort(keys)

	for i, key := range keys {
		if i > 0 {
			ownerString.WriteString("\n")
		}
		ownerString.WriteString("Owner: " + key + "\n")
		for _, filename := range ownedFiles[key] {
			ownerString.WriteString(filename + "\n")
		}
	}
	return ownerString.String()
}

func changedFilesOwners(changedFiles []string) map[string][]string {
	ownedFiles := make(map[string][]string)
	githubCodeowners = nil
	for _, filename := range changedFiles {
		if filename == "" || filename == "\"\"" {
			continue
		}
		owners := getGithubCodeOwners(filename)
		var ownersForFile string
		if len(owners) != 0 {
			for i, o := range owners {
				if i > 0 {
					ownersForFile += ","
				}
				ownersForFile += o
			}
		} else {
			ownersForFile = "unowned"
		}
		existing := ownedFiles[ownersForFile]
		if existing == nil {
			existing = make([]string, 0)
		}
		existing = append(existing, filename)
		ownedFiles[ownersForFile] = existing
	}
	return ownedFiles
}

/*
Returns changed files against main.
*/
func getChangedFiles() []string {
	gitLogArgs := []string{"--no-pager", "log", "--pretty=format:\"\"", "--name-only"}
	firstOriginCommit := firstOriginMainCommit(GetCurrentBranchName())
	if firstOriginCommit != "" {
		gitLogArgs = append(gitLogArgs, firstOriginCommit+"..HEAD")
	}
	filenamesRaw := ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", gitLogArgs...)
	return strings.Split(strings.TrimSpace(filenamesRaw), "\n")
}

var githubCodeowners *codeowners.Codeowners

func getGithubCodeOwners(filename string) []string {
	if githubCodeowners == nil {
		var err error
		if githubCodeowners, err = codeowners.FromFileWithFS(os.DirFS("."), ""); err != nil {
			slog.Info(fmt.Sprint("Could not calculate code owners: ", err))
			return []string{}
		}
	}
	return githubCodeowners.Owners(filename)
}
