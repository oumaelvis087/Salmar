package automation

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/elvisouma/salmar-ai/pkg/models"
)

// TaskScheduler manages scheduled tasks and automation
type TaskScheduler struct {
	tasks      map[string]*ScheduledTask
	running    bool
	mu         sync.Mutex
	cancelFunc context.CancelFunc
}

// ScheduledTask represents a task scheduled for execution
type ScheduledTask struct {
	ID          string
	Name        string
	Description string
	Schedule    string // cron-like expression
	Request     *models.Request
	LastRun     time.Time
	NextRun     time.Time
	Handler     TaskHandler
	Active      bool
}

// TaskHandler defines the interface for task execution
type TaskHandler interface {
	Execute(ctx context.Context, task *ScheduledTask) error
}

// NewTaskScheduler creates a new task scheduler
func NewTaskScheduler() *TaskScheduler {
	return &TaskScheduler{
		tasks:   make(map[string]*ScheduledTask),
		running: false,
	}
}

// AddTask adds a new task to the scheduler
func (s *TaskScheduler) AddTask(task *ScheduledTask) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate task
	if task.ID == "" {
		return fmt.Errorf("task ID cannot be empty")
	}

	if task.Handler == nil {
		return fmt.Errorf("task handler cannot be nil")
	}

	// Calculate next run time based on schedule
	nextRun, err := calculateNextRun(task.Schedule)
	if err != nil {
		return err
	}
	task.NextRun = nextRun

	// Add task to map
	s.tasks[task.ID] = task
	return nil
}

// RemoveTask removes a task from the scheduler
func (s *TaskScheduler) RemoveTask(taskID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tasks[taskID]; !exists {
		return fmt.Errorf("task with ID %s not found", taskID)
	}

	delete(s.tasks, taskID)
	return nil
}

// Start begins the scheduler
func (s *TaskScheduler) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("scheduler is already running")
	}

	// Create a cancellable context
	ctx, cancel := context.WithCancel(ctx)
	s.cancelFunc = cancel
	s.running = true
	s.mu.Unlock()

	// Start the scheduler loop
	go s.scheduleLoop(ctx)
	return nil
}

// Stop halts the scheduler
func (s *TaskScheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	if s.cancelFunc != nil {
		s.cancelFunc()
		s.cancelFunc = nil
	}
	s.running = false
}

// scheduleLoop runs the main scheduling loop
func (s *TaskScheduler) scheduleLoop(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.checkAndRunTasks(ctx)
		}
	}
}

// checkAndRunTasks checks for due tasks and runs them
func (s *TaskScheduler) checkAndRunTasks(ctx context.Context) {
	s.mu.Lock()
	now := time.Now()
	var tasksToRun []*ScheduledTask

	// Find tasks that are due
	for _, task := range s.tasks {
		if task.Active && !task.NextRun.After(now) {
			tasksToRun = append(tasksToRun, task)
		}
	}
	s.mu.Unlock()

	// Run the due tasks
	for _, task := range tasksToRun {
		go func(t *ScheduledTask) {
			err := t.Handler.Execute(ctx, t)
			
			s.mu.Lock()
			defer s.mu.Unlock()
			
			// Update task status
			t.LastRun = now
			nextRun, err := calculateNextRun(t.Schedule)
			if err == nil {
				t.NextRun = nextRun
			}
		}(task)
	}
}

// calculateNextRun determines the next run time based on a schedule string
// This is a simplified implementation - a real one would use cron expressions
func calculateNextRun(schedule string) (time.Time, error) {
	// For this example, we'll just add 24 hours
	// In a real implementation, this would parse cron expressions
	return time.Now().Add(24 * time.Hour), nil
}

// GetAllTasks returns all scheduled tasks
func (s *TaskScheduler) GetAllTasks() []*ScheduledTask {
	s.mu.Lock()
	defer s.mu.Unlock()

	tasks := make([]*ScheduledTask, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, task)
	}
	return tasks
}

// GetTask returns a specific task by ID
func (s *TaskScheduler) GetTask(taskID string) (*ScheduledTask, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("task with ID %s not found", taskID)
	}
	return task, nil
}