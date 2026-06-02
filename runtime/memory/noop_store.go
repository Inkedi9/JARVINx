package memory

// NoopStore is a Store that silently discards all writes and returns empty reads.
// Used as the secondary in DoubleWriteStore when SQLite is unavailable at startup.
type NoopStore struct{}

var _ Store = NoopStore{}

func (NoopStore) Add(Snapshot) error           { return nil }
func (NoopStore) AddCycle(CycleRecord) error   { return nil }
func (NoopStore) Last(int) []Snapshot          { return nil }
func (NoopStore) LastCycles(int) []CycleRecord { return nil }
func (NoopStore) CurrentCycle() int            { return 0 }
func (NoopStore) Save() error                  { return nil }
