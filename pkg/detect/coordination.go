package detect

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/QRY91/doggowoof/pkg/alerts"
	"github.com/QRY91/doggowoof/pkg/osmotic"
)

// CoordinationConfig holds settings for combined coordination detection.
type CoordinationConfig struct {
	Temporal  TemporalConfig
	Behavior  BehaviorConfig
	MinScore  float64 // minimum combined score to alert (default 0.7)
}

// CoordinationResult holds the combined detection output.
type CoordinationResult struct {
	TemporalClusters  []TemporalCluster
	BehaviorClusters  []BehaviorCluster
	CoordinatedGroups []CoordinatedGroup
	ProfilesUpdated   int
}

// CoordinatedGroup represents accounts flagged by both temporal and behavioral analysis.
type CoordinatedGroup struct {
	Accounts        []string
	TemporalScore   float64
	BehavioralScore float64
	CombinedScore   float64
	Topic           string
}

// DetectCoordination runs the full coordination detection pipeline:
// temporal clustering + behavior profiling + overlap analysis.
func DetectCoordination(ctx context.Context, reader *osmotic.Reader, store *alerts.Store, cfg CoordinationConfig, source string, since time.Time) (*CoordinationResult, error) {
	if cfg.MinScore == 0 {
		cfg.MinScore = 0.7
	}

	result := &CoordinationResult{}

	// Step 1: Temporal clustering
	temporalClusters, err := DetectTemporal(ctx, reader, cfg.Temporal, source, since)
	if err != nil {
		return nil, fmt.Errorf("temporal detection: %w", err)
	}
	result.TemporalClusters = temporalClusters

	// Step 2: Profile all authors
	updated, err := ProfileAuthors(ctx, reader, store, source)
	if err != nil {
		return nil, fmt.Errorf("behavior profiling: %w", err)
	}
	result.ProfilesUpdated = updated

	// Step 3: Behavior clustering
	behaviorClusters, err := DetectBehaviorClusters(ctx, store, cfg.Behavior)
	if err != nil {
		return nil, fmt.Errorf("behavior detection: %w", err)
	}
	result.BehaviorClusters = behaviorClusters

	// Step 4: Find overlapping accounts between temporal and behavioral clusters
	result.CoordinatedGroups = findOverlaps(temporalClusters, behaviorClusters, cfg.MinScore)

	return result, nil
}

// SaveCoordinationAlerts saves all alerts from the coordination pipeline.
func SaveCoordinationAlerts(ctx context.Context, store *alerts.Store, result *CoordinationResult) (int, error) {
	total := 0

	// Save temporal alerts
	n, err := SaveTemporalAlerts(ctx, store, result.TemporalClusters)
	if err != nil {
		return total, fmt.Errorf("saving temporal alerts: %w", err)
	}
	total += n

	// Save behavior alerts
	n, err = SaveBehaviorAlerts(ctx, store, result.BehaviorClusters)
	if err != nil {
		return total, fmt.Errorf("saving behavior alerts: %w", err)
	}
	total += n

	// Save coordination alerts (overlap findings)
	now := alerts.TimeNow()
	for _, cg := range result.CoordinatedGroups {
		accountsJSON, _ := json.Marshal(cg.Accounts)
		evidenceJSON, _ := json.Marshal(map[string]any{
			"accounts":         cg.Accounts,
			"temporal_score":   cg.TemporalScore,
			"behavioral_score": cg.BehavioralScore,
			"combined_score":   cg.CombinedScore,
			"topic":            cg.Topic,
		})

		severity := scoreSeverity(cg.CombinedScore)

		alert := &alerts.Alert{
			AlertType:   "coordination",
			Severity:    severity,
			Title:       fmt.Sprintf("%d accounts: coordinated signals (score: %.2f)", len(cg.Accounts), cg.CombinedScore),
			Description: fmt.Sprintf("Temporal + behavioral overlap on %q", cg.Topic),
			Evidence:    string(evidenceJSON),
			Confidence:  cg.CombinedScore,
			CreatedAt:   now,
		}
		if err := store.SaveAlert(ctx, alert); err != nil {
			return total, err
		}

		cluster := &alerts.Cluster{
			ClusterType:   "coordination",
			Accounts:      string(accountsJSON),
			Score:         cg.CombinedScore,
			Evidence:      string(evidenceJSON),
			Topic:         cg.Topic,
			FirstDetected: now,
			LastDetected:  now,
			Occurrences:   1,
			Status:        "active",
		}
		if err := store.SaveCluster(ctx, cluster); err != nil {
			return total, err
		}

		total++
	}

	return total, nil
}

func findOverlaps(temporal []TemporalCluster, behavioral []BehaviorCluster, minScore float64) []CoordinatedGroup {
	var groups []CoordinatedGroup

	for _, tc := range temporal {
		temporalSet := toSet(tc.Accounts)

		for _, bc := range behavioral {
			behaviorSet := toSet(bc.Accounts)

			// Find intersection
			var overlap []string
			for a := range temporalSet {
				if behaviorSet[a] {
					overlap = append(overlap, a)
				}
			}

			if len(overlap) < 2 {
				continue
			}

			sort.Strings(overlap)

			// Combined score: weighted average
			combined := 0.6*tc.Score + 0.4*bc.Similarity

			if combined >= minScore {
				groups = append(groups, CoordinatedGroup{
					Accounts:        overlap,
					TemporalScore:   tc.Score,
					BehavioralScore: bc.Similarity,
					CombinedScore:   combined,
					Topic:           tc.Topic,
				})
			}
		}
	}

	return groups
}

func toSet(items []string) map[string]bool {
	s := make(map[string]bool, len(items))
	for _, item := range items {
		s[item] = true
	}
	return s
}
