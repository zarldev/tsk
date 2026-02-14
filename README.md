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
- configurable via `~/.config/tsk/config.toml`
- storage backends: local file (default), GitHub Gist
- zero dependencies

## Usage

```bash
tsk add "buy milk"       # add a task
tsk list                 # show all tasks
tsk done 1               # mark task 1 complete
tsk rm 1                 # remove task 1
tsk config               # print current config
tsk version              # print version
```

## Docs

Full documentation at [zarldev.github.io/tsk](https://zarldev.github.io/tsk/)
