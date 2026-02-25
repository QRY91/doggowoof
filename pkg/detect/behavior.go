package detect

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/QRY91/doggowoof/pkg/alerts"
	"github.com/QRY91/doggowoof/pkg/osmotic"
)

// BehaviorConfig holds settings for behavior profiling.
type BehaviorConfig struct {
	SimilarityThreshold float64 // minimum cosine similarity to cluster (default 0.85)
	MinClusterSize      int     // minimum accounts in a cluster (default 3)
}

// BehaviorCluster represents a group of accounts with similar posting behavior.
type BehaviorCluster struct {
	Accounts   []string
	Similarity float64
	Evidence   string // description of what makes them similar
}

// ProfileAuthors builds or updates behavioral profiles for all authors in osmotic.
func ProfileAuthors(ctx context.Context, reader *osmotic.Reader, store *alerts.Store, source string) (int, error) {
	authors, err := reader.GetAuthors(ctx, source)
	if err != nil {
		return 0, fmt.Errorf("fetching authors: %w", err)
	}

	updated := 0
	now := alerts.TimeNow()

	for _, author := range authors {
		stats, err := reader.GetAuthorStats(ctx, author)
		if err != nil {
			continue
		}
		if stats == nil || stats.PostCount == 0 {
			continue
		}

		// Calculate avg posts per day
		var avgPostsPerDay float64
		firstSeen, err1 := time.Parse(time.RFC3339, stats.FirstSeen)
		lastSeen, err2 := time.Parse(time.RFC3339, stats.LastSeen)
		if err1 == nil && err2 == nil {
			days := lastSeen.Sub(firstSeen).Hours() / 24
			if days > 0 {
				avgPostsPerDay = float64(stats.PostCount) / days
			}
		}

		hourJSON, _ := json.Marshal(stats.PostsByHour)
		dayJSON, _ := json.Marshal(stats.PostsByDay)

		primarySource := ""
		if len(stats.Sources) > 0 {
			primarySource = stats.Sources[0]
		}

		profile := &alerts.AccountProfile{
			Author:           author,
			Source:           primarySource,
			PostCount:        stats.PostCount,
			FirstSeen:        stats.FirstSeen,
			LastSeen:         stats.LastSeen,
			AvgPostsPerDay:   avgPostsPerDay,
			SourceDiversity:  len(stats.Sources),
			HourDistribution: string(hourJSON),
			DayDistribution:  string(dayJSON),
			ProfileUpdated:   now,
		}

		if err := store.SaveProfile(ctx, profile); err != nil {
			continue
		}
		updated++
	}

	return updated, nil
}

