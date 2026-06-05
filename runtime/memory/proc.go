package memory

// ProcInfo holds a single process snapshot sorted by resident memory.
type ProcInfo struct {
	PID   int32  `json:"pid"`
	Name  string `json:"name"`
	MemMB uint64 `json:"mem_mb"`
}
