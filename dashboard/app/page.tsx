'use client'

import { useStatus, useHistory, useAgents, useLLMContext } from '@/lib/hooks'
import { TrendingUp, CheckCircle, XCircle, Clock } from 'lucide-react'
import { cn } from '@/lib/utils'
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
  const { data: llmCtx } = useLLMContext()

  const hasForecasts = !!(llmCtx.cpu_forecast || llmCtx.ram_forecast || llmCtx.disk_forecast)

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
      {/* Forecast banner */}
      {hasForecasts && (
        <div className="bg-amber-500/5 border border-amber-500/15 rounded-xl px-5 py-3 flex items-center gap-6 flex-wrap">
          <div className="flex items-center gap-2 shrink-0">
            <TrendingUp size={13} className="text-amber-400" />
            <span className="font-mono text-[10px] text-amber-400 uppercase tracking-widest">Forecast</span>
          </div>
          {llmCtx.cpu_forecast && (
            <span className="font-mono text-xs text-gray-300">
              <span className="text-blue-400 mr-1">CPU</span>{llmCtx.cpu_forecast}
            </span>
          )}
          {llmCtx.ram_forecast && (
            <span className="font-mono text-xs text-gray-300">
              <span className="text-amber-400 mr-1">RAM</span>{llmCtx.ram_forecast}
            </span>
          )}
          {llmCtx.disk_forecast && (
            <span className="font-mono text-xs text-gray-300">
              <span className="text-gray-400 mr-1">Disk</span>{llmCtx.disk_forecast}
            </span>
          )}
        </div>
      )}

      {/* Execute Guard */}
      {status?.exec_guard && (
        <div className="font-mono text-[10px] flex items-center gap-1.5">
          {status.exec_guard.cooldown_remaining_seconds > 0 ? (
            <span className="text-amber-400">
              ⏸ Cooldown actif — {status.exec_guard.last_cmd} — reprend dans {Math.ceil(status.exec_guard.cooldown_remaining_seconds)}s
            </span>
          ) : (
            <span className="text-emerald-600">✅ Execute disponible</span>
          )}
        </div>
      )}

      {/* Terminal — last execute result */}
      {status?.last_exec_result && (() => {
        const r = status.last_exec_result
        return (
          <div className="bg-[#0d1117] border border-border rounded-xl overflow-hidden font-mono text-xs">
            <div className="flex items-center justify-between px-4 py-2 border-b border-border bg-bg-secondary">
              <div className="flex items-center gap-2">
                {r.success
                  ? <CheckCircle size={12} className="text-emerald-400" />
                  : <XCircle size={12} className="text-red-400" />}
                <span className="text-[10px] uppercase tracking-widest text-gray-500">Last Execute</span>
                <span className={cn('text-[10px]', r.success ? 'text-emerald-400' : 'text-red-400')}>
                  {r.command}
                </span>
              </div>
              <div className="flex items-center gap-1 text-gray-600">
                <Clock size={10} />
                <span className="text-[10px]">{r.duration_ms.toFixed(0)}ms</span>
                {r.timed_out && <span className="text-amber-400 ml-1">TIMEOUT</span>}
              </div>
            </div>
            <div className="px-4 py-3 max-h-36 overflow-y-auto">
              <span className="text-gray-600 select-none">{'> '}</span>
              <span className="text-gray-400">{r.command}</span>
              {r.output && (
                <pre className="mt-1.5 text-emerald-300 whitespace-pre-wrap break-all leading-relaxed text-[11px]">
                  {r.output}
                </pre>
              )}
              {r.error && (
                <pre className="mt-1.5 text-red-400 whitespace-pre-wrap break-all leading-relaxed text-[11px]">
                  {r.error}
                </pre>
              )}
            </div>
          </div>
        )
      })()}

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