package memory

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

// SQLiteStore implements Store with an SQLite backend — unlimited history.
type SQLiteStore struct {
	db       *sql.DB
	mu       sync.Mutex
	cycleNum int
}

var _ Store = (*SQLiteStore)(nil)

// OpenSQLiteStore opens (or creates) the SQLite database at path and runs migrations.
// Use ":memory:" in tests.
func OpenSQLiteStore(path string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("sqlite open: %w", err)
	}

	db.SetMaxOpenConns(1)

	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("sqlite pragma: %w", err)
	}
	if _, err := db.Exec("PRAGMA synchronous=NORMAL"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("sqlite pragma synchronous: %w", err)
	}

	if err := sqliteMigrate(db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("sqlite migrate: %w", err)
	}

	var maxCycle sql.NullInt64
	if err := db.QueryRow(`SELECT MAX(cycle_num) FROM cycles`).Scan(&maxCycle); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("sqlite init cycle_num: %w", err)
	}

	s := &SQLiteStore{db: db}
	if maxCycle.Valid {
		s.cycleNum = int(maxCycle.Int64)
	}
	return s, nil
}

// Close releases the database connection.
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

func sqliteMigrate(db *sql.DB) error {
	const createSnapshots = `
		CREATE TABLE IF NOT EXISTS snapshots (
			id            INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp     DATETIME NOT NULL,
			cpu_percent   REAL     NOT NULL,
			mem_used_mb   INTEGER  NOT NULL,
			mem_total_mb  INTEGER  NOT NULL,
			mem_percent   REAL     NOT NULL,
			disk_used_gb  INTEGER  NOT NULL,
			disk_total_gb INTEGER  NOT NULL,
			disk_percent  REAL     NOT NULL
		)`

	const createCycles = `
		CREATE TABLE IF NOT EXISTS cycles (
			id               INTEGER PRIMARY KEY AUTOINCREMENT,
			cycle_num        INTEGER NOT NULL,
			timestamp        DATETIME NOT NULL,
			action           TEXT    NOT NULL,
			analysis         TEXT    NOT NULL,
			reason           TEXT    NOT NULL,
			command          TEXT    DEFAULT '',
			trigger_cpu      REAL    DEFAULT 0,
			trigger_ram      REAL    DEFAULT 0,
			trigger_disk     REAL    DEFAULT 0,
			confidence       REAL    DEFAULT 0,
			snap_cpu_percent  REAL    NOT NULL,
			snap_mem_used_mb  INTEGER NOT NULL,
			snap_mem_total_mb INTEGER NOT NULL,
			snap_mem_percent  REAL    NOT NULL,
			snap_disk_used_gb  INTEGER NOT NULL,
			snap_disk_total_gb INTEGER NOT NULL,
			snap_disk_percent  REAL    NOT NULL
		)`

	const idxSnapshotsTS = `CREATE INDEX IF NOT EXISTS idx_snapshots_ts ON snapshots(timestamp)`
	const idxCyclesTS = `CREATE INDEX IF NOT EXISTS idx_cycles_ts ON cycles(timestamp)`

	for _, stmt := range []string{createSnapshots, createCycles, idxSnapshotsTS, idxCyclesTS} {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

func (s *SQLiteStore) Add(snap Snapshot) error {
	_, err := s.db.Exec(
		`INSERT INTO snapshots
			(timestamp, cpu_percent, mem_used_mb, mem_total_mb, mem_percent,
			 disk_used_gb, disk_total_gb, disk_percent)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		snap.Timestamp.UTC().Format(time.RFC3339Nano),
		snap.CPUPercent, snap.MemUsed, snap.MemTotal, snap.MemPercent,
		snap.DiskUsed, snap.DiskTotal, snap.DiskPercent,
	)
	if err != nil {
		return fmt.Errorf("sqlite insert snapshot: %w", err)
	}
	return nil
}

func (s *SQLiteStore) AddCycle(record CycleRecord) error {
	s.mu.Lock()
	s.cycleNum++
	record.CycleNum = s.cycleNum
	s.mu.Unlock()

	_, err := s.db.Exec(
		`INSERT INTO cycles
			(cycle_num, timestamp, action, analysis, reason, command,
			 trigger_cpu, trigger_ram, trigger_disk, confidence,
			 snap_cpu_percent, snap_mem_used_mb, snap_mem_total_mb, snap_mem_percent,
			 snap_disk_used_gb, snap_disk_total_gb, snap_disk_percent)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		record.CycleNum,
		record.Timestamp.UTC().Format(time.RFC3339Nano),
		record.Action, record.Analysis, record.Reason, record.Command,
		record.TriggerCPU, record.TriggerRAM, record.TriggerDisk, record.Confidence,
		record.Snapshot.CPUPercent, record.Snapshot.MemUsed, record.Snapshot.MemTotal,
		record.Snapshot.MemPercent, record.Snapshot.DiskUsed, record.Snapshot.DiskTotal,
		record.Snapshot.DiskPercent,
	)
	if err != nil {
		return fmt.Errorf("sqlite insert cycle: %w", err)
	}
	return nil
}

func (s *SQLiteStore) Last(n int) []Snapshot {
	if n <= 0 {
		return nil
	}

	rows, err := s.db.Query(
		`SELECT timestamp, cpu_percent, mem_used_mb, mem_total_mb, mem_percent,
		        disk_used_gb, disk_total_gb, disk_percent
		 FROM snapshots ORDER BY id DESC LIMIT ?`, n)
	if err != nil {
		return nil
	}
	defer func() { _ = rows.Close() }()

	snaps := make([]Snapshot, 0, n)
	for rows.Next() {
		var snap Snapshot
		var ts string
		if err := rows.Scan(&ts, &snap.CPUPercent, &snap.MemUsed, &snap.MemTotal,
			&snap.MemPercent, &snap.DiskUsed, &snap.DiskTotal, &snap.DiskPercent); err != nil {
			continue
		}
		snap.Timestamp, _ = time.Parse(time.RFC3339Nano, ts)
		snaps = append(snaps, snap)
	}

	reverseSnapshots(snaps)
	return snaps
}

func (s *SQLiteStore) LastCycles(n int) []CycleRecord {
	if n <= 0 {
		return nil
	}

	rows, err := s.db.Query(
		`SELECT cycle_num, timestamp, action, analysis, reason, command,
		        trigger_cpu, trigger_ram, trigger_disk, confidence,
		        snap_cpu_percent, snap_mem_used_mb, snap_mem_total_mb, snap_mem_percent,
		        snap_disk_used_gb, snap_disk_total_gb, snap_disk_percent
		 FROM cycles ORDER BY id DESC LIMIT ?`, n)
	if err != nil {
		return nil
	}
	defer func() { _ = rows.Close() }()

	cycles := make([]CycleRecord, 0, n)
	for rows.Next() {
		var r CycleRecord
		var ts string
		if err := rows.Scan(
			&r.CycleNum, &ts, &r.Action, &r.Analysis, &r.Reason, &r.Command,
			&r.TriggerCPU, &r.TriggerRAM, &r.TriggerDisk, &r.Confidence,
			&r.Snapshot.CPUPercent, &r.Snapshot.MemUsed, &r.Snapshot.MemTotal,
			&r.Snapshot.MemPercent, &r.Snapshot.DiskUsed, &r.Snapshot.DiskTotal,
			&r.Snapshot.DiskPercent,
		); err != nil {
			continue
		}
		r.Timestamp, _ = time.Parse(time.RFC3339Nano, ts)
		r.Snapshot.Timestamp = r.Timestamp
		cycles = append(cycles, r)
	}

	reverseCycles(cycles)
	return cycles
}

func (s *SQLiteStore) CurrentCycle() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.cycleNum
}

// Save is a no-op — SQLite writes are immediate.
func (s *SQLiteStore) Save() error { return nil }

// SnapshotBuckets returns pre-aggregated metrics grouped by time bucket.
// bucketHours: 1 = hourly, 6 = 6-hourly, 24 = daily.
func (s *SQLiteStore) SnapshotBuckets(from time.Time, bucketHours int) []SnapshotBucket {
	var bucketExpr string
	switch bucketHours {
	case 1:
		bucketExpr = `strftime('%Y-%m-%dT%H:00:00Z', timestamp)`
	case 6:
		bucketExpr = `strftime('%Y-%m-%dT', timestamp) || printf('%02d:00:00Z', (CAST(strftime('%H', timestamp) AS INTEGER) / 6) * 6)`
	default: // 24
		bucketExpr = `strftime('%Y-%m-%dT00:00:00Z', timestamp)`
	}

	q := fmt.Sprintf(`
		SELECT %s AS bucket,
			ROUND(AVG(cpu_percent), 2), ROUND(MAX(cpu_percent), 2),
			ROUND(AVG(mem_percent), 2), ROUND(MAX(mem_percent), 2),
			ROUND(AVG(disk_percent), 2), ROUND(MAX(disk_percent), 2),
			COUNT(*)
		FROM snapshots
		WHERE timestamp >= ?
		GROUP BY bucket
		ORDER BY bucket ASC`, bucketExpr)

	rows, err := s.db.Query(q, from.UTC().Format(time.RFC3339))
	if err != nil {
		return nil
	}
	defer func() { _ = rows.Close() }()

	var buckets []SnapshotBucket
	for rows.Next() {
		var b SnapshotBucket
		if err := rows.Scan(
			&b.Timestamp,
			&b.CPUAvg, &b.CPUMax,
			&b.MEMAvg, &b.MEMMax,
			&b.DiskAvg, &b.DiskMax,
			&b.Count,
		); err != nil {
			continue
		}
		buckets = append(buckets, b)
	}
	return buckets
}

// TotalSnapshots returns the number of snapshots recorded since from.
func (s *SQLiteStore) TotalSnapshots(from time.Time) int {
	var count int
	if err := s.db.QueryRow(
		`SELECT COUNT(*) FROM snapshots WHERE timestamp >= ?`,
		from.UTC().Format(time.RFC3339),
	).Scan(&count); err != nil {
		return 0
	}
	return count
}

func reverseSnapshots(s []Snapshot) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

func reverseCycles(c []CycleRecord) {
	for i, j := 0, len(c)-1; i < j; i, j = i+1, j-1 {
		c[i], c[j] = c[j], c[i]
	}
}
