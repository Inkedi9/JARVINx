package tools

import "github.com/shirou/gopsutil/v3/net"

// NetCounters holds raw cumulative byte counters (aggregated across all interfaces).
type NetCounters struct {
	Recv uint64
	Sent uint64
}

// ReadNetCounters reads aggregated IO counters. Returns false on error or no data.
func ReadNetCounters() (NetCounters, bool) {
	stats, err := net.IOCounters(false)
	if err != nil || len(stats) == 0 {
		return NetCounters{}, false
	}
	return NetCounters{
		Recv: stats[0].BytesRecv,
		Sent: stats[0].BytesSent,
	}, true
}

// DeltaMBps computes MB/s from two counter snapshots and elapsed seconds.
// Returns 0,0 if elapsed ≤ 0 or counters wrapped (reboot/reset).
func DeltaMBps(prev, curr NetCounters, elapsedSec float64) (recvMBps, sentMBps float64) {
	if elapsedSec <= 0 || curr.Recv < prev.Recv || curr.Sent < prev.Sent {
		return 0, 0
	}
	recvMBps = float64(curr.Recv-prev.Recv) / elapsedSec / 1024 / 1024
	sentMBps = float64(curr.Sent-prev.Sent) / elapsedSec / 1024 / 1024
	return
}
