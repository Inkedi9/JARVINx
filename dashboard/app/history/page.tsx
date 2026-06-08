'use client'

import { useState } from 'react'
import { useHistory, useHistoryFull } from '@/lib/hooks'
import { formatTime, actionColor, cn } from '@/lib/utils'
import { Terminal, Clock, TrendingUp, Activity, BarChart2 } from 'lucide-react'
import { MetricsChart } from '@/app/components/metrics-chart'
import { CycleRecord } from '@/lib/api'

const SPARK_W = 104
const SPARK_H = 28

function Sparkline({ cycles, highlightNum }: { cycles: CycleRecord[]; highlightNum: number }) {
    if (cycles.length < 2) return <span className="block w-[104px]" />

    const ordered = [...cycles].reverse() // oldest → newest

    const toPath = (vals: number[]) =>
        vals.map((v, i) => {
            const x = (i / (ordered.length - 1)) * SPARK_W
            const y = SPARK_H - (v / 100) * SPARK_H
            return `${i === 0 ? 'M' : 'L'}${x.toFixed(1)},${y.toFixed(1)}`
        }).join(' ')

    const cpuPath = toPath(ordered.map(c => c.snapshot.cpu_percent))
    const ramPath = toPath(ordered.map(c => c.snapshot.mem_percent))

    const hiIdx = ordered.findIndex(c => c.cycle_num === highlightNum)
    const hi = hiIdx >= 0 ? ordered[hiIdx] : null
    const hiX = hi ? (hiIdx / (ordered.length - 1)) * SPARK_W : null
    const hiCpuY = hi ? SPARK_H - (hi.snapshot.cpu_percent / 100) * SPARK_H : null
    const hiRamY = hi ? SPARK_H - (hi.snapshot.mem_percent / 100) * SPARK_H : null

    return (
        <svg width={SPARK_W} height={SPARK_H} className="overflow-visible">
            <path d={ramPath} fill="none" stroke="#f59e0b" strokeWidth="1.2" strokeLinejoin="round" opacity="0.5" />
            <path d={cpuPath} fill="none" stroke="#60a5fa" strokeWidth="1.5" strokeLinejoin="round" opacity="0.85" />
            {hiX !== null && hiCpuY !== null && (
                <circle cx={hiX} cy={hiCpuY} r="2.5" fill="#60a5fa" />
            )}
            {hiX !== null && hiRamY !== null && (
                <circle cx={hiX} cy={hiRamY} r="2" fill="#f59e0b" />
            )}
        </svg>
    )
}

function ActionBadge({ action }: { action: string }) {
    return (
        <span className={cn(
            'font-mono text-[9px] px-2 py-0.5 rounded border uppercase tracking-wider',
            actionColor(action)
        )}>
            {action}
        </span>
    )
}

function ConfidenceBadge({ confidence }: { confidence: number }) {
    const pct = Math.round(confidence * 100)
    const color = confidence >= 0.75
        ? 'text-emerald-400 border-emerald-400/30 bg-emerald-400/5'
        : confidence >= 0.5
            ? 'text-amber-400 border-amber-400/30 bg-amber-400/5'
            : 'text-red-400 border-red-400/30 bg-red-400/5'
    return (
        <span className={cn('font-mono text-[9px] px-1.5 py-0.5 rounded border', color)}>
            Confiance {pct}%
        </span>
    )
}

function StatPill({
    label,
    value,
    color = 'text-white'
}: {
    label: string
    value: string | number
    color?: string
}) {
    return (
        <div className="bg-bg-tertiary border border-border rounded-lg px-3 py-2 text-center">
            <div className={cn('font-mono text-sm font-semibold', color)}>{value}</div>
            <div className="font-mono text-[9px] text-gray-600 uppercase tracking-wider mt-0.5">{label}</div>
        </div>
    )
}

