package evaluator

import (
	"dush/internal/builtins"
	"fmt"
	"os/exec"
	"sync"
)

type JobStatus int

const (
	JobRunning JobStatus = iota
	JobDone
	JobFailed
)

func (s JobStatus) String() string {
	switch s {
	case JobRunning:
		return "running"
	case JobDone:
		return "done"
	case JobFailed:
		return "failed"
	}
	return "unknown"
}

type Job struct {
	ID      int
	Command string
	Cmd     *exec.Cmd
	Status  JobStatus
	Error   error
	Done    chan struct{} // closed when the job finishes
}

type JobManager struct {
	mu     sync.Mutex
	jobs   map[int]*Job
	nextID int
}

var Jobs = &JobManager{
	jobs:   make(map[int]*Job),
	nextID: 1,
}

func (jm *JobManager) Add(command string, cmd *exec.Cmd) *Job {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	job := &Job{
		ID:      jm.nextID,
		Command: command,
		Cmd:     cmd,
		Status:  JobRunning,
		Done:    make(chan struct{}),
	}
	jm.jobs[jm.nextID] = job
	jm.nextID++
	return job
}

func (jm *JobManager) Get(id int) *Job {
	jm.mu.Lock()
	defer jm.mu.Unlock()
	return jm.jobs[id]
}

func (jm *JobManager) List() []*Job {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	// Collect and return sorted by ID
	result := make([]*Job, 0, len(jm.jobs))
	for i := 1; i < jm.nextID; i++ {
		if j, ok := jm.jobs[i]; ok {
			result = append(result, j)
		}
	}
	return result
}

func (jm *JobManager) Remove(id int) bool {
	jm.mu.Lock()
	defer jm.mu.Unlock()
	if _, ok := jm.jobs[id]; ok {
		delete(jm.jobs, id)
		return true
	}
	return false
}

// Cleanup removes all finished jobs and returns how many were removed.
func (jm *JobManager) Cleanup() int {
	jm.mu.Lock()
	defer jm.mu.Unlock()
	count := 0
	for id, j := range jm.jobs {
		if j.Status != JobRunning {
			delete(jm.jobs, id)
			count++
		}
	}
	return count
}

// MarkDone updates a job's status and closes its Done channel.
func (jm *JobManager) MarkDone(id int, err error) {
	jm.mu.Lock()
	defer jm.mu.Unlock()
	if j, ok := jm.jobs[id]; ok {
		if err != nil {
			j.Status = JobFailed
			j.Error = err
		} else {
			j.Status = JobDone
		}
		close(j.Done)
	}
}

// Kill terminates a running job.
func (jm *JobManager) Kill(id int) error {
	jm.mu.Lock()
	j, ok := jm.jobs[id]
	jm.mu.Unlock()

	if !ok {
		return fmt.Errorf("no such job: %d", id)
	}
	if j.Status != JobRunning {
		return fmt.Errorf("job %d is not running", id)
	}
	if j.Cmd == nil || j.Cmd.Process == nil {
		return fmt.Errorf("job %d has no process", id)
	}
	return j.Cmd.Process.Kill()
}

// --- builtins.JobLister interface ---

func (jm *JobManager) ListJobs() []builtins.JobInfo {
	jobs := jm.List()
	infos := make([]builtins.JobInfo, len(jobs))
	for i, j := range jobs {
		pid := 0
		if j.Cmd != nil && j.Cmd.Process != nil {
			pid = j.Cmd.Process.Pid
		}
		infos[i] = builtins.JobInfo{
			ID:      j.ID,
			Command: j.Command,
			Status:  j.Status.String(),
			PID:     pid,
		}
	}
	return infos
}

func (jm *JobManager) KillJob(id int) error {
	return jm.Kill(id)
}

func (jm *JobManager) WaitJob(id int) error {
	j := jm.Get(id)
	if j == nil {
		return fmt.Errorf("no such job: %d", id)
	}
	if j.Status != JobRunning {
		return fmt.Errorf("job %d is not running (%s)", id, j.Status)
	}
	<-j.Done
	if j.Error != nil {
		return j.Error
	}
	return nil
}

func (jm *JobManager) RemoveJob(id int) bool {
	return jm.Remove(id)
}

func (jm *JobManager) CleanupJobs() int {
	return jm.Cleanup()
}

func init() {
	builtins.SetJobLister(Jobs)
}
