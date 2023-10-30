package main

import (
	"flag"
	"fmt"
)

func main() {
	var useGithub bool
	flag.BoolVar(&useGithub, "use-github", true, "Whether to use github CODEOWNERS, or instead whether to use config/code-ownership/code_ownership.csv")
	flag.Parse()
	fmt.Println(ChangedFilesOwnersString(useGithub))
}
