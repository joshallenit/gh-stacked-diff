package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"slices"
	sd "stackeddiff"
)

type Command struct {
	flagSet    *flag.FlagSet
	onSelected func()
}

/*
Outputs abbreviated git log that only shows what has changed, useful for copying commit hashes.
Adds a checkmark beside commits that have an associated branch.
*/
func main() {

	var logFlags int
	var logLevelString string

	commands := []Command{createLogCommand()}
	//
	var commandName string
	if len(os.Args) > 1 {
		commandName = os.Args[1]
	}
	selectedIndex := slices.IndexFunc(commands, func(command Command) bool {
		return command.flagSet.Name() == commandName
	})
	if selectedIndex == -1 {
		fmt.Println("Usage: sd", getCommandNames(commands))
		os.Exit(1)
	}
	addCommonFlags(&logFlags, &logLevelString, commands[selectedIndex].flagSet)
	commands[selectedIndex].flagSet.Parse(os.Args[2:])

	log.SetFlags(logFlags)
	var logLevel slog.Level
	var unmarshallErr = logLevel.UnmarshalText([]byte(logLevelString))
	if unmarshallErr != nil {
		panic("Invalid log level " + logLevelString + ": " + unmarshallErr.Error())
	}
	slog.SetLogLoggerLevel(logLevel)

	commands[selectedIndex].onSelected()
	/*
			package main

		import (
		    "flag"
		    "fmt"
		    "os"
		)

		func main() {

		    fooCmd := flag.NewFlagSet("foo", flag.ExitOnError)
		    fooEnable := fooCmd.Bool("enable", false, "enable")
		    fooName := fooCmd.String("name", "", "name")

		    barCmd := flag.NewFlagSet("bar", flag.ExitOnError)
		    barLevel := barCmd.Int("level", 0, "level")

		    if len(os.Args) < 2 {
		        fmt.Println("expected 'foo' or 'bar' subcommands")
		        os.Exit(1)
		    }

		    switch os.Args[1] {

		    case "foo":
		        fooCmd.Parse(os.Args[2:])
		        fmt.Println("subcommand 'foo'")
		        fmt.Println("  enable:", *fooEnable)
		        fmt.Println("  name:", *fooName)
		        fmt.Println("  tail:", fooCmd.Args())
		    case "bar":
		        barCmd.Parse(os.Args[2:])
		        fmt.Println("subcommand 'bar'")
		        fmt.Println("  level:", *barLevel)
		        fmt.Println("  tail:", barCmd.Args())
		    default:
		        fmt.Println("expected 'foo' or 'bar' subcommands")
		        os.Exit(1)
		    }
		}
	*/

	//flag.Parse()

}

func addCommonFlags(logFlags *int, logLevelString *string, flagSet *flag.FlagSet) {
	flagSet.IntVar(logFlags, "log-flags", 0, "Log flags, see https://pkg.go.dev/log#pkg-constants")
	flagSet.StringVar(logLevelString, "log-level", "info", "Log level: debug, info, warn, or error")
}

func createLogCommand() Command {
	flagSet := flag.NewFlagSet("log", flag.ExitOnError)
	var xxx int
	flagSet.IntVar(&xxx, "xxx", 0, "xxx")

	return Command{flagSet: flagSet, onSelected: func() {
		sd.PrintGitLog(os.Stdout)
		fmt.Println("xxx", xxx)
	}}
}

func getCommandNames(commands []Command) []string {
	var names []string
	slices.Grow(names, len(commands))
	for _, command := range commands {
		names = append(names, command.flagSet.Name())
	}
	slices.Sort(names)
	return names
}
