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

/*
Options for [ExecuteWithOptions].
*/
type ExecuteOptions struct {
	// String to use for stdin. Useful for "git apply".
	Stdin *string
	// For example "MY_VAR=some_value"
	EnvironmentVariables []string
	// Whether to pipe output to stdout and stderr instead of returning it.
	PipeToStdout bool
}

func Execute(options ExecuteOptions, programName string, args ...string) (string, error) {
	cmd := exec.Command(programName, args...)
	if options.EnvironmentVariables != nil {
		cmd.Env = append(os.Environ(), options.EnvironmentVariables...)
	}
	if options.Stdin != nil {
		cmd.Stdin = strings.NewReader(*options.Stdin)
	}
	var out []byte
	var err error
	if options.PipeToStdout {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
	} else {
		out, err = cmd.CombinedOutput()
	}
	return string(out), err
}

func ExecuteOrDie(options ExecuteOptions, programName string, args ...string) string {
	out, err := Execute(options, programName, args...)
	if err != nil {
		debug.PrintStack()
		log.Fatal(Red+"Failed executing `", programName, " ", strings.Join(args, " "), "`: "+Reset, out, err)
	}
	return out
}

func GetMainBranch() string {
	if mainBranchName == "" {
		if _, mainErr := Execute(ExecuteOptions{}, "git", "rev-parse", "--verify", "main"); mainErr != nil {
			if _, masterErr := Execute(ExecuteOptions{}, "git", "rev-parse", "--verify", "master"); masterErr != nil {
				log.Fatal(Red + "Could not find main or master branch" + Reset)
			}
			mainBranchName = "master"
		} else {
			mainBranchName = "main"
		}
	}
	return mainBranchName
}
