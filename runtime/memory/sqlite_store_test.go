package memory

import (
	"testing"
	"time"
)

func openTestSQLite(t *testing.T) *SQLiteStore {
	t.Helper()
	s, err := OpenSQLiteStore(":memory:")
	if err != nil {
		t.Fatalf("OpenSQLiteStore: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestSQLiteStore_AddAndLast(t *testing.T) {
	s := openTestSQLite(t)

	snap := Snapshot{
		Timestamp:   time.Now().UTC().Truncate(time.Millisecond),
		CPUPercent:  42.5,
		MemUsed:     4096,
		MemTotal:    16384,
		MemPercent:  25.0,
		DiskUsed:    100,
		DiskTotal:   500,
		DiskPercent: 20.0,
	}

	if err := s.Add(snap); err != nil {
		t.Fatalf("Add: %v", err)
	}

	got := s.Last(1)
	if len(got) != 1 {
		t.Fatalf("Last(1): want 1 result, got %d", len(got))
	}
	if got[0].CPUPercent != snap.CPUPercent {
		t.Errorf("CPUPercent: want %.1f, got %.1f", snap.CPUPercent, got[0].CPUPercent)
	}
}

func TestSQLiteStore_LastReturnsChronological(t *testing.T) {
	s := openTestSQLite(t)

	for i := range 5 {
		if err := s.Add(Snapshot{
			Timestamp:  time.Now().UTC(),
			CPUPercent: float64(i + 1),
		}); err != nil {
			t.Fatalf("Add[%d]: %v", i, err)
		}
	}

	got := s.Last(5)
	if len(got) != 5 {
		t.Fatalf("want 5 snapshots, got %d", len(got))
	}
	for i := 1; i < len(got); i++ {
		if got[i].CPUPercent <= got[i-1].CPUPercent {
			t.Errorf("snapshots not in chronological order at index %d", i)
		}
	}
}

func TestSQLiteStore_LastCappedAtN(t *testing.T) {
	s := openTestSQLite(t)

	for range 10 {
		_ = s.Add(Snapshot{Timestamp: time.Now().UTC()})
	}

	got := s.Last(3)
	if len(got) != 3 {
		t.Errorf("Last(3): want 3, got %d", len(got))
	}
}

func TestSQLiteStore_AddCycleAndLastCycles(t *testing.T) {
	s := openTestSQLite(t)

	record := NewCycleRecord(
		Snapshot{Timestamp: time.Now().UTC(), CPUPercent: 50.0},
		"alert", "cpu high", "reduce load", "",
	)
	record.Confidence = 0.9

	if err := s.AddCycle(record); err != nil {
		t.Fatalf("AddCycle: %v", err)
	}

	got := s.LastCycles(1)
	if len(got) != 1 {
		t.Fatalf("LastCycles(1): want 1, got %d", len(got))
	}
	if got[0].Action != "alert" {
		t.Errorf("Action: want alert, got %s", got[0].Action)
	}
	if got[0].Confidence != 0.9 {
		t.Errorf("Confidence: want 0.9, got %.1f", got[0].Confidence)
	}
}

func TestSQLiteStore_CurrentCycleIncrements(t *testing.T) {
	s := openTestSQLite(t)

	if s.CurrentCycle() != 0 {
		t.Fatalf("initial CurrentCycle: want 0, got %d", s.CurrentCycle())
	}

	for i := 1; i <= 3; i++ {
		_ = s.AddCycle(NewCycleRecord(Snapshot{Timestamp: time.Now()}, "log", "", "", ""))
		if s.CurrentCycle() != i {
			t.Errorf("after cycle %d: want %d, got %d", i, i, s.CurrentCycle())
		}
	}
}

func TestSQLiteStore_PersistsCycleNum(t *testing.T) {
	path := t.TempDir() + "/persist.db"

	s1, err := OpenSQLiteStore(path)
	if err != nil {
		t.Fatalf("first open: %v", err)
	}
	for range 5 {
		_ = s1.AddCycle(NewCycleRecord(Snapshot{Timestamp: time.Now()}, "log", "", "", ""))
	}
	_ = s1.Close()

	s2, err := OpenSQLiteStore(path)
	if err != nil {
		t.Fatalf("second open: %v", err)
	}
	defer func() { _ = s2.Close() }()

	if s2.CurrentCycle() != 5 {
		t.Errorf("persisted CycleNum: want 5, got %d", s2.CurrentCycle())
	}
}

func TestSQLiteStore_LastZeroReturnsNil(t *testing.T) {
	s := openTestSQLite(t)
	if s.Last(0) != nil {
		t.Error("Last(0) should return nil")
	}
	if s.LastCycles(0) != nil {
		t.Error("LastCycles(0) should return nil")
	}
}

func TestSQLiteStore_Benchmark_5760(t *testing.T) {
	s := openTestSQLite(t)

	// Inserts dans une transaction pour éviter 5760 auto-commits individuels.
	base := time.Now().UTC().Add(-24 * time.Hour)
	tx, err := s.db.Begin()
	if err != nil {
		t.Fatalf("Begin: %v", err)
	}
	for i := range 5760 {
		snap := Snapshot{
			Timestamp:   base.Add(time.Duration(i) * 15 * time.Second),
			CPUPercent:  float64(i%100) * 0.9,
			MemUsed:     4096,
			MemTotal:    16384,
			MemPercent:  float64(i%100) * 0.5,
			DiskPercent: 80.0 + float64(i%5),
		}
		if _, execErr := tx.Exec(
			`INSERT INTO snapshots (timestamp,cpu_percent,mem_used_mb,mem_total_mb,mem_percent,disk_used_gb,disk_total_gb,disk_percent)
			 VALUES (?,?,?,?,?,?,?,?)`,
			snap.Timestamp.UTC().Format(time.RFC3339Nano),
			snap.CPUPercent, snap.MemUsed, snap.MemTotal, snap.MemPercent,
			snap.DiskUsed, snap.DiskTotal, snap.DiskPercent,
		); execErr != nil {
			_ = tx.Rollback()
			t.Fatalf("insert snapshot[%d]: %v", i, execErr)
		}
		if _, execErr := tx.Exec(
			`INSERT INTO cycles (cycle_num,timestamp,action,analysis,reason,command,snap_cpu_percent,snap_mem_used_mb,snap_mem_total_mb,snap_mem_percent,snap_disk_used_gb,snap_disk_total_gb,snap_disk_percent)
			 VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`,
			i+1,
			snap.Timestamp.UTC().Format(time.RFC3339Nano),
			"log", "bench", "bench", "",
			snap.CPUPercent, snap.MemUsed, snap.MemTotal, snap.MemPercent,
			snap.DiskUsed, snap.DiskTotal, snap.DiskPercent,
		); execErr != nil {
			_ = tx.Rollback()
			t.Fatalf("insert cycle[%d]: %v", i, execErr)
		}
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("Commit: %v", err)
	}

	start := time.Now()
	cycles := s.LastCycles(5760)
	elapsed := time.Since(start)

	if len(cycles) != 5760 {
		t.Errorf("LastCycles(5760): want 5760, got %d", len(cycles))
	}
	if elapsed > 3*time.Second {
		t.Errorf("LastCycles(5760) trop lent : %v (max 3s)", elapsed)
	}
	t.Logf("LastCycles(5760) : %v", elapsed)

	start = time.Now()
	buckets := s.SnapshotBuckets(base, 1)
	elapsed = time.Since(start)

	if len(buckets) == 0 {
		t.Error("SnapshotBuckets(1h) : résultat vide inattendu")
	}
	t.Logf("SnapshotBuckets(24h, 1h) : %v — %d buckets", elapsed, len(buckets))
}
