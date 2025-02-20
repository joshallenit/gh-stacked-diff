package templates

import (
	"bytes"
	_ "embed"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/template"
)

//go:embed config/branch-name.template
var branchNameTemplateText string

//go:embed config/pr-title.template
var prTitleTemplateText string

//go:embed config/pr-description.template
var prDescriptionTemplateText string

// Commit and associated branch.
type BranchInfo struct {
	// Git commit hash, might be abbreviated.
	CommitHash string
	// Name of associated branch, might exist or not.
	BranchName string
}

type pullRequestText struct {
	Title       string
	Description string
}

type branchTemplateData struct {
	UsernameCleaned      string
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

// Enum for what commitIndicator represents.
type IndicatorType string

const (
	// commitIndicator is a commit hash.
	IndicatorTypeCommit IndicatorType = "commit"
	// commitIndicator is a PR number.
	IndicatorTypePr IndicatorType = "pr"
	// commitIndicator is a list index from log (1 based).
	IndicatorTypeList IndicatorType = "list"
	// Guess based on length of commitIndicator and whether it is all numeric.
	IndicatorTypeGuess IndicatorType = "guess"
)

// Returns weather the indicator type is of a known type.
func (indicator IndicatorType) IsValid() bool {
	switch indicator {
	case IndicatorTypeCommit, IndicatorTypePr, IndicatorTypeList, IndicatorTypeGuess:
		return true
	default:
		return false
	}
}

// Returns BranchInfo for commitIndicator and indicatorType.
func GetBranchInfo(commitIndicator string, indicatorType IndicatorType) BranchInfo {
	if !indicatorType.IsValid() {
		panic("Invalid IndicatorType " + string(indicatorType))
	}
	if commitIndicator == "" {
		commitIndicator = GetMainBranchOrDie()
		indicatorType = IndicatorTypeCommit
	}
	if indicatorType == IndicatorTypeGuess {
		indicatorType = guessIndicatorType(commitIndicator)
	}

	var info BranchInfo
	switch indicatorType {
	case IndicatorTypePr:
		slog.Debug("Using commitIndicator as a pull request number " + commitIndicator)

		branchName := ex.ExecuteOrDie(ex.ExecuteOptions{}, "gh", "pr", "view", commitIndicator, "--json", "headRefName", "-q", ".headRefName")
		// Fetch the branch in case the lastest commit is only on GitHub.
		ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "fetch", "origin", branchName)
		prCommit := strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "gh", "pr", "view", commitIndicator, "--json", "commits", "-q", "[.commits[].oid] | first"))
		summary := strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "show", "--no-patch", "--format=%s", prCommit))
		thisBranchCommit := strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "log", "--grep", "^"+regexp.QuoteMeta(summary)+"$", "--format=%h"))
		if thisBranchCommit == "" {
			panic(fmt.Sprint("Could not find associated commit for PR (\"", summary, "\") in "+GetMainBranchOrDie()))
		}
		info = BranchInfo{CommitHash: thisBranchCommit, BranchName: branchName}
		slog.Info("Using pull request " + commitIndicator + ", commit " + info.CommitHash + ", branch " + info.BranchName)
	case IndicatorTypeCommit:
		slog.Debug("Using commitIndicator as a commit hash " + commitIndicator)

		info = BranchInfo{CommitHash: commitIndicator, BranchName: getBranchForCommit(commitIndicator)}
		slog.Info("Using commit " + info.CommitHash + ", branch " + info.BranchName)
	case IndicatorTypeList:
		slog.Debug("Using commitIndicator as a list index " + commitIndicator)
		newCommits := getNewCommits(GetCurrentBranchName())
		listIndex, err := strconv.Atoi(commitIndicator)
		if err != nil {
			panic("When indicator type is " + string(IndicatorTypeList) + " commit indicator must be a number, given " + commitIndicator)
		}
		// list indicators are 1 based, convert to 0 based.
		listIndex--
		if listIndex >= len(newCommits) || listIndex < 0 {
			panic("list index " + fmt.Sprint(listIndex) +
				" (parsed from " + commitIndicator + ") " +
				"out of bounds for list of new commits with size " +
				fmt.Sprint(len(newCommits)))
		}
		slog.Info("Using list index " + commitIndicator + ", commit " + newCommits[listIndex].Commit + " " + newCommits[listIndex].Subject)
		info = BranchInfo{CommitHash: newCommits[listIndex].Commit, BranchName: newCommits[listIndex].Branch}
	default:
		panic("Impossible: guessIndicatorType only returns known values, " + fmt.Sprint(indicatorType))
	}
	return info
}

func guessIndicatorType(commitIndicator string) IndicatorType {
	if _, err := strconv.Atoi(commitIndicator); err == nil {
		if len(commitIndicator) < 3 {
			return IndicatorTypeList
		}
		if len(commitIndicator) < 7 {
			return IndicatorTypePr
		}
	}
	return IndicatorTypeCommit
}

func getBranchForCommit(commit string) string {
	sanitizedSubject := strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "show", "--no-patch", "--format=%f", commit))
	return getBranchForSantizedSubject(sanitizedSubject)
}

func getBranchForSantizedSubject(sanitizedSubject string) string {
	name := runTemplate("branch-name.template", branchNameTemplateText, getBranchTemplateData(sanitizedSubject))
	// Branch names that are too long cause problems with Github.
	name = truncateString(name, 120)
	return name
}

func truncateString(str string, maxBytes int) string {
	for i := range str {
		if i >= maxBytes {
			return str[:i]
		}
	}
	return str
}

func getPullRequestText(commitHash string, featureFlag string) pullRequestText {
	data := getPullRequestTemplateData(commitHash, featureFlag)
	title := runTemplate("pr-title.template", prTitleTemplateText, data)
	description := runTemplate("pr-description.template", prDescriptionTemplateText, data)
	return pullRequestText{Description: description, Title: title}
}

func runTemplate(configFilename string, defaultTemplateText string, data any) string {
	configFile := getConfigFile(configFilename)
	var parsed *template.Template
	var err error
	if configFile != nil {
		parsed, err = template.ParseFiles(*configFile)
		if err != nil {
			panic(fmt.Sprint("Could not parse ", *configFile, err))
		}
	} else {
		parsed, err = template.New("").Parse(defaultTemplateText)
		if err != nil {
			panic(fmt.Sprint("Could not parse ", defaultTemplateText, err))
		}
	}
	var output bytes.Buffer
	parsed.Execute(&output, data)
	return output.String()
}

func getPullRequestTemplateData(commitHash string, featureFlag string) templateData {
	commitSummary := strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "--no-pager", "show", "--no-patch", "--format=%s", commitHash))
	commitBody := strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "--no-pager", "show", "--no-patch", "--format=%b", commitHash))
	commitSummaryCleaned := strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "show", "--no-patch", "--format=%f", commitHash))
	expression := regexp.MustCompile("^(\\S+-[[:digit:]]+ )?(.*)")
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

func getBranchTemplateData(sanitizedSummary string) branchTemplateData {
	// Dots are not allowed in branch names of some Github configurations.
	username := strings.ReplaceAll(getUsername(), ".", "-")
	return branchTemplateData{
		UsernameCleaned:      username,
		CommitSummaryCleaned: sanitizedSummary,
	}
}

func getConfigFile(filenameWithoutPath string) *string {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Sprint("Could not get home dir", err))
	}
	fullPath := home + "/.stacked-diff-workflow/" + filenameWithoutPath
	if _, err := os.Stat(fullPath); err == nil {
		return &fullPath
	} else {
		return nil
	}
}
