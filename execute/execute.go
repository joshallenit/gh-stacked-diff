/*
Provides a simple way to execute shell commands.
*/
package util

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"strings"
)

// Options for [ExecuteWithOptions].
type ExecuteOptions struct {
	// String to use for stdin. Useful for "git apply".
	Stdin *string
	// For example "MY_VAR=some_value"
	EnvironmentVariables []string
	// Where to send program output, or nil for [Execute] to return it.
	Output *ExecutionOutput
}

// Where output should go for [Execute].
type ExecutionOutput struct {
	Stdout io.Writer // For standard out.
	Stderr io.Writer // For standard error.
}

// Send outout to stdout and stderr.
func NewStandardOutput() *ExecutionOutput {
	return &ExecutionOutput{Stdout: os.Stdout, Stderr: os.Stderr}
}

// Allows swapping in a [TestExecutor] via Dependency Injection during tests.
type Executor interface {
	Execute(options ExecuteOptions, programName string, args ...string) (string, error)
}

var globalExecutor Executor = DefaultExecutor{}

// Default implementation of [Executor].
type DefaultExecutor struct{}

// Sets the executor that [Execute] will use.
func SetGlobalExecutor(executor Executor) {
	globalExecutor = executor
}

// Implementation of Execute that uses [exec.Command].
func (defaultExecutor DefaultExecutor) Execute(options ExecuteOptions, programName string, args ...string) (string, error) {
	cmd := exec.Command(programName, args...)
	if options.EnvironmentVariables != nil {
		cmd.Env = append(os.Environ(), options.EnvironmentVariables...)
	}
	if options.Stdin != nil {
		cmd.Stdin = strings.NewReader(*options.Stdin)
	}
	var out []byte
	var err error
	if options.Output != nil {
		cmd.Stdout = options.Output.Stdout
		cmd.Stderr = options.Output.Stderr
		err = cmd.Run()
	} else {
		out, err = cmd.CombinedOutput()
	}

	stringOut := string(out)
	slog.Debug("Executed " + getLogMessage(programName, args, stringOut, err))
	return stringOut, err
}

// Executes a shell program with arguments.
func Execute(options ExecuteOptions, programName string, args ...string) (string, error) {
	return globalExecutor.Execute(options, programName, args...)
}

// Executes a shell program with arguments. Panics if there is an error.
func ExecuteOrDie(options ExecuteOptions, programName string, args ...string) string {
	out, err := Execute(options, programName, args...)
	if err != nil {
		panic(fmt.Sprint(Red + "Failed executing " + Reset + getLogMessage(programName, args, out, err)))
	}
	return out
}

func getLogMessage(programName string, args []string, out string, err error) string {
	logMessage := programName + " " + strings.Join(args, " ")
	stringOut := string(out)
	if strings.TrimSpace(stringOut) != "" {
		logMessage = logMessage + "\n" + stringOut
	}
	if err != nil {
		logMessage = logMessage + "\nError " + err.Error()
		var exerr *exec.ExitError
		if errors.As(err, &exerr) {
			logMessage = logMessage + " (Exit code " + fmt.Sprint(exerr.ExitCode()) + ")"
		}
	}
	return logMessage
}
