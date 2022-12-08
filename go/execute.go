package main

import (
	"log"
	"os/exec"
	"runtime/debug"
	"strings"
)

type ExecuteOptions struct {
	TrimSpace bool
  IncludeStack bool
  Stdin *string
}

func Execute(programName string, args ...string) string {
	return ExecuteWithOptions(ExecuteOptions{TrimSpace: true, IncludeStack: true}, programName, args...)
}

func ExecuteWithOptions(options ExecuteOptions, programName string, args ...string) string {
	cmd := exec.Command(programName, args...)
  if (options.Stdin != nil) {
    println("stdin", "x" + *options.Stdin + "x")
    cmd.Stdin = strings.NewReader(*options.Stdin)
  }
  out, err := cmd.CombinedOutput()
	if err != nil {
    if (options.IncludeStack) {
      debug.PrintStack()
    }
		log.Fatal("Failed executing ", programName, args, ": ", string(out), err)
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
	if err != nil {
		log.Println("Failed executing ", programName, args, ": ", string(out), err)
	}
	return strings.TrimSpace(string(out)), err
}
