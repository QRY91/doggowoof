package cli

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/QRY91/doggowoof/pkg/detect"
	"github.com/gen2brain/beeep"
)

// RunSniff runs detection heuristics against osmotic's database.
func RunSniff(args []string) {
	flags := parseFlags(args)

	stores, err := LoadConfigAndStores()
	if err != nil {
		fatalf("Error: %v", err)
	}
	defer stores.Close()

	ctx := context.Background()

	source := flags["source"]
	detectType := flags["detect"] // "temporal", "behavior", "coordination", or "" (all)
	days := 1
	if d, ok := flags["days"]; ok {
		if n, err := strconv.Atoi(d); err == nil && n > 0 {
			days = n
		}
	}

	since := time.Now().UTC().AddDate(0, 0, -days)

	// Summary header
	itemCount, _ := stores.Osmotic.CountItems(ctx)
	authorCount, _ := stores.Osmotic.CountAuthors(ctx)
	fmt.Println("=== doggo sniff ===")
	fmt.Printf("Scanning osmotic database (%d items, %d authors)...\n", itemCount, authorCount)

	cfg := stores.Config

	switch detectType {
	case "temporal":
		runTemporal(ctx, stores, cfg.Thresholds.CoordinationScore, source, since)
	case "behavior":
		runBehavior(ctx, stores, source)
	case "coordination", "":
		runCoordination(ctx, stores, source, since)
	default:
		fatalf("Unknown detection type: %s (use: temporal, behavior, coordination)", detectType)
	}
}

func runTemporal(ctx context.Context, stores *Stores, minScore float64, source string, since time.Time) {
	tcfg := detect.TemporalConfig{
		WindowMinutes: stores.Config.Thresholds.TemporalClusterWindowMinutes,
		MinAccounts:   stores.Config.Thresholds.TemporalClusterMinAccounts,
		MinScore:      minScore,
	}

	fmt.Print("  temporal clustering... ")
	clusters, err := detect.DetectTemporal(ctx, stores.Osmotic, tcfg, source, since)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}

	if len(clusters) == 0 {
		fmt.Println("no clusters found")
		return
	}

	fmt.Printf("%d cluster(s) found\n", len(clusters))
	saved, _ := detect.SaveTemporalAlerts(ctx, stores.Alerts, clusters)

	for _, c := range clusters {
		sev := classifySeverity(c.Score)
		fmt.Printf("    [%s] %d accounts posted about %q within %v\n",
			sev, len(c.Accounts), c.Topic, c.Window.Round(time.Second))
	}

	notifyHighSeverity(clusters, saved)
}

func runBehavior(ctx context.Context, stores *Stores, source string) {
	fmt.Print("  behavior profiling... ")

	updated, err := detect.ProfileAuthors(ctx, stores.Osmotic, stores.Alerts, source)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}

	bcfg := detect.BehaviorConfig{
		SimilarityThreshold: 0.85,
		MinClusterSize:      stores.Config.Thresholds.TemporalClusterMinAccounts,
	}

	clusters, err := detect.DetectBehaviorClusters(ctx, stores.Alerts, bcfg)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}

	fmt.Printf("%d profiles updated", updated)
	if len(clusters) > 0 {
		fmt.Printf(", %d anomalous cluster(s)", len(clusters))
		detect.SaveBehaviorAlerts(ctx, stores.Alerts, clusters)
	}
	fmt.Println()

	for _, c := range clusters {
		fmt.Printf("    %d accounts with %.0f%% behavioral similarity\n",
			len(c.Accounts), c.Similarity*100)
	}
}

func runCoordination(ctx context.Context, stores *Stores, source string, since time.Time) {
	cfg := detect.CoordinationConfig{
		Temporal: detect.TemporalConfig{
			WindowMinutes: stores.Config.Thresholds.TemporalClusterWindowMinutes,
			MinAccounts:   stores.Config.Thresholds.TemporalClusterMinAccounts,
			MinScore:      stores.Config.Thresholds.CoordinationScore,
		},
		Behavior: detect.BehaviorConfig{
			SimilarityThreshold: 0.85,
			MinClusterSize:      stores.Config.Thresholds.TemporalClusterMinAccounts,
		},
		MinScore: stores.Config.Thresholds.CoordinationScore,
	}

	result, err := detect.DetectCoordination(ctx, stores.Osmotic, stores.Alerts, cfg, source, since)
	if err != nil {
		fatalf("Error: %v", err)
	}

	// Print temporal results
	fmt.Printf("  temporal clustering... %d cluster(s) found\n", len(result.TemporalClusters))
	for _, c := range result.TemporalClusters {
		sev := classifySeverity(c.Score)
		fmt.Printf("    [%s] %d accounts posted about %q within %v\n",
			sev, len(c.Accounts), c.Topic, c.Window.Round(time.Second))
	}

	// Print behavior results
	fmt.Printf("  behavior profiling... %d profiles updated", result.ProfilesUpdated)
	if len(result.BehaviorClusters) > 0 {
		fmt.Printf(", %d anomalous cluster(s)", len(result.BehaviorClusters))
	}
	fmt.Println()

	// Print coordination results
	if len(result.CoordinatedGroups) > 0 {
		fmt.Printf("  coordination scoring... %d coordination alert(s)\n", len(result.CoordinatedGroups))
		for _, cg := range result.CoordinatedGroups {
			sev := classifySeverity(cg.CombinedScore)
			fmt.Printf("    [%s] %d accounts: coordinated signals (score: %.2f)\n",
				sev, len(cg.Accounts), cg.CombinedScore)
		}
	} else {
		fmt.Println("  coordination scoring... no coordination detected")
	}

	// Save all alerts
	total, _ := detect.SaveCoordinationAlerts(ctx, stores.Alerts, result)

	// Summary
	fmt.Println()
	if total > 0 {
		fmt.Printf("Alerts saved: %d\n", total)

		// Desktop notification for high/critical
		hasHigh := false
		for _, tc := range result.TemporalClusters {
			if tc.Score >= 0.8 {
				hasHigh = true
				break
			}
		}
		for _, cg := range result.CoordinatedGroups {
			if cg.CombinedScore >= 0.8 {
				hasHigh = true
				break
			}
		}
		if hasHigh {
			_ = beeep.Notify("DoggoWoof Alert", fmt.Sprintf("%d detection alert(s) — check 'doggo alerts'", total), "")
		}
	} else {
		fmt.Println("No suspicious patterns detected. Good dog nap time.")
	}
}

func classifySeverity(score float64) string {
	switch {
	case score >= 0.9:
		return "critical"
	case score >= 0.8:
		return "high"
	case score >= 0.7:
		return "medium"
	default:
		return "low"
	}
}

func notifyHighSeverity(clusters []detect.TemporalCluster, saved int) {
	for _, c := range clusters {
		if c.Score >= 0.8 {
			_ = beeep.Notify("DoggoWoof Alert",
				fmt.Sprintf("%d temporal cluster(s) detected — check 'doggo alerts'", saved), "")
			return
		}
	}
}
