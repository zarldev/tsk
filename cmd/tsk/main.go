package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/zarlbot/tsk/internal/task"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	path, err := task.DefaultPath()
	if err != nil {
		fatal(err)
	}

	switch os.Args[1] {
	case "add":
		cmdAdd(path)
	case "list":
		cmdList(path)
	case "done":
		cmdDone(path)
	case "rm":
		cmdRm(path)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		usage()
		os.Exit(1)
	}
}

func cmdAdd(path string) {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: tsk add <title>")
		os.Exit(1)
	}

	tasks, err := task.Load(path)
	if err != nil {
		fatal(err)
	}

	title := os.Args[2]
	tasks = task.Add(tasks, title)

	if err := task.Save(path, tasks); err != nil {
		fatal(err)
	}

	t := tasks[len(tasks)-1]
	fmt.Printf("added task %d: %s\n", t.ID, t.Title)
}

func cmdList(path string) {
	f := task.FilterAll
	if len(os.Args) > 2 {
		switch os.Args[2] {
		case "--done":
			f = task.FilterDone
		case "--pending":
			f = task.FilterPending
		default:
			fmt.Fprintf(os.Stderr, "unknown flag: %s\n", os.Args[2])
			os.Exit(1)
		}
	}

	tasks, err := task.Load(path)
	if err != nil {
		fatal(err)
	}

	filtered := task.List(tasks, f)
	if len(filtered) == 0 {
		fmt.Println("no tasks")
		return
	}

	for _, t := range filtered {
		check := "[ ]"
		if t.Done {
			check = "[x]"
		}
		fmt.Printf("%3d %s %s  (%s)\n", t.ID, check, t.Title, age(t.CreatedAt))
	}
}

func cmdDone(path string) {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: tsk done <id>")
		os.Exit(1)
	}

	id, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid id: %s\n", os.Args[2])
		os.Exit(1)
	}

	tasks, err := task.Load(path)
	if err != nil {
		fatal(err)
	}

	if err := task.Done(tasks, id); err != nil {
		fatal(err)
	}

	if err := task.Save(path, tasks); err != nil {
		fatal(err)
	}

	fmt.Printf("task %d marked done\n", id)
}

func cmdRm(path string) {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: tsk rm <id>")
		os.Exit(1)
	}

	id, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid id: %s\n", os.Args[2])
		os.Exit(1)
	}

	tasks, err := task.Load(path)
	if err != nil {
		fatal(err)
	}

	tasks, err = task.Remove(tasks, id)
	if err != nil {
		fatal(err)
	}

	if err := task.Save(path, tasks); err != nil {
		fatal(err)
	}

	fmt.Printf("task %d removed\n", id)
}

func age(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		days := int(d.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, `usage: tsk <command> [args]

commands:
  add <title>       add a new task
  list [--done|--pending]  list tasks
  done <id>         mark a task as done
  rm <id>           remove a task`)
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
