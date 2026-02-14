---
title: "docs"
---

# tsk documentation

- [demo](#demo)
- [install](#install)
- [commands](#commands) -- [show](#show) / [add](#add) / [list (ls)](#list) / [done](#done) / [edit](#edit) / [rm](#rm) / [clear](#clear) / [export](#export) / [config](#config) / [completion](#completion) / [version](#version)
- [priority](#priority)
- [configuration](#configuration)
- [storage](#storage)

---

## demo

```
$ tsk add "buy milk"
added task 1: buy milk

$ tsk add -p h "urgent fix"
added task 2: urgent fix

$ tsk add -p m "review PR"
added task 3: review PR

$ tsk list
  1    [ ] buy milk     (just now)
  2 !! [ ] urgent fix   (just now)
  3  ! [ ] review PR    (just now)

$ tsk done 1
task 1 marked done

$ tsk list
  1    [x] buy milk     (2m ago)
  2 !! [ ] urgent fix   (2m ago)
  3  ! [ ] review PR    (2m ago)

$ tsk list --pending
  2 !! [ ] urgent fix   (2m ago)
  3  ! [ ] review PR    (2m ago)

$ tsk rm 1
task 1 removed
```

output is colored in the terminal — task IDs in cyan, done checkmarks in green, completed tasks dimmed, `!!` in red for high priority, `!` in yellow for medium. disable with `NO_COLOR=1` or in config.

---

## install

### homebrew (macOS/Linux)

```
$ brew install zarldev/tap/tsk
```

### go install

you need [Go](https://go.dev/dl/) 1.23 or later.

```
$ go install github.com/zarldev/tsk/cmd/tsk@latest
```

this places the `tsk` binary in your `$GOPATH/bin` directory. make sure that directory is in your `PATH`.

---

## commands

### show

display detailed information about a single task.

```
tsk <id>
```

pass a task ID as the first argument (no subcommand needed).

```
$ tsk 3
  id:        3
  title:     buy milk
  status:    pending
  created:   2026-02-14 19:42:25 (3h ago)

$ tsk 2
  id:        2
  title:     urgent fix
  priority:  high
  status:    pending
  created:   2026-02-14 19:42:25 (3h ago)

$ tsk 1
  id:        1
  title:     write tests
  status:    done
  created:   2026-02-14 10:30:00 (12h ago)
  completed: 2026-02-14 11:15:00 (11h ago)
```

the `priority:` line only appears for tasks that have a priority set (low, medium, or high). the `completed:` line only appears for tasks that have been marked done. if the task ID does not exist, `tsk` prints an error and exits with status 1.

### add

create a new task, optionally with a priority level.

```
tsk add [-p h|m|l] <title>
```

the title should be quoted if it contains spaces. each task gets an auto-incrementing ID. the `-p` flag sets the priority: `h` (high), `m` (medium), or `l` (low). full names also accepted. if omitted, the task has no priority.

```
$ tsk add "buy milk"
added task 1: buy milk

$ tsk add -p h "urgent fix"
added task 2: urgent fix

$ tsk add -p l "someday"
added task 3: someday
```

### list

display tasks. by default shows all tasks. `ls` is an alias for `list`.

```
tsk list [--done|--pending]
tsk ls [--done|--pending]
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

each line shows the task ID, a priority indicator (`!!` for high, `!` for medium, blank otherwise), completion status (`[ ]` or `[x]`), title, and how long ago it was created. low priority tasks do not show an indicator in the list — use `tsk <id>` to see the priority in the detail view.

**time display:**

| age | display |
|-----|---------|
| < 1 minute | `just now` |
| < 1 hour | `5m ago` |
| < 24 hours | `3h ago` |
| 1 day | `1 day ago` |
| > 1 day | `4 days ago` |

### done

mark one or more tasks as completed. accepts a single ID or comma-separated IDs.

```
tsk done <id>[,<id>,...]
```

```
$ tsk done 1
task 1 marked done

$ tsk done 1,3,5
task 1 marked done
task 3 marked done
task 5 marked done
```

if an ID does not exist, `tsk` prints an error for that ID and continues with the rest. the exit status is non-zero if any ID was not found.

### edit

rename an existing task. the task keeps its ID, creation timestamp, and completion status.

```
tsk edit <id> <title>
```

the title should be quoted if it contains spaces.

```
$ tsk edit 1 "buy oat milk"
task 1 updated: buy oat milk
```

if the ID does not exist, `tsk` prints an error and exits with status 1.

### rm

remove one or more tasks permanently. accepts a single ID or comma-separated IDs.

```
tsk rm <id>[,<id>,...]
```

```
$ tsk rm 1
task 1 removed

$ tsk rm 2,4
task 2 removed
task 4 removed
```

this deletes the task from storage entirely. there is no undo. if an ID does not exist, `tsk` prints an error for that ID and continues with the rest.

### clear

remove all completed tasks in one operation.

```
tsk clear
```

```
$ tsk clear
cleared 3 done tasks

$ tsk clear
no done tasks to clear
```

this removes every task that has been marked done. pending tasks are left untouched. there is no confirmation prompt — use `tsk list --done` first to review what will be removed.

### export

export tasks as a markdown checklist, suitable for pasting into PRs, docs, or notes.

```
tsk export [--done|--pending]
```

export all tasks:

```
$ tsk export
- [ ] buy milk
- [ ] urgent fix (high)
- [x] write tests
```

export only pending tasks:

```
$ tsk export --pending
- [ ] buy milk
- [ ] urgent fix (high)
```

export only completed tasks:

```
$ tsk export --done
- [x] write tests
```

tasks with a priority show it in parentheses after the title. tasks without a priority show the title only.

output goes to stdout with no colors, so it can be piped or redirected:

```
$ tsk export --pending > todo.md
$ tsk export | pbcopy
```

if no tasks match the filter, the output is empty (no "no tasks" message). this is intentional so `tsk export > file.md` produces an empty file rather than one containing a status message.

### config

print the current resolved configuration in TOML format.

```
tsk config
```

pipe to create a config file:

```
$ tsk config > ~/.config/tsk/config.toml
```

### completion

generate shell completion scripts. the script is printed to stdout so you can eval it in your shell config.

```
tsk completion <bash|zsh|fish>
```

#### bash

add to `~/.bashrc`:

```
eval "$(tsk completion bash)"
```

#### zsh

add to `~/.zshrc`:

```
eval "$(tsk completion zsh)"
```

#### fish

add to `~/.config/fish/config.fish`:

```
tsk completion fish | source
```

completions cover subcommands, flags (`--done`, `--pending`, `-p`), priority values (`low`, `medium`, `high`), shell names for `completion`, and task IDs for `done`, `rm`, and `edit` (fetched dynamically via `tsk list`).

### version

print the version.

```
$ tsk version
tsk v0.2.0
```

---

## priority

tasks can have an optional priority level: `h` (high), `m` (medium), or `l` (low). set it when adding a task with the `-p` flag:

```
$ tsk add -p h "deploy hotfix"
added task 1: deploy hotfix

$ tsk add -p m "review PR"
added task 2: review PR

$ tsk add -p l "update docs"
added task 3: update docs

$ tsk add "buy milk"
added task 4: buy milk
```

in the list view, high priority shows `!!` (red) and medium shows `!` (yellow). low and no-priority tasks have no indicator, keeping the list clean:

```
$ tsk list
  1 !! [ ] deploy hotfix  (just now)
  2  ! [ ] review PR      (just now)
  3    [ ] update docs    (just now)
  4    [ ] buy milk       (just now)
```

the detail view shows the priority for any task that has one set:

```
$ tsk 1
  id:        1
  title:     deploy hotfix
  priority:  high
  status:    pending
  created:   2026-02-14 19:42:25 (just now)
```

tasks without a priority omit the `priority:` line entirely. existing tasks from before this feature load fine with no priority (backwards compatible).

---

## configuration

tsk reads configuration from `~/.config/tsk/config.toml`. if the file does not exist, sensible defaults are used — tsk works out of the box with no configuration.

generate a default config file:

    $ mkdir -p ~/.config/tsk
    $ tsk config > ~/.config/tsk/config.toml

### color

    [color]
    enabled = "auto"  # "auto", "always", "never"

- `auto` — colors when stdout is a terminal (default)
- `always` — colors even when piped (useful for `less -R`)
- `never` — no colors

tsk respects the `NO_COLOR` environment variable (https://no-color.org). if set, colors are disabled regardless of config.

---

## storage

tsk supports pluggable storage backends. configure the backend in `~/.config/tsk/config.toml` under the `[storage]` section.

### file (default)

tasks are stored in a single JSON file.

    [storage]
    type = "file"
    path = "~/.tasks.json"

the file is created automatically the first time you add a task. if the file does not exist, `tsk` treats it as an empty task list.

the file contains a JSON array of task objects:

```json
[
  {
    "id": 1,
    "title": "buy milk",
    "done": false,
    "created_at": "2025-01-15T10:30:00Z"
  },
  {
    "id": 2,
    "title": "urgent fix",
    "done": false,
    "priority": "high",
    "created_at": "2025-01-15T10:30:00Z"
  },
  {
    "id": 3,
    "title": "write tests",
    "done": true,
    "created_at": "2025-01-15T10:30:00Z",
    "completed_at": "2025-01-15T11:15:00Z"
  }
]
```

**fields:**

| field | type | description |
|-------|------|-------------|
| `id` | integer | auto-incrementing task identifier |
| `title` | string | task description |
| `done` | boolean | completion status |
| `priority` | string (optional) | `"low"`, `"medium"`, or `"high"`; omitted when not set |
| `created_at` | string | RFC 3339 timestamp of when the task was created |
| `completed_at` | string (optional) | RFC 3339 timestamp of when the task was marked done; omitted for pending tasks |

because the storage is plain JSON, you can back it up, sync it across machines, edit it manually, or version control it.

### github gist

sync tasks across machines via a private gist. requires a GitHub personal access token with `gist` scope.

    [storage]
    type = "gist"
    gist_token = "ghp_..."
    gist_id = ""

or set the token via environment variable:

```
export TSK_GIST_TOKEN=ghp_...
```

the env var takes precedence over the config file value.

on first run with an empty `gist_id`, tsk creates a new private gist and prints the ID. add it to your config to reuse the same gist:

    [storage]
    type = "gist"
    gist_id = "abc123..."
