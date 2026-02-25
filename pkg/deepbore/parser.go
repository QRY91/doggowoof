package deepbore

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// EventType classifies a parsed log line.
type EventType string

const (
	EventAgentStart       EventType = "agent_start"
	EventTaskComplete     EventType = "task_complete"
	EventTaskDispatch     EventType = "task_dispatch"
	EventSandboxViolation EventType = "sandbox_violation"
	EventSelfModification EventType = "self_modification"
	EventAgentBugReport   EventType = "agent_bug_report"
	EventAgentAnomaly     EventType = "agent_anomaly"
	EventBroadcast        EventType = "agent_broadcast"
	EventBinaryMismatch   EventType = "binary_mismatch"
	EventConfigViolation  EventType = "config_violation"
)

const (
	SeverityCritical = "critical"
	SeverityHigh     = "high"
	SeverityInfo     = "low"
)

// Event represents a parsed log event.
type Event struct {
	Type      EventType
	Severity  string
	Timestamp time.Time
	Line      string
	LineNum   int
	Detail    string
	Agent     string
}

// TaskInfo holds parsed task dispatch/complete data.
type TaskInfo struct {
	Name    string
	Current int
	Total   int
	Cost    float64
	Turns   int
	Seconds int
}

// ParserConfig holds the configurable detection patterns.
type ParserConfig struct {
	ForbiddenTools     []string
	SelfModPatterns    []string
	EscalationPatterns []string
}

// Parser parses bridge log lines into events.
type Parser struct {
	cfg            ParserConfig
	forbiddenRe    *regexp.Regexp
	selfModStrings []string
	escalationRe   *regexp.Regexp
	timestampRe    *regexp.Regexp
	agentStartRe   *regexp.Regexp
	taskCompleteRe *regexp.Regexp
	taskDispatchRe *regexp.Regexp
	broadcastRe    *regexp.Regexp
	chainRe        *regexp.Regexp
	toolCallRe     *regexp.Regexp
}

// NewParser creates a parser with compiled patterns.
func NewParser(cfg ParserConfig) *Parser {
	// Build forbidden tool pattern: tool: (Bash|Edit|Read|Write|...)(
	toolAlts := strings.Join(cfg.ForbiddenTools, "|")
	forbiddenRe := regexp.MustCompile(fmt.Sprintf(`tool: (%s)\(`, toolAlts))

	// Escalation patterns as alternation
	escParts := make([]string, len(cfg.EscalationPatterns))
	for i, p := range cfg.EscalationPatterns {
		escParts[i] = regexp.QuoteMeta(p)
	}
	escalationRe := regexp.MustCompile(strings.Join(escParts, "|"))

	return &Parser{
		cfg:            cfg,
		forbiddenRe:    forbiddenRe,
		selfModStrings: cfg.SelfModPatterns,
		escalationRe:   escalationRe,
		timestampRe:    regexp.MustCompile(`^\[(\d{2}:\d{2}:\d{2})\]\s*(.*)$`),
		agentStartRe:   regexp.MustCompile(`Spawning claude \(new session\)`),
		taskCompleteRe: regexp.MustCompile(`done: \$(\d+\.?\d*) \| (\d+) turns? \| (\d+)s`),
		taskDispatchRe: regexp.MustCompile(`Chain: dispatching '([^']+)' \((\d+)/(\d+)\)`),
		broadcastRe:    regexp.MustCompile(`^\[([\w][\w-]*)\]\s+(.+)$`),
		chainRe:        regexp.MustCompile(`^\[chain -> ([\w-]+)\]\s+(.+)$`),
		toolCallRe:     regexp.MustCompile(`^tool: \w+\(`),
	}
}

