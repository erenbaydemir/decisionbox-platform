package handler

import (
	"os"
	"sync"
)

// ProcessTracker tracks spawned agent subprocesses by run ID.
// Used to cancel/kill running discoveries.
type ProcessTracker struct {
	mu        sync.RWMutex
	processes map[string]*os.Process // runID -> process
}

func NewProcessTracker() *ProcessTracker {
	return &ProcessTracker{
		processes: make(map[string]*os.Process),
	}
}

// Track registers a process for a run ID.
func (t *ProcessTracker) Track(runID string, process *os.Process) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.processes[runID] = process
}

// Remove removes tracking for a run ID (called when process completes).
func (t *ProcessTracker) Remove(runID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.processes, runID)
}

// Kill kills the process for a run ID. Returns true if process was found and killed.
func (t *ProcessTracker) Kill(runID string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	proc, ok := t.processes[runID]
	if !ok {
		return false
	}

	proc.Kill()
	delete(t.processes, runID)
	return true
}

// IsRunning checks if a process is tracked for a run ID.
func (t *ProcessTracker) IsRunning(runID string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	_, ok := t.processes[runID]
	return ok
}
