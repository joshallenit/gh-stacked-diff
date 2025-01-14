package main

import (
	"flag"
	"log"
	sd "stacked-diff-workflow/src/stacked-diff"
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

	sd.RequireMainBranch()
	sd.Stash("rebase-main")

	sd.ExecuteOrDie(sd.ExecuteOptions{
		PipeToStdout: true,
	}, "git", "fetch")
	username := sd.GetUsername()
	originCommits := strings.Split(strings.TrimSpace(sd.ExecuteOrDie(sd.ExecuteOptions{}, "git", "--no-pager", "log", "origin/main", "-n", "30", "--format=%s", "--author="+username)), "\n")
	localCommits := strings.Split(strings.TrimSpace(sd.ExecuteOrDie(sd.ExecuteOptions{}, "git", "--no-pager", "log", "origin/"+sd.GetMainBranch()+"..HEAD", "--format=%h %s")), "\n")
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
		options := sd.ExecuteOptions{
			EnvironmentVariables: environmentVariables,
			PipeToStdout:         true,
		}
		sd.ExecuteOrDie(options, "git", "rebase", "-i", "origin/main")
	} else {
		options := sd.ExecuteOptions{
			PipeToStdout: true,
		}
		sd.ExecuteOrDie(options, "git", "rebase", "origin/main")
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
