---
title: "Commands Reference"
weight: 2
---

# Commands Reference

## add

Create a new task.

```
tsk add <title>
```

The title should be quoted if it contains spaces.

```bash
$ tsk add "buy milk"
added task 1: buy milk
```

Each task gets an auto-incrementing ID. IDs are never reused within the current task list.

## list

Display tasks. By default, shows all tasks.

```
tsk list [--done|--pending]
```

### Show all tasks

```bash
$ tsk list
  1 [ ] buy milk  (just now)
```

### Show only completed tasks

```bash
$ tsk list --done
  1 [x] buy milk  (2m ago)
```

### Show only pending tasks

```bash
$ tsk list --pending
no tasks
```

When there are no tasks matching the filter, `tsk` prints `no tasks`.

Each line shows the task ID, completion status (`[ ]` or `[x]`), title, and how long ago it was created.

### Time Display

The age column shows a human-readable duration:

| Age | Display |
|-----|---------|
| < 1 minute | `just now` |
| < 1 hour | `5m ago` |
| < 24 hours | `3h ago` |
| 1 day | `1 day ago` |
| > 1 day | `4 days ago` |

## done

Mark a task as completed.

```
tsk done <id>
```

```bash
$ tsk done 1
task 1 marked done
```

If the ID does not exist, `tsk` prints an error and exits with a non-zero status.

## rm

Remove a task permanently.

```
tsk rm <id>
```

```bash
$ tsk rm 1
task 1 removed
```

This deletes the task from storage entirely. There is no undo.
