'use client'

import { useLLMContext } from '@/lib/hooks'
import { cn } from '@/lib/utils'
import { Brain, TrendingUp, TrendingDown, ArrowRight, AlertTriangle } from 'lucide-react'

function TrendBadge({ trend }: { trend: string }) {
    const t = trend.toLowerCase()
    if (t.includes('critique') || t.includes('critical') || t === 'high') {
        return (
            <span className="flex items-center gap-1 px-2 py-0.5 rounded border bg-red-500/15 border-red-500/30 text-red-400 font-mono text-[10px] font-semibold">
                <TrendingUp size={10} />{trend || '—'}
            </span>
        )
    }
    if (t.includes('hausse') || t.includes('rising') || t.includes('élevé')) {
        return (
            <span className="flex items-center gap-1 px-2 py-0.5 rounded border bg-amber-500/15 border-amber-500/30 text-amber-400 font-mono text-[10px] font-semibold">
                <TrendingUp size={10} />{trend || '—'}
            </span>
        )
    }
    if (t.includes('baisse') || t.includes('falling')) {
        return (
            <span className="flex items-center gap-1 px-2 py-0.5 rounded border bg-sky-500/15 border-sky-500/30 text-sky-400 font-mono text-[10px] font-semibold">
                <TrendingDown size={10} />{trend || '—'}
            </span>
        )
    }
    if (t.includes('stable')) {
        return (
            <span className="flex items-center gap-1 px-2 py-0.5 rounded border bg-emerald-500/15 border-emerald-500/30 text-emerald-400 font-mono text-[10px] font-semibold">
                <ArrowRight size={10} />{trend || '—'}
            </span>
        )
    }
    return (
        <span className="flex items-center gap-1 px-2 py-0.5 rounded border bg-gray-500/15 border-gray-500/30 text-gray-400 font-mono text-[10px]">
            <ArrowRight size={10} />{trend || '—'}
        </span>
    )
}

function ActionBadge({ action }: { action: string }) {
    const colors: Record<string, string> = {
        log: 'bg-emerald-500/10 border-emerald-500/20 text-emerald-400',
        suggest: 'bg-amber-500/10 border-amber-500/20 text-amber-400',
        alert: 'bg-red-500/10 border-red-500/20 text-red-400',
        execute: 'bg-blue-500/10 border-blue-500/20 text-blue-400',
    }
    return (
        <span className={cn(
            'px-2 py-0.5 rounded border font-mono text-[10px]',
            colors[action] ?? 'bg-gray-500/10 border-gray-500/20 text-gray-400'
        )}>
            {action || '—'}
        </span>
    )
}

