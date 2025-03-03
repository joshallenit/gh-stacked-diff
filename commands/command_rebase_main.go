package commands

import (
	"flag"
	"io"

	"fmt"
	"log/slog"
	"slices"
	"strings"

	ex "github.com/joshallenit/gh-testsd3/v2/execute"
	"github.com/joshallenit/gh-testsd3/v2/templates"
	"github.com/joshallenit/gh-testsd3/v2/util"
)

func createRebaseMainCommand() Command {
	flagSet := flag.NewFlagSet("rebase-main", flag.ContinueOnError)

	return Command{
		FlagSet: flagSet,
		Summary: "Bring your main branch up to date with remote",
		Description: "Rebase with origin/" + util.GetMainBranchForHelp() + ", dropping any commits who's associated\n" +
			"branches have been merged.\n" +
			"\n" +
			"This avoids having to manually call \"git reset --hard head\" whenever\n" +
			"you have merge conflicts with a commit that has already been merged\n" +
			"but has slight variation with local main because, for example, a\n" +
			"change was made with the Github Web UI.",
		Usage: "sd " + flagSet.Name(),
		OnSelected: func(command Command, stdOut io.Writer, stdErr io.Writer, sequenceEditorPrefix string, exit func(err any)) {
			if flagSet.NArg() != 0 {
				commandError(flagSet, "too many arguments", command.Usage)
			}
			rebaseMain(sequenceEditorPrefix)
		}}
}

// Bring local main branch up to date with remote
func rebaseMain(sequenceEditorPrefix string) {
	util.RequireMainBranch()
	shouldPopStash := util.Stash("rebase-main")

	slog.Info("Fetching...")
	ex.ExecuteOrDie(ex.ExecuteOptions{Output: ex.NewStandardOutput()}, "git", "fetch")
	slog.Info("Getting merged branches from Github...")
	mergedBranches := getMergedBranches()
	slog.Debug(fmt.Sprint("mergedBranches ", mergedBranches))
	localLogs := templates.GetNewCommits("HEAD")
	dropCommits := getDropCommits(localLogs, mergedBranches)
	slog.Info("Rebasing...")
	var rebaseError error
	if len(dropCommits) > 0 {
		environmentVariables := []string{
			"GIT_SEQUENCE_EDITOR=" + sequenceEditorPrefix + "sequence-editor-drop-already-merged " +
				strings.Join(util.MapSlice(dropCommits, func(gitLog templates.GitLog) string {
					return gitLog.Commit
				}), " ")}
		options := ex.ExecuteOptions{
			EnvironmentVariables: environmentVariables,
			Output:               ex.NewStandardOutput(),
		}
		_, rebaseError = ex.Execute(options, "git", "rebase", "-i", "origin/"+util.GetMainBranchOrDie())
		slog.Info("Deleting merged branches...")
		dropBranches(dropCommits)
	} else {
		options := ex.ExecuteOptions{
			Output: ex.NewStandardOutput(),
		}
		_, rebaseError = ex.Execute(options, "git", "rebase", "origin/"+util.GetMainBranchOrDie())
	}
	if rebaseError != nil {
		slog.Warn("Rebase failed, check output ^^ for details. Continue rebase manually.")
	} else {
		util.PopStash(shouldPopStash)
	}
}

func getMergedBranches() []string {
	mergedBranchesRaw := ex.ExecuteOrDie(ex.ExecuteOptions{},
		"gh", "pr", "list", "--author", "@me", "--state", "merged", "--base", util.GetMainBranchOrDie(),
		"--json", "headRefName,mergeCommit", "--jq", ".[ ] | .headRefName + \" \" +  .mergeCommit.oid")
	mergedBranchesRawLines := strings.Split(strings.TrimSpace(mergedBranchesRaw), "\n")
	mergedBranches := make([]string, 0, len(mergedBranchesRawLines))
	for _, mergedBranchRawLine := range mergedBranchesRawLines {
		fields := strings.Fields(mergedBranchRawLine)
		if len(fields) != 2 {
			panic(fmt.Sprint("Unexpected output from gh pr list: ", mergedBranchRawLine))
		}
		// Checking for ancestor is more reliable than filtering on merge date via "gh pr list --search".
		_, mergeBaseErr := ex.Execute(ex.ExecuteOptions{}, "git", "merge-base", "--is-ancestor", fields[1], "HEAD")
		if mergeBaseErr != nil {
			// Not an ancestor, so it was merged after the first origin commit.
			mergedBranches = append(mergedBranches, fields[0])
		}
	}
	return mergedBranches
}

func getDropCommits(localLogs []templates.GitLog, mergedBranches []string) []templates.GitLog {
	// Look for matching summaries between localLogs and mergedBranches
	var dropCommits []templates.GitLog
	for _, localLog := range localLogs {
		if slices.ContainsFunc(mergedBranches, func(mergedBranch string) bool {
			return mergedBranch == localLog.Branch
		}) {
			slog.Info(fmt.Sprint("Force dropping as it was already merged: ", localLog.Commit, " ", localLog.Subject))
			dropCommits = append(dropCommits, localLog)
		}
	}
	// Verify that there is only one local commit with that hash
	checkUniqueBranches(dropCommits)
	return dropCommits
}

// panics if there are duplicate branches in dropCommits.
func checkUniqueBranches(dropCommits []templates.GitLog) {
	branchToCommit := make(map[string]string)
	for _, dropCommit := range dropCommits {
		if otherCommit, ok := branchToCommit[dropCommit.Branch]; ok {
			panic(fmt.Sprint("Multiple commits, (", dropCommit.Commit, ", ", otherCommit, "), have the same branch:\n",
				dropCommit.Branch, "\n",
				"Ensure that all the commits in the diff stack have unique commit summaries."))
		}
		branchToCommit[dropCommit.Branch] = dropCommit.Commit
	}
}

func dropBranches(dropCommits []templates.GitLog) {
	for _, dropCommit := range dropCommits {
		// Ignore any error.
		ex.Execute(ex.ExecuteOptions{Output: ex.NewStandardOutput()}, "git", "branch", "-D", dropCommit.Branch)
	}
}
