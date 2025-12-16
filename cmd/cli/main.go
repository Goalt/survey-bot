package main

import (
	"context"
	"flag"
	"os"

	"github.com/google/subcommands"
)

func main() {
	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(subcommands.FlagsCommand(), "")
	subcommands.Register(subcommands.CommandsCommand(), "")
	subcommands.Register(&CreateSurveyCmd{}, "")
	subcommands.Register(&UpdateSurveyCmd{}, "")
	subcommands.Register(&DeleteUserInfoCmd{}, "")
	subcommands.Register(&GetResultsCmd{}, "")

	flag.Parse()
	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}
