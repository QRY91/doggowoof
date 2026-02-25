package alerts

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// Store wraps doggowoof's own SQLite database for alerts, profiles, and clusters.
type Store struct {
	db *sql.DB
}

// New opens (or creates) the alerts database, enables WAL mode,
// sets a busy timeout, and runs migrations.
func New(dbPath string) (*Store, error) {
	if dbPath != ":memory:" {
		dir := filepath.Dir(dbPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("creating database directory: %w", err)
		}
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	db.SetMaxOpenConns(1)

	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("setting WAL mode: %w", err)
	}
	if _, err := db.Exec("PRAGMA busy_timeout=5000"); err != nil {
		db.Close()
		return nil, fmt.Errorf("setting busy timeout: %w", err)
	}
	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("enabling foreign keys: %w", err)
	}

	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("running migrations: %w", err)
	}

	return s, nil
}

// NewMemory creates an in-memory store for testing.
func NewMemory() (*Store, error) {
	return New(":memory:")
}

// Close closes the underlying database connection.
func (s *Store) Close() error {
	return s.db.Close()
}

var migrations = []string{
	`CREATE TABLE IF NOT EXISTS alerts (
		id            INTEGER PRIMARY KEY AUTOINCREMENT,
		alert_type    TEXT NOT NULL,
		severity      TEXT NOT NULL,
		title         TEXT NOT NULL,
		description   TEXT,
		evidence      TEXT,
		confidence    REAL,
		source_filter TEXT,
		created_at    TEXT NOT NULL,
		acknowledged  INTEGER DEFAULT 0
	)`,

	`CREATE INDEX IF NOT EXISTS idx_alerts_type ON alerts(alert_type)`,
	`CREATE INDEX IF NOT EXISTS idx_alerts_severity ON alerts(severity)`,
	`CREATE INDEX IF NOT EXISTS idx_alerts_created ON alerts(created_at)`,

	`CREATE TABLE IF NOT EXISTS account_profiles (
		id                INTEGER PRIMARY KEY AUTOINCREMENT,
		author            TEXT NOT NULL UNIQUE,
		source            TEXT NOT NULL,
		post_count        INTEGER DEFAULT 0,
		first_seen        TEXT NOT NULL,
		last_seen         TEXT NOT NULL,
		avg_posts_per_day REAL DEFAULT 0,
		source_diversity  INTEGER DEFAULT 1,
		hour_distribution TEXT,
		day_distribution  TEXT,
		metadata          TEXT,
		profile_updated   TEXT NOT NULL
	)`,

	`CREATE INDEX IF NOT EXISTS idx_profiles_author ON account_profiles(author)`,
	`CREATE INDEX IF NOT EXISTS idx_profiles_source ON account_profiles(source)`,

	`CREATE TABLE IF NOT EXISTS clusters (
		id             INTEGER PRIMARY KEY AUTOINCREMENT,
		cluster_type   TEXT NOT NULL,
		accounts       TEXT NOT NULL,
		score          REAL NOT NULL,
		evidence       TEXT,
		topic          TEXT,
		first_detected TEXT NOT NULL,
		last_detected  TEXT NOT NULL,
		occurrences    INTEGER DEFAULT 1,
		status         TEXT DEFAULT 'active'
	)`,

	`CREATE INDEX IF NOT EXISTS idx_clusters_type ON clusters(cluster_type)`,
	`CREATE INDEX IF NOT EXISTS idx_clusters_status ON clusters(status)`,

	`CREATE TABLE IF NOT EXISTS deepbore_sessions (
		id             INTEGER PRIMARY KEY AUTOINCREMENT,
		session_name   TEXT NOT NULL UNIQUE,
		log_file       TEXT NOT NULL,
		cursor_offset  INTEGER DEFAULT 0,
		agent_name     TEXT,
		task_current   INTEGER DEFAULT 0,
		task_total     INTEGER DEFAULT 0,
		task_name      TEXT,
		turns          INTEGER DEFAULT 0,
		cost_dollars   REAL DEFAULT 0,
		uptime_seconds INTEGER DEFAULT 0,
		binary_hash    TEXT,
		sandbox_ok     INTEGER DEFAULT 1,
		started_at     TEXT NOT NULL,
		updated_at     TEXT NOT NULL
	)`,

	`CREATE INDEX IF NOT EXISTS idx_deepbore_session_name ON deepbore_sessions(session_name)`,
}

func (s *Store) migrate() error {
	for _, stmt := range migrations {
		if _, err := s.db.Exec(stmt); err != nil {
			return fmt.Errorf("migration failed: %w\n  SQL: %s", err, stmt)
		}
	}
	return nil
}

// --- Alert types ---

