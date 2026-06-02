package memory

import (
	"errors"
	"testing"
	"time"
)

// errStore is a Store whose writes always fail — used to test fail-silent behavior.
type errStore struct{ NoopStore }

func (errStore) Add(Snapshot) error         { return errors.New("secondary error") }
func (errStore) AddCycle(CycleRecord) error { return errors.New("secondary error") }

func TestDoubleWriteStore_PrimaryAuthoritative(t *testing.T) {
	primary := NewState("")
	secondary := NoopStore{}
	dw := NewDoubleWriteStore(primary, secondary)

	snap := Snapshot{Timestamp: time.Now(), CPUPercent: 77.0}
	if err := dw.Add(snap); err != nil {
		t.Fatalf("Add: %v", err)
	}

	got := dw.Last(1)
	if len(got) != 1 || got[0].CPUPercent != 77.0 {
		t.Errorf("Last: want CPUPercent=77, got %+v", got)
	}
}

func TestDoubleWriteStore_SecondaryFailSilent(t *testing.T) {
	primary := NewState("")
	dw := NewDoubleWriteStore(primary, errStore{})

	if err := dw.Add(Snapshot{Timestamp: time.Now()}); err != nil {
		t.Errorf("Add should succeed even when secondary fails, got: %v", err)
	}
	if err := dw.AddCycle(NewCycleRecord(Snapshot{Timestamp: time.Now()}, "log", "", "", "")); err != nil {
		t.Errorf("AddCycle should succeed even when secondary fails, got: %v", err)
	}
}

func TestDoubleWriteStore_CurrentCycleDelegatesToPrimary(t *testing.T) {
	primary := NewState("")
	dw := NewDoubleWriteStore(primary, NoopStore{})

	_ = dw.AddCycle(NewCycleRecord(Snapshot{Timestamp: time.Now()}, "log", "", "", ""))
	_ = dw.AddCycle(NewCycleRecord(Snapshot{Timestamp: time.Now()}, "log", "", "", ""))

	if dw.CurrentCycle() != primary.CurrentCycle() {
		t.Errorf("CurrentCycle mismatch: dw=%d primary=%d", dw.CurrentCycle(), primary.CurrentCycle())
	}
}

func TestDoubleWriteStore_SQLiteServes25WhenJSONCappedAt20(t *testing.T) {
	primary := NewState("")
	sq, err := OpenSQLiteStore(":memory:")
	if err != nil {
		t.Fatalf("OpenSQLiteStore: %v", err)
	}
	defer func() { _ = sq.Close() }()

	dw := NewDoubleWriteStore(primary, sq)

	for i := range 25 {
		_ = dw.AddCycle(NewCycleRecord(
			Snapshot{Timestamp: time.Now(), CPUPercent: float64(i)},
			"log", "cohérence", "test", "",
		))
	}

	// DoubleWriteStore lit depuis SQLite → 25 résultats
	all := dw.LastCycles(25)
	if len(all) != 25 {
		t.Errorf("LastCycles(25) via SQLite : want 25, got %d", len(all))
	}

	// Primary JSON reste capé à 20
	fromPrimary := primary.LastCycles(25)
	if len(fromPrimary) != 20 {
		t.Errorf("JSON State capé à 20, got %d", len(fromPrimary))
	}
}

func TestDoubleWriteStore_SaveDelegatesToPrimary(t *testing.T) {
	path := t.TempDir() + "/dw_state.json"
	primary := NewState(path)
	dw := NewDoubleWriteStore(primary, NoopStore{})

	_ = dw.Add(Snapshot{Timestamp: time.Now(), CPUPercent: 10.0})
	if err := dw.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Reload and verify persistence
	reloaded := NewState(path)
	if len(reloaded.Last(1)) == 0 {
		t.Error("Save did not persist to primary")
	}
}
