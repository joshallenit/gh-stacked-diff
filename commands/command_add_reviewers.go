package commands

import (
	"flag"
	"fmt"
	"log/slog"

	"github.com/joshallenit/gh-stacked-diff/v2/interactive"
	"github.com/joshallenit/gh-stacked-diff/v2/util"

	"slices"
	"strings"
	"sync"
	"time"

	"github.com/joshallenit/gh-stacked-diff/v2/templates"
)

func createAddReviewersCommand() Command {
	flagSet := flag.NewFlagSet("add-reviewers", flag.ContinueOnError)
	indicatorTypeString := addIndicatorFlag(flagSet)

	whenChecksPass := flagSet.Bool("when-checks-pass", true, "Poll until all checks pass before adding reviewers")
	defaultPollFrequency := 30 * time.Second
	pollFrequency := flagSet.Duration("poll-frequency", defaultPollFrequency,
		"Frequency which to poll checks. For valid formats see https://pkg.go.dev/time#ParseDuration")
	reviewers, silent, minChecks := addReviewersFlags(flagSet)

	return Command{
		FlagSet: flagSet,
		Summary: "Add reviewers to Pull Request on Github once its checks have passed",
		Description: "Add reviewers to Pull Request on Github once its checks have passed.\n" +
			"\n" +
			"If PR is marked as a Draft, it is first marked as \"Ready for Review\".",
		Usage: "sd " + flagSet.Name() + " [flags] [commitIndicator [commitIndicator]...]",
		OnSelected: func(asyncConfig util.AsyncAppConfig, command Command) {
			selectPrsOptions := interactive.CommitSelectionOptions{
				Prompt:      "What PR do you want to add reviewers too?",
				CommitType:  interactive.CommitTypePr,
				MultiSelect: true,
			}
			targetCommits := getTargetCommits(asyncConfig.App, command, flagSet.Args(), indicatorTypeString, selectPrsOptions)
			if *reviewers == "" {
				*reviewers = interactive.UserSelection(asyncConfig)
				if *reviewers == "" {
					commandError(
						asyncConfig.App,
						flagSet,
						"reviewers not specified.",
						command.Usage)
				}
				slog.Info("Using reviewers " + *reviewers)
			} else {
				util.SetHistory(asyncConfig.App, interactive.REVIEWERS_HISTORY_FILE,
					util.AddToHistory(
						util.ReadHistory(asyncConfig.App, interactive.REVIEWERS_HISTORY_FILE), *reviewers))
			}
			addReviewersToPr(asyncConfig, targetCommits, *whenChecksPass, *silent, *minChecks, *reviewers, *pollFrequency)
		}}
}

// Adds reviewers to a PR once checks have passed via Github CLI.
func addReviewersToPr(asyncConfig util.AsyncAppConfig, targetCommits []templates.GitLog, whenChecksPass bool, silent bool, minChecks int, reviewers string, pollFrequency time.Duration) {
	if reviewers == "" {
		panic("Reviewers cannot be empty")
	}
	var wg sync.WaitGroup
	for _, targetCommit := range targetCommits {
		wg.Add(1)
		go checkBranch(asyncConfig, &wg, targetCommit, whenChecksPass, silent, minChecks, reviewers, pollFrequency)
	}
	wg.Wait()
}

func checkBranch(asyncConfig util.AsyncAppConfig, wg *sync.WaitGroup, targetCommit templates.GitLog, whenChecksPass bool, silent bool, minChecks int, reviewers string, pollFrequency time.Duration) {
	defer asyncConfig.GracefulRecover()
	if whenChecksPass {
		for {
			summary := util.GetChecksStatus(targetCommit.Branch, minChecks)
			if summary.Failing > 0 {
				if !silent {
					util.ExecuteOrDie(util.ExecuteOptions{}, "say", "Checks failed")
				}
				slog.Error(fmt.Sprint("Checks failed for ", targetCommit, ". "+
					"Total: ", summary.Total(),
					" | Passed: ", summary.Passing,
					" | Pending: ", summary.Pending,
					" | Failed: ", summary.Failing,
					"\n"))
				asyncConfig.App.Exit(1)
			}

			if summary.Total() < summary.MinChecks {
				slog.Info(fmt.Sprint("Waiting for at least ", summary.MinChecks, " checks to be added to PR. Currently only ", summary.Total()))
			} else if summary.Passing == summary.Total() {
				slog.Info(fmt.Sprint("All ", summary.Total(), " checks passed"))
				break
			} else if summary.Passing == 0 {
				slog.Info(fmt.Sprint("Checks pending for ", targetCommit, ". Completed: 0%"))
			} else {
				slog.Info(fmt.Sprint("Checks pending for ", targetCommit, ". Completed: ", int(summary.PercentageComplete()*100), "%"))
			}
			util.Sleep(pollFrequency)
		}
	}
	slog.Info("Marking PR as ready for review")
	util.ExecuteOrDie(util.ExecuteOptions{}, "gh", "pr", "ready", targetCommit.Branch)
	slog.Info("Waiting 10 seconds for any automatically assigned reviewers to be added...")
	util.Sleep(10 * time.Second)
	slog.Info("Checking if user has already approved latest commit")
	approvingUsers, nonApprovingUsers := getNonApprovingUsers(targetCommit, reviewers)
	if nonApprovingUsers != reviewers {
		slog.Warn(fmt.Sprint("Skipping reviewers that have already approved: " + approvingUsers))
	}
	if len(nonApprovingUsers) > 0 {
		prUrl := strings.TrimSpace(
			util.ExecuteOrDie(util.ExecuteOptions{},
				"gh", "pr", "edit", targetCommit.Branch, "--add-reviewer", nonApprovingUsers,
			),
		)
		slog.Info(fmt.Sprint("Added reviewers ", nonApprovingUsers, " to ", prUrl))
	}
	wg.Done()
}

func getNonApprovingUsers(commit templates.GitLog, reviewers string) (string, string) {
	allApprovingUsers := util.GetAllApprovingUsers(commit.Branch)
	approvingUsers := make([]string, 0)
	nonApprovingUsers := make([]string, 0)
	for _, reviewer := range strings.Split(reviewers, ",") {
		if slices.Contains(allApprovingUsers, reviewer) {
			approvingUsers = append(approvingUsers, reviewer)
		} else {
			nonApprovingUsers = append(nonApprovingUsers, reviewer)
		}
	}
	return strings.Join(approvingUsers, ","), strings.Join(nonApprovingUsers, ",")
}
