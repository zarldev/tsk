---
title: "tsk"
layout: landing
---

<div class="book-hero">

# tsk {anchor=false}

Track tasks from your terminal. No database, no config, no fuss.

```bash
go install github.com/zarldev/tsk/cmd/tsk@latest
```

{{< button relref="/docs/getting-started" >}}Get Started{{< /button >}}
{{< button href="https://github.com/zarldev/tsk" >}}GitHub{{< /button >}}

</div>

<div class="landing-features">

{{% columns %}}

- {{< card >}}
  ### Instant
  Add, complete, and remove tasks in seconds. No workflow to learn.
  {{< /card >}}

- {{< card >}}
  ### Zero Dependencies
  Single Go binary. No database, no config file, no runtime needed.
  {{< /card >}}

- {{< card >}}
  ### Plain JSON
  Tasks live in `~/.tasks.json`. Read them, script them, back them up.
  {{< /card >}}

- {{< card >}}
  ### Filtered Views
  List everything, or show only pending or completed tasks.
  {{< /card >}}

{{% /columns %}}

</div>

<div class="landing-demo">

## See it in action

```bash
$ tsk add "buy milk"
added task 1: buy milk

$ tsk add "write report"
added task 2: write report

$ tsk list
  1 [ ] buy milk  (just now)
  2 [ ] write report  (just now)

$ tsk done 1
task 1 marked done

$ tsk list
  1 [x] buy milk  (2m ago)
  2 [ ] write report  (2m ago)

$ tsk rm 1
task 1 removed
```

</div>
