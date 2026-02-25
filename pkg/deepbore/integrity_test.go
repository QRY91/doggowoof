package deepbore

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckBinaryHash(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test_binary")
	os.WriteFile(path, []byte("test content for hashing"), 0755)

	// Should succeed and return a hash
	check := CheckBinaryHash(path)
	if !check.Exists {
		t.Fatal("expected file to exist")
	}
	if check.Error != nil {
		t.Fatalf("unexpected error: %v", check.Error)
	}
	if check.Hash == "" {
		t.Fatal("expected non-empty hash")
	}

	// Same content = same hash
	check2 := CheckBinaryHash(path)
	if check2.Hash != check.Hash {
		t.Error("expected same hash for same content")
	}

	// Different content = different hash
	path2 := filepath.Join(dir, "other_binary")
	os.WriteFile(path2, []byte("different content"), 0755)
	check3 := CheckBinaryHash(path2)
	if check3.Hash == check.Hash {
		t.Error("expected different hash for different content")
	}

	// Missing file
	check4 := CheckBinaryHash(filepath.Join(dir, "nonexistent"))
	if check4.Exists {
		t.Error("expected file to not exist")
	}
}

func TestValidateAgentConfigs(t *testing.T) {
	dir := t.TempDir()

	// Valid config
	validJSON := `{"name":"doug-nauvis","sandbox":true,"disallowedTools":["Bash","Edit","Read","Write","Grep","Glob","Task"]}`
	os.WriteFile(filepath.Join(dir, "doug.json"), []byte(validJSON), 0644)

	// Invalid config — sandbox false
	invalidJSON := `{"name":"evil-doug","sandbox":false,"disallowedTools":[]}`
	os.WriteFile(filepath.Join(dir, "evil.json"), []byte(invalidJSON), 0644)

	// Invalid config — missing tools
	missingJSON := `{"name":"lazy-doug","sandbox":true,"disallowedTools":["Bash"]}`
	os.WriteFile(filepath.Join(dir, "lazy.json"), []byte(missingJSON), 0644)

	required := []string{"Bash", "Edit", "Read", "Write", "Grep", "Glob", "Task"}
	configs, violations := ValidateAgentConfigs(dir, required)

	if len(configs) != 3 {
		t.Errorf("expected 3 configs, got %d", len(configs))
	}

	if len(violations) == 0 {
		t.Fatal("expected violations")
	}

	// Check for sandbox violation
	foundSandbox := false
	foundMissing := false
	for _, v := range violations {
		if v.Type != EventConfigViolation {
			t.Errorf("expected config_violation event type, got %s", v.Type)
		}
		if v.Agent == "evil-doug" {
			foundSandbox = true
		}
		if v.Agent == "lazy-doug" {
			foundMissing = true
		}
	}
	if !foundSandbox {
		t.Error("expected sandbox violation for evil-doug")
	}
	if !foundMissing {
		t.Error("expected missing tools violation for lazy-doug")
	}
}

func TestValidateAgentConfigsEmptyDir(t *testing.T) {
	dir := t.TempDir()
	configs, violations := ValidateAgentConfigs(dir, []string{"Bash"})
	if len(configs) != 0 {
		t.Errorf("expected 0 configs, got %d", len(configs))
	}
	if len(violations) != 0 {
		t.Errorf("expected 0 violations, got %d", len(violations))
	}
}

func TestValidateAgentConfigsAllValid(t *testing.T) {
	dir := t.TempDir()
	validJSON := `{"name":"good-doug","sandbox":true,"disallowedTools":["Bash","Edit","Read","Write","Grep","Glob","Task"]}`
	os.WriteFile(filepath.Join(dir, "good.json"), []byte(validJSON), 0644)

	required := []string{"Bash", "Edit"}
	configs, violations := ValidateAgentConfigs(dir, required)

	if len(configs) != 1 {
		t.Errorf("expected 1 config, got %d", len(configs))
	}
	if len(violations) != 0 {
		t.Errorf("expected 0 violations for valid config, got %d", len(violations))
	}
}

func TestValidateAgentConfigsMissingSandboxField(t *testing.T) {
	dir := t.TempDir()

	// Config with no sandbox field at all — this is the cross-language bug.
	// Go unmarshals missing bool as false, Python defaults to True.
	// The fix: treat missing as a distinct violation from explicit false.
	noSandboxJSON := `{"name":"ambiguous-doug","disallowedTools":["Bash","Edit"]}`
	os.WriteFile(filepath.Join(dir, "ambiguous.json"), []byte(noSandboxJSON), 0644)

	_, violations := ValidateAgentConfigs(dir, []string{"Bash", "Edit"})

	if len(violations) != 1 {
		t.Fatalf("expected 1 violation for missing sandbox field, got %d", len(violations))
	}
	v := violations[0]
	if v.Agent != "ambiguous-doug" {
		t.Errorf("expected agent 'ambiguous-doug', got %q", v.Agent)
	}
	// Must say "missing" not just "sandbox=false"
	if v.Detail == "" {
		t.Fatal("expected non-empty detail")
	}
	if !strings.Contains(v.Detail, "missing") {
		t.Errorf("expected detail to mention 'missing', got: %s", v.Detail)
	}
}
