package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	sd "stackeddiff"

	ex "stackeddiff/execute"
)

func main() {
	var logFlags int
	flag.IntVar(&logFlags, "logFlags", 0, "Log flags, see https://pkg.go.dev/log#pkg-constants")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr,
			ex.Reset+"Add one or more commits to a PR.\n"+
				"\n"+
				"update-pr <pr-commit> [fixup commit (defaults to top commit)] [other fixup commit...]\n"+
				ex.White+"Flags:"+ex.Reset+"\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}
	var otherCommits []string
	if len(flag.Args()) > 1 {
		otherCommits = flag.Args()[1:]
	}
	log.SetFlags(logFlags)
	sd.UpdatePr(flag.Arg(0), otherCommits, log.Default())
}
