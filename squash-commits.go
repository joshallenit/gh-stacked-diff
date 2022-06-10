package main

import (
    "fmt"
	"log"
    "os"
	"strings"
	"io/ioutil"
)

func main() {

    rebaseFilename := os.Args[1]

	data, err := ioutil.ReadFile(rebaseFilename)

	if err != nil {
		log.Fatal(err)
    }

	originalText := string(data)
	var newText strings.Builder

	for i, line := range strings.Split(strings.TrimSuffix(originalText, "\n"), "\n") {
		if (i > 0 && strings.HasPrefix(line, "pick ")) {
			newText.WriteString(strings.Replace(line, "pick", "squash", 1))
		} else {
			newText.WriteString(line)
		}
		newText.WriteString("\n")
	}

	fmt.Println(newText.String())

	err = os.WriteFile(rebaseFilename, []byte(newText.String()), 0644)
    if err != nil {
        log.Fatal(err)
    }
}