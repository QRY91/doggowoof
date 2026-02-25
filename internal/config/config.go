package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// Config is the top-level configuration for doggowoof.
type Config struct {
	Osmotic       OsmoticConfig       `toml:"osmotic"`
	Alerts        AlertsConfig        `toml:"alerts"`
	Thresholds    ThresholdsConfig    `toml:"thresholds"`
	Notifications NotificationsConfig `toml:"notifications"`
	DeepBore      DeepBoreConfig      `toml:"deepbore"`
}

// OsmoticConfig points to osmotic's database (read-only).
type OsmoticConfig struct {
	DatabasePath string `toml:"database_path"`
}

// AlertsConfig holds the path to doggowoof's own database.
type AlertsConfig struct {
	DatabasePath string `toml:"database_path"`
}

// ThresholdsConfig holds detection sensitivity settings.
type ThresholdsConfig struct {
	CoordinationScore            float64 `toml:"coordination_score"`
	TemporalClusterWindowMinutes int     `toml:"temporal_cluster_window_minutes"`
	TemporalClusterMinAccounts   int     `toml:"temporal_cluster_min_accounts"`
}

// NotificationsConfig holds notification preferences.
type NotificationsConfig struct {
	Enabled bool `toml:"enabled"`
}

// DeepBoreConfig holds Deep Bore agent monitoring settings.
type DeepBoreConfig struct {
	Enabled              bool     `toml:"enabled"`
	LogDirectory         string   `toml:"log_directory"`
	BinaryPath           string   `toml:"binary_path"`
	AgentConfigDirectory string   `toml:"agent_config_directory"`
	PollIntervalSeconds  int      `toml:"poll_interval_seconds"`
	BinaryCheckSeconds   int      `toml:"binary_check_seconds"`
	ConfigCheckSeconds   int      `toml:"config_check_seconds"`
	ForbiddenTools       []string `toml:"forbidden_tools"`
	SelfModPatterns      []string `toml:"self_modification_patterns"`
	EscalationPatterns   []string `toml:"escalation_patterns"`
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	return &Config{
		Osmotic: OsmoticConfig{
			DatabasePath: "~/.local/share/osmotic/osmotic.sqlite",
		},
		Alerts: AlertsConfig{
			DatabasePath: "~/.local/share/doggowoof/alerts.sqlite",
		},
		Thresholds: ThresholdsConfig{
			CoordinationScore:            0.7,
			TemporalClusterWindowMinutes: 5,
			TemporalClusterMinAccounts:   3,
		},
		Notifications: NotificationsConfig{
			Enabled: true,
		},
		DeepBore: DeepBoreConfig{
			Enabled:              false,
			LogDirectory:         "~/projects/claude-in-factorio/logs/",
			BinaryPath:           "~/projects/claude-in-factorio/factorioctl/target/release/mcp",
			AgentConfigDirectory: "~/projects/claude-in-factorio/bridge/agents/",
			PollIntervalSeconds:  1,
			BinaryCheckSeconds:   60,
			ConfigCheckSeconds:   30,
			ForbiddenTools:       []string{"Bash", "Edit", "Read", "Write", "Grep", "Glob", "Task"},
			SelfModPatterns:      []string{"cargo build", "cargo test", "rustc", "gcc", "make"},
			EscalationPatterns:   []string{"BUG REPORT", "DEFECT", "BLOCKED", "Self-Repair"},
		},
	}
}

// ConfigDir returns the doggowoof config directory path.
func ConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "doggowoof"), nil
}

// ConfigPath returns the config file path.
func ConfigPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.toml"), nil
}

// ExpandPath replaces a leading ~ with the user's home directory.
func ExpandPath(path string) string {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[1:])
	}
	return path
}

// Load reads and parses the config file, applying defaults for missing values.
func Load() (*Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}

	cfg := DefaultConfig()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("config not found at %s — run 'doggo init'", path)
	}

	if _, err := toml.DecodeFile(path, cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	cfg.Osmotic.DatabasePath = ExpandPath(cfg.Osmotic.DatabasePath)
	cfg.Alerts.DatabasePath = ExpandPath(cfg.Alerts.DatabasePath)
	cfg.DeepBore.LogDirectory = ExpandPath(cfg.DeepBore.LogDirectory)
	cfg.DeepBore.BinaryPath = ExpandPath(cfg.DeepBore.BinaryPath)
	cfg.DeepBore.AgentConfigDirectory = ExpandPath(cfg.DeepBore.AgentConfigDirectory)

	return cfg, nil
}

// Init creates the default config file and database directories.
func Init() error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}

	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("config already exists at %s", path)
	}

	dir, err := ConfigDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}

	cfg := DefaultConfig()

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating config file: %w", err)
	}
	defer f.Close()

	enc := toml.NewEncoder(f)
	if err := enc.Encode(cfg); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	// Create the alerts database directory
	alertsDir := filepath.Dir(ExpandPath(cfg.Alerts.DatabasePath))
	if err := os.MkdirAll(alertsDir, 0755); err != nil {
		return fmt.Errorf("creating alerts database dir: %w", err)
	}

	fmt.Printf("Config created at %s\n", path)
	fmt.Printf("Alerts database directory at %s\n", alertsDir)

	return nil
}
