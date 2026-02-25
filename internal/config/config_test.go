package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Osmotic.DatabasePath == "" {
		t.Error("expected default osmotic database path")
	}
	if cfg.Alerts.DatabasePath == "" {
		t.Error("expected default alerts database path")
	}
	if cfg.Thresholds.CoordinationScore != 0.7 {
		t.Errorf("expected coordination score 0.7, got %f", cfg.Thresholds.CoordinationScore)
	}
	if cfg.Thresholds.TemporalClusterWindowMinutes != 5 {
		t.Errorf("expected temporal window 5, got %d", cfg.Thresholds.TemporalClusterWindowMinutes)
	}
	if cfg.Thresholds.TemporalClusterMinAccounts != 3 {
		t.Errorf("expected min accounts 3, got %d", cfg.Thresholds.TemporalClusterMinAccounts)
	}
	if !cfg.Notifications.Enabled {
		t.Error("expected notifications enabled by default")
	}

	// Deep Bore defaults
	if cfg.DeepBore.Enabled {
		t.Error("expected deepbore disabled by default")
	}
	if cfg.DeepBore.PollIntervalSeconds != 1 {
		t.Errorf("expected poll interval 1s, got %d", cfg.DeepBore.PollIntervalSeconds)
	}
	if cfg.DeepBore.BinaryCheckSeconds != 60 {
		t.Errorf("expected binary check 60s, got %d", cfg.DeepBore.BinaryCheckSeconds)
	}
	if len(cfg.DeepBore.ForbiddenTools) != 7 {
		t.Errorf("expected 7 forbidden tools, got %d", len(cfg.DeepBore.ForbiddenTools))
	}
	if len(cfg.DeepBore.SelfModPatterns) != 5 {
		t.Errorf("expected 5 self-mod patterns, got %d", len(cfg.DeepBore.SelfModPatterns))
	}
	if len(cfg.DeepBore.EscalationPatterns) != 4 {
		t.Errorf("expected 4 escalation patterns, got %d", len(cfg.DeepBore.EscalationPatterns))
	}
}

func TestExpandPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home directory")
	}

	tests := []struct {
		input string
		want  string
	}{
		{"~/foo/bar", filepath.Join(home, "foo/bar")},
		{"/absolute/path", "/absolute/path"},
		{"relative/path", "relative/path"},
	}

	for _, tt := range tests {
		got := ExpandPath(tt.input)
		if got != tt.want {
			t.Errorf("ExpandPath(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestInitAndLoad(t *testing.T) {
	// Use a temp directory so we don't affect real config
	tmpDir := t.TempDir()

	// Override config dir for testing
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Init should create config
	err := Init()
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	// Init again should fail (already exists)
	err = Init()
	if err == nil {
		t.Error("expected error on second Init()")
	}

	// Load should work
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.Thresholds.CoordinationScore != 0.7 {
		t.Errorf("expected default coordination score after load")
	}
}
