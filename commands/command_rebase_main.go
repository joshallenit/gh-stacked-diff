package commands

import (
	"flag"
	sd "stackeddiff"

	"github.com/joshallenit/stacked-diff/util"
)

func createRebaseMainCommand() Command {
	flagSet := flag.NewFlagSet("rebase-main", flag.ContinueOnError)

	return Command{
		FlagSet: flagSet,
		Summary: "Bring your main branch up to date with remote",
		Description: "Rebase with origin/" + util.GetMainBranchForHelp() + ", dropping any commits who's associated\n" +
			"branches have been merged.\n" +
			"\n" +
			"This avoids having to manually call \"git reset --hard head\" whenever\n" +
			"you have merge conflicts with a commit that has already been merged\n" +
			"but has slight variation with local main because, for example, a\n" +
			"change was made with the Github Web UI.",
		Usage: "sd " + flagSet.Name(),
		OnSelected: func(command Command) {
			if flagSet.NArg() != 0 {
				commandError(flagSet, "too many arguments", command.Usage)
			}
			sd.RebaseMain()
		}}
}
