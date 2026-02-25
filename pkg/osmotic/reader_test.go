package osmotic

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

// createTestDB creates a temp SQLite database with osmotic's schema and test data.
func createTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("opening test db: %v", err)
	}

	// Create osmotic schema
	ddl := []string{
		`CREATE TABLE IF NOT EXISTS items (
			id TEXT PRIMARY KEY, source TEXT NOT NULL, source_id TEXT NOT NULL,
			source_url TEXT, author TEXT, author_meta TEXT, content TEXT,
			content_raw TEXT, content_type TEXT NOT NULL, title TEXT,
			parent_id TEXT, tags TEXT, metadata TEXT, collected_at TEXT NOT NULL,
			published_at TEXT, checksum TEXT NOT NULL, UNIQUE(source, source_id))`,
		`CREATE INDEX IF NOT EXISTS idx_items_source ON items(source)`,
		`CREATE INDEX IF NOT EXISTS idx_items_author ON items(author)`,
		`CREATE TABLE IF NOT EXISTS analyses (
			id INTEGER PRIMARY KEY AUTOINCREMENT, item_id TEXT NOT NULL,
			pass_name TEXT NOT NULL, result TEXT NOT NULL, confidence REAL,
			model_used TEXT, analyzed_at TEXT NOT NULL, UNIQUE(item_id, pass_name))`,
		`CREATE TABLE IF NOT EXISTS entities (
			id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL,
			entity_type TEXT NOT NULL, aliases TEXT, first_seen TEXT NOT NULL,
			last_seen TEXT NOT NULL, mention_count INTEGER DEFAULT 1,
			UNIQUE(name, entity_type))`,
		`CREATE TABLE IF NOT EXISTS patterns (
			id INTEGER PRIMARY KEY AUTOINCREMENT, pattern_type TEXT NOT NULL,
			description TEXT NOT NULL, evidence TEXT NOT NULL, confidence REAL NOT NULL,
			first_seen TEXT NOT NULL, last_seen TEXT NOT NULL,
			occurrences INTEGER DEFAULT 1, status TEXT DEFAULT 'active')`,
		`CREATE TABLE IF NOT EXISTS sources (
			id INTEGER PRIMARY KEY AUTOINCREMENT, source_type TEXT NOT NULL,
			config TEXT NOT NULL, enabled INTEGER DEFAULT 1, last_collected TEXT,
			collect_interval_minutes INTEGER DEFAULT 60)`,
	}

	for _, stmt := range ddl {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("creating schema: %v", err)
		}
	}

	// Insert test items
	now := time.Now().UTC()
	items := []struct {
		id, source, sourceID, author, content, tags, publishedAt string
	}{
		{"a1", "reddit", "r1", "alice", "Post about politics", `["r/politics"]`, now.Add(-10 * time.Minute).Format(time.RFC3339)},
		{"a2", "reddit", "r2", "bob", "Another politics post", `["r/politics"]`, now.Add(-9 * time.Minute).Format(time.RFC3339)},
		{"a3", "reddit", "r3", "charlie", "Politics discussion", `["r/politics"]`, now.Add(-8 * time.Minute).Format(time.RFC3339)},
		{"a4", "hn", "h1", "alice", "Tech news item", `["tech"]`, now.Add(-2 * time.Hour).Format(time.RFC3339)},
		{"a5", "reddit", "r4", "dave", "Unrelated post", `["r/funny"]`, now.Add(-24 * time.Hour).Format(time.RFC3339)},
	}

	for _, item := range items {
		_, err := db.Exec(
			`INSERT INTO items (id, source, source_id, author, content, content_type, tags, collected_at, published_at, checksum)
			 VALUES (?, ?, ?, ?, ?, 'post', ?, ?, ?, 'chk')`,
			item.id, item.source, item.sourceID, item.author, item.content,
			item.tags, item.publishedAt, item.publishedAt)
		if err != nil {
			t.Fatalf("inserting test item: %v", err)
		}
	}

	// Insert test entity
	_, err = db.Exec(
		`INSERT INTO entities (name, entity_type, first_seen, last_seen, mention_count)
		 VALUES ('Test Corp', 'organization', ?, ?, 5)`,
		now.Format(time.RFC3339), now.Format(time.RFC3339))
	if err != nil {
		t.Fatalf("inserting test entity: %v", err)
	}

	return db
}

