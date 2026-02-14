---
title: "tsk"
type: docs
---

# tsk

A simple CLI task tracker with zero dependencies. Track what you need to do, right from the terminal.

## Install

```bash
go install github.com/zarldev/tsk/cmd/tsk@latest
```

## Features

- **Instant task tracking** -- add, complete, and remove tasks in seconds
- **Zero dependencies** -- single Go binary, no database, no config required
- **Human-readable storage** -- tasks live in `~/.tasks.json` as plain JSON
- **Filtered views** -- list all tasks, or show only done/pending

## Quick Start

```bash
$ tsk add "buy milk"
added task 1: buy milk

$ tsk list
  1 [ ] buy milk  (just now)

$ tsk done 1
task 1 marked done

$ tsk list
  1 [x] buy milk  (2m ago)
```

{{< button relref="/docs/getting-started" >}}Get Started{{< /button >}}
