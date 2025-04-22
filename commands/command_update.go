package commands

import (
	"flag"
	"log/slog"
	"strings"

	"fmt"

	"github.com/joshallenit/gh-stacked-diff/v2/interactive"
	"github.com/joshallenit/gh-stacked-diff/v2/templates"
	"github.com/joshallenit/gh-stacked-diff/v2/util"

	"slices"
	"time"
)

func createUpdateCommand() Command {
	flagSet := flag.NewFlagSet("update", flag.ContinueOnError)
	indicatorTypeString := addIndicatorFlag(flagSet)
	reviewers, silent, minChecks := addReviewersFlags(flagSet, "")
	return Command{
		FlagSet: flagSet,
		Summary: "Add commits from " + util.GetMainBranchForHelp() + " to an existing PR",
		Description: "Add commits from local " + util.GetMainBranchForHelp() + " branch to an existing PR.\n" +
			"\n" +
			"Can also add reviewers once PR checks have passed, see \"--reviewers\" flag.",
		Usage: "sd " + flagSet.Name() + " [flags] [PR commitIndicator [fixup commitIndicator [fixup commitIndicator...]]]\n" +
			"\nIf commitIndicators are missing then you will be prompted to select commits:\n" +
			"\n" +
			"   [enter]    confirms selection\n" +
			"   [space]    adds to selection when selecting commits to add\n" +
			"   [up,k]     moves cursor up\n" +
			"   [down,j]   moves cursor down\n" +
			"   [q,esc]    cancels\n",
		OnSelected: func(appConfig util.AppConfig, command Command) {
			destCommit := getDestCommit(appConfig, command, indicatorTypeString)
			commitsToCherryPick := getCommitsToCherryPick(appConfig, command, indicatorTypeString)
			if *reviewers == "" && flagSet.NArg() < 2 {
				*reviewers = interactive.UserSelection(appConfig)
				if *reviewers != "" {
					slog.Info("Using reviewers " + *reviewers)
				}
			}
			updatePr(appConfig, destCommit, commitsToCherryPick)
			if *reviewers != "" {
				addReviewersToPr(appConfig, []templates.GitLog{destCommit}, true, *silent, *minChecks, *reviewers, 30*time.Second)
			}
		}}
}

func getDestCommit(appConfig util.AppConfig, command Command, indicatorTypeString *string) templates.GitLog {
	selectPrOptions := interactive.CommitSelectionOptions{
		Prompt:      "What PR do you want to update?",
		CommitType:  interactive.CommitTypePr,
		MultiSelect: false,
	}
	targetCommits := getTargetCommits(appConfig, command, []string{command.FlagSet.Arg(0)}, indicatorTypeString, selectPrOptions)
	return targetCommits[0]
}

func getCommitsToCherryPick(appConfig util.AppConfig, command Command, indicatorTypeString *string) []templates.GitLog {
	selectCommitsOptions := interactive.CommitSelectionOptions{
		Prompt:      "What commits do you want to add?",
		CommitType:  interactive.CommitTypeNoPr,
		MultiSelect: true,
	}
	var commitsFromCommandLine []string
	if command.FlagSet.NArg() > 1 {
		commitsFromCommandLine = command.FlagSet.Args()[1:]
	}
	return getTargetCommits(appConfig, command, commitsFromCommandLine, indicatorTypeString, selectCommitsOptions)
}