func TestReaderQueryOnly(t *testing.T) {
	db := createTestDB(t)
	db.Close()

	// We can't easily test query_only with :memory: since it's a different connection.
	// Instead, test that the Reader type works correctly with a temp file.
	tmpFile := t.TempDir() + "/test.sqlite"

	// Create DB with data
	srcDB, err := sql.Open("sqlite", tmpFile)
	if err != nil {
		t.Fatalf("creating source db: %v", err)
	}
	srcDB.Exec(`CREATE TABLE items (id TEXT PRIMARY KEY, source TEXT NOT NULL, source_id TEXT NOT NULL,
		source_url TEXT, author TEXT, author_meta TEXT, content TEXT,
		content_raw TEXT, content_type TEXT NOT NULL, title TEXT,
		parent_id TEXT, tags TEXT, metadata TEXT, collected_at TEXT NOT NULL,
		published_at TEXT, checksum TEXT NOT NULL, UNIQUE(source, source_id))`)
	srcDB.Exec(`CREATE TABLE analyses (id INTEGER PRIMARY KEY AUTOINCREMENT, item_id TEXT NOT NULL,
		pass_name TEXT NOT NULL, result TEXT NOT NULL, confidence REAL,
		model_used TEXT, analyzed_at TEXT NOT NULL, UNIQUE(item_id, pass_name))`)
	srcDB.Exec(`CREATE TABLE entities (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL,
		entity_type TEXT NOT NULL, aliases TEXT, first_seen TEXT NOT NULL,
		last_seen TEXT NOT NULL, mention_count INTEGER DEFAULT 1, UNIQUE(name, entity_type))`)
	srcDB.Exec(`CREATE TABLE patterns (id INTEGER PRIMARY KEY AUTOINCREMENT, pattern_type TEXT NOT NULL,
		description TEXT NOT NULL, evidence TEXT NOT NULL, confidence REAL NOT NULL,
		first_seen TEXT NOT NULL, last_seen TEXT NOT NULL, occurrences INTEGER DEFAULT 1,
		status TEXT DEFAULT 'active')`)

	now := time.Now().UTC().Format(time.RFC3339)
	srcDB.Exec(`INSERT INTO items (id, source, source_id, author, content, content_type, collected_at, published_at, checksum)
		VALUES ('t1', 'reddit', 'r1', 'testuser', 'test content', 'post', ?, ?, 'chk')`, now, now)
	srcDB.Close()

	// Open as reader
	reader, err := NewReader(tmpFile)
	if err != nil {
		t.Fatalf("NewReader() failed: %v", err)
	}
	defer reader.Close()

	ctx := context.Background()

	// Reads should work
	count, err := reader.CountItems(ctx)
	if err != nil {
		t.Fatalf("CountItems() failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 item, got %d", count)
	}

	// Writes should fail
	_, err = reader.db.Exec("INSERT INTO items (id, source, source_id, content_type, collected_at, checksum) VALUES ('x', 'x', 'x', 'x', 'x', 'x')")
	if err == nil {
		t.Error("expected write to fail with query_only mode")
	}
}

func TestGetItems(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	// Wrap the in-memory db in a Reader manually for testing
	reader := &Reader{db: db}

	ctx := context.Background()

	// All items
	items, err := reader.GetItems(ctx, ItemFilter{})
	if err != nil {
		t.Fatalf("GetItems() failed: %v", err)
	}
	if len(items) != 5 {
		t.Errorf("expected 5 items, got %d", len(items))
	}

	// Filter by source
	redditItems, err := reader.GetItems(ctx, ItemFilter{Source: "reddit"})
	if err != nil {
		t.Fatalf("GetItems(reddit) failed: %v", err)
	}
	if len(redditItems) != 4 {
		t.Errorf("expected 4 reddit items, got %d", len(redditItems))
	}

	// Filter by author
	aliceItems, err := reader.GetItems(ctx, ItemFilter{Author: "alice"})
	if err != nil {
		t.Fatalf("GetItems(alice) failed: %v", err)
	}
	if len(aliceItems) != 2 {
		t.Errorf("expected 2 alice items, got %d", len(aliceItems))
	}

	// With limit
	limited, err := reader.GetItems(ctx, ItemFilter{Limit: 2})
	if err != nil {
		t.Fatalf("GetItems(limit=2) failed: %v", err)
	}
	if len(limited) != 2 {
		t.Errorf("expected 2 items with limit, got %d", len(limited))
	}
}

