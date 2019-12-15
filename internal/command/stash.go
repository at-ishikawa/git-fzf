package command

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

type stashCli struct {
	listOptions []string
	fzfOption   string
}

const (
	stashFzfPreviewCommand = "git stash show --color -p '{{.stash}}'"
)

func NewStashSubcommand() *cobra.Command {
	return &cobra.Command{
		Use:   "stash [-- <git options>]",
		Short: "git stash list with fzf",
		Args:  cobra.MaximumNArgs(100),
		RunE: func(cmd *cobra.Command, args []string) error {
			flags := cmd.Flags()
			fzfQuery, err := flags.GetString("query")
			if err != nil {
				return err
			}

			cli, err := newStashCli(args, fzfQuery)
			if err != nil {
				return err
			}
			if err := cli.Run(context.Background(), os.Stdin, os.Stdout, os.Stderr); err != nil {
				return err
			}
			return nil
		},
	}
}

func newStashCli(gitOptions []string, fzfQuery string) (*stashCli, error) {
	previewCommand, err := commandFromTemplate("preview", stashFzfPreviewCommand, map[string]interface{}{
		"stash": "{{1}}",
	})
	if err != nil {
		return nil, fmt.Errorf("invalid fzf preview command: %w", err)
	}

	fzfOption, err := getFzfOption(previewCommand)
	if err != nil {
		return nil, fmt.Errorf("failed to get fzf option: %w", err)
	}
	if fzfQuery != "" {
		fzfOption = fzfOption + " --query " + fzfQuery
	}

	return &stashCli{
		listOptions: gitOptions,
		fzfOption:   fzfOption,
	}, nil
}

func (c stashCli) Run(ctx context.Context, ioIn io.Reader, ioOut io.Writer, ioErr io.Writer) error {
	command := fmt.Sprintf("git stash list --format='%%gd %%gs' %s | fzf %s", strings.Join(c.listOptions, " "), c.fzfOption)
	out, err := runCommandWithFzf(ctx, command, ioIn, ioErr)
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			// Script canceled by Ctrl-c
			// Only for bash?: http://tldp.org/LDP/abs/html/exitcodes.html
			if exitErr.ExitCode() == 130 {
				return nil
			}
		}
		return fmt.Errorf("failed to run the command %s: %w", command, err)
	}
	if err := writeFzfResult(ioOut, out, 0); err != nil {
		return err
	}
	return nil
}
