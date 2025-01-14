package stacked_diff

import (
	"encoding/csv"
	"io"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/hairyhenderson/go-codeowners"
)

func ChangedFilesOwnersString(useGithub bool) string {
	ownerString := ""
	ownedFiles := ChangedFilesOwners(useGithub, GetChangedFiles("HEAD"))
	for key, val := range ownedFiles {
		ownerString += "Owner: " + key + "\n"
		for _, filename := range val {
			ownerString += filename + "\n"
		}
		ownerString += "\n"
	}
	return ownerString
}

func ChangedFilesOwners(useGithub bool, changedFiles []string) map[string][]string {
	return getCodeOwners(useGithub, changedFiles)
}

/*
Returns changed files against main.
*/
func GetChangedFiles(commit string) []string {
	filenamesRaw := ExecuteOrDie(ExecuteOptions{}, "git", "--no-pager", "log", commit+"~.."+commit, "--pretty=format:\"\"", "--name-only")
	return strings.Split(strings.TrimSpace(filenamesRaw), "\n")
}

func getCodeOwners(useGithub bool, filenames []string) map[string][]string {
	ownedFiles := make(map[string][]string)
	githubCodeowners = nil
	customCodeowners = make(map[string][]regexp.Regexp)
	for _, filename := range filenames {
		if filename == "" || filename == "\"\"" {
			continue
		}
		var owners []string
		if useGithub {
			owners = getGithubCodeOwners(filename)
		} else {
			owners = getCustomCodeOwners(filename)
		}
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

var githubCodeowners *codeowners.Codeowners

func getGithubCodeOwners(filename string) []string {
	if githubCodeowners == nil {
		var err error
		var cwd string
		if cwd, err = os.Getwd(); err != nil {
			log.Fatal("Could not get cwd ", err)
		} else {
			if githubCodeowners, err = codeowners.FromFile(cwd); err != nil {
				log.Println("Note: Could not calculate code owners:", err)
				return nil
			}
		}
	}
	return githubCodeowners.Owners(filename)
}

var customCodeowners map[string][]regexp.Regexp

func getCustomCodeOwners(filename string) []string {
	if len(customCodeowners) == 0 {
		if csvFile, err := os.Open("config/code-ownership/code_ownership.csv"); err != nil {
			log.Fatal("Could not open csv", err)
		} else {
			reader := csv.NewReader(csvFile)
			for {
				record, err := reader.Read()
				if err == io.EOF {
					break
				}
				if err != nil {
					log.Fatal(err)
				}
				existing := customCodeowners[record[1]]
				if existing == nil {
					existing = make([]regexp.Regexp, 0)
				}

				if re, regexError := regexp.Compile("(?m)^" + record[0]); regexError != nil {
					log.Println(Red+"Warning, cannot use regex"+Reset, record[0], regexError)
				} else {
					existing = append(existing, *re)
					customCodeowners[record[1]] = existing
				}

			}
		}
	}
	for key, val := range customCodeowners {
		for _, re := range val {
			if re.MatchString(filename) {
				return []string{key}
			}
		}
	}
	return []string{}
}
