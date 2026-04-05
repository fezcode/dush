package builtins

import (
	"context"
	"fmt"
	"io"
)

// JobLister is implemented by the evaluator's job manager.
// We use an interface to avoid circular imports.
type JobLister interface {
	ListJobs() []JobInfo
	KillJob(id int) error
	WaitJob(id int) error
	RemoveJob(id int) bool
	CleanupJobs() int
}

type JobInfo struct {
	ID      int
	Command string
	Status  string
	PID     int
}

// Global job lister, set by the evaluator package at init time.
var jobLister JobLister

func SetJobLister(jl JobLister) {
	jobLister = jl
}

// JobsCommand lists background jobs.
type JobsCommand struct{}

func (c *JobsCommand) Execute(ctx context.Context, args []string, out io.Writer, errOut io.Writer) error {
	if jobLister == nil {
		fmt.Fprintln(out, "no job manager available")
		return nil
	}

	jobs := jobLister.ListJobs()
	if len(jobs) == 0 {
		fmt.Fprintln(out, "no background jobs")
		return nil
	}

	for _, j := range jobs {
		if j.PID > 0 {
			fmt.Fprintf(out, "[%d]\t%s\tPID %d\t%s\n", j.ID, j.Status, j.PID, j.Command)
		} else {
			fmt.Fprintf(out, "[%d]\t%s\t\t%s\n", j.ID, j.Status, j.Command)
		}
	}
	return nil
}

// FgCommand brings a background job to the foreground (waits for it).
type FgCommand struct{}

func (c *FgCommand) Execute(ctx context.Context, args []string, out io.Writer, errOut io.Writer) error {
	if jobLister == nil {
		return fmt.Errorf("no job manager available")
	}
	if len(args) == 0 {
		return fmt.Errorf("usage: fg <job-id>")
	}

	id := 0
	_, err := fmt.Sscanf(args[0], "%d", &id)
	if err != nil {
		return fmt.Errorf("invalid job id: %s", args[0])
	}

	return jobLister.WaitJob(id)
}

// KillJobCommand kills a background job.
type KillJobCommand struct{}

func (c *KillJobCommand) Execute(ctx context.Context, args []string, out io.Writer, errOut io.Writer) error {
	if jobLister == nil {
		return fmt.Errorf("no job manager available")
	}
	if len(args) == 0 {
		return fmt.Errorf("usage: killjob <job-id>")
	}

	id := 0
	_, err := fmt.Sscanf(args[0], "%d", &id)
	if err != nil {
		return fmt.Errorf("invalid job id: %s", args[0])
	}

	return jobLister.KillJob(id)
}

// CleanJobsCommand removes completed jobs from the list.
type CleanJobsCommand struct{}

func (c *CleanJobsCommand) Execute(ctx context.Context, args []string, out io.Writer, errOut io.Writer) error {
	if jobLister == nil {
		return fmt.Errorf("no job manager available")
	}

	n := jobLister.CleanupJobs()
	fmt.Fprintf(out, "removed %d finished job(s)\n", n)
	return nil
}

// WaitCommand waits for background jobs to finish.
type WaitCommand struct{}

func (c *WaitCommand) Execute(ctx context.Context, args []string, out io.Writer, errOut io.Writer) error {
	if jobLister == nil {
		return fmt.Errorf("no job manager available")
	}

	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			fmt.Fprintln(out, "Usage: wait [job-id...]\nWait for background jobs to finish. With no arguments, waits for all running jobs.")
			return nil
		}
	}

	if len(args) == 0 {
		// Wait for all running jobs
		jobs := jobLister.ListJobs()
		for _, j := range jobs {
			if j.Status == "running" {
				if err := jobLister.WaitJob(j.ID); err != nil {
					fmt.Fprintf(errOut, "wait: job %d: %v\n", j.ID, err)
				}
			}
		}
		return nil
	}

	for _, arg := range args {
		id := 0
		_, err := fmt.Sscanf(arg, "%d", &id)
		if err != nil {
			fmt.Fprintf(errOut, "wait: invalid job id: %s\n", arg)
			continue
		}
		if err := jobLister.WaitJob(id); err != nil {
			fmt.Fprintf(errOut, "wait: %v\n", err)
		}
	}
	return nil
}

// DisownCommand removes a job from the shell's job table.
type DisownCommand struct{}

func (c *DisownCommand) Execute(ctx context.Context, args []string, out io.Writer, errOut io.Writer) error {
	if jobLister == nil {
		return fmt.Errorf("no job manager available")
	}

	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			fmt.Fprintln(out, "Usage: disown <job-id>\nRemove a job from the shell's job table.")
			return nil
		}
	}

	if len(args) == 0 {
		return fmt.Errorf("usage: disown <job-id>")
	}

	id := 0
	_, err := fmt.Sscanf(args[0], "%d", &id)
	if err != nil {
		return fmt.Errorf("invalid job id: %s", args[0])
	}

	if !jobLister.RemoveJob(id) {
		return fmt.Errorf("disown: no such job: %d", id)
	}
	return nil
}

func init() {
	RegisterBuiltin("jobs", &JobsCommand{})
	RegisterBuiltin("fg", &FgCommand{})
	RegisterBuiltin("killjob", &KillJobCommand{})
	RegisterBuiltin("cleanjobs", &CleanJobsCommand{})
	RegisterBuiltin("wait", &WaitCommand{})
	RegisterBuiltin("disown", &DisownCommand{})
}