export default function LLMContextPage() {
    const { data } = useLLMContext()

    const alertRateColor =
        data.alert_rate > 30 ? 'text-red-400' :
            data.alert_rate > 10 ? 'text-amber-400' :
                'text-emerald-400'

    return (
        <div className="space-y-6 max-w-[1400px]">

            {/* Header */}
            <div>
                <div className="font-mono text-[10px] text-gray-600 tracking-widest uppercase mb-1">
                    LLM Context · Prompt Adaptatif
                </div>
                <h1 className="text-xl font-semibold text-white">Contexte LLM</h1>
            </div>

            {data.cycle_count === 0 && (
                <div className="bg-bg-secondary border border-border rounded-xl p-8 text-center">
                    <Brain size={32} className="text-gray-700 mx-auto mb-3" />
                    <div className="font-mono text-sm text-gray-600">
                        Aucun cycle analysé — le contexte apparaîtra après le premier cycle
                    </div>
                </div>
            )}

            {data.cycle_count > 0 && (
                <div className="grid grid-cols-2 gap-6">

                    {/* Colonne gauche */}
                    <div className="space-y-4">

                        {/* Vue globale */}
                        <div className="bg-bg-secondary border border-border rounded-xl p-5">
                            <div className="font-mono text-[10px] text-gray-600 uppercase tracking-widest mb-4">
                                Vue globale
                            </div>
                            <div className="space-y-3">
                                <div className="flex items-center justify-between">
                                    <span className="font-mono text-xs text-gray-400">Cycles analysés</span>
                                    <span className="font-mono text-xs text-white font-semibold">
                                        {data.cycle_count}
                                    </span>
                                </div>
                                <div className="flex items-center justify-between">
                                    <span className="font-mono text-xs text-gray-400">Action dominante</span>
                                    <ActionBadge action={data.dominant_action} />
                                </div>
                                <div className="flex items-center justify-between">
                                    <span className="font-mono text-xs text-gray-400">Taux d&apos;alerte</span>
                                    <span className={cn('font-mono text-xs font-semibold', alertRateColor)}>
                                        {data.alert_rate.toFixed(1)}%
                                    </span>
                                </div>
                            </div>

                            {/* Alert rate bar */}
                            <div className="mt-4">
                                <div className="h-1.5 bg-bg-tertiary rounded-full overflow-hidden">
                                    <div
                                        className={cn(
                                            'h-full rounded-full transition-all duration-700',
                                            data.alert_rate > 30 ? 'bg-red-500' :
                                                data.alert_rate > 10 ? 'bg-amber-500' : 'bg-emerald-500'
                                        )}
                                        style={{ width: `${Math.min(data.alert_rate, 100)}%` }}
                                    />
                                </div>
                                <div className="flex justify-between mt-1">
                                    <span className="font-mono text-[9px] text-gray-700">0%</span>
                                    <span className="font-mono text-[9px] text-gray-700">100%</span>
                                </div>
                            </div>
                        </div>

                        {/* Tendances métriques */}
                        <div className="bg-bg-secondary border border-border rounded-xl p-5">
                            <div className="font-mono text-[10px] text-gray-600 uppercase tracking-widest mb-4">
                                Tendances observées
                            </div>
                            <div className="space-y-3">
                                {[
                                    { label: 'CPU', trend: data.cpu_trend },
                                    { label: 'RAM', trend: data.ram_trend },
                                    { label: 'Disk', trend: data.disk_trend },
                                ].map(({ label, trend }) => (
                                    <div key={label} className="flex items-center justify-between">
                                        <span className="font-mono text-xs text-gray-500 w-8">{label}</span>
                                        <TrendBadge trend={trend} />
                                    </div>
                                ))}
                            </div>
                        </div>
                    </div>

                    {/* Colonne droite */}
                    <div className="space-y-4">

                        {/* Alertes récentes */}
                        <div className="bg-bg-secondary border border-border rounded-xl p-5">
                            <div className="font-mono text-[10px] text-gray-600 uppercase tracking-widest mb-4">
                                Alertes récentes dans le contexte
                            </div>

                            {data.recent_alerts.length === 0 ? (
                                <div className="flex items-center gap-2 py-4 justify-center">
                                    <span className="font-mono text-xs text-gray-600">
                                        Aucune alerte récente — système stable
                                    </span>
                                </div>
                            ) : (
                                <div className="space-y-2">
                                    {data.recent_alerts.map((alert, i) => (
                                        <div
                                            key={i}
                                            className="flex items-start gap-2 bg-red-500/5 border border-red-500/10 rounded-lg p-3"
                                        >
                                            <AlertTriangle size={11} className="text-red-400 flex-shrink-0 mt-0.5" />
                                            <span className="font-mono text-[10px] text-red-300 leading-relaxed">
                                                {alert}
                                            </span>
                                        </div>
                                    ))}
                                </div>
                            )}
                        </div>

                        {/* Prompt adaptatif info */}
                        <div className="bg-bg-secondary border border-border rounded-xl p-5">
                            <div className="font-mono text-[10px] text-gray-600 uppercase tracking-widest mb-4">
                                Prompt adaptatif
                            </div>
                            <div className="space-y-2">
                                <div className="font-mono text-[10px] text-gray-500 leading-relaxed">
                                    Le prompt système est enrichi automatiquement à chaque cycle avec ce contexte.
                                    L&apos;agent LLM reçoit les tendances et l&apos;historique d&apos;alertes pour affiner ses décisions.
                                </div>
                                <div className="mt-3 pt-3 border-t border-border">
                                    <div className="flex items-center justify-between">
                                        <span className="font-mono text-[10px] text-gray-600">Snapshots analysés</span>
                                        <span className="font-mono text-[10px] text-gray-400">10 derniers</span>
                                    </div>
                                    <div className="flex items-center justify-between mt-1.5">
                                        <span className="font-mono text-[10px] text-gray-600">Cycles analysés</span>
                                        <span className="font-mono text-[10px] text-gray-400">20 derniers</span>
                                    </div>
                                    <div className="flex items-center justify-between mt-1.5">
                                        <span className="font-mono text-[10px] text-gray-600">Refresh</span>
                                        <span className="font-mono text-[10px] text-gray-400">15s</span>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            )}
        </div>
    )
}