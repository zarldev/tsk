```
  _        _
 | |_ ___ | | __
 | __/ __|| |/ /
 | |_\__ \|   <
  \__|___/|_|\_\
```

A simple CLI task tracker with zero dependencies.

## Install

### Homebrew

```bash
brew install zarldev/tap/tsk
```

### Go

```bash
go install github.com/zarldev/tsk/cmd/tsk@latest
```

## Features

- colored output (respects `NO_COLOR`)
- shell completions for bash, zsh, and fish
- configurable via `~/.config/tsk/config.toml`
- storage backends: local file (default), GitHub Gist
- zero dependencies

## Usage

```bash
tsk add "buy milk"             # add a task
tsk add -p h "urgent"          # add with priority (h=high, m=medium, l=low)
tsk list                       # show all tasks
tsk ls                         # same as list
tsk 1                          # show task 1 details
tsk done 1                     # mark task 1 complete
tsk done 1,3,5                 # mark multiple tasks complete
tsk edit 1 "buy oat milk"      # rename task 1
tsk rm 1                       # remove task 1
tsk rm 2,4                     # remove multiple tasks
tsk clear                      # remove all done tasks
tsk export                     # export tasks as markdown
tsk export --pending           # export only pending tasks
tsk config                     # print current config
tsk completion bash            # generate bash completions
tsk version                    # print version
```

## Docs

Full documentation at [zarldev.github.io/tsk](https://zarldev.github.io/tsk/)
