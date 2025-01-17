package main

import (
	"flag"
	"log"
	ex "stacked-diff-workflow/src/execute"
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

	ex.ExecuteOrDie(ex.ExecuteOptions{
		Output: ex.NewStandardOutput(),
	}, "git", "fetch")
	username := sd.GetUsername()
	originCommits := strings.Split(strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "--no-pager", "log", "origin/main", "-n", "30", "--format=%s", "--author="+username)), "\n")
	localCommits := strings.Split(strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "--no-pager", "log", "origin/"+ex.GetMainBranch()+"..HEAD", "--format=%h %s")), "\n")
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
		options := ex.ExecuteOptions{
			EnvironmentVariables: environmentVariables,
			Output:               ex.NewStandardOutput(),
		}
		ex.ExecuteOrDie(options, "git", "rebase", "-i", "origin/main")
	} else {
		options := ex.ExecuteOptions{
			Output: ex.NewStandardOutput(),
		}
		ex.ExecuteOrDie(options, "git", "rebase", "origin/main")
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
