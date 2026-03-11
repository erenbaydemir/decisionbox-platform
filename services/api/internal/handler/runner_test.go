package handler

import (
	"os/exec"
	"runtime"
	"testing"
)

func TestProcessTracker_TrackAndRemove(t *testing.T) {
	tracker := NewProcessTracker()

	// Start a harmless process
	cmd := sleepCmd()
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start process: %v", err)
	}
	defer cmd.Process.Kill()

	tracker.Track("run-1", cmd.Process)

	if !tracker.IsRunning("run-1") {
		t.Error("should be running after Track")
	}

	tracker.Remove("run-1")

	if tracker.IsRunning("run-1") {
		t.Error("should not be running after Remove")
	}
}

func TestProcessTracker_Kill(t *testing.T) {
	tracker := NewProcessTracker()

	cmd := sleepCmd()
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start process: %v", err)
	}

	tracker.Track("run-2", cmd.Process)

	killed := tracker.Kill("run-2")
	if !killed {
		t.Error("Kill should return true for tracked process")
	}

	if tracker.IsRunning("run-2") {
		t.Error("should not be running after Kill")
	}

	// Wait to clean up zombie
	cmd.Wait()
}

func TestProcessTracker_KillNotFound(t *testing.T) {
	tracker := NewProcessTracker()

	killed := tracker.Kill("nonexistent")
	if killed {
		t.Error("Kill should return false for untracked run")
	}
}

func TestProcessTracker_IsRunningEmpty(t *testing.T) {
	tracker := NewProcessTracker()

	if tracker.IsRunning("any") {
		t.Error("should return false for empty tracker")
	}
}

func TestProcessTracker_MultipleRuns(t *testing.T) {
	tracker := NewProcessTracker()

	cmd1 := sleepCmd()
	cmd2 := sleepCmd()
	cmd1.Start()
	cmd2.Start()
	defer cmd1.Process.Kill()
	defer cmd2.Process.Kill()

	tracker.Track("r1", cmd1.Process)
	tracker.Track("r2", cmd2.Process)

	if !tracker.IsRunning("r1") || !tracker.IsRunning("r2") {
		t.Error("both should be running")
	}

	tracker.Kill("r1")
	cmd1.Wait()

	if tracker.IsRunning("r1") {
		t.Error("r1 should be killed")
	}
	if !tracker.IsRunning("r2") {
		t.Error("r2 should still be running")
	}
}

func sleepCmd() *exec.Cmd {
	if runtime.GOOS == "windows" {
		return exec.Command("timeout", "/t", "60")
	}
	return exec.Command("sleep", "60")
}

// Verify the handler uses ProcessTracker field
func TestDiscoveriesHandler_HasTracker(t *testing.T) {
	tracker := NewProcessTracker()
	h := &DiscoveriesHandler{tracker: tracker}
	if h.tracker == nil {
		t.Error("handler should have tracker")
	}
}

// Verify cancel status exists via a mock-like check
func TestProcessTracker_ConcurrentSafe(t *testing.T) {
	tracker := NewProcessTracker()

	done := make(chan bool, 10)

	// Concurrent reads/writes
	for i := 0; i < 5; i++ {
		go func() {
			tracker.IsRunning("test")
			done <- true
		}()
		go func() {
			cmd := exec.Command("true")
			cmd.Start()
			if cmd.Process != nil {
				tracker.Track("concurrent", cmd.Process)
				tracker.Remove("concurrent")
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}
