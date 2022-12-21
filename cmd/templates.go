package main

import (
	"bytes"
	_ "embed"
	"log"
	"os"
	"regexp"
	"strings"
	"text/template"
)

//go:embed config/branch-name.template
var branchNameTemplateText string

//go:embed config/pr-title.template
var prTitleTemplateText string

//go:embed config/pr-description.template
var prDescriptionTemplateText string

type BranchInfo struct {
	CommitHash string
	BranchName string
}

type PullRequestText struct {
	Title       string
	Description string
}

type branchTemplateData struct {
	Username             string
	CommitSummaryCleaned string
}

type templateData struct {
	TicketNumber               string
	Username                   string
	CommitBody                 string
	CommitSummary              string
	CommitSummaryCleaned       string
	CommitSummaryWithoutTicket string
	FeatureFlag                string
}

func GetBranchInfo(commitOrBranch string) BranchInfo {
	if commitOrBranch == "" {
		commitOrBranch = "main"
	}
	result, _ := ExecuteFailable("git", "cat-file", "-t", commitOrBranch)
	if result == "commit" {
		return BranchInfo{CommitHash: commitOrBranch, BranchName: GetBranchForCommit(commitOrBranch)}
	} else {
		branchName := Execute("gh", "pr", "view", commitOrBranch, "--json", "headRefName", "-q", ".headRefName")
		prCommit := Execute("gh", "pr", "view", commitOrBranch, "--json", "commits", "-q", "[.commits[].oid] | first")
		summary := Execute("git", "show", "--no-patch", "--format=%s", prCommit)
		thisBranchCommit := Execute("git", "log", "--grep", summary, "--format=%h")
		if thisBranchCommit == "" {
			log.Fatal("Could not find associated commit for PR (\"", summary, "\") in main")
		}
		return BranchInfo{CommitHash: thisBranchCommit, BranchName: branchName}
	}
}

func GetBranchForCommit(commit string) string {
	return runTemplate("branch-name.template", branchNameTemplateText, getBranchTemplateData(commit))
}

func GetPullRequestText(commitHash string, featureFlag string) PullRequestText {
	data := getTemplateData(commitHash, featureFlag)
	title := runTemplate("pr-title.template", prTitleTemplateText, data)
	description := runTemplate("pr-description.template", prDescriptionTemplateText, data)
	return PullRequestText{Description: description, Title: title}
}

func runTemplate(configFilename string, defaultTemplateText string, data any) string {
	configFile := getConfigFile(configFilename)
	var parsed *template.Template
	var err error
	if configFile != nil {
		parsed, err = template.ParseFiles(*configFile)
		if err != nil {
			log.Fatal("Could not parse ", *configFile, err)
		}
	} else {
		parsed, err = template.New("").Parse(defaultTemplateText)
		if err != nil {
			log.Fatal("Could not parse ", defaultTemplateText, err)
		}
	}
	var output bytes.Buffer
	parsed.Execute(&output, data)
	return output.String()
}

func getTemplateData(commitHash string, featureFlag string) templateData {
	commitSummary := Execute("git", "--no-pager", "show", "--no-patch", "--format=%s", commitHash)
	commitBody := Execute("git", "--no-pager", "show", "--no-patch", "--format=%b", commitHash)
	commitSummaryCleaned := Execute("git", "show", "--no-patch", "--format=%f", commitHash)
	expression := regexp.MustCompile("^(\\S+[[:digit:]]+ )?(.*)")
	summaryMatches := expression.FindStringSubmatch(commitSummary)
	return templateData{
		Username:                   getUsername(),
		TicketNumber:               strings.TrimSpace(summaryMatches[1]),
		CommitBody:                 commitBody,
		CommitSummary:              commitSummary,
		CommitSummaryWithoutTicket: summaryMatches[2],
		CommitSummaryCleaned:       commitSummaryCleaned,
		FeatureFlag:                featureFlag,
	}
}

func getBranchTemplateData(commitHash string) branchTemplateData {
	return branchTemplateData{
		Username:             getUsername(),
		CommitSummaryCleaned: Execute("git", "show", "--no-patch", "--format=%f", commitHash),
	}
}

func getUsername() string {
	email := Execute("git", "config", "user.email")
	return email[0:strings.Index(email, "@")]
}

func getConfigFile(filenameWithoutPath string) *string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Could not get home dir", err)
	}
	fullPath := home + "/.stacked-diff-workflow/" + filenameWithoutPath
	if _, err := os.Stat(fullPath); err == nil {
		return &fullPath
	} else {
		return nil
	}
}

// Returns first commit of the given branch that is on origin/main.
func FirstOriginMainCommit(branchName string) string {
	allNewCommits := strings.Fields(Execute("git", "--no-pager", "log", "origin/main.."+branchName, "--pretty=format:%h", "--abbrev-commit"))
	if len(allNewCommits) == 0 {
		log.Fatal("No commits on ", branchName, "other than what is on main, nothing to create a commit from")
	}
	return allNewCommits[len(allNewCommits)-1] + "~1"
}

func RequireMainBranch() {
	if GetCurrentBranchName() != "main" {
		log.Fatal("Must be run from main branch")
	}
}

func GetCurrentBranchName() string {
	return Execute("git", "rev-parse", "--abbrev-ref", "HEAD")
}
