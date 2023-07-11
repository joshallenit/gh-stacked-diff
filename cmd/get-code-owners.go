package main

import (
	"log"
	"os"
	"strings"

	"github.com/hairyhenderson/go-codeowners"
)

func ChangedFilesOwners() string {
	return getCodeOwners(getChangedFiles())
}

/*
Returns changed files against main.
*/
func getChangedFiles() []string {
	filenamesRaw := Execute("git", "--no-pager", "log", "origin/"+GetMainBranch()+"..HEAD", "--pretty=format:\"\"", "--name-only")
	return strings.Split(filenamesRaw, "\n")
}

func getCodeOwners(filenames []string) string {
	ownerString := ""
	if cwd, err := os.Getwd(); err != nil {
		log.Fatal("Could not get cwd ", err)
	} else {
		if c, err := codeowners.FromFile(cwd); err != nil {
			log.Println("Note: Could not calculate code owners:", err)
		} else {
			for _, filename := range filenames {
				owners := c.Owners(filename)
				for i, o := range owners {
					if len(owners) > 1 {
						ownerString += "Owner #" + string(i) + " of " + filename + " is " + o + "\n"
					} else {
						ownerString += "Owner of " + filename + " is " + o + "\n"
					}
				}
			}
		}
	}
	return ownerString
}
