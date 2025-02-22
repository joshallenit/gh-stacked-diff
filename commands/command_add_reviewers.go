package commands

import (
	"bufio"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/fatih/color"
	ex "github.com/joshallenit/stacked-diff/v2/execute"
	"github.com/joshallenit/stacked-diff/v2/util"

	"strings"
	"sync"
	"time"

	"github.com/joshallenit/stacked-diff/v2/templates"
)

// Next `gh pr view 83824 --json latestReviews` and ensure developer is already not approved so that the review is not dismissed
type pullRequestChecksStatus struct {
	Pending int
	Failing int
	Passing int
	Total   int
}

func createAddReviewersCommand() Command {
	flagSet := flag.NewFlagSet("add-reviewers", flag.ContinueOnError)
	var indicatorTypeString *string = addIndicatorFlag(flagSet)

	var whenChecksPass *bool = flagSet.Bool("when-checks-pass", true, "Poll until all checks pass before adding reviewers")
	var defaultPollFrequency time.Duration = 30 * time.Second
	var pollFrequency *time.Duration = flagSet.Duration("poll-frequency", defaultPollFrequency,
		"Frequency which to poll checks. For valid formats see https://pkg.go.dev/time#ParseDuration")
	reviewers, silent, minChecks := addReviewersFlags(flagSet, "Falls back to "+color.HiWhiteString("PR_REVIEWERS")+" environment variable.")

	return Command{
		FlagSet: flagSet,
		Summary: "Add reviewers to Pull Request on Github once its checks have passed",
		Description: "Add reviewers to Pull Request on Github once its checks have passed.\n" +
			"\n" +
			"If PR is marked as a Draft, it is first marked as \"Ready for Review\".",
		Usage: "sd " + flagSet.Name() + " [flags] [commitIndicator [commitIndicator]...]",
		OnSelected: func(command Command) {
			commitIndicators := flagSet.Args()
			if len(commitIndicators) == 0 {
				slog.Debug("Using main branch because commitIndicators is empty")
				commitIndicators = []string{util.GetMainBranchOrDie()}
				*indicatorTypeString = string(templates.IndicatorTypeCommit)
			}
			if *reviewers == "" {
				*reviewers = os.Getenv("PR_REVIEWERS")
				if *reviewers == "" {
					commandError(flagSet, "reviewers not specified. Use reviewers flag or set PR_REVIEWERS environment variable", command.Usage)
				}
			}
			indicatorType := checkIndicatorFlag(command, indicatorTypeString)
			addReviewersToPr(commitIndicators, indicatorType, *whenChecksPass, *silent, *minChecks, *reviewers, *pollFrequency)
		}}
}

// Adds reviewers to a PR once checks have passed via Github CLI.
func addReviewersToPr(commitIndicators []string, indicatorType templates.IndicatorType, whenChecksPass bool, silent bool, minChecks int, reviewers string, pollFrequency time.Duration) {
	if reviewers == "" {
		panic("Reviewers cannot be empty")
	}
	var wg sync.WaitGroup
	for _, commitIndicator := range commitIndicators {
		wg.Add(1)
		go checkBranch(&wg, commitIndicator, indicatorType, whenChecksPass, silent, minChecks, reviewers, pollFrequency)
	}
	wg.Wait()
}

func checkBranch(wg *sync.WaitGroup, commitIndicator string, indicatorType templates.IndicatorType, whenChecksPass bool, silent bool, minChecks int, reviewers string, pollFrequency time.Duration) {
	branchName := templates.GetBranchInfo(commitIndicator, indicatorType).Branch
	if whenChecksPass {
		for {
			summary := getChecksStatus(branchName)
			if summary.Failing > 0 {
				if !silent {
					ex.ExecuteOrDie(ex.ExecuteOptions{}, "say", "Checks failed")
				}
				slog.Error(fmt.Sprint("Checks failed for ", commitIndicator, ". "+
					"Total: ", summary.Total,
					" | Passed: ", summary.Passing,
					" | Pending: ", summary.Pending,
					" | Failed: ", summary.Failing,
					"\n"))
				os.Exit(1)
			}

			if summary.Total < minChecks {
				slog.Info(fmt.Sprint("Waiting for at least ", minChecks, " checks to be added to PR. Currently only ", summary.Total))
			} else if summary.Passing == summary.Total {
				slog.Info(fmt.Sprint("All ", summary.Total, " checks passed"))
				break
			} else if summary.Passing == 0 {
				slog.Info(fmt.Sprint("Checks pending for ", commitIndicator, ". Completed: 0%"))
			} else {
				slog.Info(fmt.Sprint("Checks pending for ", commitIndicator, ". Completed: ", int32(float32(summary.Passing)/float32(summary.Total)*100), "%"))
			}
			util.Sleep(pollFrequency)
		}
	}
	slog.Info("Marking PR as ready for review")
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "gh", "pr", "ready", branchName)
	slog.Info("Waiting 10 seconds for any automatically assigned reviewers to be added...")
	util.Sleep(10 * time.Second)
	prUrl := strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "gh", "pr", "edit", branchName, "--add-reviewer", reviewers))
	slog.Info(fmt.Sprint("Added reviewers ", reviewers, " to ", prUrl))
	wg.Done()
}

/*
 * Logic copied from https://github.com/cli/cli/blob/57fbe4f317ca7d0849eeeedb16c1abc21a81913b/api/queries_pr.go#L258-L274
 */
func getChecksStatus(branchName string) pullRequestChecksStatus {
	var summary pullRequestChecksStatus
	stateString := ex.ExecuteOrDie(ex.ExecuteOptions{}, "gh", "pr", "view", branchName, "--json", "statusCheckRollup", "--jq", ".statusCheckRollup[] | .status, .conclusion, .state")
	scanner := bufio.NewScanner(strings.NewReader(strings.TrimSpace(stateString)))
	for scanner.Scan() {
		status := scanner.Text()
		scanner.Scan()
		conclusion := scanner.Text()
		scanner.Scan()
		state := scanner.Text()
		if state == "" {
			if status == "COMPLETED" {
				state = conclusion
			} else {
				state = status
			}
		}
		switch state {
		case "SUCCESS", "NEUTRAL", "SKIPPED":
			summary.Passing++
		case "ERROR", "FAILURE", "CANCELLED", "TIMED_OUT", "ACTION_REQUIRED":
			summary.Failing++
		default: // "EXPECTED", "REQUESTED", "WAITING", "QUEUED", "PENDING", "IN_PROGRESS", "STALE"
			summary.Pending++
		}
		summary.Total++
	}
	return summary
}
