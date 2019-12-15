package command

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLogSubcommand(t *testing.T) {
	assert.NotNil(t, NewLogSubcommand())
}

func TestNewLogCommand(t *testing.T) {
	testCases := []struct {
		name       string
		gitOptions []string
		fzfQuery   string
		envVars    map[string]string
		want       *logCli
		wantErr    error
	}{
		{
			name:       "no options",
			gitOptions: []string{},
			fzfQuery:   "",
			want: &logCli{
				listOptions: []string{},
				fzfOption:   fmt.Sprintf("--multi --ansi --inline-info --layout reverse --preview '%s' --preview-window down:70%% --bind %s", "git show --color  {{1}}", defaultFzfBindOption),
			},
			wantErr: nil,
		},
		{
			name: "all options",
			gitOptions: []string{
				"origin/master",
				"--diff-filter",
				"A",
			},
			fzfQuery: "config",
			want: &logCli{
				listOptions: []string{
					"origin/master",
					"--diff-filter",
					"A",
				},
				fzfOption: fmt.Sprintf("--multi --ansi --inline-info --layout reverse --preview '%s' --preview-window down:70%% --bind %s --query config", "git show --color origin/master {{1}}", defaultFzfBindOption),
			},
			wantErr: nil,
		},
		{
			name:       "GIT_FZF_FZF_OPTION includes invalid env",
			gitOptions: []string{},
			fzfQuery:   "",
			envVars: map[string]string{
				envNameFzfOption: "$UNKNOWN_ENV1, $UNKNOWN_ENV2",
			},
			want:    nil,
			wantErr: fmt.Errorf("failed to get fzf option: %w", fmt.Errorf("%s has invalid environment variables: %s", envNameFzfOption, "UNKNOWN_ENV1,UNKNOWN_ENV2")),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if len(tc.envVars) > 0 {
				defer func() {
					for k := range tc.envVars {
						require.NoError(t, os.Unsetenv(k))
					}
				}()
				for k, v := range tc.envVars {
					require.NoError(t, os.Setenv(k, v))
				}
			}
			got, gotErr := newLogCli(tc.gitOptions, tc.fzfQuery)
			assert.Equal(t, tc.want, got)
			assert.Equal(t, tc.wantErr, gotErr)
		})
	}
}

func TestLogCli_Run(t *testing.T) {
	fzfOption := "--inline-info"
	defaultRunCommand := func(ctx context.Context, commandLine string, ioIn io.Reader, ioErr io.Writer) (i []byte, e error) {
		assert.Equal(t, fmt.Sprintf("%s | fzf %s",
			"git log --color --oneline origin/master",
			fzfOption,
		), commandLine)
		return bytes.NewBufferString("abc Commit message1\nxyz Commit message2\n").Bytes(), nil
	}
	defaultWantErr := errors.New("want error")
	exitErr := exec.ExitError{}

	testCases := []struct {
		name              string
		runCommandWithFzf func(ctx context.Context, commandLine string, ioIn io.Reader, ioErr io.Writer) (i []byte, e error)
		sut               logCli
		wantErr           error
		wantIO            string
		wantIOErr         string
	}{
		{
			name: "name output",
			sut: logCli{
				listOptions: []string{
					"origin/master",
				},
				fzfOption: fzfOption,
			},
			runCommandWithFzf: defaultRunCommand,
			wantErr:           nil,
			wantIO:            "abc\nxyz\n",
			wantIOErr:         "",
		},
		{
			name: "command with fzf error",
			sut: logCli{
				listOptions: []string{},
				fzfOption:   fzfOption,
			},
			runCommandWithFzf: func(ctx context.Context, commandLine string, ioIn io.Reader, ioErr io.Writer) (i []byte, e error) {
				return nil, defaultWantErr
			},
			wantErr:   defaultWantErr,
			wantIO:    "",
			wantIOErr: "",
		},
		{
			name: "command with fzf exit error (not 130)",
			sut: logCli{
				listOptions: []string{},
				fzfOption:   fzfOption,
			},
			runCommandWithFzf: func(ctx context.Context, commandLine string, ioIn io.Reader, ioErr io.Writer) (i []byte, e error) {
				return nil, &exitErr
			},
			wantErr:   &exitErr,
			wantIO:    "",
			wantIOErr: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runCommandWithFzf = tc.runCommandWithFzf

			var gotIOOut bytes.Buffer
			var gotIOErr bytes.Buffer
			gotErr := tc.sut.Run(context.Background(), strings.NewReader("in"), &gotIOOut, &gotIOErr)
			assert.True(t, errors.Is(gotErr, tc.wantErr))
			assert.Equal(t, tc.wantIO, gotIOOut.String())
			assert.Equal(t, tc.wantIOErr, gotIOErr.String())
		})
	}
}
