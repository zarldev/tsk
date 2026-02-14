package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
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
	case "gist":
		token := cfg.Storage.GistToken
		if env := os.Getenv("TSK_GIST_TOKEN"); env != "" {
			token = env
		}
		if token == "" {
			fatal(fmt.Errorf("gist storage requires gist_token in config or TSK_GIST_TOKEN env var"))
		}
		store = task.NewGistStore(token, cfg.Storage.GistID)
	default:
		fmt.Fprintf(os.Stderr, "unknown storage type: %s\n", cfg.Storage.Type)
		os.Exit(1)
	}

	c := color.New(cfg.Color.Enabled)

	switch os.Args[1] {
	case "add":
		cmdAdd(store, c)
	case "list", "ls":
		cmdList(store, c)
	case "done":
		cmdDone(store, c)
	case "edit":
		cmdEdit(store, c)
	case "rm":
		cmdRm(store, c)
	case "clear":
		cmdClear(store, c)
	case "export":
		cmdExport(store)
	case "config":
		cmdConfig(cfg)
	case "version":
		fmt.Printf("tsk %s\n", version)
	default:
		if id, err := strconv.Atoi(os.Args[1]); err == nil {
			cmdShow(store, c, id)
			return
		}
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		usage()
		os.Exit(1)
	}
}

func cmdAdd(store task.Store, c color.Palette) {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: tsk add [-p <priority>] <title>")
		os.Exit(1)
	}

	var title string
	var priority task.Priority

	if os.Args[2] == "-p" {
		if len(os.Args) < 5 {
			fmt.Fprintln(os.Stderr, "usage: tsk add -p <priority> <title>")
			os.Exit(1)
		}
		p, ok := task.ValidPriority(os.Args[3])
		if !ok {
			fmt.Fprintf(os.Stderr, "invalid priority: %s (use low, medium, high)\n", os.Args[3])
			os.Exit(1)
		}
		priority = p
		title = os.Args[4]
	} else {
		title = os.Args[2]
	}

	tasks, err := store.Load()
	if err != nil {
		fatal(err)
	}

	tasks = task.Add(tasks, title, priority)

	if err := store.Save(tasks); err != nil {
		fatal(err)
	}

	t := tasks[len(tasks)-1]
	fmt.Printf("added task %s: %s\n", c.BoldCyan(strconv.Itoa(t.ID)), t.Title)
}

func cmdShow(store task.Store, c color.Palette, id int) {
	tasks, err := store.Load()
	if err != nil {
		fatal(err)
	}

	t := task.Find(tasks, id)
	if t == nil {
		fmt.Fprintf(os.Stderr, "task %d: not found\n", id)
		os.Exit(1)
	}

	status := "pending"
	if t.Done {
		status = c.Green("done")
	}

	created := t.CreatedAt.Format("2006-01-02 15:04:05")
	createdAge := age(t.CreatedAt)

	fmt.Printf("  %s  %s\n", c.Dim("id:"), c.BoldCyan(strconv.Itoa(t.ID)))
	fmt.Printf("  %s  %s\n", c.Dim("title:"), t.Title)

	if t.Priority != task.PriorityNone {
		pv := colorPriority(c, t.Priority)
		fmt.Printf("  %s  %s\n", c.Dim("priority:"), pv)
	}

	fmt.Printf("  %s  %s\n", c.Dim("status:"), status)
	fmt.Printf("  %s  %s %s\n", c.Dim("created:"), created, c.Dim("("+createdAge+")"))

	if t.CompletedAt != nil {
		completed := t.CompletedAt.Format("2006-01-02 15:04:05")
		completedAge := age(*t.CompletedAt)
		fmt.Printf("  %s  %s %s\n", c.Dim("completed:"), completed, c.Dim("("+completedAge+")"))
	}
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
		pri := priorityIndicator(c, t.Priority)

		if t.Done {
			check := c.Green("[x]")
			title := c.DimStrikethrough(t.Title)
			fmt.Printf("%s %s %s %s  %s\n", id, pri, check, title, a)
		} else {
			fmt.Printf("%s %s [ ] %s  %s\n", id, pri, t.Title, a)
		}
	}
}

