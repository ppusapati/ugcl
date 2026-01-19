package engine

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"

	"p9e.in/ugcl/scheduler/models"
)

// CronEngine manages cron-based job scheduling
type CronEngine struct {
	cron      *cron.Cron
	jobMap    map[uuid.UUID]cron.EntryID // job ID -> cron entry ID
	jobMutex  sync.RWMutex
	executor  JobExecutor
	running   bool
}

// JobExecutor interface for executing jobs
type JobExecutor interface {
	ExecuteJob(ctx context.Context, job *models.Job) error
}

// NewCronEngine creates a new cron engine
func NewCronEngine(executor JobExecutor) *CronEngine {
	// Create cron with second precision and timezone support
	c := cron.New(cron.WithSeconds(), cron.WithLocation(time.UTC))

	return &CronEngine{
		cron:     c,
		jobMap:   make(map[uuid.UUID]cron.EntryID),
		executor: executor,
	}
}

// Start starts the cron engine
func (e *CronEngine) Start(ctx context.Context) error {
	e.jobMutex.Lock()
	defer e.jobMutex.Unlock()

	if e.running {
		return nil
	}

	log.Println("Starting cron engine...")
	e.cron.Start()
	e.running = true

	// Monitor context cancellation
	go func() {
		<-ctx.Done()
		e.Stop()
	}()

	return nil
}

// Stop stops the cron engine
func (e *CronEngine) Stop() error {
	e.jobMutex.Lock()
	defer e.jobMutex.Unlock()

	if !e.running {
		return nil
	}

	log.Println("Stopping cron engine...")
	ctx := e.cron.Stop()
	<-ctx.Done() // Wait for all running jobs to complete

	e.running = false
	return nil
}

// AddJob adds a job to the cron scheduler
func (e *CronEngine) AddJob(job *models.Job) error {
	e.jobMutex.Lock()
	defer e.jobMutex.Unlock()

	// Remove existing job if it exists
	if entryID, exists := e.jobMap[job.ID]; exists {
		e.cron.Remove(entryID)
		delete(e.jobMap, job.ID)
	}

	// Only add active jobs
	if !job.IsActive {
		return nil
	}

	// Create job function
	jobFunc := func() {
		ctx := context.Background()
		if err := e.executor.ExecuteJob(ctx, job); err != nil {
			log.Printf("Job execution failed for %s: %v", job.Name, err)
		}
	}

	// Add to cron scheduler
	entryID, err := e.cron.AddFunc(job.CronExpression, jobFunc)
	if err != nil {
		return err
	}

	// Store mapping
	e.jobMap[job.ID] = entryID

	log.Printf("Job %s (%s) added to scheduler with cron: %s", job.Name, job.ID, job.CronExpression)
	return nil
}

// RemoveJob removes a job from the cron scheduler
func (e *CronEngine) RemoveJob(jobID uuid.UUID) error {
	e.jobMutex.Lock()
	defer e.jobMutex.Unlock()

	if entryID, exists := e.jobMap[jobID]; exists {
		e.cron.Remove(entryID)
		delete(e.jobMap, jobID)
		log.Printf("Job %s removed from scheduler", jobID)
	}

	return nil
}

// UpdateJob updates an existing job in the scheduler
func (e *CronEngine) UpdateJob(job *models.Job) error {
	// Remove and re-add the job
	if err := e.RemoveJob(job.ID); err != nil {
		return err
	}

	return e.AddJob(job)
}

// GetNextRunTime calculates the next run time for a cron expression
func (e *CronEngine) GetNextRunTime(cronExpr string) (*time.Time, error) {
	schedule, err := cron.ParseStandard(cronExpr)
	if err != nil {
		return nil, err
	}

	nextRun := schedule.Next(time.Now())
	return &nextRun, nil
}

// ValidateCronExpression validates a cron expression
func (e *CronEngine) ValidateCronExpression(cronExpr string) error {
	_, err := cron.ParseStandard(cronExpr)
	return err
}

// GetActiveJobs returns the count of active jobs in the scheduler
func (e *CronEngine) GetActiveJobs() int {
	e.jobMutex.RLock()
	defer e.jobMutex.RUnlock()

	return len(e.jobMap)
}

// IsRunning returns whether the cron engine is running
func (e *CronEngine) IsRunning() bool {
	e.jobMutex.RLock()
	defer e.jobMutex.RUnlock()

	return e.running
}