// Alert represents a detection finding.
type Alert struct {
	ID           int
	AlertType    string
	Severity     string // "critical", "high", "medium", "low"
	Title        string
	Description  string
	Evidence     string // JSON
	Confidence   float64
	SourceFilter string
	CreatedAt    string
	Acknowledged bool
}

// AccountProfile holds behavioral data for an author.
type AccountProfile struct {
	ID               int
	Author           string
	Source           string
	PostCount        int
	FirstSeen        string
	LastSeen         string
	AvgPostsPerDay   float64
	SourceDiversity  int
	HourDistribution string // JSON [24]int
	DayDistribution  string // JSON [7]int
	Metadata         string // JSON
	ProfileUpdated   string
}

// Cluster represents a group of accounts with similar behavior.
type Cluster struct {
	ID            int
	ClusterType   string // "temporal", "behavioral", "coordination"
	Accounts      string // JSON array
	Score         float64
	Evidence      string // JSON
	Topic         string
	FirstDetected string
	LastDetected  string
	Occurrences   int
	Status        string // "active", "resolved", "archived"
}

// --- Alert CRUD ---

// SaveAlert inserts a new alert.
func (s *Store) SaveAlert(ctx context.Context, a *Alert) error {
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO alerts (alert_type, severity, title, description, evidence, confidence, source_filter, created_at, acknowledged)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		a.AlertType, a.Severity, a.Title, a.Description, a.Evidence,
		a.Confidence, a.SourceFilter, a.CreatedAt, boolToInt(a.Acknowledged),
	)
	if err != nil {
		return fmt.Errorf("saving alert: %w", err)
	}
	id, _ := res.LastInsertId()
	a.ID = int(id)
	return nil
}

// ListAlerts returns alerts filtered by severity, ordered by most recent.
func (s *Store) ListAlerts(ctx context.Context, severity string, limit int) ([]Alert, error) {
	query := `SELECT id, alert_type, severity, title, description, evidence, confidence, source_filter, created_at, acknowledged
		FROM alerts WHERE 1=1`
	var args []any

	if severity != "" {
		query += " AND severity = ?"
		args = append(args, severity)
	}

	query += " ORDER BY created_at DESC"

	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}

	return s.scanAlerts(ctx, query, args...)
}

// RecentAlerts returns the N most recent alerts.
func (s *Store) RecentAlerts(ctx context.Context, limit int) ([]Alert, error) {
	return s.ListAlerts(ctx, "", limit)
}

// AcknowledgeAlert marks an alert as acknowledged.
func (s *Store) AcknowledgeAlert(ctx context.Context, id int) error {
	res, err := s.db.ExecContext(ctx,
		"UPDATE alerts SET acknowledged = 1 WHERE id = ?", id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("alert %d not found", id)
	}
	return nil
}

// CountAlerts returns total alert count.
func (s *Store) CountAlerts(ctx context.Context) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM alerts").Scan(&count)
	return count, err
}

// SeverityCount holds a severity level and its count.
type SeverityCount struct {
	Severity string
	Count    int
}

