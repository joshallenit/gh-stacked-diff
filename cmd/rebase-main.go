package main

import (
	"log"
	"strings"
)

/*
Find out if any of the commits have already been merged and automatically drop
them to avoid having to deal with merge conflicts that have already been fixed
in main.
*/
func main() {
	RequireMainBranch()
	Stash("rebase-main")

	Execute("git", "fetch")

	username := GetUsername()
	originCommits := strings.Split(Execute("git", "--no-pager", "log", "origin/main", "-n", "30", "--format=%s", "--author="+username), "\n")
	localCommits := strings.Split(Execute("git", "--no-pager", "log", "origin/"+GetMainBranch()+"..HEAD", "--format=%h %s"), "\n")
	// Look for matching summaries between localCommits and originCommits
	var dropCommits []string
	for _, localLine := range localCommits {
		localCommit := localLine[0:strings.Index(localLine, " ")]
		localSummary := localLine[len(localCommit)+1 : len(localLine)]
		if contains(originCommits, localSummary) {
			log.Println("Force dropping as it was already merged:", localCommit, localSummary)
			dropCommits = append(dropCommits, localCommit)
		}
	}

	if len(dropCommits) > 0 {
		environmentVariables := []string{"GIT_SEQUENCE_EDITOR=sequence-editor-drop-already-merged " + strings.Join(dropCommits, " ")}
		options := ExecuteOptions{
			EnvironmentVariables: environmentVariables,
			AbortOnFailure:       false,
		}
		out := ExecuteWithOptions(options, "git", "rebase", "-i", "origin/main")
		log.Println(out)
	} else {
		out, _ := ExecuteFailable("git", "rebase", "origin/main")
		log.Println(out)
	}
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}
