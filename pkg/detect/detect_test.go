package detect

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/QRY91/doggowoof/pkg/alerts"
	"github.com/QRY91/doggowoof/pkg/osmotic"

	_ "modernc.org/sqlite"
)

// createOsmoticTestDB creates a temp SQLite database with osmotic's schema and test data.
func createOsmoticTestDB(t *testing.T) string {
	t.Helper()

	tmpFile := t.TempDir() + "/osmotic_test.sqlite"
	db, err := sql.Open("sqlite", tmpFile)
	if err != nil {
		t.Fatalf("opening test db: %v", err)
	}
	defer db.Close()

	ddl := []string{
		`CREATE TABLE items (
			id TEXT PRIMARY KEY, source TEXT NOT NULL, source_id TEXT NOT NULL,
			source_url TEXT, author TEXT, author_meta TEXT, content TEXT,
			content_raw TEXT, content_type TEXT NOT NULL, title TEXT,
			parent_id TEXT, tags TEXT, metadata TEXT, collected_at TEXT NOT NULL,
			published_at TEXT, checksum TEXT NOT NULL, UNIQUE(source, source_id))`,
		`CREATE INDEX idx_items_source ON items(source)`,
		`CREATE INDEX idx_items_author ON items(author)`,
		`CREATE TABLE analyses (
			id INTEGER PRIMARY KEY AUTOINCREMENT, item_id TEXT NOT NULL,
			pass_name TEXT NOT NULL, result TEXT NOT NULL, confidence REAL,
			model_used TEXT, analyzed_at TEXT NOT NULL, UNIQUE(item_id, pass_name))`,
		`CREATE TABLE entities (
			id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL,
			entity_type TEXT NOT NULL, aliases TEXT, first_seen TEXT NOT NULL,
			last_seen TEXT NOT NULL, mention_count INTEGER DEFAULT 1,
			UNIQUE(name, entity_type))`,
		`CREATE TABLE patterns (
			id INTEGER PRIMARY KEY AUTOINCREMENT, pattern_type TEXT NOT NULL,
			description TEXT NOT NULL, evidence TEXT NOT NULL, confidence REAL NOT NULL,
			first_seen TEXT NOT NULL, last_seen TEXT NOT NULL,
			occurrences INTEGER DEFAULT 1, status TEXT DEFAULT 'active')`,
		`CREATE TABLE sources (
			id INTEGER PRIMARY KEY AUTOINCREMENT, source_type TEXT NOT NULL,
			config TEXT NOT NULL, enabled INTEGER DEFAULT 1, last_collected TEXT,
			collect_interval_minutes INTEGER DEFAULT 60)`,
	}

	for _, stmt := range ddl {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("creating schema: %v", err)
		}
	}

	return tmpFile
}

func insertItems(t *testing.T, dbPath string, items []struct {
	id, source, sourceID, author, content, tags, publishedAt string
}) {
	t.Helper()

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("opening db for insert: %v", err)
	}
	defer db.Close()

	for _, item := range items {
		_, err := db.Exec(
			`INSERT INTO items (id, source, source_id, author, content, content_type, tags, collected_at, published_at, checksum)
			 VALUES (?, ?, ?, ?, ?, 'post', ?, ?, ?, 'chk')`,
			item.id, item.source, item.sourceID, item.author, item.content,
			item.tags, item.publishedAt, item.publishedAt)
		if err != nil {
			t.Fatalf("inserting test item %s: %v", item.id, err)
		}
	}
}

func TestTemporalDetection_Clustered(t *testing.T) {
	dbPath := createOsmoticTestDB(t)

	// 5 accounts posting within 2 minutes on the same topic — should be detected
	now := time.Now().UTC()
	items := []struct {
		id, source, sourceID, author, content, tags, publishedAt string
	}{
		{"t1", "reddit", "r1", "account_a", "Topic X post", `["r/news"]`, now.Add(-5 * time.Minute).Format(time.RFC3339)},
		{"t2", "reddit", "r2", "account_b", "Topic X post", `["r/news"]`, now.Add(-4 * time.Minute).Format(time.RFC3339)},
		{"t3", "reddit", "r3", "account_c", "Topic X post", `["r/news"]`, now.Add(-4 * time.Minute).Add(30 * time.Second).Format(time.RFC3339)},
		{"t4", "reddit", "r4", "account_d", "Topic X post", `["r/news"]`, now.Add(-3 * time.Minute).Format(time.RFC3339)},
		{"t5", "reddit", "r5", "account_e", "Topic X post", `["r/news"]`, now.Add(-3 * time.Minute).Add(30 * time.Second).Format(time.RFC3339)},
	}
	insertItems(t, dbPath, items)

	reader, err := osmotic.NewReader(dbPath)
	if err != nil {
		t.Fatalf("NewReader() failed: %v", err)
	}
	defer reader.Close()

	ctx := context.Background()
	cfg := TemporalConfig{
		WindowMinutes: 5,
		MinAccounts:   3,
		MinScore:      0.5,
	}

	clusters, err := DetectTemporal(ctx, reader, cfg, "", now.Add(-10*time.Minute))
	if err != nil {
		t.Fatalf("DetectTemporal() failed: %v", err)
	}

	if len(clusters) == 0 {
		t.Fatal("expected at least 1 temporal cluster, got 0")
	}

	// Should have found all 5 accounts
	if len(clusters[0].Accounts) < 3 {
		t.Errorf("expected at least 3 accounts in cluster, got %d", len(clusters[0].Accounts))
	}
}