// CountAlertsBySeverity returns alert counts grouped by severity.
func (s *Store) CountAlertsBySeverity(ctx context.Context) ([]SeverityCount, error) {
	rows, err := s.db.QueryContext(ctx,
		"SELECT severity, COUNT(*) FROM alerts GROUP BY severity ORDER BY COUNT(*) DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var counts []SeverityCount
	for rows.Next() {
		var sc SeverityCount
		if err := rows.Scan(&sc.Severity, &sc.Count); err != nil {
			return nil, err
		}
		counts = append(counts, sc)
	}
	return counts, rows.Err()
}

func (s *Store) scanAlerts(ctx context.Context, query string, args ...any) ([]Alert, error) {
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []Alert
	for rows.Next() {
		var a Alert
		var acked int
		var desc, evidence, sourceFilter sql.NullString
		var confidence sql.NullFloat64
		if err := rows.Scan(&a.ID, &a.AlertType, &a.Severity, &a.Title,
			&desc, &evidence, &confidence, &sourceFilter,
			&a.CreatedAt, &acked); err != nil {
			return nil, err
		}
		a.Description = desc.String
		a.Evidence = evidence.String
		a.Confidence = confidence.Float64
		a.SourceFilter = sourceFilter.String
		a.Acknowledged = acked != 0
		alerts = append(alerts, a)
	}
	return alerts, rows.Err()
}

// --- Profile CRUD ---

// SaveProfile inserts or updates an account profile (upsert on author).
func (s *Store) SaveProfile(ctx context.Context, p *AccountProfile) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO account_profiles (author, source, post_count, first_seen, last_seen,
			avg_posts_per_day, source_diversity, hour_distribution, day_distribution, metadata, profile_updated)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(author) DO UPDATE SET
			post_count        = excluded.post_count,
			last_seen         = excluded.last_seen,
			avg_posts_per_day = excluded.avg_posts_per_day,
			source_diversity  = excluded.source_diversity,
			hour_distribution = excluded.hour_distribution,
			day_distribution  = excluded.day_distribution,
			metadata          = excluded.metadata,
			profile_updated   = excluded.profile_updated`,
		p.Author, p.Source, p.PostCount, p.FirstSeen, p.LastSeen,
		p.AvgPostsPerDay, p.SourceDiversity, p.HourDistribution,
		p.DayDistribution, p.Metadata, p.ProfileUpdated,
	)
	return err
}

// GetProfile retrieves a profile by author name.
func (s *Store) GetProfile(ctx context.Context, author string) (*AccountProfile, error) {
	var p AccountProfile
	var hourDist, dayDist, meta sql.NullString
	err := s.db.QueryRowContext(ctx,
		`SELECT id, author, source, post_count, first_seen, last_seen,
			avg_posts_per_day, source_diversity, hour_distribution, day_distribution,
			metadata, profile_updated
		 FROM account_profiles WHERE author = ?`, author).Scan(
		&p.ID, &p.Author, &p.Source, &p.PostCount, &p.FirstSeen, &p.LastSeen,
		&p.AvgPostsPerDay, &p.SourceDiversity, &hourDist, &dayDist,
		&meta, &p.ProfileUpdated,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	p.HourDistribution = hourDist.String
	p.DayDistribution = dayDist.String
	p.Metadata = meta.String
	return &p, nil
}

// ListProfiles returns all account profiles, ordered by post count.
func (s *Store) ListProfiles(ctx context.Context, limit int) ([]AccountProfile, error) {
	query := `SELECT id, author, source, post_count, first_seen, last_seen,
		avg_posts_per_day, source_diversity, hour_distribution, day_distribution,
		metadata, profile_updated
		FROM account_profiles ORDER BY post_count DESC`
	var args []any

	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var profiles []AccountProfile
	for rows.Next() {
		var p AccountProfile
		var hourDist, dayDist, meta sql.NullString
		if err := rows.Scan(&p.ID, &p.Author, &p.Source, &p.PostCount,
			&p.FirstSeen, &p.LastSeen, &p.AvgPostsPerDay, &p.SourceDiversity,
			&hourDist, &dayDist, &meta, &p.ProfileUpdated); err != nil {
			return nil, err
		}
		p.HourDistribution = hourDist.String
		p.DayDistribution = dayDist.String
		p.Metadata = meta.String
		profiles = append(profiles, p)
	}
	return profiles, rows.Err()
}

// CountProfiles returns total profile count.
func (s *Store) CountProfiles(ctx context.Context) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM account_profiles").Scan(&count)
	return count, err
}

// --- Cluster CRUD ---

// SaveCluster inserts a new cluster.
func (s *Store) SaveCluster(ctx context.Context, c *Cluster) error {
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO clusters (cluster_type, accounts, score, evidence, topic, first_detected, last_detected, occurrences, status)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		c.ClusterType, c.Accounts, c.Score, c.Evidence, c.Topic,
		c.FirstDetected, c.LastDetected, c.Occurrences, c.Status,
	)
	if err != nil {
		return fmt.Errorf("saving cluster: %w", err)
	}
	id, _ := res.LastInsertId()
	c.ID = int(id)
	return nil
}

// ListClusters returns clusters filtered by status.
func (s *Store) ListClusters(ctx context.Context, status string, limit int) ([]Cluster, error) {
	query := `SELECT id, cluster_type, accounts, score, evidence, topic,
		first_detected, last_detected, occurrences, status
		FROM clusters WHERE 1=1`
	var args []any

	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}

	query += " ORDER BY score DESC, last_detected DESC"

	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clusters []Cluster
	for rows.Next() {
		var c Cluster
		var evidence, topic sql.NullString
		if err := rows.Scan(&c.ID, &c.ClusterType, &c.Accounts, &c.Score,
			&evidence, &topic, &c.FirstDetected, &c.LastDetected,
			&c.Occurrences, &c.Status); err != nil {
			return nil, err
		}
		c.Evidence = evidence.String
		c.Topic = topic.String
		clusters = append(clusters, c)
	}
	return clusters, rows.Err()
}

// ActiveClusters returns clusters with status "active".
func (s *Store) ActiveClusters(ctx context.Context) ([]Cluster, error) {
	return s.ListClusters(ctx, "active", 0)
}

// CountClusters returns total cluster count by status.
func (s *Store) CountClusters(ctx context.Context, status string) (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM clusters"
	var args []any
	if status != "" {
		query += " WHERE status = ?"
		args = append(args, status)
	}
	err := s.db.QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
}

// --- Deep Bore Session CRUD ---

// DeepBoreSession holds the mutable state of a Deep Bore monitoring session.
type DeepBoreSession struct {
	ID            int
	SessionName   string
	LogFile       string
	CursorOffset  int64
	AgentName     string
	TaskCurrent   int
	TaskTotal     int
	TaskName      string
	Turns         int
	CostDollars   float64
	UptimeSeconds int
	BinaryHash    string
	SandboxOK     bool
	StartedAt     string
	UpdatedAt     string
}

// UpsertDeepBoreSession creates or updates a Deep Bore session.
func (s *Store) UpsertDeepBoreSession(ctx context.Context, sess *DeepBoreSession) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO deepbore_sessions (session_name, log_file, cursor_offset, agent_name,
			task_current, task_total, task_name, turns, cost_dollars, uptime_seconds,
			binary_hash, sandbox_ok, started_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(session_name) DO UPDATE SET
			log_file       = excluded.log_file,
			cursor_offset  = excluded.cursor_offset,
			agent_name     = excluded.agent_name,
			task_current   = excluded.task_current,
			task_total     = excluded.task_total,
			task_name      = excluded.task_name,
			turns          = excluded.turns,
			cost_dollars   = excluded.cost_dollars,
			uptime_seconds = excluded.uptime_seconds,
			binary_hash    = excluded.binary_hash,
			sandbox_ok     = excluded.sandbox_ok,
			updated_at     = excluded.updated_at`,
		sess.SessionName, sess.LogFile, sess.CursorOffset, sess.AgentName,
		sess.TaskCurrent, sess.TaskTotal, sess.TaskName, sess.Turns,
		sess.CostDollars, sess.UptimeSeconds, sess.BinaryHash,
		boolToInt(sess.SandboxOK), sess.StartedAt, sess.UpdatedAt,
	)
	return err
}

