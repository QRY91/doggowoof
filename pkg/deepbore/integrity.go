package deepbore

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// BinaryCheck holds the result of a binary integrity check.
type BinaryCheck struct {
	Path   string
	Hash   string
	Exists bool
	Error  error
}

// CheckBinaryHash computes SHA-256 of the file at path.
func CheckBinaryHash(path string) BinaryCheck {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return BinaryCheck{Path: path, Exists: false}
		}
		return BinaryCheck{Path: path, Error: err}
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return BinaryCheck{Path: path, Exists: true, Error: err}
	}

	return BinaryCheck{
		Path:   path,
		Hash:   hex.EncodeToString(h.Sum(nil)),
		Exists: true,
	}
}

// AgentConfig represents a parsed agent configuration file.
// Sandbox is a pointer to distinguish missing (nil) from explicit false.
// A missing sandbox field is a config error, not a default.
type AgentConfig struct {
	Name            string   `json:"name"`
	Sandbox         *bool    `json:"sandbox"`
	DisallowedTools []string `json:"disallowedTools"`
	FilePath        string   `json:"-"`
}

// ValidateAgentConfigs reads and validates all agent config files in a directory.
// Returns parsed configs and a slice of violation events for any failures.
func ValidateAgentConfigs(dir string, requiredTools []string) ([]AgentConfig, []Event) {
	pattern := filepath.Join(dir, "*.json")
	matches, err := filepath.Glob(pattern)
	if err != nil || len(matches) == 0 {
		return nil, nil
	}

	now := time.Now().UTC()
	var configs []AgentConfig
	var violations []Event

	for _, path := range matches {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var cfg AgentConfig
		if err := json.Unmarshal(data, &cfg); err != nil {
			continue
		}
		cfg.FilePath = path
		configs = append(configs, cfg)

		// Check sandbox field exists and is enabled
		if cfg.Sandbox == nil {
			violations = append(violations, Event{
				Type:      EventConfigViolation,
				Severity:  SeverityCritical,
				Timestamp: now,
				Detail:    fmt.Sprintf("agent %q: missing required 'sandbox' field in %s", cfg.Name, filepath.Base(path)),
				Agent:     cfg.Name,
			})
		} else if !*cfg.Sandbox {
			violations = append(violations, Event{
				Type:      EventConfigViolation,
				Severity:  SeverityCritical,
				Timestamp: now,
				Detail:    fmt.Sprintf("agent %q: sandbox=false in %s", cfg.Name, filepath.Base(path)),
				Agent:     cfg.Name,
			})
		}

		// Check disallowed tools
		disallowed := make(map[string]bool)
		for _, tool := range cfg.DisallowedTools {
			disallowed[tool] = true
		}
		for _, required := range requiredTools {
			if !disallowed[required] {
				violations = append(violations, Event{
					Type:      EventConfigViolation,
					Severity:  SeverityCritical,
					Timestamp: now,
					Detail:    fmt.Sprintf("agent %q: missing %q in disallowedTools", cfg.Name, required),
					Agent:     cfg.Name,
				})
				break // one violation per config is enough
			}
		}
	}

	return configs, violations
}
