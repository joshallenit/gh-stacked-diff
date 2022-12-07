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
	// So next step would be to parse a file instead of embedded text
	result := Execute("git", "cat-file", "-t", commitOrBranch)
	if result == "commit" {
		branchName := runTemplate("branch-name.template", branchNameTemplateText, getTemplateData(commitOrBranch))
		return BranchInfo{CommitHash: commitOrBranch, BranchName: branchName}
	} else {
		branchName := Execute("gh", "pr", "view", commitOrBranch, "--json", "headRefName", "-q", "'.headRefName'")
		commitHash := Execute("gh", "pr", "view", commitOrBranch, "--json", "commits", "-q", "'[.commits[].oid] | first'")
		return BranchInfo{CommitHash: commitHash, BranchName: branchName}
	}
}

func GetPullRequestText(commitHash string) PullRequestText {
	data := getTemplateData(commitHash)
	title := runTemplate("pr-title.template", prTitleTemplateText, data)
	description := runTemplate("pr-description.template", prDescriptionTemplateText, data)
	return PullRequestText{Description: description, Title: title}
}

func runTemplate(configFilename string, defaultTemplateText string, data templateData) string {
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

func getTemplateData(commitHash string) templateData {
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
		FeatureFlag:                "",
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
