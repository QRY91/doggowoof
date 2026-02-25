package deepbore

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/QRY91/doggowoof/pkg/alerts"
	"github.com/gen2brain/beeep"
)

// TailerConfig holds runtime config for the log tailer.
type TailerConfig struct {
	LogDirectory        string
	PollInterval        time.Duration
	BinaryCheckInterval time.Duration
	ConfigCheckInterval time.Duration
	BinaryPath          string
	AgentConfigDir      string
	ForbiddenTools      []string
	Parser              *Parser
}

// Tailer watches bridge logs and emits events.
type Tailer struct {
	cfg          TailerConfig
	store        *alerts.Store
	currentFile  string
	cursorOffset int64
	sessionName  string
	baselineHash string
	lineNum      int
}

// NewTailer creates a tailer that will write alerts to the given store.
func NewTailer(cfg TailerConfig, store *alerts.Store) *Tailer {
	return &Tailer{
		cfg:   cfg,
		store: store,
	}
}

// Run starts the watch loop. Blocks until ctx is cancelled.
func (t *Tailer) Run(ctx context.Context) error {
	// Find latest log file
	logFile, err := t.findLatestLog()
	if err != nil {
		return fmt.Errorf("finding log file: %w", err)
	}
	t.currentFile = logFile
	t.sessionName = sessionNameFromFile(logFile)

	// Restore cursor from DB if resuming
	t.restoreCursor(ctx)

	fmt.Printf("Tailing: %s (offset: %d)\n", filepath.Base(t.currentFile), t.cursorOffset)

	// Initial checks
	t.checkBinary(ctx)
	t.checkConfigs(ctx)

	// Create tickers
	pollTicker := time.NewTicker(t.cfg.PollInterval)
	defer pollTicker.Stop()

	binaryTicker := time.NewTicker(t.cfg.BinaryCheckInterval)
	defer binaryTicker.Stop()

	configTicker := time.NewTicker(t.cfg.ConfigCheckInterval)
	defer configTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			t.persistCursor(ctx)
			return nil
		case <-pollTicker.C:
			t.pollLog(ctx)
		case <-binaryTicker.C:
			t.checkBinary(ctx)
		case <-configTicker.C:
			t.checkConfigs(ctx)
		}
	}
}

// findLatestLog globs bridge-*.log files and returns the newest.
func (t *Tailer) findLatestLog() (string, error) {
	pattern := filepath.Join(t.cfg.LogDirectory, "bridge-*.log")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return "", fmt.Errorf("globbing logs: %w", err)
	}
	if len(matches) == 0 {
		return "", fmt.Errorf("no bridge-*.log files in %s", t.cfg.LogDirectory)
	}
	sort.Strings(matches) // lexicographic = chronological for YYYY-MM-DD_HHMMSS
	return matches[len(matches)-1], nil
}

// pollLog reads new bytes from the current log file.
func (t *Tailer) pollLog(ctx context.Context) {
	// Check for log rotation
	latest, err := t.findLatestLog()
	if err == nil && latest != t.currentFile {
		fmt.Printf("Log rotation: %s -> %s\n", filepath.Base(t.currentFile), filepath.Base(latest))
		t.persistCursor(ctx)
		t.currentFile = latest
		t.cursorOffset = 0
		t.lineNum = 0
		t.sessionName = sessionNameFromFile(latest)
	}

	f, err := os.Open(t.currentFile)
	if err != nil {
		return
	}
	defer f.Close()

	// Check if file was truncated
	fi, err := f.Stat()
	if err != nil {
		return
	}
	if fi.Size() < t.cursorOffset {
		t.cursorOffset = 0
		t.lineNum = 0
	}

	// Seek to cursor
	if t.cursorOffset > 0 {
		if _, err := f.Seek(t.cursorOffset, io.SeekStart); err != nil {
			return
		}
	}

	scanner := bufio.NewScanner(f)
	eventsThisPoll := 0

	for scanner.Scan() {
		t.lineNum++
		line := scanner.Text()
		event := t.cfg.Parser.ParseLine(line, t.lineNum)
		if event != nil {
			t.handleEvent(ctx, event)
			eventsThisPoll++
		}
	}

	// Update cursor to current position
	newOffset, err := f.Seek(0, io.SeekCurrent)
	if err == nil {
		t.cursorOffset = newOffset
	}

	if eventsThisPoll > 0 {
		t.persistCursor(ctx)
	}
}