func TestGetAuthors(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	reader := &Reader{db: db}
	ctx := context.Background()

	authors, err := reader.GetAuthors(ctx, "")
	if err != nil {
		t.Fatalf("GetAuthors() failed: %v", err)
	}
	if len(authors) != 4 {
		t.Errorf("expected 4 authors, got %d", len(authors))
	}

	// Filter by source
	hnAuthors, err := reader.GetAuthors(ctx, "hn")
	if err != nil {
		t.Fatalf("GetAuthors(hn) failed: %v", err)
	}
	if len(hnAuthors) != 1 {
		t.Errorf("expected 1 hn author, got %d", len(hnAuthors))
	}
}

func TestGetAuthorStats(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	reader := &Reader{db: db}
	ctx := context.Background()

	stats, err := reader.GetAuthorStats(ctx, "alice")
	if err != nil {
		t.Fatalf("GetAuthorStats() failed: %v", err)
	}
	if stats == nil {
		t.Fatal("expected stats, got nil")
	}
	if stats.PostCount != 2 {
		t.Errorf("expected 2 posts for alice, got %d", stats.PostCount)
	}
	if len(stats.Sources) != 2 {
		t.Errorf("expected 2 sources for alice, got %d", len(stats.Sources))
	}

	// Nonexistent author
	stats, err = reader.GetAuthorStats(ctx, "nobody")
	if err != nil {
		t.Fatalf("GetAuthorStats(nobody) failed: %v", err)
	}
	if stats != nil {
		t.Error("expected nil stats for nonexistent author")
	}
}

func TestCountItems(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	reader := &Reader{db: db}
	ctx := context.Background()

	count, err := reader.CountItems(ctx)
	if err != nil {
		t.Fatalf("CountItems() failed: %v", err)
	}
	if count != 5 {
		t.Errorf("expected 5 items, got %d", count)
	}

	bySource, err := reader.CountItemsBySource(ctx)
	if err != nil {
		t.Fatalf("CountItemsBySource() failed: %v", err)
	}
	if len(bySource) != 2 {
		t.Errorf("expected 2 sources, got %d", len(bySource))
	}

	authorCount, err := reader.CountAuthors(ctx)
	if err != nil {
		t.Fatalf("CountAuthors() failed: %v", err)
	}
	if authorCount != 4 {
		t.Errorf("expected 4 distinct authors, got %d", authorCount)
	}
}

func TestGetEntities(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	reader := &Reader{db: db}
	ctx := context.Background()

	entities, err := reader.GetEntities(ctx, 10)
	if err != nil {
		t.Fatalf("GetEntities() failed: %v", err)
	}
	if len(entities) != 1 {
		t.Errorf("expected 1 entity, got %d", len(entities))
	}
	if entities[0].Name != "Test Corp" {
		t.Errorf("expected 'Test Corp', got %q", entities[0].Name)
	}
}

func TestGetItemsInTimeWindow(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	reader := &Reader{db: db}
	ctx := context.Background()

	now := time.Now().UTC()
	// Should get the 3 recent reddit posts about politics (within last 15 min)
	items, err := reader.GetItemsInTimeWindow(ctx, now.Add(-15*time.Minute), now, "reddit")
	if err != nil {
		t.Fatalf("GetItemsInTimeWindow() failed: %v", err)
	}
	if len(items) != 3 {
		t.Errorf("expected 3 recent reddit items, got %d", len(items))
	}
}
