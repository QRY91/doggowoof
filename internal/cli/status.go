package cli

import (
	"context"
	"fmt"
)

// RunStatus shows a dashboard of Deep Bore session state and alerts.
func RunStatus(args []string) {
	stores, err := LoadStores()
	if err != nil {
		fatalf("Error: %v", err)
	}
	defer stores.Close()

	ctx := context.Background()

	fmt.Println("=== doggo status ===")
	fmt.Println()

	showDeepBoreSession(ctx, stores)
	showAlertsSummary(ctx, stores)
	showRecentAlerts(ctx, stores, 5)
}

func showDeepBoreSession(ctx context.Context, stores *Stores) {
	cfg := stores.Config.DeepBore
	fmt.Println("Deep Bore:")
	if !cfg.Enabled {
		fmt.Println("  disabled — set [deepbore] enabled = true in config to activate")
		fmt.Println()
		return
	}

	sess, err := stores.Alerts.GetLatestDeepBoreSession(ctx)
	if err != nil {
		fmt.Printf("  (error reading: %v)\n\n", err)
		return
	}
	if sess == nil {
		fmt.Println("  enabled, no session yet — run 'doggo bore --watch'")
		fmt.Println()
		return
	}

	fmt.Printf("  Session: %s\n", sess.SessionName)
	if sess.AgentName != "" {
		fmt.Printf("  Agent:   %s\n", sess.AgentName)
	}
	if sess.SandboxOK {
		fmt.Println("  Sandbox: [ok] enabled")
	} else {
		fmt.Println("  Sandbox: [!!] VIOLATION DETECTED")
	}
	if sess.BinaryHash != "" {
		hash := sess.BinaryHash
		if len(hash) > 12 {
			hash = hash[:12]
		}
		fmt.Printf("  Binary:  [ok] verified (sha256: %s...)\n", hash)
	}
	if sess.TaskTotal > 0 {
		fmt.Printf("  Task:    %d/%d (%s)\n", sess.TaskCurrent, sess.TaskTotal, sess.TaskName)
	}
	if t, err := ParseISO(sess.UpdatedAt); err == nil {
		fmt.Printf("  Updated: %s\n", FormatTimeAgo(t))
	}
	fmt.Println()
}

func showAlertsSummary(ctx context.Context, stores *Stores) {
	fmt.Println("Alerts:")

	total, _ := stores.Alerts.CountAlerts(ctx)
	if total == 0 {
		fmt.Println("  No alerts yet.")
		fmt.Println()
		return
	}

	bySeverity, _ := stores.Alerts.CountAlertsBySeverity(ctx)

	fmt.Printf("  Total: %d", total)
	if len(bySeverity) > 0 {
		fmt.Print(" (")
		for i, sc := range bySeverity {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Printf("%s %s", severityIcon(sc.Severity), sc.Severity)
			if sc.Count > 1 {
				fmt.Printf(": %d", sc.Count)
			}
		}
		fmt.Print(")")
	}
	fmt.Println()
	fmt.Println()
}

func showRecentAlerts(ctx context.Context, stores *Stores, limit int) {
	recent, _ := stores.Alerts.RecentAlerts(ctx, limit)
	if len(recent) == 0 {
		return
	}

	fmt.Printf("Recent Alerts (last %d):\n", len(recent))
	for _, a := range recent {
		ago := ""
		if t, err := ParseISO(a.CreatedAt); err == nil {
			ago = FormatTimeAgo(t)
		}
		ack := ""
		if a.Acknowledged {
			ack = " (acked)"
		}
		fmt.Printf("  %s #%d %s — %s%s\n",
			severityIcon(a.Severity), a.ID, a.Title, ago, ack)
	}
	fmt.Println()
}

