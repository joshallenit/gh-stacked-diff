package commands

import (
	"bufio"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/fatih/color"

	"github.com/joshallenit/gh-stacked-diff/v2/interactive"
	"github.com/joshallenit/gh-stacked-diff/v2/util"

	"slices"
	"strings"
	"sync"
	"time"

	"github.com/joshallenit/gh-stacked-diff/v2/templates"
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
	indicatorTypeString := addIndicatorFlag(flagSet)

	whenChecksPass := flagSet.Bool("when-checks-pass", true, "Poll until all checks pass before adding reviewers")
	defaultPollFrequency := 30 * time.Second
	pollFrequency := flagSet.Duration("poll-frequency", defaultPollFrequency,
		"Frequency which to poll checks. For valid formats see https://pkg.go.dev/time#ParseDuration")
	reviewers, silent, minChecks := addReviewersFlags(flagSet, "Falls back to "+color.HiWhiteString("PR_REVIEWERS")+" environment variable.")

	return Command{
		FlagSet: flagSet,
		Summary: "Add reviewers to Pull Request on Github once its checks have passed",
		Description: "Add reviewers to Pull Request on Github once its checks have passed.\n" +
			"\n" +
			"If PR is marked as a Draft, it is first marked as \"Ready for Review\".",
		Usage: "sd " + flagSet.Name() + " [flags] [commitIndicator [commitIndicator]...]",
		OnSelected: func(appConfig util.AppConfig, command Command) {
			selectPrsOptions := interactive.CommitSelectionOptions{
				Prompt:      "What PR do you want to add reviewers too?",
				CommitType:  interactive.CommitTypePr,
				MultiSelect: true,
			}
			targetCommits := getTargetCommits(appConfig, command, flagSet.Args(), indicatorTypeString, selectPrsOptions)
			if *reviewers == "" {
				*reviewers = os.Getenv("PR_REVIEWERS")
				if *reviewers == "" {
					commandError(appConfig, flagSet, "reviewers not specified. Use reviewers flag or set PR_REVIEWERS environment variable", command.Usage)
				}
			}
			addReviewersToPr(appConfig, targetCommits, *whenChecksPass, *silent, *minChecks, *reviewers, *pollFrequency)
		}}
}

// Adds reviewers to a PR once checks have passed via Github CLI.
func addReviewersToPr(appConfig util.AppConfig, targetCommits []templates.GitLog, whenChecksPass bool, silent bool, minChecks int, reviewers string, pollFrequency time.Duration) {
	if reviewers == "" {
		panic("Reviewers cannot be empty")
	}
	var wg sync.WaitGroup
	for _, targetCommit := range targetCommits {
		wg.Add(1)
		go checkBranch(appConfig, &wg, targetCommit, whenChecksPass, silent, minChecks, reviewers, pollFrequency)
	}
	wg.Wait()
}

func checkBranch(appConfig util.AppConfig, wg *sync.WaitGroup, targetCommit templates.GitLog, whenChecksPass bool, silent bool, minChecks int, reviewers string, pollFrequency time.Duration) {
	if whenChecksPass {
		for {
			summary := getChecksStatus(targetCommit.Branch)
			if summary.Failing > 0 {
				if !silent {
					util.ExecuteOrDie(util.ExecuteOptions{}, "say", "Checks failed")
				}
				slog.Error(fmt.Sprint("Checks failed for ", targetCommit, ". "+
					"Total: ", summary.Total,
					" | Passed: ", summary.Passing,
					" | Pending: ", summary.Pending,
					" | Failed: ", summary.Failing,
					"\n"))
				appConfig.Exit(1)
			}

			if summary.Total < minChecks {
				slog.Info(fmt.Sprint("Waiting for at least ", minChecks, " checks to be added to PR. Currently only ", summary.Total))
			} else if summary.Passing == summary.Total {
				slog.Info(fmt.Sprint("All ", summary.Total, " checks passed"))
				break
			} else if summary.Passing == 0 {
				slog.Info(fmt.Sprint("Checks pending for ", targetCommit, ". Completed: 0%"))
			} else {
				slog.Info(fmt.Sprint("Checks pending for ", targetCommit, ". Completed: ", int32(float32(summary.Passing)/float32(summary.Total)*100), "%"))
			}
			util.Sleep(pollFrequency)
		}
	}
	slog.Info("Marking PR as ready for review")
	util.ExecuteOrDie(util.ExecuteOptions{}, "gh", "pr", "ready", targetCommit.Branch)
	slog.Info("Waiting 10 seconds for any automatically assigned reviewers to be added...")
	util.Sleep(10 * time.Second)
	slog.Info("Checking if user has already been reviewed")
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

/*
 * Logic copied from https://github.com/cli/cli/blob/57fbe4f317ca7d0849eeeedb16c1abc21a81913b/api/queries_pr.go#L258-L274
 */
func getChecksStatus(branchName string) pullRequestChecksStatus {
	var summary pullRequestChecksStatus
	stateString := util.ExecuteOrDie(util.ExecuteOptions{}, "gh", "pr", "view", branchName, "--json", "statusCheckRollup", "--jq", ".statusCheckRollup[] | .status, .conclusion, .state")
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

func getNonApprovingUsers(commit templates.GitLog, reviewers string) (string, string) {
	allApprovingUsers := getAllApprovingUsers(commit)
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

/*
Returns users that have already approved latest commit.

Example output of gh pr view:

$ gh pr view mybranch --json "reviews"

	{
	  "reviews": [
	    {
	      "id": "PRR_kwDODeVIac6f37Qq",
	      "author": {
	        "login": "mybestie"
	      },
	      "authorAssociation": "MEMBER",
	      "body": "",
	      "submittedAt": "2025-03-13T14:47:31Z",
	      "includesCreatedEdit": false,
	      "reactionGroups": [],
	      "state": "COMMENTED",
	      "commit": {
	        "oid": "af01bdf8eb5649956096a608717f7de5eeb97e45"
	      }
	    },
	    {
	      "id": "PRR_kwDODeVIac6f5jeG",
	      "author": {
	        "login": "myfave"
	      },
	      "authorAssociation": "MEMBER",
	      "body": "",
	      "submittedAt": "2025-03-13T16:32:44Z",
	      "includesCreatedEdit": false,
	      "reactionGroups": [],
	      "state": "APPROVED",
	      "commit": {
	        "oid": "af01bdf8eb5649956096a608717f7de5eeb97e45"
	      }
	    }
	  ]
	}
*/
func getAllApprovingUsers(commit templates.GitLog) []string {
	jq := ".reviews[] | select(.state == \"APPROVED\" and .commit.oid == \"" + commit.CommitFull + "\") | .author.login"
	approvedUsers := util.ExecuteOrDie(util.ExecuteOptions{}, "gh", "pr", "view", commit.Branch, "--json", "reviews", "--jq", jq)
	return strings.Fields(approvedUsers)
}
