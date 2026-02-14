---
title: "Configuration"
weight: 3
---

# Configuration

## Storage Location

Tasks are stored in a single JSON file at:

```
~/.tasks.json
```

This file is created automatically the first time you add a task. If the file does not exist, `tsk` treats it as an empty task list.

## File Format

The file contains a JSON array of task objects, pretty-printed with 2-space indentation:

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
    "title": "write report",
    "done": true,
    "created_at": "2025-01-15T11:00:00Z"
  }
]
```

### Fields

| Field | Type | Description |
|-------|------|-------------|
| `id` | integer | Auto-incrementing task identifier |
| `title` | string | Task description |
| `done` | boolean | Completion status |
| `created_at` | string | RFC 3339 timestamp of when the task was created |

## Portability

Because the storage is plain JSON, you can:

- Back it up by copying `~/.tasks.json`
- Sync across machines using any file sync tool
- Edit it manually with a text editor
- Version control it if you want history

## No Other Configuration

`tsk` has no configuration file, no environment variables, and no flags beyond those documented in the [commands reference]({{< relref "commands" >}}). It works out of the box.
