package memory

import "time"

// Store abstracts the state persistence layer — implemented by *State (JSON) and future SQLite backend.
type Store interface {
	Add(Snapshot) error
	AddCycle(CycleRecord) error
	Last(n int) []Snapshot
	LastCycles(n int) []CycleRecord
	CurrentCycle() int
	Save() error
}

// EventLog abstracts the append-only metrics logger.
type EventLog interface {
	Write(LogEntry) error
	Status() LogStatus
}

// SnapshotBucket holds pre-aggregated snapshot metrics for a time window.
type SnapshotBucket struct {
	Timestamp string  `json:"timestamp"`
	CPUAvg    float64 `json:"cpu_avg"`
	CPUMax    float64 `json:"cpu_max"`
	MEMAvg    float64 `json:"mem_avg"`
	MEMMax    float64 `json:"mem_max"`
	DiskAvg   float64 `json:"disk_avg"`
	DiskMax   float64 `json:"disk_max"`
	Count     int     `json:"count"`
}

// HistoryReader provides time-range historical queries — implemented by *SQLiteStore.
type HistoryReader interface {
	SnapshotBuckets(from time.Time, bucketHours int) []SnapshotBucket
	TotalSnapshots(from time.Time) int
}
