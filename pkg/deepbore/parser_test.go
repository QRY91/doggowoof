package deepbore

import (
	"testing"
)

func testParser() *Parser {
	return NewParser(ParserConfig{
		ForbiddenTools:     []string{"Bash", "Edit", "Read", "Write", "Grep", "Glob", "Task"},
		SelfModPatterns:    []string{"cargo build", "cargo test", "rustc", "gcc", "make"},
		EscalationPatterns: []string{"BUG REPORT", "DEFECT", "BLOCKED", "Self-Repair"},
	})
}

func TestParseLineEvents(t *testing.T) {
	p := testParser()

	tests := []struct {
		name     string
		line     string
		wantType EventType
		wantSev  string
	}{
		{
			name:     "agent start",
			line:     "[20:25:15] Spawning claude (new session)",
			wantType: EventAgentStart,
			wantSev:  SeverityInfo,
		},
		{
			name:     "task complete",
			line:     "[20:49:37] done: $2.14 | 45 turns | 312s",
			wantType: EventTaskComplete,
			wantSev:  SeverityInfo,
		},
		{
			name:     "task dispatch",
			line:     "[20:50:00] Chain: dispatching 'belt-to-assembler' (24/33)",
			wantType: EventTaskDispatch,
			wantSev:  SeverityInfo,
		},
		{
			name:     "sandbox violation bash",
			line:     `[20:50:38] tool: Bash({"command":"cd /home/qry/projects/..."})`,
			wantType: EventSandboxViolation,
			wantSev:  SeverityCritical,
		},
		{
			name:     "sandbox violation edit",
			line:     `[20:50:13] tool: Edit({"file_path":"/home/qry/projects/..."})`,
			wantType: EventSandboxViolation,
			wantSev:  SeverityCritical,
		},
		{
			name:     "sandbox violation glob",
			line:     `[21:00:00] tool: Glob({"pattern":"**/*.rs"})`,
			wantType: EventSandboxViolation,
			wantSev:  SeverityCritical,
		},
		{
			name:     "self modification cargo build",
			line:     `[20:51:00] tool: Bash({"command":"cargo build --release"})`,
			wantType: EventSelfModification,
			wantSev:  SeverityCritical,
		},
		{
			name:     "self modification cargo test",
			line:     `[20:51:00] tool: Bash({"command":"cargo test"})`,
			wantType: EventSelfModification,
			wantSev:  SeverityCritical,
		},
		{
			name:     "bug report in broadcast",
			line:     "[21:00:00] [DOUG-NAUVIS] BUG REPORT: iron inserter not placing",
			wantType: EventAgentBugReport,
			wantSev:  SeverityHigh,
		},
		{
			name:     "defect in text",
			line:     "[21:05:00] text: DEFECT detected in belt placement",
			wantType: EventAgentBugReport,
			wantSev:  SeverityHigh,
		},
		{
			name:     "broadcast",
			line:     "[20:52:43] [DOUG-NAUVIS] COAL DEDUCTED. 121 to 119.",
			wantType: EventBroadcast,
			wantSev:  SeverityInfo,
		},
		{
			name:     "chain broadcast",
			line:     "[21:10:00] [chain -> doug-nauvis] DEEP BORE DIRECTIVE: mine iron",
			wantType: EventBroadcast,
			wantSev:  SeverityInfo,
		},
		{
			name: "normal tool call ignored",
			line: "[20:25:15] tool: get_character({})",
		},
		{
			name: "empty line ignored",
			line: "",
		},
		{
			name: "plain text ignored",
			line: "[20:25:15] text: thinking about iron",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := p.ParseLine(tt.line, 1)
			if tt.wantType == "" {
				if event != nil {
					t.Errorf("expected nil event, got %s", event.Type)
				}
				return
			}
			if event == nil {
				t.Fatal("expected event, got nil")
			}
			if event.Type != tt.wantType {
				t.Errorf("type = %s, want %s", event.Type, tt.wantType)
			}
			if event.Severity != tt.wantSev {
				t.Errorf("severity = %s, want %s", event.Severity, tt.wantSev)
			}
		})
	}
}

func TestSelfModificationPriority(t *testing.T) {
	p := testParser()

	// cargo build inside a Bash call should be self_modification, not sandbox_violation
	line := `[20:50:38] tool: Bash({"command":"cd /home/qry && cargo build --release"})`
	event := p.ParseLine(line, 1)
	if event == nil {
		t.Fatal("expected event, got nil")
	}
	if event.Type != EventSelfModification {
		t.Errorf("expected self_modification, got %s", event.Type)
	}
}

func TestParseTimestamp(t *testing.T) {
	p := testParser()

	tests := []struct {
		line     string
		wantHour int
		wantMin  int
		wantSec  int
		wantBody string
	}{
		{"[20:25:15] some text", 20, 25, 15, "some text"},
		{"[00:00:00] midnight", 0, 0, 0, "midnight"},
		{"no timestamp here", 0, 0, 0, "no timestamp here"},
	}

	for _, tt := range tests {
		ts, body := p.parseTimestamp(tt.line)
		if tt.wantBody != body {
			t.Errorf("parseTimestamp(%q) body = %q, want %q", tt.line, body, tt.wantBody)
		}
		if tt.wantHour > 0 && ts.Hour() != tt.wantHour {
			t.Errorf("parseTimestamp(%q) hour = %d, want %d", tt.line, ts.Hour(), tt.wantHour)
		}
	}
}

func TestDateFromLogFilename(t *testing.T) {
	tests := []struct {
		filename string
		wantErr  bool
		wantYear int
	}{
		{"bridge-2026-02-23_202508.log", false, 2026},
		{"bridge-2025-12-31_235959.log", false, 2025},
		{"random-file.log", true, 0},
		{"bridge.log", true, 0},
	}

	for _, tt := range tests {
		ts, err := DateFromLogFilename(tt.filename)
		if tt.wantErr {
			if err == nil {
				t.Errorf("DateFromLogFilename(%q) expected error", tt.filename)
			}
			continue
		}
		if err != nil {
			t.Errorf("DateFromLogFilename(%q) unexpected error: %v", tt.filename, err)
			continue
		}
		if ts.Year() != tt.wantYear {
			t.Errorf("DateFromLogFilename(%q) year = %d, want %d", tt.filename, ts.Year(), tt.wantYear)
		}
	}
}

func TestParseTaskInfo(t *testing.T) {
	info := ParseTaskInfo("task 'belt-to-assembler' (24/33)")
	if info == nil {
		t.Fatal("expected task info, got nil")
	}
	if info.Name != "belt-to-assembler" {
		t.Errorf("name = %q, want 'belt-to-assembler'", info.Name)
	}
	if info.Current != 24 || info.Total != 33 {
		t.Errorf("progress = %d/%d, want 24/33", info.Current, info.Total)
	}
}

func TestParseTaskComplete(t *testing.T) {
	info := ParseTaskComplete("cost=$2.14 turns=45 duration=312s")
	if info == nil {
		t.Fatal("expected task info, got nil")
	}
	if info.Cost != 2.14 {
		t.Errorf("cost = %f, want 2.14", info.Cost)
	}
	if info.Turns != 45 {
		t.Errorf("turns = %d, want 45", info.Turns)
	}
	if info.Seconds != 312 {
		t.Errorf("seconds = %d, want 312", info.Seconds)
	}
}