// handleEvent processes a parsed event.
func (t *Tailer) handleEvent(ctx context.Context, event *Event) {
	// Map severity
	alertSeverity := event.Severity
	switch event.Severity {
	case SeverityHigh:
		alertSeverity = "high"
	case SeverityCritical:
		alertSeverity = "critical"
	case SeverityInfo:
		alertSeverity = "low"
	}

	// Save alert
	alert := &alerts.Alert{
		AlertType:   string(event.Type),
		Severity:    alertSeverity,
		Title:       formatEventTitle(event),
		Description: event.Detail,
		Confidence:  1.0,
		CreatedAt:   alerts.TimeNow(),
	}
	t.store.SaveAlert(ctx, alert)

	// Update session state from task events
	t.updateSessionFromEvent(ctx, event)

	// Print to stdout
	icon := severityIcon(alertSeverity)
	fmt.Printf("%s %s %s\n", icon, event.Type, event.Detail)

	// Desktop notification for critical/high
	if alertSeverity == "critical" || alertSeverity == "high" {
		_ = beeep.Notify(
			"DoggoWoof — Deep Bore",
			fmt.Sprintf("%s: %s", event.Type, event.Detail),
			"",
		)
	}
}

// checkBinary computes SHA-256 and compares to baseline.
func (t *Tailer) checkBinary(ctx context.Context) {
	if t.cfg.BinaryPath == "" {
		return
	}

	check := CheckBinaryHash(t.cfg.BinaryPath)
	if check.Error != nil {
		fmt.Printf("[?] binary check error: %v\n", check.Error)
		return
	}
	if !check.Exists {
		fmt.Printf("[?] binary not found: %s\n", t.cfg.BinaryPath)
		return
	}

	if t.baselineHash == "" {
		// First check — set baseline
		t.baselineHash = check.Hash
		fmt.Printf("[ok] binary baseline: %s...%s\n", check.Hash[:8], check.Hash[len(check.Hash)-8:])
		t.updateSessionBinaryHash(ctx, check.Hash)
		return
	}

	if check.Hash != t.baselineHash {
		event := &Event{
			Type:      EventBinaryMismatch,
			Severity:  SeverityCritical,
			Timestamp: time.Now().UTC(),
			Detail:    fmt.Sprintf("hash changed: %s... -> %s...", t.baselineHash[:12], check.Hash[:12]),
		}
		t.handleEvent(ctx, event)
		t.baselineHash = check.Hash
		t.updateSessionBinaryHash(ctx, check.Hash)
	}
}

// checkConfigs validates agent config files.
func (t *Tailer) checkConfigs(ctx context.Context) {
	if t.cfg.AgentConfigDir == "" {
		return
	}

	_, violations := ValidateAgentConfigs(t.cfg.AgentConfigDir, t.cfg.ForbiddenTools)
	for i := range violations {
		t.handleEvent(ctx, &violations[i])
	}

	if len(violations) > 0 {
		t.updateSessionSandbox(ctx, false)
	}
}

// restoreCursor loads the cursor from the database if a session exists.
func (t *Tailer) restoreCursor(ctx context.Context) {
	sess, err := t.store.GetDeepBoreSession(ctx, t.sessionName)
	if err != nil || sess == nil {
		return
	}
	if sess.LogFile == filepath.Base(t.currentFile) {
		t.cursorOffset = sess.CursorOffset
		t.baselineHash = sess.BinaryHash
	}
}

