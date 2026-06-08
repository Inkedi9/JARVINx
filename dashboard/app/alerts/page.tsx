'use client'

import { useState, useMemo } from 'react'
import { useAlerts } from '@/lib/hooks'
import { AlertEntry } from '@/lib/api'
import { cn } from '@/lib/utils'
import { Bell, AlertTriangle, AlertOctagon, RefreshCw } from 'lucide-react'

type LevelFilter = 'all' | 'warning' | 'critical'

const LEVEL_META = {
    warning: {
        label: 'WARNING',
        icon: AlertTriangle,
        row: 'border-amber-500/20 bg-amber-500/5',
        badge: 'bg-amber-500/15 text-amber-400 border-amber-500/30',
    },
    critical: {
        label: 'CRITICAL',
        icon: AlertOctagon,
        row: 'border-red-500/20 bg-red-500/5',
        badge: 'bg-red-500/15 text-red-400 border-red-500/30',
    },
}

function AlertRow({ alert }: { alert: AlertEntry }) {
    const meta = LEVEL_META[alert.level]
    const Icon = meta.icon
    const ts = new Date(alert.timestamp)
    const date = ts.toLocaleDateString('fr-FR', { day: '2-digit', month: 'short' })
    const time = ts.toLocaleTimeString('fr-FR', { hour: '2-digit', minute: '2-digit', second: '2-digit' })

    return (
        <div className={cn('flex items-start gap-4 px-4 py-3 border rounded-lg', meta.row)}>
            <Icon size={15} className={alert.level === 'critical' ? 'text-red-400 mt-0.5 shrink-0' : 'text-amber-400 mt-0.5 shrink-0'} />

            <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2 flex-wrap">
                    <span className={cn('font-mono text-[10px] font-semibold px-1.5 py-0.5 rounded border', meta.badge)}>
                        {meta.label}
                    </span>
                    <span className="font-mono text-xs font-bold text-white">{alert.metric}</span>
                    <span className="font-mono text-xs text-gray-400">
                        {alert.value.toFixed(1)}% / {alert.threshold.toFixed(0)}% threshold
                    </span>
                    {alert.cycles_above > 0 && (
                        <span className="font-mono text-[10px] text-gray-600">
                            {alert.cycles_above} cycle{alert.cycles_above > 1 ? 's' : ''}
                        </span>
                    )}
                </div>
                <p className="text-sm text-gray-300 mt-1 truncate">{alert.message}</p>
            </div>

            <div className="font-mono text-[10px] text-gray-500 text-right shrink-0">
                <div>{date}</div>
                <div>{time}</div>
            </div>
        </div>
    )
}

export default function AlertsPage() {
    const { data, error } = useAlerts()
    const [filter, setFilter] = useState<LevelFilter>('all')

    const counts = useMemo(() => ({
        all: data.total,
        warning: data.alerts.filter(a => a.level === 'warning').length,
        critical: data.alerts.filter(a => a.level === 'critical').length,
    }), [data])

    const visible = useMemo(() =>
        filter === 'all' ? data.alerts : data.alerts.filter(a => a.level === filter),
        [data.alerts, filter]
    )

    const filters: { key: LevelFilter; label: string; count: number }[] = [
        { key: 'all', label: 'All', count: counts.all },
        { key: 'warning', label: 'Warning', count: counts.warning },
        { key: 'critical', label: 'Critical', count: counts.critical },
    ]

    return (
        <div className="p-6 space-y-6">

            {/* Header */}
            <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                    <Bell size={20} className="text-accent-blue" />
                    <div>
                        <h1 className="font-mono text-lg font-semibold text-white">Alerts</h1>
                        <p className="text-xs text-gray-500 font-mono">{data.total} total · alerts.jsonl</p>
                    </div>
                </div>
                {error && (
                    <div className="flex items-center gap-1.5 text-red-400 font-mono text-xs">
                        <RefreshCw size={12} />
                        <span>Runtime unreachable</span>
                    </div>
                )}
            </div>

            {/* Filter pills */}
            <div className="flex items-center gap-2">
                {filters.map(({ key, label, count }) => (
                    <button
                        key={key}
                        onClick={() => setFilter(key)}
                        className={cn(
                            'flex items-center gap-1.5 px-3 py-1.5 rounded-lg border font-mono text-xs transition-all',
                            filter === key
                                ? key === 'critical'
                                    ? 'bg-red-500/15 border-red-500/30 text-red-400'
                                    : key === 'warning'
                                        ? 'bg-amber-500/15 border-amber-500/30 text-amber-400'
                                        : 'bg-accent-blue/15 border-accent-blue/30 text-accent-blue'
                                : 'bg-bg-tertiary border-border text-gray-400 hover:text-white'
                        )}
                    >
                        {label}
                        <span className={cn(
                            'px-1 rounded text-[10px]',
                            filter === key ? 'bg-white/10' : 'bg-bg-secondary'
                        )}>
                            {count}
                        </span>
                    </button>
                ))}
            </div>

            {/* Alert list */}
            {visible.length === 0 ? (
                <div className="flex flex-col items-center justify-center py-16 text-gray-600">
                    <Bell size={32} className="mb-3 opacity-30" />
                    <p className="font-mono text-sm">No alerts</p>
                    {filter !== 'all' && (
                        <p className="font-mono text-xs mt-1">for level &ldquo;{filter}&rdquo;</p>
                    )}
                </div>
            ) : (
                <div className="space-y-2">
                    {visible.map((alert, i) => (
                        <AlertRow key={`${alert.timestamp}-${i}`} alert={alert} />
                    ))}
                </div>
            )}
        </div>
    )
}
