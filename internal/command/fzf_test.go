package command

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	backupRunCommandWithFzf := runCommandWithFzf
	defer func() {
		runCommandWithFzf = backupRunCommandWithFzf
	}()
	os.Exit(m.Run())
}

func TestGetFzfOption(t *testing.T) {
	testCases := []struct {
		name           string
		previewCommand string
		envVars        map[string]string
		want           string
		wantErr        error
	}{
		{
			name:           "no env vars",
			previewCommand: "git diff {1}",
			want:           fmt.Sprintf("--multi --ansi --inline-info --layout reverse --preview '%s' --preview-window down:70%% --bind %s", "git diff {1}", defaultFzfBindOption),
		},
		{
			name:           "all correct env vars",
			previewCommand: "git diff {1}",
			envVars: map[string]string{
				envNameFzfOption:     fmt.Sprintf("--preview '$GIT_FZF_FZF_PREVIEW_OPTION' --bind $%s", envNameFzfBindOption),
				envNameFzfBindOption: "ctrl-k:kill-line",
			},
			want: fmt.Sprintf("--preview '%s' --bind %s", "git diff {1}", "ctrl-k:kill-line"),
		},
		{
			name:           "no env vars",
			previewCommand: "unused preview command",
			envVars: map[string]string{
				envNameFzfOption:     "--inline-info",
				envNameFzfBindOption: "unused",
			},
			want: "--inline-info",
		},
		{
			name:           "invalid env vars in GIT_FZF_FZF_OPTION",
			previewCommand: "unused preview command",
			envVars: map[string]string{
				envNameFzfOption:     "--inline-info $UNKNOWN_ENV_NAME",
				envNameFzfBindOption: "unused",
			},
			want:    "",
			wantErr: fmt.Errorf("%s has invalid environment variables: UNKNOWN_ENV_NAME", envNameFzfOption),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				for k := range tc.envVars {
					require.NoError(t, os.Unsetenv(k))
				}
			}()
			for k, v := range tc.envVars {
				require.NoError(t, os.Setenv(k, v))
			}
			got, gotErr := getFzfOption(tc.previewCommand)
			assert.Equal(t, tc.want, got)
			assert.Equal(t, tc.wantErr, gotErr)
		})
	}
}

func TestBuildCommand(t *testing.T) {
	testCases := []struct {
		name         string
		templateName string
		command      string
		data         map[string]interface{}
		want         string
		wantIsErr    bool
	}{
		{
			name:         "template",
			templateName: "template",
			command:      "git {{ .command }} {{ .commit }}",
			data: map[string]interface{}{
				"command": "diff",
				"commit":  "abc",
			},
			want:      "git diff abc",
			wantIsErr: false,
		},
		{
			name:         "no template",
			templateName: "",
			command:      "{{ .name }}",
			data: map[string]interface{}{
				"name": "fzf",
			},
			want:      "fzf",
			wantIsErr: false,
		},
		{
			name:         "invalid command",
			templateName: "template",
			command:      "{{ .name }",
			data: map[string]interface{}{
				"name": "name",
			},
			want:      "",
			wantIsErr: true,
		},
		{
			name:         "wrong parameter",
			templateName: "template",
			command:      "wrong {{ .name }}",
			data: map[string]interface{}{
				"unknown": "unknown",
			},
			want:      "",
			wantIsErr: true,
		},
		{
			name:         "no parameter",
			templateName: "template",
			command:      "no {{ .name }}",
			data:         nil,
			want:         "",
			wantIsErr:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, gotErr := commandFromTemplate(tc.templateName, tc.command, tc.data)
			assert.Equal(t, tc.want, got)
			assert.Equal(t, tc.wantIsErr, gotErr != nil)
		})
	}
}
