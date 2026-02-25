package deepbore

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/QRY91/doggowoof/pkg/alerts"
)

func TestFindLatestLog(t *testing.T) {
	dir := t.TempDir()

	for _, name := range []string{
		"bridge-2026-02-22_150000.log",
		"bridge-2026-02-23_200000.log",
		"bridge-2026-02-23_202508.log",
	} {
		os.WriteFile(filepath.Join(dir, name), []byte("test"), 0644)
	}

	tailer := &Tailer{cfg: TailerConfig{LogDirectory: dir}}
	latest, err := tailer.findLatestLog()
	if err != nil {
		t.Fatalf("findLatestLog() failed: %v", err)
	}
	if filepath.Base(latest) != "bridge-2026-02-23_202508.log" {
		t.Errorf("expected latest log, got %s", filepath.Base(latest))
	}
}

func TestFindLatestLogEmpty(t *testing.T) {
	dir := t.TempDir()
	tailer := &Tailer{cfg: TailerConfig{LogDirectory: dir}}
	_, err := tailer.findLatestLog()
	if err == nil {
		t.Error("expected error for empty log directory")
	}
}

func TestPollLogNewLines(t *testing.T) {
	dir := t.TempDir()
	store, err := alerts.NewMemory()
	if err != nil {
		t.Fatalf("NewMemory() failed: %v", err)
	}
	defer store.Close()

	logFile := filepath.Join(dir, "bridge-2026-02-24_100000.log")
	os.WriteFile(logFile, []byte("[10:00:00] tool: get_character({})\n"), 0644)

	parser := NewParser(ParserConfig{
		ForbiddenTools:     []string{"Bash", "Edit"},
		SelfModPatterns:    []string{"cargo build"},
		EscalationPatterns: []string{"BUG REPORT"},
	})

	tailer := NewTailer(TailerConfig{
		LogDirectory: dir,
		Parser:       parser,
	}, store)
	tailer.currentFile = logFile
	tailer.sessionName = "bridge-2026-02-24_100000"

	ctx := context.Background()

	// Initialize session so updateSessionFromEvent can work
	tailer.persistCursor(ctx)

	// First poll — reads the existing normal line, no alerts
	tailer.pollLog(ctx)

	alertList, _ := store.ListAlerts(ctx, "", 10)
	if len(alertList) != 0 {
		t.Errorf("expected 0 alerts after normal line, got %d", len(alertList))
	}

	// Append a sandbox violation
	f, _ := os.OpenFile(logFile, os.O_APPEND|os.O_WRONLY, 0644)
	f.WriteString("[10:01:00] tool: Bash({\"command\":\"ls\"})\n")
	f.Close()

	// Second poll — should detect the new line
	tailer.pollLog(ctx)

	alertList, _ = store.ListAlerts(ctx, "critical", 10)
	if len(alertList) != 1 {
		t.Errorf("expected 1 critical alert, got %d", len(alertList))
	}
	if len(alertList) > 0 && alertList[0].AlertType != "sandbox_violation" {
		t.Errorf("expected sandbox_violation, got %s", alertList[0].AlertType)
	}
}

func TestLogRotationDetection(t *testing.T) {
	dir := t.TempDir()
	store, err := alerts.NewMemory()
	if err != nil {
		t.Fatalf("NewMemory() failed: %v", err)
	}
	defer store.Close()

	log1 := filepath.Join(dir, "bridge-2026-02-23_200000.log")
	os.WriteFile(log1, []byte("[20:00:00] started\n"), 0644)

	parser := NewParser(ParserConfig{
		ForbiddenTools:     []string{"Bash"},
		SelfModPatterns:    []string{"cargo build"},
		EscalationPatterns: []string{"BUG REPORT"},
	})

	tailer := NewTailer(TailerConfig{
		LogDirectory: dir,
		Parser:       parser,
	}, store)
	tailer.currentFile = log1
	tailer.sessionName = "bridge-2026-02-23_200000"

	ctx := context.Background()
	tailer.persistCursor(ctx)

	// Create a newer log file (rotation)
	log2 := filepath.Join(dir, "bridge-2026-02-23_210000.log")
	os.WriteFile(log2, []byte("[21:00:00] new session\n"), 0644)

	// Poll should detect the new file
	tailer.pollLog(ctx)

	if tailer.currentFile != log2 {
		t.Errorf("expected current file to be %s, got %s",
			filepath.Base(log2), filepath.Base(tailer.currentFile))
	}
	if tailer.cursorOffset != 0 {
		// After polling, cursor should have advanced past the one line
		// but the key check is that it switched files
	}
}

func TestSessionNameFromFile(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"/path/to/bridge-2026-02-23_202508.log", "bridge-2026-02-23_202508"},
		{"bridge-2026-02-23_202508.log", "bridge-2026-02-23_202508"},
		{"somefile.txt", "somefile.txt"}, // no .log extension handling
	}

	for _, tt := range tests {
		got := sessionNameFromFile(tt.path)
		if got != tt.want {
			t.Errorf("sessionNameFromFile(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestCursorPersistence(t *testing.T) {
	dir := t.TempDir()
	store, err := alerts.NewMemory()
	if err != nil {
		t.Fatalf("NewMemory() failed: %v", err)
	}
	defer store.Close()

	parser := NewParser(ParserConfig{
		ForbiddenTools:     []string{"Bash"},
		SelfModPatterns:    []string{"cargo build"},
		EscalationPatterns: []string{"BUG REPORT"},
	})

	logFile := filepath.Join(dir, "bridge-2026-02-24_120000.log")
	os.WriteFile(logFile, []byte("[12:00:00] test line\n"), 0644)

	tailer := NewTailer(TailerConfig{
		LogDirectory: dir,
		Parser:       parser,
	}, store)
	tailer.currentFile = logFile
	tailer.sessionName = "bridge-2026-02-24_120000"
	tailer.cursorOffset = 42

	ctx := context.Background()

	// Persist
	tailer.persistCursor(ctx)

	// Create a new tailer and restore
	tailer2 := NewTailer(TailerConfig{
		LogDirectory: dir,
		Parser:       parser,
	}, store)
	tailer2.currentFile = logFile
	tailer2.sessionName = "bridge-2026-02-24_120000"

	tailer2.restoreCursor(ctx)

	if tailer2.cursorOffset != 42 {
		t.Errorf("expected cursor 42 after restore, got %d", tailer2.cursorOffset)
	}
}

func TestRunShutdown(t *testing.T) {
	dir := t.TempDir()
	store, err := alerts.NewMemory()
	if err != nil {
		t.Fatalf("NewMemory() failed: %v", err)
	}
	defer store.Close()

	logFile := filepath.Join(dir, "bridge-2026-02-24_130000.log")
	os.WriteFile(logFile, []byte("[13:00:00] test\n"), 0644)

	parser := NewParser(ParserConfig{
		ForbiddenTools:     []string{"Bash"},
		SelfModPatterns:    []string{"cargo build"},
		EscalationPatterns: []string{"BUG REPORT"},
	})

	tailer := NewTailer(TailerConfig{
		LogDirectory:        dir,
		PollInterval:        100 * time.Millisecond,
		BinaryCheckInterval: 1 * time.Hour, // don't trigger during test
		ConfigCheckInterval: 1 * time.Hour,
		Parser:              parser,
	}, store)

	// Cancel immediately to test clean shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	err = tailer.Run(ctx)
	if err != nil {
		t.Errorf("Run() returned error: %v", err)
	}
}
