'use client'

import { useEffect, useState } from 'react'
import { api, LLMContextResponse } from '@/lib/api'
import { cn } from '@/lib/utils'
import { Brain, TrendingUp, TrendingDown, Minus, AlertTriangle, CheckCircle } from 'lucide-react'

function useLLMContext() {
    const [data, setData] = useState<LLMContextResponse | null>(null)

    useEffect(() => {
        const fetch = async () => {
            try {
                const result = await api.llmContext()
                setData(result)
            } catch { }
        }
        fetch()
        const id = setInterval(fetch, 15_000)
        return () => clearInterval(id)
    }, [])

    return { data }
}

function TrendPill({ label, trend }: { label: string; trend: string }) {
    const isHigh = trend.includes('critique') || trend.includes('élevé')
    const isRising = trend.includes('hausse')
    const isFalling = trend.includes('baisse')
    const isStable = trend.includes('stable')

    const color =
        trend.includes('critique') ? 'text-red-400' :
            isHigh || isRising ? 'text-amber-400' :
                isFalling || isStable ? 'text-emerald-400' :
                    'text-gray-400'

    const Icon =
        trend.includes('critique') || isRising ? TrendingUp :
            isFalling ? TrendingDown :
                Minus

    return (
        <div className="flex items-center gap-1.5 bg-bg-tertiary rounded-lg px-2.5 py-1.5">
            <span className="font-mono text-[9px] text-gray-600 uppercase">{label}</span>
            <Icon size={10} className={color} />
            <span className={cn('font-mono text-[10px]', color)}>
                {trend.split('(')[1]?.replace(')', '') ?? trend.split(' ').pop()}
            </span>
        </div>
    )
}

export default function AIAnalysis() {
    const { data } = useLLMContext()

    if (!data || data.cycle_count === 0) return null

    const isHealthy = data.alert_rate < 10
    const isWarning = data.alert_rate >= 10 && data.alert_rate < 30
    const isCritical = data.alert_rate >= 30

    const statusColor =
        isCritical ? 'border-red-500/30 bg-red-500/5' :
            isWarning ? 'border-amber-500/30 bg-amber-500/5' :
                'border-emerald-500/30 bg-emerald-500/5'

    const statusIcon = isCritical || isWarning
        ? <AlertTriangle size={14} className={isCritical ? 'text-red-400' : 'text-amber-400'} />
        : <CheckCircle size={14} className="text-emerald-400" />

    const statusText =
        isCritical ? `Système sous pression — ${data.alert_rate.toFixed(0)}% de cycles en alerte` :
            isWarning ? `Attention requise — ${data.alert_rate.toFixed(0)}% de cycles en alerte` :
                'Système stable — aucune anomalie détectée'

    return (
        <div className={cn(
            'border rounded-xl p-4 space-y-3',
            statusColor
        )}>
            {/* Header */}
            <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                    <Brain size={13} className="text-accent-blue" />
                    <span className="font-mono text-[10px] text-gray-500 uppercase tracking-widest">
                        Analyse IA
                    </span>
                </div>
                <span className="font-mono text-[9px] text-gray-600">
                    {data.cycle_count} cycles · action dominante{' '}
                    <span className="text-white">{data.dominant_action}</span>
                </span>
            </div>

            {/* Status */}
            <div className="flex items-center gap-2">
                {statusIcon}
                <span className="font-mono text-xs text-gray-300">{statusText}</span>
            </div>

            {/* Tendances */}
            <div className="flex items-center gap-2 flex-wrap">
                {data.cpu_trend && <TrendPill label="CPU" trend={data.cpu_trend} />}
                {data.ram_trend && <TrendPill label="RAM" trend={data.ram_trend} />}
                {data.disk_trend && <TrendPill label="Disk" trend={data.disk_trend} />}
            </div>

            {/* Alertes récentes */}
            {data.recent_alerts.length > 0 && (
                <div className="pt-2 border-t border-white/5 space-y-1">
                    {data.recent_alerts.slice(0, 2).map((alert, i) => (
                        <div key={i} className="flex items-start gap-1.5">
                            <AlertTriangle size={10} className="text-red-400 flex-shrink-0 mt-0.5" />
                            <span className="font-mono text-[10px] text-red-300 leading-relaxed truncate">
                                {alert}
                            </span>
                        </div>
                    ))}
                </div>
            )}
        </div>
    )
}