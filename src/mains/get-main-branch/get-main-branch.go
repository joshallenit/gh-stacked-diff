package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	sd "stacked-diff-workflow/src/stacked-diff"
)

func main() {
	if len(os.Args) != 1 {
		fmt.Println("Outputs name of the main branch: main or master")
		os.Exit(1)
	}
	log.SetOutput(ioutil.Discard)
	branchName := sd.GetMainBranch()
	fmt.Print(branchName)
}
