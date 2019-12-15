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

type logCli struct {
	listOptions []string
	fzfOption   string
}

const (
	logFzfPreviewCommand = "git show --color {{.objectRange}} {{.path}}"
)

func NewLogSubcommand() *cobra.Command {
	return &cobra.Command{
		Use:   "log [<commit>[..<commit>]] [-- <git options>]",
		Short: "git log with fzf",
		Args:  cobra.MaximumNArgs(100),
		RunE: func(cmd *cobra.Command, args []string) error {
			flags := cmd.Flags()
			fzfQuery, err := flags.GetString("query")
			if err != nil {
				return err
			}

			cli, err := newLogCli(args, fzfQuery)
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

func newLogCli(gitOptions []string, fzfQuery string) (*logCli, error) {
	gitObjectRange := ""
	if len(gitOptions) > 0 {
		// gitObjectRange may not have ..<commit>
		gitObjectRange = gitOptions[0]
	}
	previewCommand, err := commandFromTemplate("preview", logFzfPreviewCommand, map[string]interface{}{
		"path":        "{{1}}",
		"objectRange": gitObjectRange,
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

	return &logCli{
		listOptions: gitOptions,
		fzfOption:   fzfOption,
	}, nil
}

func (c logCli) Run(ctx context.Context, ioIn io.Reader, ioOut io.Writer, ioErr io.Writer) error {
	command := fmt.Sprintf("git log --color --oneline %s | fzf %s", strings.Join(c.listOptions, " "), c.fzfOption)
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