// GetDeepBoreSession retrieves a session by name. Returns nil, nil if not found.
func (s *Store) GetDeepBoreSession(ctx context.Context, sessionName string) (*DeepBoreSession, error) {
	var sess DeepBoreSession
	var agentName, taskName, binaryHash sql.NullString
	var sandboxOK int
	err := s.db.QueryRowContext(ctx,
		`SELECT id, session_name, log_file, cursor_offset, agent_name,
			task_current, task_total, task_name, turns, cost_dollars,
			uptime_seconds, binary_hash, sandbox_ok, started_at, updated_at
		 FROM deepbore_sessions WHERE session_name = ?`, sessionName).Scan(
		&sess.ID, &sess.SessionName, &sess.LogFile, &sess.CursorOffset, &agentName,
		&sess.TaskCurrent, &sess.TaskTotal, &taskName, &sess.Turns,
		&sess.CostDollars, &sess.UptimeSeconds, &binaryHash, &sandboxOK,
		&sess.StartedAt, &sess.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	sess.AgentName = agentName.String
	sess.TaskName = taskName.String
	sess.BinaryHash = binaryHash.String
	sess.SandboxOK = sandboxOK != 0
	return &sess, nil
}

// GetLatestDeepBoreSession retrieves the most recently updated session.
// Returns nil, nil if no sessions exist.
func (s *Store) GetLatestDeepBoreSession(ctx context.Context) (*DeepBoreSession, error) {
	var sess DeepBoreSession
	var agentName, taskName, binaryHash sql.NullString
	var sandboxOK int
	err := s.db.QueryRowContext(ctx,
		`SELECT id, session_name, log_file, cursor_offset, agent_name,
			task_current, task_total, task_name, turns, cost_dollars,
			uptime_seconds, binary_hash, sandbox_ok, started_at, updated_at
		 FROM deepbore_sessions ORDER BY updated_at DESC LIMIT 1`).Scan(
		&sess.ID, &sess.SessionName, &sess.LogFile, &sess.CursorOffset, &agentName,
		&sess.TaskCurrent, &sess.TaskTotal, &taskName, &sess.Turns,
		&sess.CostDollars, &sess.UptimeSeconds, &binaryHash, &sandboxOK,
		&sess.StartedAt, &sess.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	sess.AgentName = agentName.String
	sess.TaskName = taskName.String
	sess.BinaryHash = binaryHash.String
	sess.SandboxOK = sandboxOK != 0
	return &sess, nil
}

// ListAlertsByType returns alerts of a specific type, ordered by most recent.
func (s *Store) ListAlertsByType(ctx context.Context, alertType string, limit int) ([]Alert, error) {
	query := `SELECT id, alert_type, severity, title, description, evidence, confidence, source_filter, created_at, acknowledged
		FROM alerts WHERE alert_type = ? ORDER BY created_at DESC`
	args := []any{alertType}
	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}
	return s.scanAlerts(ctx, query, args...)
}

// TimeNow returns the current time as ISO 8601.
func TimeNow() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
