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

// TemporalConfig holds settings for temporal clustering detection.
type TemporalConfig struct {
	WindowMinutes int     // sliding window size (default 5)
	MinAccounts   int     // minimum unique authors to flag (default 3)
	MinScore      float64 // minimum score to create alert (default 0.7)
}

// TemporalCluster represents a group of accounts posting within a suspicious time window.
type TemporalCluster struct {
	Accounts []string
	Topic    string
	Window   time.Duration
	Start    time.Time
	End      time.Time
	Score    float64
	ItemIDs  []string
}

// DetectTemporal finds statistically improbable temporal clusters of posts.
func DetectTemporal(ctx context.Context, reader *osmotic.Reader, cfg TemporalConfig, source string, since time.Time) ([]TemporalCluster, error) {
	if cfg.WindowMinutes == 0 {
		cfg.WindowMinutes = 5
	}
	if cfg.MinAccounts == 0 {
		cfg.MinAccounts = 3
	}
	if cfg.MinScore == 0 {
		cfg.MinScore = 0.7
	}

	end := time.Now().UTC()
	items, err := reader.GetItemsInTimeWindow(ctx, since, end, source)
	if err != nil {
		return nil, fmt.Errorf("fetching items: %w", err)
	}

	if len(items) == 0 {
		return nil, nil
	}

	// Group items by topic (using tags as proxy)
	topicGroups := groupByTopic(items)

	window := time.Duration(cfg.WindowMinutes) * time.Minute
	var clusters []TemporalCluster

	for topic, topicItems := range topicGroups {
		// Sort by published_at
		sort.Slice(topicItems, func(i, j int) bool {
			return topicItems[i].PublishedAt < topicItems[j].PublishedAt
		})

		// Sliding window
		found := findTemporalClusters(topicItems, window, cfg.MinAccounts)
		for _, c := range found {
			c.Topic = topic
			c.Score = temporalScore(c, cfg.MinAccounts, cfg.WindowMinutes)
			if c.Score >= cfg.MinScore {
				clusters = append(clusters, c)
			}
		}
	}

	return clusters, nil
}

// SaveTemporalAlerts creates alerts and clusters from detected temporal clusters.
func SaveTemporalAlerts(ctx context.Context, store *alerts.Store, clusters []TemporalCluster) (int, error) {
	saved := 0
	now := alerts.TimeNow()

	for _, tc := range clusters {
		accountsJSON, _ := json.Marshal(tc.Accounts)
		evidenceJSON, _ := json.Marshal(map[string]any{
			"item_ids":  tc.ItemIDs,
			"window":    tc.Window.String(),
			"start":     tc.Start.Format(time.RFC3339),
			"end":       tc.End.Format(time.RFC3339),
			"num_items": len(tc.ItemIDs),
		})

		severity := scoreSeverity(tc.Score)

		alert := &alerts.Alert{
			AlertType:   "temporal_cluster",
			Severity:    severity,
			Title:       fmt.Sprintf("%d accounts posted about %q within %v", len(tc.Accounts), tc.Topic, tc.Window.Round(time.Second)),
			Description: fmt.Sprintf("Accounts: %v", tc.Accounts),
			Evidence:    string(evidenceJSON),
			Confidence:  tc.Score,
			CreatedAt:   now,
		}
		if err := store.SaveAlert(ctx, alert); err != nil {
			return saved, err
		}

		cluster := &alerts.Cluster{
			ClusterType:   "temporal",
			Accounts:      string(accountsJSON),
			Score:         tc.Score,
			Evidence:      string(evidenceJSON),
			Topic:         tc.Topic,
			FirstDetected: now,
			LastDetected:  now,
			Occurrences:   1,
			Status:        "active",
		}
		if err := store.SaveCluster(ctx, cluster); err != nil {
			return saved, err
		}

		saved++
	}

	return saved, nil
}

func groupByTopic(items []osmotic.Item) map[string][]osmotic.Item {
	groups := make(map[string][]osmotic.Item)
	for _, item := range items {
		topic := extractTopic(item)
		groups[topic] = append(groups[topic], item)
	}
	return groups
}

func extractTopic(item osmotic.Item) string {
	// Use first tag as topic proxy
	if item.Tags != "" {
		var tags []string
		if err := json.Unmarshal([]byte(item.Tags), &tags); err == nil && len(tags) > 0 {
			return tags[0]
		}
	}
	return item.Source
}

func findTemporalClusters(items []osmotic.Item, window time.Duration, minAccounts int) []TemporalCluster {
	var clusters []TemporalCluster

	for i := 0; i < len(items); i++ {
		startTime, err := time.Parse(time.RFC3339, items[i].PublishedAt)
		if err != nil {
			continue
		}

		// Collect items within window
		authors := make(map[string]bool)
		var windowItems []osmotic.Item

		for j := i; j < len(items); j++ {
			t, err := time.Parse(time.RFC3339, items[j].PublishedAt)
			if err != nil {
				continue
			}
			if t.Sub(startTime) > window {
				break
			}
			if items[j].Author != "" {
				authors[items[j].Author] = true
			}
			windowItems = append(windowItems, items[j])
		}

		if len(authors) >= minAccounts {
			endTime := startTime
			var itemIDs []string
			var accountList []string

			for _, wi := range windowItems {
				itemIDs = append(itemIDs, wi.ID)
				t, _ := time.Parse(time.RFC3339, wi.PublishedAt)
				if t.After(endTime) {
					endTime = t
				}
			}
			for a := range authors {
				accountList = append(accountList, a)
			}
			sort.Strings(accountList)

			clusters = append(clusters, TemporalCluster{
				Accounts: accountList,
				Window:   endTime.Sub(startTime),
				Start:    startTime,
				End:      endTime,
				ItemIDs:  itemIDs,
			})

			// Skip past this window to avoid overlapping clusters
			for i < len(items)-1 {
				t, err := time.Parse(time.RFC3339, items[i+1].PublishedAt)
				if err != nil || t.After(endTime) {
					break
				}
				i++
			}
		}
	}

	return clusters
}

func temporalScore(c TemporalCluster, minAccounts, windowMinutes int) float64 {
	// Score: more accounts and tighter windows = higher score
	accountFactor := float64(len(c.Accounts)) / float64(minAccounts)
	if accountFactor > 3.0 {
		accountFactor = 3.0
	}

	windowFactor := 1.0
	if c.Window > 0 {
		windowFactor = float64(windowMinutes) * float64(time.Minute) / float64(c.Window)
		if windowFactor > 3.0 {
			windowFactor = 3.0
		}
	}

	score := (accountFactor * 0.6) + (windowFactor * 0.4)

	// Normalize to [0, 1]
	if score > 1.0 {
		score = 1.0
	}
	return score
}

func scoreSeverity(score float64) string {
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
