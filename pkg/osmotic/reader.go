package osmotic

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// Reader provides read-only access to osmotic's SQLite database.
// It uses its own types to avoid import dependency on osmotic.
type Reader struct {
	db *sql.DB
}

// NewReader opens osmotic's database in read-only mode.
func NewReader(dbPath string) (*Reader, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening osmotic database: %w", err)
	}

	db.SetMaxOpenConns(1)

	if _, err := db.Exec("PRAGMA query_only=ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("setting query_only mode: %w", err)
	}
	if _, err := db.Exec("PRAGMA busy_timeout=5000"); err != nil {
		db.Close()
		return nil, fmt.Errorf("setting busy timeout: %w", err)
	}

	return &Reader{db: db}, nil
}

// Close closes the database connection.
func (r *Reader) Close() error {
	return r.db.Close()
}

// --- Types (mirrors osmotic schema without import dependency) ---

// Item represents a normalized content item from osmotic.
type Item struct {
	ID          string
	Source      string
	SourceID    string
	SourceURL   string
	Author      string
	AuthorMeta  string // JSON
	Content     string
	ContentType string
	Title       string
	Tags        string // JSON array
	PublishedAt string
	CollectedAt string
}

// Entity represents an extracted named entity.
type Entity struct {
	ID           int
	Name         string
	EntityType   string
	MentionCount int
}

// Pattern represents a detected pattern from osmotic's analysis.
type Pattern struct {
	ID          int
	PatternType string
	Description string
	Confidence  float64
	Status      string
	FirstSeen   string
	LastSeen    string
	Occurrences int
}

// AuthorStats holds computed statistics for a single author.
type AuthorStats struct {
	Author     string
	PostCount  int
	FirstSeen  string
	LastSeen   string
	Sources    []string
	PostsByHour [24]int
	PostsByDay  [7]int
}

// SourceCount holds a source name and its item count.
type SourceCount struct {
	Source string
	Count  int
}

// ItemFilter controls what GetItems returns.
type ItemFilter struct {
	Source string
	Author string
	Since  time.Time
	Limit  int
}

// --- Query methods ---

// GetItems returns items matching the filter.
func (r *Reader) GetItems(ctx context.Context, f ItemFilter) ([]Item, error) {
	query := `SELECT id, source, source_id, source_url, author, author_meta,
		content, content_type, title, tags, published_at, collected_at
		FROM items WHERE 1=1`
	var args []any

	if f.Source != "" {
		query += " AND source = ?"
		args = append(args, f.Source)
	}
	if f.Author != "" {
		query += " AND author = ?"
		args = append(args, f.Author)
	}
	if !f.Since.IsZero() {
		query += " AND collected_at >= ?"
		args = append(args, f.Since.UTC().Format(time.RFC3339))
	}

	query += " ORDER BY collected_at DESC"

	if f.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, f.Limit)
	}

	return r.scanItems(ctx, query, args...)
}

// GetItemsByAuthor returns items by a specific author.
func (r *Reader) GetItemsByAuthor(ctx context.Context, author string, limit int) ([]Item, error) {
	return r.GetItems(ctx, ItemFilter{Author: author, Limit: limit})
}

// GetAuthors returns distinct authors, optionally filtered by source.
func (r *Reader) GetAuthors(ctx context.Context, source string) ([]string, error) {
	query := "SELECT DISTINCT author FROM items WHERE author IS NOT NULL AND author != ''"
	var args []any

	if source != "" {
		query += " AND source = ?"
		args = append(args, source)
	}
	query += " ORDER BY author"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var authors []string
	for rows.Next() {
		var a string
		if err := rows.Scan(&a); err != nil {
			return nil, err
		}
		authors = append(authors, a)
	}
	return authors, rows.Err()
}

