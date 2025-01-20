package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	sd "stackeddiff"
	"strings"
	"time"

	ex "stackeddiff/execute"
)

var slackToGithubTeamMap map[string]string
var changedFiles []string

func main() {
	slackToGithubTeamMap = make(map[string]string)
	slackToGithubTeamMap["Expansion - Activation"] = "activation"
	slackToGithubTeamMap["UI Infrastructure"] = "ui-infra"
	slackToGithubTeamMap["Android Application Infrastructure"] = "infra"
	slackToGithubTeamMap["Platform - App Interactivity"] = "platform"
	slackToGithubTeamMap["Enterprise - Mobile"] = "enterprise"
	// There is no Android files team at the moment.
	slackToGithubTeamMap["Files"] = "csc"
	var shouldCreateBranches bool
	var shouldPush bool
	var createPrs bool
	var processTeam string
	flag.BoolVar(&shouldCreateBranches, "create-branches", true, "Create branches")
	flag.BoolVar(&shouldPush, "push-branches", true, "Push branches")
	flag.BoolVar(&createPrs, "create-prs", true, "Create pull requests for each branch")
	flag.StringVar(&processTeam, "process-team", "", "team to process if not all")
	flag.Parse()
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr,
			ex.Reset+"Split a commit by code owners.\n"+
				"split-by-code-owners <commit hash>\n")
		flag.PrintDefaults()
	}
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}
	changedFiles = sd.GetChangedFiles(flag.Arg(0))
	deleteFromSlice(changedFiles, "\"\"")
	if len(changedFiles) != 0 {
		log.Println("Splitting up", len(changedFiles), "files")
		branchInfo := sd.GetBranchInfo(flag.Arg(0))

		branches := createBranches(branchInfo, true, shouldCreateBranches, processTeam)
		branches = append(branches, createBranches(branchInfo, false, shouldCreateBranches, processTeam)...)
		if len(changedFiles) != 0 {
			log.Fatal("Impossible, not all files included:", len(changedFiles), changedFiles)
		}

		if shouldPush {
			originalBranch := sd.GetCurrentBranchName()
			for _, branchName := range branches {
				log.Println("Pushing branch", branchName)
				ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "switch", branchName)
				// Sleep to avoid github crapping out with kex_exchange_identification or LFS lock.
				time.Sleep(10 * time.Second)
				ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "-c", "push.default=current", "push", "-f")
				if createPrs {
					log.Println("Creating PR", branchName)
					commitHash := strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "rev-parse", "HEAD"))
					prText := sd.GetPullRequestText(commitHash, "")
					time.Sleep(10 * time.Second)
					if url, err := ex.Execute(ex.ExecuteOptions{}, "gh", "pr", "create", "--title", prText.Title, "--body", prText.Description, "--fill", "--draft"); err != nil {
						log.Println("Could not create PR", err)
					} else {
						log.Println("Created PR", url)
					}
				}
			}
			log.Println("Switching back to original branch")
			ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "switch", originalBranch)
		}
	}
}

func createBranches(branchInfo sd.BranchInfo, useGithub bool, shouldCreateBranches bool, processBranch string) []string {
	originalBranch := sd.GetCurrentBranchName()
	var branches []string
	for ownerName, filenames := range sd.ChangedFilesOwners(useGithub, changedFiles) {
		if useGithub && ownerName == "unowned" {
			continue
		}
		shortTeamName := gitTeamToShortName(ownerName)
		branchName := branchInfo.BranchName + "-for-" + shortTeamName
		branches = append(branches, branchName)
		shouldCreateThisBranch := shouldCreateBranches && (processBranch == "" || processBranch == shortTeamName)
		if shouldCreateThisBranch {
			log.Println("Creating branch", branchName, "with", len(filenames), "files")
			if _, err := ex.Execute(ex.ExecuteOptions{}, "git", "checkout", "-b", branchName); err != nil {
				if useGithub {
					log.Println("Resetting existing branch")
					ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "checkout", branchName)
					ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "reset", "--hard", "origin/"+ex.GetMainBranch())
				} else {
					log.Println("Adding to existing branch")
					ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "checkout", branchName)
					ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "reset", "--hard", "head")
				}

			} else {
				ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "reset", "--hard", "origin/"+ex.GetMainBranch())
			}
			diff := ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "diff", "--binary", branchInfo.CommitHash+"~", branchInfo.CommitHash)
			ex.ExecuteOrDie(ex.ExecuteOptions{Stdin: &diff}, "git", "apply", "--reject")
		}
		gitAddArgs := []string{"add"}
		for _, filename := range filenames {
			changedFiles = deleteFromSlice(changedFiles, filename)
			gitAddArgs = append(gitAddArgs, filename)
		}
		if shouldCreateThisBranch {
			ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", gitAddArgs...)
			summary := strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "show", "--no-patch", "--format=%s", branchInfo.CommitHash))
			ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "commit", "-m", summary+" for "+shortTeamName)
			ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "reset", "--hard", "HEAD")
			ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "clean", "--force")
		}
	}
	log.Println("Switching back to original branch")
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "switch", originalBranch)
	return branches
}

func gitTeamToShortName(gitTeamName string) string {
	// @tinyspeck/android-commons-codeowners to commons
	mapped := slackToGithubTeamMap[gitTeamName]
	if mapped != "" {
		return mapped
	}
	shortName := strings.Replace(gitTeamName, "@tinyspeck/android-", "", -1)
	shortName = strings.Replace(shortName, "-codeowners", "", -1)
	shortName = strings.Replace(shortName, "&", "and", -1)
	shortName = strings.Replace(shortName, ",", "-and-", -1)
	shortName = strings.Replace(shortName, " - ", "-", -1)
	shortName = strings.Replace(shortName, " ", "-", -1)
	shortName = strings.ToLower(shortName)
	return shortName
}

func deleteFromSlice(slice []string, elem string) []string {
	for i, next := range slice {
		if next == elem {
			modified := slice[0:i]
			if i+1 < len(slice) {
				modified = append(modified, slice[i+1:len(slice)]...)
			}
			return modified
		}
	}
	return slice
}