export default function HistoryPage() {
    const { data: history } = useHistory()
    const [chartRange, setChartRange] = useState<'7d' | '30d' | '90d'>('7d')
    const historyFull = useHistoryFull(chartRange)

    // Stats globales
    const counts = history.cycles.reduce((acc, c) => {
        acc[c.action] = (acc[c.action] ?? 0) + 1
        return acc
    }, {} as Record<string, number>)

    const withCommand = history.cycles.filter(c => c.command).length

    return (
        <div className="space-y-6 max-w-[1400px]">

            {/* Header */}
            <div>
                <div className="font-mono text-[10px] text-gray-600 tracking-widest uppercase mb-1">
                    History · Cycle Log
                </div>
                <h1 className="text-xl font-semibold text-white">Cycle History</h1>
            </div>

            {/* Stats */}
            <div className="grid grid-cols-6 gap-3">
                <StatPill label="Total" value={history.total} />
                <StatPill label="Log" value={counts.log ?? 0} color="text-emerald-400" />
                <StatPill label="Suggest" value={counts.suggest ?? 0} color="text-amber-400" />
                <StatPill label="Alert" value={counts.alert ?? 0} color="text-red-400" />
                <StatPill label="Execute" value={counts.execute ?? 0} color="text-blue-400" />
                <StatPill label="Actions" value={withCommand} color="text-accent-blue" />
            </div>

            {/* Graphes temporels */}
            <div className="bg-bg-secondary border border-border rounded-xl p-5">
                <div className="flex items-center justify-between mb-4">
                    <div className="flex items-center gap-2">
                        <BarChart2 size={14} className="text-accent-blue" />
                        <span className="font-mono text-xs text-gray-400">Métriques système — moyennes par période</span>
                        {historyFull.available && historyFull.total_snapshots > 0 && (
                            <span className="font-mono text-[9px] text-gray-600">
                                · {historyFull.total_snapshots.toLocaleString('fr-FR')} snapshots
                            </span>
                        )}
                    </div>
                    <div className="flex gap-1">
                        {(['7d', '30d', '90d'] as const).map(r => (
                            <button
                                key={r}
                                onClick={() => setChartRange(r)}
                                className={cn(
                                    'font-mono text-[9px] px-2 py-1 rounded border uppercase tracking-wider transition-colors cursor-pointer',
                                    chartRange === r
                                        ? 'border-accent-blue text-accent-blue bg-accent-blue/10'
                                        : 'border-border text-gray-500 hover:border-gray-500'
                                )}
                            >
                                {r}
                            </button>
                        ))}
                    </div>
                </div>

                {!historyFull.available ? (
                    <div className="flex flex-col items-center justify-center h-48 gap-2 text-gray-700">
                        <BarChart2 size={28} />
                        <div className="font-mono text-xs text-gray-600">SQLite non configuré</div>
                        <div className="font-mono text-[10px] text-gray-700">
                            Active <span className="text-gray-500">JARVINX_SQLITE_PATH=jarvinx.db</span> pour les graphes historiques
                        </div>
                    </div>
                ) : (
                    <MetricsChart buckets={historyFull.buckets} bucketHours={historyFull.bucket_hours} />
                )}
            </div>

            {/* Table */}
            <div className="bg-bg-secondary border border-border rounded-xl overflow-hidden">

                {/* Table header */}
                <div className="grid grid-cols-[80px_100px_60px_60px_60px_120px_1fr_140px] gap-4 px-5 py-3 border-b border-border bg-bg-tertiary">
                    {['Cycle', 'Time', 'CPU', 'RAM', 'Disk', 'Trend', 'Analysis', 'Action'].map(h => (
                        <div key={h} className="font-mono text-[9px] text-gray-600 uppercase tracking-widest">
                            {h}
                        </div>
                    ))}
                </div>

                {/* Rows */}
                <div className="divide-y divide-border">
                    {history.cycles.length === 0 && (
                        <div className="flex flex-col items-center justify-center py-16 gap-3">
                            <Activity size={32} className="text-gray-700" />
                            <div className="font-mono text-sm text-gray-600">
                                En attente des premiers cycles...
                            </div>
                        </div>
                    )}

                    {history.cycles.map((cycle) => {
                        const snap = cycle.snapshot
                        const cpuColor = snap.cpu_percent >= 85 ? 'text-red-400'
                            : snap.cpu_percent >= 70 ? 'text-amber-400'
                                : 'text-emerald-400'
                        const ramColor = snap.mem_percent >= 90 ? 'text-red-400'
                            : snap.mem_percent >= 70 ? 'text-amber-400'
                                : 'text-emerald-400'
                        const diskColor = snap.disk_percent >= 85 ? 'text-red-400'
                            : snap.disk_percent >= 70 ? 'text-amber-400'
                                : 'text-emerald-400'

                        return (
                            <div
                                key={cycle.cycle_num}
                                className="grid grid-cols-[80px_100px_60px_60px_60px_120px_1fr_140px] gap-4 px-5 py-3.5 hover:bg-bg-tertiary transition-all group"
                            >
                                {/* Cycle num */}
                                <div className="flex items-center">
                                    <span className="font-mono text-xs text-accent-blue">
                                        #{cycle.cycle_num}
                                    </span>
                                </div>

                                {/* Time */}
                                <div className="flex items-center gap-1.5">
                                    <Clock size={10} className="text-gray-600" />
                                    <span className="font-mono text-[11px] text-gray-400">
                                        {formatTime(cycle.timestamp)}
                                    </span>
                                </div>

                                {/* CPU */}
                                <div className="flex items-center">
                                    <span className={cn('font-mono text-xs font-semibold', cpuColor)}>
                                        {snap.cpu_percent.toFixed(1)}%
                                    </span>
                                </div>

                                {/* RAM */}
                                <div className="flex items-center">
                                    <span className={cn('font-mono text-xs font-semibold', ramColor)}>
                                        {snap.mem_percent.toFixed(0)}%
                                    </span>
                                </div>

                                {/* Disk */}
                                <div className="flex items-center">
                                    <span className={cn('font-mono text-xs font-semibold', diskColor)}>
                                        {snap.disk_percent.toFixed(1)}%
                                    </span>
                                </div>

                                {/* Sparkline */}
                                <div className="flex items-center">
                                    <Sparkline cycles={history.cycles} highlightNum={cycle.cycle_num} />
                                </div>

                                {/* Analysis */}
                                <div className="flex flex-col justify-center min-w-0">
                                    <div className="font-mono text-[11px] text-gray-300 truncate">
                                        {cycle.analysis}
                                    </div>
                                    {cycle.command && (
                                        <div className="flex items-center gap-1 mt-0.5">
                                            <Terminal size={9} className="text-amber-400 flex-shrink-0" />
                                            <span className="font-mono text-[9px] text-amber-400 truncate">
                                                {cycle.command}
                                            </span>
                                        </div>
                                    )}
                                    {(cycle.trigger_cpu || cycle.trigger_ram || cycle.trigger_disk) && (
                                        <div className="font-mono text-[9px] text-gray-600 mt-0.5">
                                            {`Déclenché à ${[
                                                cycle.trigger_cpu ? `CPU ${cycle.trigger_cpu.toFixed(0)}%` : null,
                                                cycle.trigger_ram ? `RAM ${cycle.trigger_ram.toFixed(0)}%` : null,
                                                cycle.trigger_disk ? `Disk ${cycle.trigger_disk.toFixed(0)}%` : null,
                                            ].filter(Boolean).join(' / ')}`}
                                        </div>
                                    )}
                                </div>

                                {/* Action */}
                                <div className="flex flex-col items-end gap-1">
                                    <ActionBadge action={cycle.action} />
                                    {(cycle.action === 'execute' || cycle.action === 'suggest') && cycle.confidence !== undefined && (
                                        <ConfidenceBadge confidence={cycle.confidence} />
                                    )}
                                </div>
                            </div>
                        )
                    })}
                </div>
            </div>

            {/* Note */}
            <div className="flex items-center gap-2 text-gray-600">
                <TrendingUp size={12} />
                <span className="font-mono text-[10px]">
                    Affiche les {history.cycles.length} derniers cycles · mis à jour toutes les 15s
                </span>
            </div>

        </div>
    )
}