// GetAuthorStats computes statistics for a single author.
func (r *Reader) GetAuthorStats(ctx context.Context, author string) (*AuthorStats, error) {
	stats := &AuthorStats{Author: author}

	// Post count and date range
	var firstSeen, lastSeen sql.NullString
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*), MIN(published_at), MAX(published_at)
		 FROM items WHERE author = ?`, author).Scan(
		&stats.PostCount, &firstSeen, &lastSeen,
	)
	if err != nil {
		return nil, fmt.Errorf("author stats for %q: %w", author, err)
	}
	if stats.PostCount == 0 {
		return nil, nil
	}
	stats.FirstSeen = firstSeen.String
	stats.LastSeen = lastSeen.String

	// Sources
	rows, err := r.db.QueryContext(ctx,
		"SELECT DISTINCT source FROM items WHERE author = ?", author)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var src string
		if err := rows.Scan(&src); err != nil {
			return nil, err
		}
		stats.Sources = append(stats.Sources, src)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Hour and day distributions from published_at
	timeRows, err := r.db.QueryContext(ctx,
		`SELECT published_at FROM items
		 WHERE author = ? AND published_at IS NOT NULL AND published_at != ''`,
		author)
	if err != nil {
		return nil, err
	}
	defer timeRows.Close()
	for timeRows.Next() {
		var ts string
		if err := timeRows.Scan(&ts); err != nil {
			return nil, err
		}
		t, err := time.Parse(time.RFC3339, ts)
		if err != nil {
			continue
		}
		stats.PostsByHour[t.Hour()]++
		stats.PostsByDay[t.Weekday()]++
	}

	return stats, timeRows.Err()
}

// GetItemsInTimeWindow returns items within a time range, optionally filtered by source.
func (r *Reader) GetItemsInTimeWindow(ctx context.Context, start, end time.Time, source string) ([]Item, error) {
	query := `SELECT id, source, source_id, source_url, author, author_meta,
		content, content_type, title, tags, published_at, collected_at
		FROM items WHERE published_at >= ? AND published_at <= ?`
	args := []any{
		start.UTC().Format(time.RFC3339),
		end.UTC().Format(time.RFC3339),
	}

	if source != "" {
		query += " AND source = ?"
		args = append(args, source)
	}

	query += " ORDER BY published_at ASC"

	return r.scanItems(ctx, query, args...)
}

// GetEntities returns top entities by mention count.
func (r *Reader) GetEntities(ctx context.Context, limit int) ([]Entity, error) {
	query := `SELECT id, name, entity_type, mention_count
		FROM entities ORDER BY mention_count DESC`
	var args []any

	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entities []Entity
	for rows.Next() {
		var e Entity
		if err := rows.Scan(&e.ID, &e.Name, &e.EntityType, &e.MentionCount); err != nil {
			return nil, err
		}
		entities = append(entities, e)
	}
	return entities, rows.Err()
}

// GetPatterns returns patterns filtered by status.
func (r *Reader) GetPatterns(ctx context.Context, status string) ([]Pattern, error) {
	query := `SELECT id, pattern_type, description, confidence, status, first_seen, last_seen, occurrences
		FROM patterns WHERE 1=1`
	var args []any

	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}

	query += " ORDER BY confidence DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var patterns []Pattern
	for rows.Next() {
		var p Pattern
		if err := rows.Scan(&p.ID, &p.PatternType, &p.Description, &p.Confidence,
			&p.Status, &p.FirstSeen, &p.LastSeen, &p.Occurrences); err != nil {
			return nil, err
		}
		patterns = append(patterns, p)
	}
	return patterns, rows.Err()
}

// CountItems returns total item count.
func (r *Reader) CountItems(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM items").Scan(&count)
	return count, err
}

// CountItemsBySource returns item counts grouped by source.
func (r *Reader) CountItemsBySource(ctx context.Context) ([]SourceCount, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT source, COUNT(*) FROM items GROUP BY source ORDER BY COUNT(*) DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var counts []SourceCount
	for rows.Next() {
		var sc SourceCount
		if err := rows.Scan(&sc.Source, &sc.Count); err != nil {
			return nil, err
		}
		counts = append(counts, sc)
	}
	return counts, rows.Err()
}

// CountAuthors returns the number of distinct authors.
func (r *Reader) CountAuthors(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(DISTINCT author) FROM items WHERE author IS NOT NULL AND author != ''").Scan(&count)
	return count, err
}

// GetAnalysisTopics returns items that have topic analysis, with the topic result.
// Used by temporal clustering to group items by topic.
func (r *Reader) GetAnalysisTopics(ctx context.Context, itemIDs []string) (map[string]string, error) {
	if len(itemIDs) == 0 {
		return nil, nil
	}

	topics := make(map[string]string)
	for _, id := range itemIDs {
		var result string
		err := r.db.QueryRowContext(ctx,
			"SELECT result FROM analyses WHERE item_id = ? AND pass_name = 'topics'",
			id).Scan(&result)
		if err == sql.ErrNoRows {
			continue
		}
		if err != nil {
			return nil, err
		}
		topics[id] = result
	}
	return topics, nil
}

func (r *Reader) scanItems(ctx context.Context, query string, args ...any) ([]Item, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		var item Item
		var sourceURL, authorMeta, content, title, tags, publishedAt sql.NullString
		if err := rows.Scan(&item.ID, &item.Source, &item.SourceID,
			&sourceURL, &item.Author, &authorMeta,
			&content, &item.ContentType, &title, &tags,
			&publishedAt, &item.CollectedAt); err != nil {
			return nil, err
		}
		item.SourceURL = sourceURL.String
		item.AuthorMeta = authorMeta.String
		item.Content = content.String
		item.Title = title.String
		item.Tags = tags.String
		item.PublishedAt = publishedAt.String
		items = append(items, item)
	}
	return items, rows.Err()
}
