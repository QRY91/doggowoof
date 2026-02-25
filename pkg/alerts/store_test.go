package alerts

import (
	"context"
	"testing"
)

func TestMigrationIdempotency(t *testing.T) {
	store, err := NewMemory()
	if err != nil {
		t.Fatalf("NewMemory() failed: %v", err)
	}
	defer store.Close()

	// Run migrations again — should not fail
	if err := store.migrate(); err != nil {
		t.Fatalf("second migration failed: %v", err)
	}
}

func TestAlertCRUD(t *testing.T) {
	store, err := NewMemory()
	if err != nil {
		t.Fatalf("NewMemory() failed: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Save alert
	alert := &Alert{
		AlertType:   "temporal_cluster",
		Severity:    "high",
		Title:       "Test alert",
		Description: "Testing",
		Confidence:  0.85,
		CreatedAt:   TimeNow(),
	}
	if err := store.SaveAlert(ctx, alert); err != nil {
		t.Fatalf("SaveAlert() failed: %v", err)
	}
	if alert.ID == 0 {
		t.Error("expected non-zero alert ID after save")
	}

	// Count
	count, err := store.CountAlerts(ctx)
	if err != nil {
		t.Fatalf("CountAlerts() failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 alert, got %d", count)
	}

	// List
	alerts, err := store.ListAlerts(ctx, "", 10)
	if err != nil {
		t.Fatalf("ListAlerts() failed: %v", err)
	}
	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(alerts))
	}
	if alerts[0].Title != "Test alert" {
		t.Errorf("expected title 'Test alert', got %q", alerts[0].Title)
	}

	// Filter by severity
	highAlerts, err := store.ListAlerts(ctx, "high", 10)
	if err != nil {
		t.Fatalf("ListAlerts(high) failed: %v", err)
	}
	if len(highAlerts) != 1 {
		t.Errorf("expected 1 high alert, got %d", len(highAlerts))
	}
	lowAlerts, err := store.ListAlerts(ctx, "low", 10)
	if err != nil {
		t.Fatalf("ListAlerts(low) failed: %v", err)
	}
	if len(lowAlerts) != 0 {
		t.Errorf("expected 0 low alerts, got %d", len(lowAlerts))
	}

	// Acknowledge
	if err := store.AcknowledgeAlert(ctx, alert.ID); err != nil {
		t.Fatalf("AcknowledgeAlert() failed: %v", err)
	}
	alerts, _ = store.ListAlerts(ctx, "", 10)
	if !alerts[0].Acknowledged {
		t.Error("expected alert to be acknowledged")
	}

	// Count by severity
	bySev, err := store.CountAlertsBySeverity(ctx)
	if err != nil {
		t.Fatalf("CountAlertsBySeverity() failed: %v", err)
	}
	if len(bySev) != 1 || bySev[0].Severity != "high" {
		t.Errorf("unexpected severity counts: %v", bySev)
	}
}

