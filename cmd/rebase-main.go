package main

import (
	"flag"
	"log"
	"strings"
)

/*
Find out if any of the commits have already been merged and automatically drop
them to avoid having to deal with merge conflicts that have already been fixed
in main.
*/
func main() {
	var logFlags int
	flag.IntVar(&logFlags, "logFlags", 0, "Log flags, see https://pkg.go.dev/log#pkg-constants")
	flag.Parse()
	log.SetFlags(logFlags)

	RequireMainBranch()
	Stash("rebase-main")

	fetchOptions := ExecuteOptions{
		PipeToStdout:   true,
		AbortOnFailure: true,
	}
	Execute(fetchOptions, "git", "fetch")
	username := GetUsername()
	originCommits := strings.Split(Execute(AbortOnFailureOptions(), "git", "--no-pager", "log", "origin/main", "-n", "30", "--format=%s", "--author="+username), "\n")
	localCommits := strings.Split(Execute(AbortOnFailureOptions(), "git", "--no-pager", "log", "origin/"+GetMainBranch()+"..HEAD", "--format=%h %s"), "\n")
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
			PipeToStdout:         true,
			AbortOnFailure:       false,
		}
		Execute(options, "git", "rebase", "-i", "origin/main")
	} else {
		options := ExecuteOptions{
			PipeToStdout:   true,
			AbortOnFailure: false,
		}
		Execute(options, "git", "rebase", "origin/main")
	}
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if strings.HasPrefix(v, str+" (#") {
			return true
		}
	}
	return false
}
