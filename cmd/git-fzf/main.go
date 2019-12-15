package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/at-ishikawa/git-fzf/internal/command"
)

func main() {
	cli := cobra.Command{
		Use:   "git-fzf [command]",
		Short: "git commands with fzf",
	}
	cli.AddCommand(command.NewDiffSubcommand())
	cli.AddCommand(command.NewLogSubcommand())
	cli.AddCommand(command.NewStashSubcommand())
	if err := cli.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(0)
}