func TestProfileCRUD(t *testing.T) {
	store, err := NewMemory()
	if err != nil {
		t.Fatalf("NewMemory() failed: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	profile := &AccountProfile{
		Author:           "testuser",
		Source:           "reddit",
		PostCount:        42,
		FirstSeen:        "2025-01-01T00:00:00Z",
		LastSeen:         "2025-06-01T00:00:00Z",
		AvgPostsPerDay:   1.5,
		SourceDiversity:  2,
		HourDistribution: "[1,2,3,4,5,6,7,8,9,10,11,12,11,10,9,8,7,6,5,4,3,2,1,0]",
		DayDistribution:  "[5,6,7,8,7,6,5]",
		ProfileUpdated:   TimeNow(),
	}

	// Save
	if err := store.SaveProfile(ctx, profile); err != nil {
		t.Fatalf("SaveProfile() failed: %v", err)
	}

	// Get
	got, err := store.GetProfile(ctx, "testuser")
	if err != nil {
		t.Fatalf("GetProfile() failed: %v", err)
	}
	if got == nil {
		t.Fatal("expected profile, got nil")
	}
	if got.PostCount != 42 {
		t.Errorf("expected post count 42, got %d", got.PostCount)
	}

	// Upsert
	profile.PostCount = 100
	if err := store.SaveProfile(ctx, profile); err != nil {
		t.Fatalf("SaveProfile(upsert) failed: %v", err)
	}
	got, _ = store.GetProfile(ctx, "testuser")
	if got.PostCount != 100 {
		t.Errorf("expected post count 100 after upsert, got %d", got.PostCount)
	}

	// Get nonexistent
	missing, err := store.GetProfile(ctx, "nobody")
	if err != nil {
		t.Fatalf("GetProfile(nobody) failed: %v", err)
	}
	if missing != nil {
		t.Error("expected nil for nonexistent profile")
	}

	// List
	profiles, err := store.ListProfiles(ctx, 10)
	if err != nil {
		t.Fatalf("ListProfiles() failed: %v", err)
	}
	if len(profiles) != 1 {
		t.Errorf("expected 1 profile, got %d", len(profiles))
	}

	// Count
	count, err := store.CountProfiles(ctx)
	if err != nil {
		t.Fatalf("CountProfiles() failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 profile count, got %d", count)
	}
}

func TestClusterCRUD(t *testing.T) {
	store, err := NewMemory()
	if err != nil {
		t.Fatalf("NewMemory() failed: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	cluster := &Cluster{
		ClusterType:   "temporal",
		Accounts:      `["user1","user2","user3"]`,
		Score:         0.85,
		Evidence:      `{"window":"5m"}`,
		Topic:         "politics",
		FirstDetected: TimeNow(),
		LastDetected:  TimeNow(),
		Occurrences:   1,
		Status:        "active",
	}

	// Save
	if err := store.SaveCluster(ctx, cluster); err != nil {
		t.Fatalf("SaveCluster() failed: %v", err)
	}
	if cluster.ID == 0 {
		t.Error("expected non-zero cluster ID")
	}

	// List active
	active, err := store.ActiveClusters(ctx)
	if err != nil {
		t.Fatalf("ActiveClusters() failed: %v", err)
	}
	if len(active) != 1 {
		t.Errorf("expected 1 active cluster, got %d", len(active))
	}

	// Count
	count, err := store.CountClusters(ctx, "active")
	if err != nil {
		t.Fatalf("CountClusters() failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 active cluster count, got %d", count)
	}
}

func TestDeepBoreSessionCRUD(t *testing.T) {
	store, err := NewMemory()
	if err != nil {
		t.Fatalf("NewMemory() failed: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	sess := &DeepBoreSession{
		SessionName:   "bridge-2026-02-23_202508",
		LogFile:       "bridge-2026-02-23_202508.log",
		CursorOffset:  0,
		AgentName:     "doug-nauvis",
		TaskCurrent:   12,
		TaskTotal:     33,
		TaskName:      "belt-to-assembler",
		Turns:         87,
		CostDollars:   6.50,
		UptimeSeconds: 1800,
		BinaryHash:    "abc123def456",
		SandboxOK:     true,
		StartedAt:     TimeNow(),
		UpdatedAt:     TimeNow(),
	}

	// Save
	if err := store.UpsertDeepBoreSession(ctx, sess); err != nil {
		t.Fatalf("UpsertDeepBoreSession() failed: %v", err)
	}

	// Get by name
	got, err := store.GetDeepBoreSession(ctx, "bridge-2026-02-23_202508")
	if err != nil {
		t.Fatalf("GetDeepBoreSession() failed: %v", err)
	}
	if got == nil {
		t.Fatal("expected session, got nil")
	}
	if got.AgentName != "doug-nauvis" {
		t.Errorf("expected agent 'doug-nauvis', got %q", got.AgentName)
	}
	if got.TaskCurrent != 12 {
		t.Errorf("expected task 12, got %d", got.TaskCurrent)
	}
	if !got.SandboxOK {
		t.Error("expected sandbox ok")
	}

	// Upsert (update)
	sess.TaskCurrent = 24
	sess.Turns = 187
	sess.CostDollars = 14.23
	sess.UpdatedAt = TimeNow()
	if err := store.UpsertDeepBoreSession(ctx, sess); err != nil {
		t.Fatalf("UpsertDeepBoreSession(update) failed: %v", err)
	}
	got, _ = store.GetDeepBoreSession(ctx, "bridge-2026-02-23_202508")
	if got.TaskCurrent != 24 {
		t.Errorf("expected task 24 after upsert, got %d", got.TaskCurrent)
	}
	if got.Turns != 187 {
		t.Errorf("expected 187 turns after upsert, got %d", got.Turns)
	}

	// Get nonexistent
	missing, err := store.GetDeepBoreSession(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("GetDeepBoreSession(nonexistent) failed: %v", err)
	}
	if missing != nil {
		t.Error("expected nil for nonexistent session")
	}

	// Get latest
	latest, err := store.GetLatestDeepBoreSession(ctx)
	if err != nil {
		t.Fatalf("GetLatestDeepBoreSession() failed: %v", err)
	}
	if latest == nil {
		t.Fatal("expected latest session, got nil")
	}
	if latest.SessionName != "bridge-2026-02-23_202508" {
		t.Errorf("expected session name, got %q", latest.SessionName)
	}
}

func TestListAlertsByType(t *testing.T) {
	store, err := NewMemory()
	if err != nil {
		t.Fatalf("NewMemory() failed: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Save alerts of different types
	for _, at := range []string{"sandbox_violation", "sandbox_violation", "agent_broadcast", "agent_start"} {
		store.SaveAlert(ctx, &Alert{
			AlertType: at,
			Severity:  "critical",
			Title:     "test " + at,
			CreatedAt: TimeNow(),
		})
	}

	// Filter by type
	violations, err := store.ListAlertsByType(ctx, "sandbox_violation", 10)
	if err != nil {
		t.Fatalf("ListAlertsByType() failed: %v", err)
	}
	if len(violations) != 2 {
		t.Errorf("expected 2 sandbox_violation alerts, got %d", len(violations))
	}

	broadcasts, err := store.ListAlertsByType(ctx, "agent_broadcast", 10)
	if err != nil {
		t.Fatalf("ListAlertsByType() failed: %v", err)
	}
	if len(broadcasts) != 1 {
		t.Errorf("expected 1 agent_broadcast alert, got %d", len(broadcasts))
	}

	// Limit
	limited, err := store.ListAlertsByType(ctx, "sandbox_violation", 1)
	if err != nil {
		t.Fatalf("ListAlertsByType(limit) failed: %v", err)
	}
	if len(limited) != 1 {
		t.Errorf("expected 1 limited alert, got %d", len(limited))
	}
}