func cmdDone(store task.Store, c color.Palette) {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: tsk done <id>[,<id>,...]")
		os.Exit(1)
	}

	ids, err := parseIDs(os.Args[2])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	tasks, err := store.Load()
	if err != nil {
		fatal(err)
	}

	var hadErr bool
	for _, id := range ids {
		if err := task.Done(tasks, id); err != nil {
			fmt.Fprintf(os.Stderr, "task %d: not found\n", id)
			hadErr = true
			continue
		}
		fmt.Printf("task %s marked %s\n", c.BoldCyan(strconv.Itoa(id)), c.Green("done"))
	}

	if err := store.Save(tasks); err != nil {
		fatal(err)
	}

	if hadErr {
		os.Exit(1)
	}
}

func cmdEdit(store task.Store, c color.Palette) {
	if len(os.Args) < 4 {
		fmt.Fprintln(os.Stderr, "usage: tsk edit <id> <title>")
		os.Exit(1)
	}

	id, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid id: %s\n", os.Args[2])
		os.Exit(1)
	}

	title := os.Args[3]

	tasks, err := store.Load()
	if err != nil {
		fatal(err)
	}

	if err := task.Edit(tasks, id, title); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := store.Save(tasks); err != nil {
		fatal(err)
	}

	fmt.Printf("task %s updated: %s\n", c.BoldCyan(strconv.Itoa(id)), title)
}

func cmdRm(store task.Store, c color.Palette) {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: tsk rm <id>[,<id>,...]")
		os.Exit(1)
	}

	ids, err := parseIDs(os.Args[2])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	tasks, err := store.Load()
	if err != nil {
		fatal(err)
	}

	var hadErr bool
	for _, id := range ids {
		var rmErr error
		tasks, rmErr = task.Remove(tasks, id)
		if rmErr != nil {
			fmt.Fprintf(os.Stderr, "task %d: not found\n", id)
			hadErr = true
			continue
		}
		fmt.Printf("task %s removed\n", c.BoldCyan(strconv.Itoa(id)))
	}

	if err := store.Save(tasks); err != nil {
		fatal(err)
	}

	if hadErr {
		os.Exit(1)
	}
}

func cmdClear(store task.Store, c color.Palette) {
	tasks, err := store.Load()
	if err != nil {
		fatal(err)
	}

	removed, tasks := task.ClearDone(tasks)

	if removed == 0 {
		fmt.Println("no done tasks to clear")
		return
	}

	if err := store.Save(tasks); err != nil {
		fatal(err)
	}

	fmt.Printf("cleared %s done %s\n",
		c.BoldCyan(strconv.Itoa(removed)),
		pluralize(removed, "task", "tasks"))
}

func cmdExport(store task.Store) {
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
		return
	}

	for _, t := range filtered {
		check := " "
		if t.Done {
			check = "x"
		}
		line := fmt.Sprintf("- [%s] %s", check, t.Title)
		if t.Priority != "" {
			line += fmt.Sprintf(" (%s)", t.Priority)
		}
		fmt.Println(line)
	}
}

func pluralize(n int, singular, plural string) string {
	if n == 1 {
		return singular
	}
	return plural
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

// priorityIndicator returns a 2-char wide indicator for the list view.
func priorityIndicator(c color.Palette, p task.Priority) string {
	switch p {
	case task.PriorityHigh:
		return c.BoldRed("!!")
	case task.PriorityMedium:
		return c.BoldYellow(" !")
	default:
		return "  "
	}
}

// colorPriority returns the priority value colored for the detail view.
func colorPriority(c color.Palette, p task.Priority) string {
	switch p {
	case task.PriorityHigh:
		return c.Red(string(p))
	case task.PriorityMedium:
		return c.Yellow(string(p))
	case task.PriorityLow:
		return c.Dim(string(p))
	default:
		return string(p)
	}
}

func cmdConfig(cfg config.Config) {
	fmt.Print(cfg.String())
}

// parseIDs splits a comma-separated string into a slice of task IDs.
// All segments must be valid integers; returns an error on the first invalid one.
func parseIDs(arg string) ([]int, error) {
	parts := strings.Split(arg, ",")
	ids := make([]int, 0, len(parts))
	for _, p := range parts {
		id, err := strconv.Atoi(p)
		if err != nil {
			return nil, fmt.Errorf("invalid id: %s", p)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func usage() {
	fmt.Fprintln(os.Stderr, `usage: tsk <command> [args]

commands:
  <id>                         show task details
  add [-p <priority>] <title>  add a new task (priority: low, medium, high)
  list, ls [--done|--pending]  list tasks
  done <id>[,<id>,...]         mark tasks as done
  edit <id> <title>            rename a task
  rm <id>[,<id>,...]           remove tasks
  clear                        remove all done tasks
  export [--done|--pending]    export tasks as markdown
  config                       show current configuration
  version                      print version`)
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
