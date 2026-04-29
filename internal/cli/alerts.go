package cli

import (
	"context"
	"fmt"
	"strconv"
)

// RunAlerts lists or acknowledges alerts.
func RunAlerts(args []string) {
	flags := parseFlags(args)

	stores, err := LoadStores()
	if err != nil {
		fatalf("Error: %v", err)
	}
	defer stores.Close()

	ctx := context.Background()

	// Acknowledge mode
	if ackStr, ok := flags["ack"]; ok {
		id, err := strconv.Atoi(ackStr)
		if err != nil {
			fatalf("Invalid alert ID: %s", ackStr)
		}
		if err := stores.Alerts.AcknowledgeAlert(ctx, id); err != nil {
			fatalf("Error: %v", err)
		}
		fmt.Printf("Alert #%d acknowledged.\n", id)
		return
	}

	// List mode
	severity := flags["severity"]
	limit := 20
	if l, ok := flags["limit"]; ok {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}

	alertList, err := stores.Alerts.ListAlerts(ctx, severity, limit)
	if err != nil {
		fatalf("Error: %v", err)
	}

	if len(alertList) == 0 {
		if severity != "" {
			fmt.Printf("No %s alerts found.\n", severity)
		} else {
			fmt.Println("No alerts found. Run 'doggo bore --watch' to start monitoring.")
		}
		return
	}

	fmt.Printf("Alerts")
	if severity != "" {
		fmt.Printf(" (severity: %s)", severity)
	}
	fmt.Printf(":\n\n")

	for _, a := range alertList {
		ago := ""
		if t, err := ParseISO(a.CreatedAt); err == nil {
			ago = FormatTimeAgo(t)
		}
		ack := ""
		if a.Acknowledged {
			ack = " [acked]"
		}
		fmt.Printf("  %s #%-4d %s\n", severityIcon(a.Severity), a.ID, a.Title)
		if a.Description != "" {
			fmt.Printf("           %s\n", a.Description)
		}
		fmt.Printf("           %s  confidence: %.0f%%%s\n", ago, a.Confidence*100, ack)
		fmt.Println()
	}
}
