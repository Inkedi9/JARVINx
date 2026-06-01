package agents

import (
	"context"
	"fmt"
	"time"

	jxlog "github.com/Inkedi9/jarvinx/jxlog"
	"github.com/Inkedi9/jarvinx/memory"
)

// ReportData contient les stats agrégées pour le rapport
type ReportData struct {
	Date         string
	TotalCycles  int
	Actions      map[string]int
	CPUAvg       float64
	CPUMax       float64
	RAMAvg       float64
	RAMMax       float64
	DiskMax      float64
	AlertCount   int
	ExecuteCount int
}

// DailyReporter envoie un rapport quotidien via le dispatcher
type DailyReporter struct {
	dispatcher *NotifierDispatcher
	state      memory.Store
	hour       int
	minute     int
	dryRun     bool
	lastSent   time.Time
}

func NewDailyReporter(
	dispatcher *NotifierDispatcher,
	state memory.Store,
	hour, minute int,
	dryRun bool,
) *DailyReporter {
	return &DailyReporter{
		dispatcher: dispatcher,
		state:      state,
		hour:       hour,
		minute:     minute,
		dryRun:     dryRun,
	}
}

// Start lance la goroutine de rapport — bloque jusqu'à ctx.Done()
func (r *DailyReporter) Start(ctx context.Context) {
	jxlog.Info("DAILY REPORT", fmt.Sprintf(
		"Rapport quotidien activé — envoi à %02d:%02d", r.hour, r.minute,
	))

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case t := <-ticker.C:
			if t.Hour() == r.hour && t.Minute() == r.minute {
				// Évite d'envoyer deux fois dans la même minute
				if time.Since(r.lastSent) < 2*time.Minute {
					continue
				}
				r.send()
				r.lastSent = t
			}
		}
	}
}

func (r *DailyReporter) send() {
	data := r.buildReport()

	msg := r.formatReport(data)

	if r.dryRun {
		jxlog.Info("DRY-RUN", "Rapport quotidien simulé — non envoyé")
		jxlog.Info("DRY-RUN", msg)
		return
	}

	// Envoie via le dispatcher comme une alerte spéciale
	alert := Alert{
		Timestamp: time.Now(),
		Level:     AlertWarning, // level neutre pour le formatting
		Metric:    "DAILY REPORT",
		Message:   msg,
	}

	r.dispatcher.Dispatch(alert)
	jxlog.Info("DAILY REPORT", "Rapport envoyé")
}

func (r *DailyReporter) buildReport() ReportData {
	cycles := r.state.LastCycles(5760) // max 24h à 15s d'intervalle

	data := ReportData{
		Date:    time.Now().Format("02 January 2006"),
		Actions: make(map[string]int),
	}

	if len(cycles) == 0 {
		return data
	}

	data.TotalCycles = len(cycles)

	var cpuSum, ramSum float64

	for _, c := range cycles {
		data.Actions[c.Action]++

		snap := c.Snapshot
		cpuSum += snap.CPUPercent
		ramSum += snap.MemPercent

		if snap.CPUPercent > data.CPUMax {
			data.CPUMax = snap.CPUPercent
		}
		if snap.MemPercent > data.RAMMax {
			data.RAMMax = snap.MemPercent
		}
		if snap.DiskPercent > data.DiskMax {
			data.DiskMax = snap.DiskPercent
		}
	}

	data.CPUAvg = cpuSum / float64(data.TotalCycles)
	data.RAMAvg = ramSum / float64(data.TotalCycles)
	data.AlertCount = data.Actions["alert"]
	data.ExecuteCount = data.Actions["execute"]

	return data
}

func (r *DailyReporter) formatReport(d ReportData) string {
	return fmt.Sprintf(
		"JARVINx — Rapport du %s\n\n"+
			"Cycles        : %d\n"+
			"Actions       : %d log · %d suggest · %d alert · %d execute\n\n"+
			"CPU           : moy %.1f%% · pic %.1f%%\n"+
			"RAM           : moy %.1f%% · pic %.1f%%\n"+
			"Disk (max)    : %.1f%%\n\n"+
			"Alertes       : %d\n"+
			"Commands     : %d",
		d.Date,
		d.TotalCycles,
		d.Actions["log"],
		d.Actions["suggest"],
		d.AlertCount,
		d.ExecuteCount,
		d.CPUAvg, d.CPUMax,
		d.RAMAvg, d.RAMMax,
		d.DiskMax,
		d.AlertCount,
		d.ExecuteCount,
	)
}

func (r *DailyReporter) LastSent() time.Time {
	return r.lastSent
}

func (r *DailyReporter) SendNow() {
	r.send()
	r.lastSent = time.Now()
}
