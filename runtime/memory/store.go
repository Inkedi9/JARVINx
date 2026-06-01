package memory

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
