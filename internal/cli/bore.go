package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/QRY91/doggowoof/pkg/deepbore"
)

// RunBore handles the "doggo bore" command.
// Default: show Deep Bore status.
// --watch: start long-running log watcher.
func RunBore(args []string) {
	flags := parseFlags(args)

	stores, err := LoadStores()
	if err != nil {
		fatalf("Error: %v", err)
	}
	defer stores.Close()

	if !stores.Config.DeepBore.Enabled {
		fatalf("Deep Bore monitoring not enabled. Set [deepbore] enabled = true in config.")
	}

	if _, ok := flags["watch"]; ok {
		runBoreWatch(stores)
	} else {
		runBoreStatus(stores)
	}
}

func runBoreStatus(stores *Stores) {
	ctx := context.Background()
	cfg := stores.Config.DeepBore

	fmt.Println("=== DEEP BORE STATUS ===")
	fmt.Println()

	// Session info
	sess, err := stores.Alerts.GetLatestDeepBoreSession(ctx)
	if err != nil {
		fatalf("Error: %v", err)
	}
	if sess == nil {
		fmt.Println("No Deep Bore sessions found.")
		fmt.Println("Start monitoring with: doggo bore --watch")
		return
	}

	fmt.Printf("Session: %s\n", sess.SessionName)
	if sess.AgentName != "" {
		fmt.Printf("Agent:   %s\n", sess.AgentName)
	}

	// Sandbox status
	if sess.SandboxOK {
		fmt.Println("Sandbox: [ok] ENABLED")
	} else {
		fmt.Println("Sandbox: [!!] VIOLATION DETECTED")
	}

	// Binary status
	if sess.BinaryHash != "" {
		hashPreview := sess.BinaryHash
		if len(hashPreview) > 12 {
			hashPreview = hashPreview[:12]
		}
		fmt.Printf("Binary:  [ok] VERIFIED (sha256: %s...)\n", hashPreview)
	} else {
		fmt.Println("Binary:  [?] NOT CHECKED")
	}

	// Task progress
	if sess.TaskTotal > 0 {
		fmt.Printf("Task:    %d/%d (%s)\n", sess.TaskCurrent, sess.TaskTotal, sess.TaskName)
	}

	// Uptime, turns, cost
	if sess.Turns > 0 || sess.CostDollars > 0 {
		fmt.Printf("Stats:   %s | %d turns | $%.2f\n",
			formatDuration(time.Duration(sess.UptimeSeconds)*time.Second),
			sess.Turns, sess.CostDollars)
	}

	// Updated
	if t, err := ParseISO(sess.UpdatedAt); err == nil {
		fmt.Printf("Updated: %s\n", FormatTimeAgo(t))
	}
	fmt.Println()

	// Recent broadcasts
	broadcasts, _ := stores.Alerts.ListAlertsByType(ctx, "agent_broadcast", 5)
	if len(broadcasts) > 0 {
		fmt.Println("Recent broadcasts:")
		for _, b := range broadcasts {
			ts := ""
			if t, err := ParseISO(b.CreatedAt); err == nil {
				ts = t.Format("15:04:05")
			}
			fmt.Printf("  [%s] %s\n", ts, b.Description)
		}
		fmt.Println()
	}

	// Alert summary (deepbore types only)
	showBoreAlertSummary(ctx, stores)

	// Binary/config check paths
	fmt.Println("Config:")
	fmt.Printf("  Log dir:    %s\n", cfg.LogDirectory)
	fmt.Printf("  Binary:     %s\n", cfg.BinaryPath)
	fmt.Printf("  Agent cfgs: %s\n", cfg.AgentConfigDirectory)
}

func runBoreWatch(stores *Stores) {
	cfg := stores.Config.DeepBore

	parserCfg := deepbore.ParserConfig{
		ForbiddenTools:     cfg.ForbiddenTools,
		SelfModPatterns:    cfg.SelfModPatterns,
		EscalationPatterns: cfg.EscalationPatterns,
	}
	parser := deepbore.NewParser(parserCfg)

	tailerCfg := deepbore.TailerConfig{
		LogDirectory:        cfg.LogDirectory,
		PollInterval:        time.Duration(cfg.PollIntervalSeconds) * time.Second,
		BinaryCheckInterval: time.Duration(cfg.BinaryCheckSeconds) * time.Second,
		ConfigCheckInterval: time.Duration(cfg.ConfigCheckSeconds) * time.Second,
		BinaryPath:          cfg.BinaryPath,
		AgentConfigDir:      cfg.AgentConfigDirectory,
		ForbiddenTools:      cfg.ForbiddenTools,
		Parser:              parser,
	}

	tailer := deepbore.NewTailer(tailerCfg, stores.Alerts)

	// Signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		fmt.Printf("\nReceived %s, shutting down...\n", sig)
		cancel()
	}()

	fmt.Println("=== DEEP BORE WATCH ===")
	fmt.Printf("Monitoring: %s\n", cfg.LogDirectory)
	if cfg.BinaryPath != "" {
		fmt.Printf("Binary:     %s\n", cfg.BinaryPath)
	}
	fmt.Println("Press Ctrl+C to stop.")
	fmt.Println()

	if err := tailer.Run(ctx); err != nil {
		fatalf("Watch error: %v", err)
	}

	fmt.Println("Watch stopped. Good dog.")
}

func showBoreAlertSummary(ctx context.Context, stores *Stores) {
	// Count deepbore-related alerts by severity
	deepboreTypes := []string{
		"sandbox_violation", "self_modification", "binary_mismatch",
		"config_violation", "agent_bug_report", "agent_anomaly",
		"agent_start", "task_dispatch", "task_complete",
	}

	var critical, high, info int
	for _, at := range deepboreTypes {
		alertList, _ := stores.Alerts.ListAlertsByType(ctx, at, 0)
		for _, a := range alertList {
			switch a.Severity {
			case "critical":
				critical++
			case "high":
				high++
			default:
				info++
			}
		}
	}

	total := critical + high + info
	if total == 0 {
		fmt.Println("Alerts: none")
		return
	}

	fmt.Printf("Alerts: %d critical | %d warning | %d info\n", critical, high, info)

	// Show recent critical/high alerts
	if critical > 0 || high > 0 {
		fmt.Println()
		for _, at := range []string{"sandbox_violation", "self_modification", "binary_mismatch", "config_violation"} {
			alertList, _ := stores.Alerts.ListAlertsByType(ctx, at, 3)
			for _, a := range alertList {
				ago := ""
				if t, err := ParseISO(a.CreatedAt); err == nil {
					ago = FormatTimeAgo(t)
				}
				fmt.Printf("  %s %s — %s\n", severityIcon(a.Severity), a.Title, ago)
			}
		}
	}
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	if m == 0 {
		return fmt.Sprintf("%dh", h)
	}
	return fmt.Sprintf("%dh %dm", h, m)
}