// persistCursor saves the current cursor position to the database.
func (t *Tailer) persistCursor(ctx context.Context) {
	now := alerts.TimeNow()
	sess := &alerts.DeepBoreSession{
		SessionName:  t.sessionName,
		LogFile:      filepath.Base(t.currentFile),
		CursorOffset: t.cursorOffset,
		BinaryHash:   t.baselineHash,
		SandboxOK:    true,
		StartedAt:    now,
		UpdatedAt:    now,
	}
	t.store.UpsertDeepBoreSession(ctx, sess)
}

// updateSessionFromEvent updates session state based on event data.
func (t *Tailer) updateSessionFromEvent(ctx context.Context, event *Event) {
	sess, _ := t.store.GetDeepBoreSession(ctx, t.sessionName)
	if sess == nil {
		return
	}

	switch event.Type {
	case EventTaskDispatch:
		info := ParseTaskInfo(event.Detail)
		if info != nil {
			sess.TaskCurrent = info.Current
			sess.TaskTotal = info.Total
			sess.TaskName = info.Name
		}
	case EventTaskComplete:
		info := ParseTaskComplete(event.Detail)
		if info != nil {
			sess.Turns += info.Turns
			sess.CostDollars += info.Cost
			sess.UptimeSeconds += info.Seconds
		}
	case EventBroadcast:
		if event.Agent != "" {
			sess.AgentName = event.Agent
		}
	case EventSandboxViolation, EventSelfModification:
		sess.SandboxOK = false
	}

	sess.UpdatedAt = alerts.TimeNow()
	t.store.UpsertDeepBoreSession(ctx, sess)
}

func (t *Tailer) updateSessionBinaryHash(ctx context.Context, hash string) {
	sess, _ := t.store.GetDeepBoreSession(ctx, t.sessionName)
	if sess == nil {
		return
	}
	sess.BinaryHash = hash
	sess.UpdatedAt = alerts.TimeNow()
	t.store.UpsertDeepBoreSession(ctx, sess)
}

func (t *Tailer) updateSessionSandbox(ctx context.Context, ok bool) {
	sess, _ := t.store.GetDeepBoreSession(ctx, t.sessionName)
	if sess == nil {
		return
	}
	sess.SandboxOK = ok
	sess.UpdatedAt = alerts.TimeNow()
	t.store.UpsertDeepBoreSession(ctx, sess)
}

// sessionNameFromFile derives a session name from a log file path.
func sessionNameFromFile(path string) string {
	base := filepath.Base(path)
	// Remove .log extension
	if len(base) > 4 && base[len(base)-4:] == ".log" {
		return base[:len(base)-4]
	}
	return base
}

// formatEventTitle creates a human-readable title for an event.
func formatEventTitle(event *Event) string {
	switch event.Type {
	case EventSandboxViolation:
		return fmt.Sprintf("SANDBOX VIOLATION: %s", event.Detail)
	case EventSelfModification:
		return "SELF-MODIFICATION DETECTED"
	case EventBinaryMismatch:
		return "BINARY INTEGRITY MISMATCH"
	case EventConfigViolation:
		return fmt.Sprintf("CONFIG VIOLATION: %s", event.Detail)
	case EventAgentBugReport:
		return "Agent bug report"
	case EventAgentAnomaly:
		return "Agent anomaly"
	case EventAgentStart:
		return "Agent session started"
	case EventTaskDispatch:
		return fmt.Sprintf("Task dispatched: %s", event.Detail)
	case EventTaskComplete:
		return fmt.Sprintf("Task completed: %s", event.Detail)
	case EventBroadcast:
		return event.Detail
	default:
		return string(event.Type)
	}
}

func severityIcon(severity string) string {
	switch severity {
	case "critical":
		return "[!!]"
	case "high":
		return "[!]"
	case "medium":
		return "[~]"
	case "low":
		return "[.]"
	default:
		return "[?]"
	}
}
