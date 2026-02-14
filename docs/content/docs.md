---
title: "docs"
---

# tsk documentation

- [install](#install)
- [commands](#commands) -- [add](#add) / [list](#list) / [done](#done) / [rm](#rm)
- [storage](#storage)

---

## install

you need [Go](https://go.dev/dl/) 1.23 or later.

```
$ go install github.com/zarldev/tsk/cmd/tsk@latest
```

this places the `tsk` binary in your `$GOPATH/bin` directory. make sure that directory is in your `PATH`.

---

## commands

### add

create a new task.

```
tsk add <title>
```

the title should be quoted if it contains spaces. each task gets an auto-incrementing ID.

```
$ tsk add "buy milk"
added task 1: buy milk
```

### list

display tasks. by default shows all tasks.

```
tsk list [--done|--pending]
```

show all tasks:

```
$ tsk list
  1 [ ] buy milk  (just now)
```

show only completed tasks:

```
$ tsk list --done
  1 [x] buy milk  (2m ago)
```

show only pending tasks:

```
$ tsk list --pending
no tasks
```

when there are no tasks matching the filter, `tsk` prints `no tasks`.

each line shows the task ID, completion status (`[ ]` or `[x]`), title, and how long ago it was created.

**time display:**

| age | display |
|-----|---------|
| < 1 minute | `just now` |
| < 1 hour | `5m ago` |
| < 24 hours | `3h ago` |
| 1 day | `1 day ago` |
| > 1 day | `4 days ago` |

### done

mark a task as completed.

```
tsk done <id>
```

```
$ tsk done 1
task 1 marked done
```

if the ID does not exist, `tsk` prints an error and exits with a non-zero status.

### rm

remove a task permanently.

```
tsk rm <id>
```

```
$ tsk rm 1
task 1 removed
```

this deletes the task from storage entirely. there is no undo.

---

## storage

tasks are stored in a single JSON file at:

```
~/.tasks.json
```

this file is created automatically the first time you add a task. if the file does not exist, `tsk` treats it as an empty task list.

the file contains a JSON array of task objects:

```json
[
  {
    "id": 1,
    "title": "buy milk",
    "done": false,
    "created_at": "2025-01-15T10:30:00Z"
  }
]
```

**fields:**

| field | type | description |
|-------|------|-------------|
| `id` | integer | auto-incrementing task identifier |
| `title` | string | task description |
| `done` | boolean | completion status |
| `created_at` | string | RFC 3339 timestamp of when the task was created |

because the storage is plain JSON, you can back it up, sync it across machines, edit it manually, or version control it.

`tsk` has no configuration file, no environment variables, and no flags beyond those documented above. it works out of the box.
