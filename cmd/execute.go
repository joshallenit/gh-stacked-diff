package main

import (
	"log"
	"os"
	"os/exec"
	"runtime/debug"
	"strings"
)

var Red = "\033[31m"
var Reset = "\033[0m"
var White = "\033[97m"
var mainBranchName string

type ExecuteOptions struct {
	TrimSpace    bool
	IncludeStack bool
	Stdin        *string
	// For example "MY_VAR=some_value"
	EnvironmentVariables []string
}

func Execute(programName string, args ...string) string {
	return ExecuteWithOptions(ExecuteOptions{TrimSpace: true, IncludeStack: true}, programName, args...)
}

func ExecuteWithOptions(options ExecuteOptions, programName string, args ...string) string {
	cmd := exec.Command(programName, args...)
	if options.EnvironmentVariables != nil {
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, options.EnvironmentVariables...)
	}
	if options.Stdin != nil {
		cmd.Stdin = strings.NewReader(*options.Stdin)
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		if options.IncludeStack {
			debug.PrintStack()
		}
		log.Fatal(Red+"Failed executing `", programName, " ", strings.Join(args, " "), "`: "+Reset, string(out), err)
	}
	if options.TrimSpace {
		return strings.TrimSpace(string(out))
	} else {
		return string(out)
	}
}

func ExecuteFailable(programName string, args ...string) (string, error) {
	cmd := exec.Command(programName, args...)
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func GetMainBranch() string {
	if mainBranchName == "" {
		if _, mainErr := ExecuteFailable("git", "rev-parse", "--verify", "main"); mainErr != nil {
			if _, masterErr := ExecuteFailable("git", "rev-parse", "--verify", "master"); masterErr != nil {
				log.Fatal(Red + "Could not find main or master branch" + Reset)
			}
			mainBranchName = "master"
		} else {
			mainBranchName = "main"
		}
	}
	return mainBranchName
}
