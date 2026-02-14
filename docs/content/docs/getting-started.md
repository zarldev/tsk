---
title: "Getting Started"
weight: 1
---

# Getting Started

## Install

You need [Go](https://go.dev/dl/) 1.23 or later.

```bash
go install github.com/zarldev/tsk/cmd/tsk@latest
```

This places the `tsk` binary in your `$GOPATH/bin` directory. Make sure that directory is in your `PATH`.

## Your First Task

Add a task:

```bash
$ tsk add "buy milk"
added task 1: buy milk
```

List your tasks:

```bash
$ tsk list
  1 [ ] buy milk  (just now)
```

## Basic Workflow

A typical workflow looks like this:

```bash
# add some tasks
$ tsk add "buy milk"
added task 1: buy milk

$ tsk add "write report"
added task 2: write report

# see what needs doing
$ tsk list
  1 [ ] buy milk  (just now)
  2 [ ] write report  (just now)

# finish a task
$ tsk done 1
task 1 marked done

# check what's left
$ tsk list --pending
  2 [ ] write report  (5m ago)

# clean up completed tasks
$ tsk rm 1
task 1 removed
```

That's it. No projects, no tags, no priorities -- just tasks you need to do.

## Next Steps

- [Commands Reference]({{< relref "commands" >}}) -- every command explained
- [Configuration]({{< relref "configuration" >}}) -- where tasks are stored
