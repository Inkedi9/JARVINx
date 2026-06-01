'use client'

import { useHistory } from '@/lib/hooks'
import { formatTime, actionColor, cn } from '@/lib/utils'
import { Terminal, Clock, TrendingUp, Activity } from 'lucide-react'

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

            {/* Table */}
            <div className="bg-bg-secondary border border-border rounded-xl overflow-hidden">

                {/* Table header */}
                <div className="grid grid-cols-[80px_100px_80px_80px_80px_1fr_140px] gap-4 px-5 py-3 border-b border-border bg-bg-tertiary">
                    {['Cycle', 'Time', 'CPU', 'RAM', 'Disk', 'Analysis', 'Action'].map(h => (
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
                                className="grid grid-cols-[80px_100px_80px_80px_80px_1fr_140px] gap-4 px-5 py-3.5 hover:bg-bg-tertiary transition-all group"
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