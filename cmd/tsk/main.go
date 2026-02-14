package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/zarldev/tsk/internal/color"
	"github.com/zarldev/tsk/internal/config"
	"github.com/zarldev/tsk/internal/task"
)

var version = "dev"

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	cfg, err := config.Load()
	if err != nil {
		fatal(err)
	}

	var store task.Store
	switch cfg.Storage.Type {
	case "file":
		store = task.NewFileStore(cfg.Storage.Path)
	default:
		fmt.Fprintf(os.Stderr, "unknown storage type: %s\n", cfg.Storage.Type)
		os.Exit(1)
	}

	c := color.New(cfg.Color.Enabled)

	switch os.Args[1] {
	case "add":
		cmdAdd(store, c)
	case "list":
		cmdList(store, c)
	case "done":
		cmdDone(store, c)
	case "rm":
		cmdRm(store, c)
	case "config":
		cmdConfig(cfg)
	case "version":
		fmt.Printf("tsk %s\n", version)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		usage()
		os.Exit(1)
	}
}

func cmdAdd(store task.Store, c color.Palette) {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: tsk add <title>")
		os.Exit(1)
	}

	tasks, err := store.Load()
	if err != nil {
		fatal(err)
	}

	title := os.Args[2]
	tasks = task.Add(tasks, title)

	if err := store.Save(tasks); err != nil {
		fatal(err)
	}

	t := tasks[len(tasks)-1]
	fmt.Printf("added task %s: %s\n", c.BoldCyan(strconv.Itoa(t.ID)), t.Title)
}

func cmdList(store task.Store, c color.Palette) {
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

	tasks, err := store.Load()
	if err != nil {
		fatal(err)
	}

	filtered := task.List(tasks, f)
	if len(filtered) == 0 {
		fmt.Println("no tasks")
		return
	}

	for _, t := range filtered {
		id := c.BoldCyan(fmt.Sprintf("%3d", t.ID))
		a := c.Dim(fmt.Sprintf("(%s)", age(t.CreatedAt)))

		if t.Done {
			check := c.Green("[x]")
			title := c.DimStrikethrough(t.Title)
			fmt.Printf("%s %s %s  %s\n", id, check, title, a)
		} else {
			fmt.Printf("%s [ ] %s  %s\n", id, t.Title, a)
		}
	}
}

func cmdDone(store task.Store, c color.Palette) {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: tsk done <id>")
		os.Exit(1)
	}

	id, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid id: %s\n", os.Args[2])
		os.Exit(1)
	}

	tasks, err := store.Load()
	if err != nil {
		fatal(err)
	}

	if err := task.Done(tasks, id); err != nil {
		fatal(err)
	}

	if err := store.Save(tasks); err != nil {
		fatal(err)
	}

	fmt.Printf("task %s marked %s\n", c.BoldCyan(strconv.Itoa(id)), c.Green("done"))
}

func cmdRm(store task.Store, c color.Palette) {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: tsk rm <id>")
		os.Exit(1)
	}

	id, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid id: %s\n", os.Args[2])
		os.Exit(1)
	}

	tasks, err := store.Load()
	if err != nil {
		fatal(err)
	}

	tasks, err = task.Remove(tasks, id)
	if err != nil {
		fatal(err)
	}

	if err := store.Save(tasks); err != nil {
		fatal(err)
	}

	fmt.Printf("task %s removed\n", c.BoldCyan(strconv.Itoa(id)))
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

func cmdConfig(cfg config.Config) {
	fmt.Print(cfg.String())
}

func usage() {
	fmt.Fprintln(os.Stderr, `usage: tsk <command> [args]

commands:
  add <title>              add a new task
  list [--done|--pending]  list tasks
  done <id>                mark a task as done
  rm <id>                  remove a task
  config                   show current configuration
  version                  print version`)
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
