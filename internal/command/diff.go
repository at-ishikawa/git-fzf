package command

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
)

type diffCli struct {
	listOptions []string
	fzfOption   string
}

const (
	envNameFzfOption     = "GIT_FZF_FZF_OPTION"
	envNameFzfBindOption = "GIT_FZF_FZF_BIND_OPTION"
	defaultFzfBindOption = "ctrl-k:kill-line,ctrl-alt-t:toggle-preview,ctrl-alt-n:preview-down,ctrl-alt-p:preview-up,ctrl-alt-v:preview-page-down"

	fzfPreviewCommand = "git diff --color {{.objectRange}} {{.path}}"
	defaultFzfOption  = "--inline-info --ansi --preview '$GIT_FZF_FZF_PREVIEW_OPTION' --bind $GIT_FZF_FZF_BIND_OPTION"
)

var (
	runCommandWithFzf = func(ctx context.Context, commandLine string, ioIn io.Reader, ioErr io.Writer) ([]byte, error) {
		cmd := exec.CommandContext(ctx, "sh", "-c", commandLine)
		cmd.Stderr = ioErr
		cmd.Stdin = ioIn
		return cmd.Output()
	}
)

func NewDiffSubcommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "diff [<commit>[..<commit>]] [-- <git options>]",
		Short: "git diff with fzf",
		Args:  cobra.MaximumNArgs(100),
		RunE: func(cmd *cobra.Command, args []string) error {
			flags := cmd.Flags()
			fzfQuery, err := flags.GetString("query")
			if err != nil {
				return err
			}

			cli, err := newDiffCli(args, fzfQuery)
			if err != nil {
				return err
			}
			if err := cli.Run(context.Background(), os.Stdin, os.Stdout, os.Stderr); err != nil {
				return err
			}
			return nil
		},
	}
	flags := command.Flags()
	flags.StringP("query", "q", "", "Start the fzf with this query")
	return command
}

func newDiffCli(gitOptions []string, fzfQuery string) (*diffCli, error) {
	gitObjectRange := ""
	if len(gitOptions) > 0 {
		// gitObjectRange may not have ..<commit>
		gitObjectRange = gitOptions[0]
	}
	previewCommand, err := commandFromTemplate("preview", fzfPreviewCommand, map[string]interface{}{
		"path":        "{{2}}",
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

	return &diffCli{
		listOptions: gitOptions,
		fzfOption:   fzfOption,
	}, nil
}

func getFzfOption(previewCommand string) (string, error) {
	fzfOption := os.Getenv(envNameFzfOption)
	if fzfOption == "" {
		fzfOption = defaultFzfOption
	}

	options := map[string][]string{
		"GIT_FZF_FZF_PREVIEW_OPTION": {
			previewCommand,
		},
		envNameFzfBindOption: {
			os.Getenv(envNameFzfBindOption),
			defaultFzfBindOption,
		},
	}
	var invalidEnvVars []string
	fzfOption = os.Expand(fzfOption, func(envName string) string {
		for _, opt := range options[envName] {
			if opt != "" {
				return opt
			}
		}
		invalidEnvVars = append(invalidEnvVars, envName)
		return ""
	})
	if len(invalidEnvVars) != 0 {
		return "", fmt.Errorf("%s has invalid environment variables: %s", envNameFzfOption, strings.Join(invalidEnvVars, ","))
	}
	return fzfOption, nil
}

func (c diffCli) Run(ctx context.Context, ioIn io.Reader, ioOut io.Writer, ioErr io.Writer) error {
	command := fmt.Sprintf("git diff --color --name-status %s | fzf %s", strings.Join(c.listOptions, " "), c.fzfOption)
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
	if _, err := ioOut.Write(out); err != nil {
		return fmt.Errorf("failed to output the result: %w", err)
	}
	return nil
}

func commandFromTemplate(name string, command string, data map[string]interface{}) (string, error) {
	tmpl, err := template.New(name).Option("missingkey=error").Parse(command)
	if err != nil {
		return "", fmt.Errorf("failed to parse the command: %w", err)
	}
	builder := strings.Builder{}
	tmpl.Templates()[0].Option()
	if err = tmpl.Execute(&builder, data); err != nil {
		return "", fmt.Errorf("failed to set data on the template of command: %w", err)
	}
	return builder.String(), nil
}