func TestTemporalDetection_Spread(t *testing.T) {
	dbPath := createOsmoticTestDB(t)

	// 5 accounts spread over 2 hours — should NOT be detected
	now := time.Now().UTC()
	items := []struct {
		id, source, sourceID, author, content, tags, publishedAt string
	}{
		{"s1", "reddit", "s1", "user_a", "Spread post", `["r/news"]`, now.Add(-120 * time.Minute).Format(time.RFC3339)},
		{"s2", "reddit", "s2", "user_b", "Spread post", `["r/news"]`, now.Add(-90 * time.Minute).Format(time.RFC3339)},
		{"s3", "reddit", "s3", "user_c", "Spread post", `["r/news"]`, now.Add(-60 * time.Minute).Format(time.RFC3339)},
		{"s4", "reddit", "s4", "user_d", "Spread post", `["r/news"]`, now.Add(-30 * time.Minute).Format(time.RFC3339)},
		{"s5", "reddit", "s5", "user_e", "Spread post", `["r/news"]`, now.Format(time.RFC3339)},
	}
	insertItems(t, dbPath, items)

	reader, err := osmotic.NewReader(dbPath)
	if err != nil {
		t.Fatalf("NewReader() failed: %v", err)
	}
	defer reader.Close()

	ctx := context.Background()
	cfg := TemporalConfig{
		WindowMinutes: 5,
		MinAccounts:   3,
		MinScore:      0.5,
	}

	clusters, err := DetectTemporal(ctx, reader, cfg, "", now.Add(-3*time.Hour))
	if err != nil {
		t.Fatalf("DetectTemporal() failed: %v", err)
	}

	if len(clusters) != 0 {
		t.Errorf("expected 0 clusters for spread posts, got %d", len(clusters))
	}
}