// Add commits from main to an existing PR.
func updatePr(appConfig util.AppConfig, destCommit templates.GitLog, commitsToCherryPick []templates.GitLog) {
	util.RequireMainBranch()
	templates.RequireCommitOnMain(destCommit.Commit)
	checkNotMerged(appConfig, destCommit.Branch)
	shouldPopStash := util.Stash("before update-pr " + destCommit.Commit + " " + destCommit.Subject)
	rollbackManager := &util.GitRollbackManager{}
	rollbackManager.SaveState()
	util.ExecuteOrDie(util.ExecuteOptions{}, "git", "switch", destCommit.Branch)
	rollbackManager.SaveState() // Save state again for associated branch.
	defer func() {
		r := recover()
		if r != nil {
			rollbackManager.Restore(r)
		}
		util.PopStash(shouldPopStash)
		if r != nil {
			panic(r)
		}
	}()
	slog.Info("Fast forwarding in case there were any commits made via github web interface")
	util.ExecuteOrDie(util.ExecuteOptions{}, "git", "fetch", "origin", destCommit.Branch)
	forcePush := false
	if _, err := util.Execute(util.ExecuteOptions{Io: appConfig.Io}, "git", "merge", "--ff-only", "origin/"+destCommit.Branch); err != nil {
		slog.Info(fmt.Sprint("Could not fast forward to match origin. Rebasing instead. ", err))
		util.ExecuteOrDie(util.ExecuteOptions{Io: appConfig.Io}, "git", "rebase", "origin", destCommit.Branch)
		// As we rebased, a force push may be required.
		forcePush = true
	}

	slog.Info(fmt.Sprint("Cherry picking ", commitsToCherryPick))
	cherryPickArgs := make([]string, 1+len(commitsToCherryPick))
	cherryPickArgs[0] = "cherry-pick"
	for i, commit := range commitsToCherryPick {
		cherryPickArgs[i+1] = commit.Commit
	}
	_, cherryPickError := util.Execute(util.ExecuteOptions{}, "git", cherryPickArgs...)
	if cherryPickError != nil {
		slog.Info("First attempt at cherry-pick failed")
		util.ExecuteOrDie(util.ExecuteOptions{}, "git", "cherry-pick", "--abort")
		rebaseCommit := util.FirstOriginMainCommit(util.GetMainBranchOrDie())
		slog.Info(fmt.Sprint("Rebasing with the base commit on "+util.GetMainBranchOrDie()+" branch, ", rebaseCommit,
			", in case the local "+util.GetMainBranchOrDie()+" was rebased with origin/"+util.GetMainBranchOrDie()))
		util.ExecuteOrDie(util.ExecuteOptions{Io: appConfig.Io}, "git", "rebase", rebaseCommit)
		slog.Info(fmt.Sprint("Cherry picking again ", commitsToCherryPick))
		util.ExecuteOrDie(util.ExecuteOptions{Io: appConfig.Io}, "git", cherryPickArgs...)
		forcePush = true
	}
	slog.Info("Pushing to remote")
	if forcePush {
		if _, err := util.Execute(util.ExecuteOptions{}, "git", "push", "origin", destCommit.Branch); err != nil {
			slog.Info("Regular push failed, force pushing instead.")
			util.ExecuteOrDie(util.ExecuteOptions{}, "git", "push", "-f", "origin", destCommit.Branch)
		}
	} else {
		util.ExecuteOrDie(util.ExecuteOptions{}, "git", "push", "origin", destCommit.Branch)
	}
	slog.Info("Switching back to " + util.GetMainBranchOrDie())
	util.ExecuteOrDie(util.ExecuteOptions{}, "git", "switch", util.GetMainBranchOrDie())
	slog.Info(fmt.Sprint("Rebasing, marking as fixup ", commitsToCherryPick, " for target ", destCommit.Commit))
	commitHashes := util.MapSlice(commitsToCherryPick, func(commit templates.GitLog) string {
		return commit.Commit
	})
	environmentVariables := []string{
		"GIT_SEQUENCE_EDITOR=" + appConfig.AppExecutable + " sequence-editor-mark-as-fixup " +
			destCommit.Commit + " " +
			strings.Join(commitHashes, " "),
	}
	slog.Debug(fmt.Sprint("Using sequence editor ", environmentVariables))
	options := util.ExecuteOptions{EnvironmentVariables: environmentVariables, Io: appConfig.Io}
	util.ExecuteOrDie(options, "git", "rebase", "-i", destCommit.Commit+"^")
	rollbackManager.Clear()
}

func checkNotMerged(appConfig util.AppConfig, branchName string) {
	slog.Info("Checking if PR was already merged...")
	mergedBranches := getMergedBranches()
	if slices.Contains(mergedBranches, branchName) {
		interactive.ConfirmOrDie(appConfig, "It looks like this PR was already merged. Try running \"sd rebase-main\". Are you sure you want to update this PR?")
	}
}
