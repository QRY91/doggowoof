package cli

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/QRY91/doggowoof/internal/config"
	"github.com/QRY91/doggowoof/pkg/alerts"
	"github.com/QRY91/doggowoof/pkg/osmotic"
)

// Stores holds the config, osmotic reader, and alerts store.
type Stores struct {
	Config  *config.Config
	Osmotic *osmotic.Reader
	Alerts  *alerts.Store
}

// LoadConfig loads just the config file.
func LoadConfig() (*config.Config, error) {
	return config.Load()
}

// LoadConfigAndStores opens the config, osmotic reader (read-only), and alerts store.
func LoadConfigAndStores() (*Stores, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	osmoticReader, err := osmotic.NewReader(cfg.Osmotic.DatabasePath)
	if err != nil {
		return nil, fmt.Errorf("opening osmotic database: %w", err)
	}

	alertsStore, err := alerts.New(cfg.Alerts.DatabasePath)
	if err != nil {
		osmoticReader.Close()
		return nil, fmt.Errorf("opening alerts database: %w", err)
	}

	return &Stores{
		Config:  cfg,
		Osmotic: osmoticReader,
		Alerts:  alertsStore,
	}, nil
}

// LoadConfigAndAlerts opens the config and alerts store only (no osmotic).
// Used by commands that don't need osmotic data (e.g., bore).
func LoadConfigAndAlerts() (*Stores, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	alertsStore, err := alerts.New(cfg.Alerts.DatabasePath)
	if err != nil {
		return nil, fmt.Errorf("opening alerts database: %w", err)
	}

	return &Stores{
		Config: cfg,
		Alerts: alertsStore,
	}, nil
}

// Close closes all open database connections.
func (s *Stores) Close() {
	if s.Osmotic != nil {
		s.Osmotic.Close()
	}
	if s.Alerts != nil {
		s.Alerts.Close()
	}
}

// FormatTimeAgo returns a human-friendly relative time string.
func FormatTimeAgo(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		mins := int(diff.Minutes())
		if mins == 1 {
			return "1m ago"
		}
		return fmt.Sprintf("%dm ago", mins)
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		if hours == 1 {
			return "1h ago"
		}
		return fmt.Sprintf("%dh ago", hours)
	case diff < 7*24*time.Hour:
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "yesterday"
		}
		return fmt.Sprintf("%dd ago", days)
	default:
		return t.Format("Jan 2")
	}
}

// ParseISO parses an ISO 8601 / RFC 3339 timestamp string.
func ParseISO(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}

// parseFlags parses --key=value flags from os.Args.
func parseFlags(args []string) map[string]string {
	flags := make(map[string]string)
	for _, arg := range args {
		if strings.HasPrefix(arg, "--") {
			arg = arg[2:]
			if idx := strings.Index(arg, "="); idx >= 0 {
				flags[arg[:idx]] = arg[idx+1:]
			} else {
				flags[arg] = "true"
			}
		}
	}
	return flags
}

// fatalf prints to stderr and exits.
func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

// severityIcon returns a symbol for a severity level.
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