func TestBehaviorProfiling(t *testing.T) {
	dbPath := createOsmoticTestDB(t)

	// Create accounts with identical hour distributions
	now := time.Now().UTC()
	var items []struct {
		id, source, sourceID, author, content, tags, publishedAt string
	}

	// 5 accounts all posting at the same hours
	for i, author := range []string{"bot_a", "bot_b", "bot_c", "bot_d", "bot_e"} {
		for j := 0; j < 10; j++ {
			// All post at 14:00 UTC
			postTime := now.AddDate(0, 0, -j).Truncate(24*time.Hour).Add(14 * time.Hour)
			id := author + "_" + time.Now().Format("150405") + "_" + string(rune('0'+j))
			items = append(items, struct {
				id, source, sourceID, author, content, tags, publishedAt string
			}{
				id:          id,
				source:      "reddit",
				sourceID:    id,
				author:      author,
				content:     "Post content",
				tags:        `["r/test"]`,
				publishedAt: postTime.Format(time.RFC3339),
			})
			_ = i // use i
		}
	}
	insertItems(t, dbPath, items)

	reader, err := osmotic.NewReader(dbPath)
	if err != nil {
		t.Fatalf("NewReader() failed: %v", err)
	}
	defer reader.Close()

	store, err := alerts.NewMemory()
	if err != nil {
		t.Fatalf("NewMemory() failed: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Profile authors
	updated, err := ProfileAuthors(ctx, reader, store, "")
	if err != nil {
		t.Fatalf("ProfileAuthors() failed: %v", err)
	}
	if updated != 5 {
		t.Errorf("expected 5 profiles updated, got %d", updated)
	}

	// Detect behavior clusters
	bcfg := BehaviorConfig{
		SimilarityThreshold: 0.8,
		MinClusterSize:      3,
	}

	clusters, err := DetectBehaviorClusters(ctx, store, bcfg)
	if err != nil {
		t.Fatalf("DetectBehaviorClusters() failed: %v", err)
	}

	if len(clusters) == 0 {
		t.Error("expected at least 1 behavior cluster for identical posting patterns")
	}
}

func TestCoordinationDetection(t *testing.T) {
	dbPath := createOsmoticTestDB(t)

	// Create accounts that cluster both temporally and behaviorally
	now := time.Now().UTC()
	var items []struct {
		id, source, sourceID, author, content, tags, publishedAt string
	}

	// 4 accounts posting together on same topic within 2 min window
	// Also give them similar historical patterns
	for i, author := range []string{"coord_a", "coord_b", "coord_c", "coord_d"} {
		// Recent burst (temporal signal)
		burstTime := now.Add(-time.Duration(i) * 30 * time.Second)
		items = append(items, struct {
			id, source, sourceID, author, content, tags, publishedAt string
		}{
			id:          "burst_" + author,
			source:      "reddit",
			sourceID:    "burst_" + author,
			author:      author,
			content:     "Breaking news about topic",
			tags:        `["r/worldnews"]`,
			publishedAt: burstTime.Format(time.RFC3339),
		})

		// Historical posts at same hours (behavioral signal)
		for j := 1; j <= 5; j++ {
			postTime := now.AddDate(0, 0, -j).Truncate(24*time.Hour).Add(10 * time.Hour)
			items = append(items, struct {
				id, source, sourceID, author, content, tags, publishedAt string
			}{
				id:          author + "_hist_" + string(rune('0'+j)),
				source:      "reddit",
				sourceID:    author + "_hist_" + string(rune('0'+j)),
				author:      author,
				content:     "Historical post",
				tags:        `["r/worldnews"]`,
				publishedAt: postTime.Format(time.RFC3339),
			})
		}
	}
	insertItems(t, dbPath, items)

	reader, err := osmotic.NewReader(dbPath)
	if err != nil {
		t.Fatalf("NewReader() failed: %v", err)
	}
	defer reader.Close()

	store, err := alerts.NewMemory()
	if err != nil {
		t.Fatalf("NewMemory() failed: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	cfg := CoordinationConfig{
		Temporal: TemporalConfig{
			WindowMinutes: 5,
			MinAccounts:   3,
			MinScore:      0.5,
		},
		Behavior: BehaviorConfig{
			SimilarityThreshold: 0.8,
			MinClusterSize:      3,
		},
		MinScore: 0.5,
	}

	result, err := DetectCoordination(ctx, reader, store, cfg, "", now.Add(-24*time.Hour))
	if err != nil {
		t.Fatalf("DetectCoordination() failed: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	// Should have found temporal clusters
	if len(result.TemporalClusters) == 0 {
		t.Error("expected temporal clusters from coordinated accounts")
	}

	// Should have profiled all accounts
	if result.ProfilesUpdated < 4 {
		t.Errorf("expected at least 4 profiles updated, got %d", result.ProfilesUpdated)
	}

	// Save alerts and verify
	total, err := SaveCoordinationAlerts(ctx, store, result)
	if err != nil {
		t.Fatalf("SaveCoordinationAlerts() failed: %v", err)
	}
	if total == 0 {
		t.Error("expected at least 1 alert saved")
	}

	// Verify alerts were saved
	alertCount, _ := store.CountAlerts(ctx)
	if alertCount == 0 {
		t.Error("expected alerts in database after save")
	}
}

func TestCosineSimilarity(t *testing.T) {
	// Identical vectors
	a := []float64{1, 0, 0, 1}
	sim := cosineSimilarity(a, a)
	if sim < 0.99 {
		t.Errorf("expected ~1.0 for identical vectors, got %f", sim)
	}

	// Orthogonal vectors
	b := []float64{0, 1, 1, 0}
	sim = cosineSimilarity(a, b)
	if sim > 0.01 {
		t.Errorf("expected ~0.0 for orthogonal vectors, got %f", sim)
	}

	// Zero vector
	z := []float64{0, 0, 0, 0}
	sim = cosineSimilarity(a, z)
	if sim != 0 {
		t.Errorf("expected 0 for zero vector, got %f", sim)
	}
}

func TestScoreSeverity(t *testing.T) {
	tests := []struct {
		score float64
		want  string
	}{
		{0.95, "critical"},
		{0.9, "critical"},
		{0.85, "high"},
		{0.8, "high"},
		{0.75, "medium"},
		{0.7, "medium"},
		{0.5, "low"},
	}
	for _, tt := range tests {
		got := scoreSeverity(tt.score)
		if got != tt.want {
			t.Errorf("scoreSeverity(%f) = %q, want %q", tt.score, got, tt.want)
		}
	}
}
