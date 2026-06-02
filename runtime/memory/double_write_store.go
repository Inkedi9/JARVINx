package memory

// DoubleWriteStore fans writes to a primary and a secondary Store.
// The primary is the source of truth: all reads and Save() delegate to it.
// Secondary failures are silently discarded.
type DoubleWriteStore struct {
	primary   Store
	secondary Store
}

var _ Store = (*DoubleWriteStore)(nil)

func NewDoubleWriteStore(primary, secondary Store) *DoubleWriteStore {
	return &DoubleWriteStore{primary: primary, secondary: secondary}
}

func (d *DoubleWriteStore) Add(snap Snapshot) error {
	err := d.primary.Add(snap)
	_ = d.secondary.Add(snap)
	return err
}

func (d *DoubleWriteStore) AddCycle(record CycleRecord) error {
	err := d.primary.AddCycle(record)
	_ = d.secondary.AddCycle(record)
	return err
}

func (d *DoubleWriteStore) Last(n int) []Snapshot {
	if result := d.secondary.Last(n); len(result) > 0 {
		return result
	}
	return d.primary.Last(n)
}

func (d *DoubleWriteStore) LastCycles(n int) []CycleRecord {
	if result := d.secondary.LastCycles(n); len(result) > 0 {
		return result
	}
	return d.primary.LastCycles(n)
}
func (d *DoubleWriteStore) CurrentCycle() int              { return d.primary.CurrentCycle() }
func (d *DoubleWriteStore) Save() error                    { return d.primary.Save() }
