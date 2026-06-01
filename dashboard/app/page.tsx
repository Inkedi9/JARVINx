'use client'

import { useStatus, useHistory, useAgents } from '@/lib/hooks'
import StatCard from './components/ui/stat-card'
import RuntimeCycle from './components/runtime-cycle'
import AgentList from './components/agent-list'
import DecisionFeed from './components/decision-feed'
import MetricsBar from './components/metrics-bar'
import AIAnalysis from './components/ai-analysis'
import DailyReporter from './components/daily-reporter'

export default function Overview() {
  const { data: status } = useStatus()
  const { data: history } = useHistory()
  const { data: agents } = useAgents()

  const lastCycle = status?.last_cycle
  const snap = lastCycle?.snapshot

  // Détermine l'étape active selon la dernière action
  const activeStep = lastCycle
    ? lastCycle.action === 'execute' ? 4
      : lastCycle.action === 'alert' ? 3
        : 2
    : 1

  // Health globale — ratio cycles sans alerte
  const alertCycles = history.cycles.filter(c => c.action === 'alert').length
  const healthPct = history.cycles.length > 0
    ? Math.round((1 - alertCycles / history.cycles.length) * 100)
    : 100

  const activeAgents = agents.agents.filter(a => a.enabled).length

  return (
    <div className="space-y-4 max-w-[1400px]">

      {/* Section label */}
      <div className="font-mono text-[10px] text-gray-600 tracking-widest uppercase">
        Overview · JARVINx Runtime
      </div>

      {/* Stat cards */}
      <div className="grid grid-cols-4 gap-4">
        <StatCard
          label="System Health"
          value={`${healthPct}%`}
          sub="vs last 10 cycles"
          accent={healthPct < 80}
        />

        <StatCard
          label="Active Agents"
          value={activeAgents}
          sub={`${agents.total} total`}
        >
          <div className="flex gap-3 font-mono text-[10px]">
            <span className="text-emerald-400">● {activeAgents} Running</span>
            <span className="text-gray-600">
              ● {agents.total - activeAgents} Sleeping
            </span>
          </div>
        </StatCard>

        <StatCard
          label="Decisions Today"
          value={history.total}
          sub="cycles enregistrés"
        />

        <StatCard
          label="Interval"
          value={status?.interval ?? '—'}
          sub={`Cycle #${status?.cycle_num ?? 0}`}
          accent
        />
      </div>
      <AIAnalysis />

      {/* Runtime cycle + Agents */}
      <div className="grid grid-cols-[1fr_340px] gap-4">
        <RuntimeCycle activeStep={activeStep} />
        <AgentList />
      </div>

      {/* Decisions + Metrics */}
      <div className="grid grid-cols-[1fr_340px] gap-4">
        <DecisionFeed />

        {/* System Metrics */}
        <div className="bg-bg-secondary border border-border rounded-xl p-5">
          <div className="flex items-center justify-between mb-4">
            <span className="font-mono text-[10px] text-gray-500 tracking-widest uppercase">
              System Metrics
            </span>
            <span className="font-mono text-[10px] text-emerald-400">Live</span>
          </div>

          {snap ? (
            <div>
              <MetricsBar
                label="CPU"
                value={snap.cpu_percent}
                warn={70} crit={85}
              />
              <MetricsBar
                label="Memory"
                value={snap.mem_percent}
                detail={`${(snap.mem_used_mb / 1024).toFixed(1)} / ${(snap.mem_total_mb / 1024).toFixed(1)} GB`}
                warn={70} crit={90}
              />
              <MetricsBar
                label="Disk"
                value={snap.disk_percent}
                detail={`${snap.disk_used_gb} / ${snap.disk_total_gb} GB`}
                warn={70} crit={85}
              />
            </div>
          ) : (
            <div className="font-mono text-[11px] text-gray-600 text-center py-8">
              En attente des données...
            </div>
          )}
        </div>
        <DailyReporter />
      </div>

    </div>
  )
}