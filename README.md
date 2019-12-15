# git fzf plugin

[![GitHub](https://img.shields.io/github/license/at-ishikawa/git-fzf)](https://github.com/at-ishikawa/git-fzf/blob/master/LICENSE)
[![Go workflow](https://github.com/at-ishikawa/git-fzf/workflows/Go/badge.svg)](https://github.com/at-ishikawa/git-fzf)

## Install
You must install `go >= v1.13` beforehand.
```shell script
$ go get -u github.com/at-ishikawa/git-fzf/cmd/git-fzf
```

# git fzf diff command
## Usage
```shell script
$ git fzf diff --help
git diff with fzf

Usage:
  git-fzf diff [<commit>[..<commit>]] [-- <git options>] [flags]

Flags:
  -h, --help           help for diff
  -q, --query string   Start the fzf with this query
```

# Requirements
* go (version 1.13)
* git
* fzf

# Environment variables
* `GIT_FZF_FZF_BIND_OPTION`
    * The bind option for fzf
    * Default: `ctrl-k:kill-line,ctrl-alt-t:toggle-preview,ctrl-alt-n:preview-down,ctrl-alt-p:preview-up,ctrl-alt-v:preview-page-down`
* `GIT_FZF_FZF_OPTION`
    * The entire option for fzf. This option may use `GIT_FZF_FZF_BIND_OPTION` environment variable.
    * Default: `--multi --ansi --inline-info --layout reverse --preview '$GIT_FZF_FZF_PREVIEW_OPTION' --preview-window down:70% --bind $GIT_FZF_FZF_BIND_OPTION`
    * `$GIT_FZF_FZF_PREVIEW_OPTION` is replaced with preview command. This cannot be injected by environment variable `GIT_FZF_FZF_PREVIEW_OPTION`.