// DetectBehaviorClusters finds groups of accounts with suspiciously similar posting patterns.
func DetectBehaviorClusters(ctx context.Context, store *alerts.Store, cfg BehaviorConfig) ([]BehaviorCluster, error) {
	if cfg.SimilarityThreshold == 0 {
		cfg.SimilarityThreshold = 0.85
	}
	if cfg.MinClusterSize == 0 {
		cfg.MinClusterSize = 3
	}

	profiles, err := store.ListProfiles(ctx, 0)
	if err != nil {
		return nil, fmt.Errorf("listing profiles: %w", err)
	}

	if len(profiles) < cfg.MinClusterSize {
		return nil, nil
	}

	// Parse hour distributions
	type profileVec struct {
		Author string
		Hours  [24]float64
		Days   [7]float64
		Freq   float64
	}

	var vecs []profileVec
	for _, p := range profiles {
		var pv profileVec
		pv.Author = p.Author
		pv.Freq = p.AvgPostsPerDay

		var hours [24]int
		if p.HourDistribution != "" {
			json.Unmarshal([]byte(p.HourDistribution), &hours)
		}
		var days [7]int
		if p.DayDistribution != "" {
			json.Unmarshal([]byte(p.DayDistribution), &days)
		}

		// Normalize to probability distributions
		totalH := 0
		for _, h := range hours {
			totalH += h
		}
		if totalH > 0 {
			for i, h := range hours {
				pv.Hours[i] = float64(h) / float64(totalH)
			}
		}

		totalD := 0
		for _, d := range days {
			totalD += d
		}
		if totalD > 0 {
			for i, d := range days {
				pv.Days[i] = float64(d) / float64(totalD)
			}
		}

		vecs = append(vecs, pv)
	}

	// Compute pairwise similarity and cluster
	// Simple approach: find connected components above threshold
	n := len(vecs)
	similar := make([][]bool, n)
	for i := range similar {
		similar[i] = make([]bool, n)
	}

	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			hourSim := cosineSimilarity(vecs[i].Hours[:], vecs[j].Hours[:])
			daySim := cosineSimilarity(vecs[i].Days[:], vecs[j].Days[:])
			freqSim := freqSimilarity(vecs[i].Freq, vecs[j].Freq)

			// Weighted combination
			combined := 0.5*hourSim + 0.3*daySim + 0.2*freqSim

			if combined >= cfg.SimilarityThreshold {
				similar[i][j] = true
				similar[j][i] = true
			}
		}
	}

	// Find connected components
	visited := make([]bool, n)
	var clusters []BehaviorCluster

	for i := 0; i < n; i++ {
		if visited[i] {
			continue
		}

		// BFS to find component
		component := []int{i}
		visited[i] = true
		queue := []int{i}

		for len(queue) > 0 {
			curr := queue[0]
			queue = queue[1:]
			for j := 0; j < n; j++ {
				if !visited[j] && similar[curr][j] {
					visited[j] = true
					component = append(component, j)
					queue = append(queue, j)
				}
			}
		}

		if len(component) >= cfg.MinClusterSize {
			var accounts []string
			var totalSim float64
			comparisons := 0

			for _, idx := range component {
				accounts = append(accounts, vecs[idx].Author)
			}
			sort.Strings(accounts)

			// Average pairwise similarity within cluster
			for a := 0; a < len(component); a++ {
				for b := a + 1; b < len(component); b++ {
					i, j := component[a], component[b]
					hourSim := cosineSimilarity(vecs[i].Hours[:], vecs[j].Hours[:])
					daySim := cosineSimilarity(vecs[i].Days[:], vecs[j].Days[:])
					freqSim := freqSimilarity(vecs[i].Freq, vecs[j].Freq)
					totalSim += 0.5*hourSim + 0.3*daySim + 0.2*freqSim
					comparisons++
				}
			}

			avgSim := totalSim / float64(comparisons)

			clusters = append(clusters, BehaviorCluster{
				Accounts:   accounts,
				Similarity: avgSim,
				Evidence:   fmt.Sprintf("%d accounts with %.0f%% behavioral similarity", len(accounts), avgSim*100),
			})
		}
	}

	return clusters, nil
}

// SaveBehaviorAlerts creates alerts from detected behavior clusters.
func SaveBehaviorAlerts(ctx context.Context, store *alerts.Store, clusters []BehaviorCluster) (int, error) {
	saved := 0
	now := alerts.TimeNow()

	for _, bc := range clusters {
		accountsJSON, _ := json.Marshal(bc.Accounts)
		evidenceJSON, _ := json.Marshal(map[string]any{
			"accounts":   bc.Accounts,
			"similarity": bc.Similarity,
			"evidence":   bc.Evidence,
		})

		severity := scoreSeverity(bc.Similarity)

		alert := &alerts.Alert{
			AlertType:   "behavior_cluster",
			Severity:    severity,
			Title:       fmt.Sprintf("%d accounts with similar posting patterns (%.0f%% match)", len(bc.Accounts), bc.Similarity*100),
			Description: bc.Evidence,
			Evidence:    string(evidenceJSON),
			Confidence:  bc.Similarity,
			CreatedAt:   now,
		}
		if err := store.SaveAlert(ctx, alert); err != nil {
			return saved, err
		}

		cluster := &alerts.Cluster{
			ClusterType:   "behavioral",
			Accounts:      string(accountsJSON),
			Score:         bc.Similarity,
			Evidence:      string(evidenceJSON),
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

func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

func freqSimilarity(a, b float64) float64 {
	if a == 0 && b == 0 {
		return 1.0
	}
	max := math.Max(a, b)
	if max == 0 {
		return 0
	}
	return 1.0 - math.Abs(a-b)/max
}
