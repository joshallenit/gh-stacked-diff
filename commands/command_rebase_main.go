package commands

import (
	"flag"

	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/joshallenit/gh-stacked-diff/v2/templates"
	"github.com/joshallenit/gh-stacked-diff/v2/util"
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
		OnSelected: func(appConfig util.AppConfig, command Command) {
			if flagSet.NArg() != 0 {
				commandError(appConfig, flagSet, "too many arguments", command.Usage)
			}
			rebaseMain(appConfig)
		}}
}

// Bring local main branch up to date with remote
func rebaseMain(appConfig util.AppConfig) {
	util.RequireMainBranch()
	shouldPopStash := util.Stash("rebase-main")

	slog.Info("Fetching...")
	util.ExecuteOrDie(util.ExecuteOptions{Io: appConfig.Io}, "git", "fetch")
	slog.Info("Getting merged branches from Github...")
	mergedBranches := getMergedBranches()
	slog.Debug(fmt.Sprint("mergedBranches ", mergedBranches))
	localLogs := templates.GetNewCommits("HEAD")
	dropCommits := getDropCommits(localLogs, mergedBranches)
	slog.Info("Rebasing...")
	var rebaseError error
	if len(dropCommits) > 0 {
		environmentVariables := []string{
			"GIT_SEQUENCE_EDITOR=" + appConfig.AppExecutable + " sequence-editor-drop-already-merged " +
				strings.Join(util.MapSlice(dropCommits, func(gitLog templates.GitLog) string {
					return gitLog.Commit
				}), " ")}
		options := util.ExecuteOptions{
			EnvironmentVariables: environmentVariables,
			Io:                   appConfig.Io,
		}
		_, rebaseError = util.Execute(options, "git", "rebase", "-i", "origin/"+util.GetMainBranchOrDie())
		slog.Info("Deleting merged branches...")
		deleteBranches(appConfig.Io, dropCommits, mergedBranches)
	} else {
		options := util.ExecuteOptions{Io: appConfig.Io}
		_, rebaseError = util.Execute(options, "git", "rebase", "origin/"+util.GetMainBranchOrDie())
	}
	if rebaseError != nil {
		slog.Warn("Rebase failed, check output ^^ for details. Continue rebase manually.")
	} else {
		util.PopStash(shouldPopStash)
	}
}

func getMergedBranches() []templates.GitLog {
	mergedBranchesRaw := util.ExecuteOrDie(util.ExecuteOptions{},
		"gh", "pr", "list", "--author", "@me", "--state", "merged", "--base", util.GetMainBranchOrDie(),
		"--json", "headRefName,mergeCommit", "--jq", ".[ ] | .headRefName + \" \" +  .mergeCommit.oid")
	mergedBranchesRawLines := strings.Split(strings.TrimSpace(mergedBranchesRaw), "\n")
	mergedBranches := make([]templates.GitLog, 0, len(mergedBranchesRawLines))
	for _, mergedBranchRawLine := range mergedBranchesRawLines {
		fields := strings.Fields(mergedBranchRawLine)
		if len(fields) != 2 {
			break
		}
		// Checking for ancestor is more reliable than filtering on merge date via "gh pr list --search".
		_, mergeBaseErr := util.Execute(util.ExecuteOptions{}, "git", "merge-base", "--is-ancestor", fields[1], "HEAD")
		if mergeBaseErr != nil {
			// Not an ancestor, so it was merged after the first origin commit.
			mergedBranches = append(mergedBranches, templates.GitLog{Branch: fields[0], Commit: fields[1], Subject: ""})
		}
	}
	return mergedBranches
}

func getDropCommits(localLogs []templates.GitLog, mergedBranches []templates.GitLog) []templates.GitLog {
	// Look for matching summaries between localLogs and mergedBranches
	var dropCommits []templates.GitLog
	for _, localLog := range localLogs {
		if slices.ContainsFunc(mergedBranches, func(mergedBranch templates.GitLog) bool {
			return mergedBranch.Branch == localLog.Branch
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

func deleteBranches(stdIo util.StdIo, dropCommits []templates.GitLog, mergedCommits []templates.GitLog) {
	for _, dropCommit := range dropCommits {
		if out, err := util.Execute(util.ExecuteOptions{Io: stdIo}, "git", "branch", "-D", dropCommit.Branch); err == nil {
			fmt.Fprint(stdIo.Out, out)
		}
		index := slices.IndexFunc(mergedCommits, func(mergedCommit templates.GitLog) bool {
			return mergedCommit.Branch == dropCommit.Branch
		})
		println("index ", index)
		if index != -1 {
			println(mergedCommits[index].Commit, getBranchLatestCommit("origin/"+dropCommit.Branch))
			// Only delete remote branch if it is on the same commit to avoid accidentally deleting
			// a branch that is not merged.
			if mergedCommits[index].Commit == getBranchLatestCommit("origin/"+dropCommit.Branch) {
				util.Execute(util.ExecuteOptions{Io: stdIo}, "git", "push", "--delete", "origin", dropCommit.Branch)
			}
		}
	}
}

// Returns full commit hash of branch with name of branchName, or "" if no such branch.
func getBranchLatestCommit(branchName string) string {
	out, err := util.Execute(util.ExecuteOptions{}, "git", "log", "-n", "1", "--pretty=format:%H", branchName)
	if err != nil {
		return ""
	} else {
		return strings.TrimSpace(out)
	}
}
