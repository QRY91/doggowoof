package cli

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "📊 Check what your guard dog has been up to",
	Long: `See what DOGGOWOOF has been monitoring:

- Daemon status (is your guard dog awake?)
- Recent alerts caught
- Alert breakdown by severity
- Quick statistics

Your guard dog's activity report!`,
	Run: func(cmd *cobra.Command, args []string) {
		showStatus()
	},
}

type AlertStats struct {
	Total    int
	High     int
	Medium   int
	Low      int
	Critical int
	Today    int
}

type RecentAlert struct {
	ID        int
	Source    string
	Severity  string
	Title     string
	Timestamp string
}

func showStatus() {
	fmt.Println("🚨🐕 DOGGOWOOF STATUS REPORT 🐕🚨")
	fmt.Println()

	// Check daemon status
	showDaemonStatus()
	fmt.Println()

	// Check database and show stats
	showAlertStats()
	fmt.Println()

	// Show recent alerts
	showRecentAlerts()
}

func showDaemonStatus() {
	fmt.Println("🐕 GUARD DOG STATUS:")

	isRunning, pid := isDaemonRunning()
	if isRunning {
		fmt.Printf("   🟢 DAEMON: RUNNING (PID: %d) - Guard dog is alert!\n", pid)
		fmt.Println("   📡 Listening on: http://localhost:8080")
	} else {
		fmt.Println("   🔴 DAEMON: SLEEPING - No guard dog watching")
		fmt.Println("   💡 Start with: doggo daemon start")
	}
}

func showAlertStats() {
	stats, err := getAlertStats()
	if err != nil {
		fmt.Printf("❌ Error reading alert stats: %v\n", err)
		return
	}

	fmt.Println("📊 ALERT STATISTICS:")
	fmt.Printf("   📈 Total alerts: %d\n", stats.Total)
	fmt.Printf("   🔥 Today: %d\n", stats.Today)
	fmt.Println()
	fmt.Println("   📋 By severity:")
	if stats.Critical > 0 {
		fmt.Printf("      🚨 CRITICAL: %d\n", stats.Critical)
	}
	if stats.High > 0 {
		fmt.Printf("      🔴 HIGH: %d\n", stats.High)
	}
	if stats.Medium > 0 {
		fmt.Printf("      🟡 MEDIUM: %d\n", stats.Medium)
	}
	if stats.Low > 0 {
		fmt.Printf("      🟢 LOW: %d\n", stats.Low)
	}

	if stats.Total == 0 {
		fmt.Println("   😴 No alerts yet - good boy is waiting!")
		fmt.Println("   💡 Test with: curl -X POST http://localhost:8080/webhook/github")
	}
}

func showRecentAlerts() {
	alerts, err := getRecentAlerts(5) // Last 5 alerts
	if err != nil {
		fmt.Printf("❌ Error reading recent alerts: %v\n", err)
		return
	}

	if len(alerts) == 0 {
		return // Already handled in stats
	}

	fmt.Println("🔔 RECENT ALERTS:")
	for _, alert := range alerts {
		icon := getSeverityIcon(alert.Severity)
		fmt.Printf("   %s [%s] %s: %s\n",
			icon, alert.Timestamp, alert.Source, alert.Title)
	}
}

func getSeverityIcon(severity string) string {
	switch severity {
	case "CRITICAL":
		return "🚨"
	case "HIGH":
		return "🔴"
	case "MEDIUM":
		return "🟡"
	case "LOW":
		return "🟢"
	default:
		return "📝"
	}
}

func getAlertStats() (*AlertStats, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	dbPath := filepath.Join(home, ".doggowoof", "alerts.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	stats := &AlertStats{}

	// Total alerts
	err = db.QueryRow("SELECT COUNT(*) FROM alerts").Scan(&stats.Total)
	if err != nil {
		return nil, err
	}

	// Today's alerts
	today := time.Now().Format("2006-01-02")
	err = db.QueryRow("SELECT COUNT(*) FROM alerts WHERE DATE(timestamp) = ?", today).Scan(&stats.Today)
	if err != nil {
		return nil, err
	}

	// By severity
	rows, err := db.Query("SELECT severity, COUNT(*) FROM alerts GROUP BY severity")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var severity string
		var count int
		if err := rows.Scan(&severity, &count); err != nil {
			continue
		}

		switch severity {
		case "CRITICAL":
			stats.Critical = count
		case "HIGH":
			stats.High = count
		case "MEDIUM":
			stats.Medium = count
		case "LOW":
			stats.Low = count
		}
	}

	return stats, nil
}

func getRecentAlerts(limit int) ([]RecentAlert, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	dbPath := filepath.Join(home, ".doggowoof", "alerts.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	query := `
		SELECT id, source, severity, title, 
		       strftime('%m-%d %H:%M', timestamp) as formatted_time
		FROM alerts 
		ORDER BY timestamp DESC 
		LIMIT ?
	`

	rows, err := db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []RecentAlert
	for rows.Next() {
		var alert RecentAlert
		err := rows.Scan(&alert.ID, &alert.Source, &alert.Severity,
			&alert.Title, &alert.Timestamp)
		if err != nil {
			continue
		}
		alerts = append(alerts, alert)
	}

	return alerts, nil
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
