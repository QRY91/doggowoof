package cli

import (
	"context"
	"fmt"
)

// RunStatus shows a dashboard of osmotic data and doggowoof alerts.
func RunStatus(args []string) {
	stores, err := LoadConfigAndStores()
	if err != nil {
		fatalf("Error: %v", err)
	}
	defer stores.Close()

	ctx := context.Background()

	fmt.Println("=== doggo status ===")
	fmt.Println()

	// Osmotic database summary
	showOsmoticSummary(ctx, stores)

	// Alerts summary
	showAlertsSummary(ctx, stores)

	// Active clusters
	showActiveClusters(ctx, stores)

	// Recent alerts
	showRecentAlerts(ctx, stores, 5)
}

func showOsmoticSummary(ctx context.Context, stores *Stores) {
	fmt.Println("Osmotic Database:")
	itemCount, err := stores.Osmotic.CountItems(ctx)
	if err != nil {
		fmt.Printf("  (error reading: %v)\n\n", err)
		return
	}

	sourceCounts, _ := stores.Osmotic.CountItemsBySource(ctx)
	authorCount, _ := stores.Osmotic.CountAuthors(ctx)

	fmt.Printf("  Items: %d", itemCount)
	if len(sourceCounts) > 0 {
		fmt.Print(" (")
		for i, sc := range sourceCounts {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Printf("%s: %d", sc.Source, sc.Count)
		}
		fmt.Print(")")
	}
	fmt.Println()
	fmt.Printf("  Authors: %d\n", authorCount)
	fmt.Println()
}

func showAlertsSummary(ctx context.Context, stores *Stores) {
	fmt.Println("Alerts:")

	total, _ := stores.Alerts.CountAlerts(ctx)
	if total == 0 {
		fmt.Println("  No alerts yet. Run 'doggo sniff' to start detection.")
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

func showActiveClusters(ctx context.Context, stores *Stores) {
	clusters, _ := stores.Alerts.ActiveClusters(ctx)
	if len(clusters) == 0 {
		return
	}

	fmt.Println("Active Clusters:")
	for _, c := range clusters {
		fmt.Printf("  %s %s (score: %.2f, seen: %dx)\n",
			severityIcon(classifySeverity(c.Score)), c.ClusterType, c.Score, c.Occurrences)
	}
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
