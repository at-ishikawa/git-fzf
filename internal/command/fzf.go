package command

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io"
	"os"
	"os/exec"
	"strings"
)

const (
	envNameFzfOption     = "GIT_FZF_FZF_OPTION"
	envNameFzfBindOption = "GIT_FZF_FZF_BIND_OPTION"
	defaultFzfBindOption = "ctrl-k:kill-line,ctrl-alt-t:toggle-preview,ctrl-alt-n:preview-down,ctrl-alt-p:preview-up,ctrl-alt-v:preview-page-down"

	defaultFzfOption = "--multi --ansi --inline-info --layout reverse --preview '$GIT_FZF_FZF_PREVIEW_OPTION' --preview-window down:70% --bind $GIT_FZF_FZF_BIND_OPTION"
)

var (
	runCommandWithFzf = func(ctx context.Context, commandLine string, ioIn io.Reader, ioErr io.Writer) ([]byte, error) {
		cmd := exec.CommandContext(ctx, "sh", "-c", commandLine)
		cmd.Stderr = ioErr
		cmd.Stdin = ioIn
		return cmd.Output()
	}
)

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

func writeFzfResult(ioOut io.Writer, out []byte, column int) error {
	lineSeparator := "\n"
	lines := strings.Split(strings.TrimSpace(string(out)), lineSeparator)
	filePaths := make([]string, len(lines))
	for i, line := range lines {
		fields := strings.Fields(line)
		filePath := strings.TrimSpace(fields[column])
		filePaths[i] = filePath
	}
	buf := bytes.NewBufferString(strings.Join(filePaths, lineSeparator) + "\n")
	if _, err := ioOut.Write(buf.Bytes()); err != nil {
		return fmt.Errorf("failed to output the result: %w", err)
	}
	return nil
}