// ParseLine parses a single log line and returns an event, or nil if unremarkable.
func (p *Parser) ParseLine(line string, lineNum int) *Event {
	if line == "" {
		return nil
	}

	ts, body := p.parseTimestamp(line)

	// Priority 1: Self-modification (most specific critical)
	// Only within tool calls to avoid false positives on agent text
	if p.toolCallRe.MatchString(body) {
		for _, pat := range p.selfModStrings {
			if strings.Contains(body, pat) {
				return &Event{
					Type:      EventSelfModification,
					Severity:  SeverityCritical,
					Timestamp: ts,
					Line:      line,
					LineNum:   lineNum,
					Detail:    body,
				}
			}
		}
	}

	// Priority 2: Sandbox violation (forbidden tool usage)
	if matches := p.forbiddenRe.FindStringSubmatch(body); matches != nil {
		return &Event{
			Type:      EventSandboxViolation,
			Severity:  SeverityCritical,
			Timestamp: ts,
			Line:      line,
			LineNum:   lineNum,
			Detail:    fmt.Sprintf("forbidden tool: %s", matches[1]),
		}
	}

	// Priority 3: Structured info events
	if p.agentStartRe.MatchString(body) {
		return &Event{
			Type:      EventAgentStart,
			Severity:  SeverityInfo,
			Timestamp: ts,
			Line:      line,
			LineNum:   lineNum,
			Detail:    "new agent session",
		}
	}

	if matches := p.taskCompleteRe.FindStringSubmatch(body); matches != nil {
		return &Event{
			Type:      EventTaskComplete,
			Severity:  SeverityInfo,
			Timestamp: ts,
			Line:      line,
			LineNum:   lineNum,
			Detail:    fmt.Sprintf("cost=$%s turns=%s duration=%ss", matches[1], matches[2], matches[3]),
		}
	}

	if matches := p.taskDispatchRe.FindStringSubmatch(body); matches != nil {
		return &Event{
			Type:      EventTaskDispatch,
			Severity:  SeverityInfo,
			Timestamp: ts,
			Line:      line,
			LineNum:   lineNum,
			Detail:    fmt.Sprintf("task '%s' (%s/%s)", matches[1], matches[2], matches[3]),
		}
	}

	// Priority 4: Escalation patterns (in any text)
	if p.escalationRe.MatchString(body) {
		// Check if it's more specifically in a broadcast
		evType := EventAgentBugReport
		if strings.Contains(body, "ANOMALY") {
			evType = EventAgentAnomaly
		}
		return &Event{
			Type:      evType,
			Severity:  SeverityHigh,
			Timestamp: ts,
			Line:      line,
			LineNum:   lineNum,
			Detail:    body,
		}
	}

	// Priority 5: Broadcasts (agent messages)
	if matches := p.chainRe.FindStringSubmatch(body); matches != nil {
		return &Event{
			Type:      EventBroadcast,
			Severity:  SeverityInfo,
			Timestamp: ts,
			Line:      line,
			LineNum:   lineNum,
			Detail:    matches[2],
			Agent:     matches[1],
		}
	}

	if matches := p.broadcastRe.FindStringSubmatch(body); matches != nil {
		// Skip matches that look like log-level prefixes or tool calls
		name := matches[1]
		if name == "tool" || name == "text" || name == "done" {
			return nil
		}
		return &Event{
			Type:      EventBroadcast,
			Severity:  SeverityInfo,
			Timestamp: ts,
			Line:      line,
			LineNum:   lineNum,
			Detail:    matches[2],
			Agent:     strings.ToLower(name),
		}
	}

	return nil
}

// ParseTaskInfo extracts task info from a task_dispatch event detail string.
func ParseTaskInfo(detail string) *TaskInfo {
	re := regexp.MustCompile(`task '([^']+)' \((\d+)/(\d+)\)`)
	matches := re.FindStringSubmatch(detail)
	if matches == nil {
		return nil
	}
	current, _ := strconv.Atoi(matches[2])
	total, _ := strconv.Atoi(matches[3])
	return &TaskInfo{
		Name:    matches[1],
		Current: current,
		Total:   total,
	}
}

// ParseTaskComplete extracts cost/turns/duration from a task_complete event detail.
func ParseTaskComplete(detail string) *TaskInfo {
	re := regexp.MustCompile(`cost=\$(\d+\.?\d*) turns=(\d+) duration=(\d+)s`)
	matches := re.FindStringSubmatch(detail)
	if matches == nil {
		return nil
	}
	cost, _ := strconv.ParseFloat(matches[1], 64)
	turns, _ := strconv.Atoi(matches[2])
	secs, _ := strconv.Atoi(matches[3])
	return &TaskInfo{
		Cost:    cost,
		Turns:   turns,
		Seconds: secs,
	}
}

// parseTimestamp extracts time from [HH:MM:SS] prefix.
// Returns zero time and the full line if no timestamp found.
func (p *Parser) parseTimestamp(line string) (time.Time, string) {
	matches := p.timestampRe.FindStringSubmatch(line)
	if matches == nil {
		return time.Time{}, line
	}

	t, err := time.Parse("15:04:05", matches[1])
	if err != nil {
		return time.Time{}, matches[2]
	}

	// Anchor to today's date
	now := time.Now().UTC()
	t = time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), t.Second(), 0, time.UTC)
	return t, matches[2]
}

// DateFromLogFilename extracts the date from a bridge log filename.
// Expected format: bridge-YYYY-MM-DD_HHMMSS.log
func DateFromLogFilename(filename string) (time.Time, error) {
	re := regexp.MustCompile(`bridge-(\d{4}-\d{2}-\d{2})_(\d{2})(\d{2})(\d{2})\.log`)
	matches := re.FindStringSubmatch(filename)
	if matches == nil {
		return time.Time{}, fmt.Errorf("cannot parse date from filename: %s", filename)
	}
	dateStr := fmt.Sprintf("%sT%s:%s:%sZ", matches[1], matches[2], matches[3], matches[4])
	return time.Parse(time.RFC3339, dateStr)
}
