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
	// Whether to trim to whitespace from the output. Simplifies parsing output.
	TrimSpace bool
	// String to use for stdin. Useful for "git apply".
	Stdin *string
	// For example "MY_VAR=some_value"
	EnvironmentVariables []string
	// Whether to call [log.Fatal] on failure.
	AbortOnFailure bool
	// Whether to pipe output to stdout and stderr instead of returning it.
	PipeToStdout bool
}

func DefaultExecuteOptions() ExecuteOptions {
	return ExecuteOptions{TrimSpace: true, AbortOnFailure: true}
}

func ExecuteFailable(options ExecuteOptions, programName string, args ...string) (string, error) {
	cmd := exec.Command(programName, args...)
	if options.EnvironmentVariables != nil {
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, options.EnvironmentVariables...)
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
	if options.AbortOnFailure && err != nil {
		debug.PrintStack()
		log.Fatal(Red+"Failed executing `", programName, " ", strings.Join(args, " "), "`: "+Reset, string(out), err)
	}
	if options.TrimSpace {
		return strings.TrimSpace(string(out)), err
	} else {
		return string(out), err
	}
}

/*
Simplified [ExecuteFailable] that discards err.
*/
func Execute(options ExecuteOptions, programName string, args ...string) string {
	out, _ := ExecuteFailable(options, programName, args...)
	return out
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
