import { cn, metricColor, metricBg } from '@/lib/utils'

interface MetricsBarProps {
    label: string
    value: number
    unit?: string
    detail?: string
    warn?: number
    crit?: number
}

export default function MetricsBar({
    label,
    value,
    unit = '%',
    detail,
    warn = 70,
    crit = 85,
}: MetricsBarProps) {
    return (
        <div className="flex items-center gap-4 py-2 border-b border-border last:border-0">
            {/* Label */}
            <div className="font-mono text-[10px] text-gray-500 w-16 flex-shrink-0 uppercase tracking-wider">
                {label}
            </div>

            {/* Bar */}
            <div className="flex-1 h-1 bg-bg-tertiary rounded-full overflow-hidden">
                <div
                    className={cn('h-full rounded-full transition-all duration-700', metricBg(value, warn, crit))}
                    style={{ width: `${Math.min(value, 100)}%` }}
                />
            </div>

            {/* Value */}
            <div className={cn('font-mono text-xs font-semibold w-12 text-right', metricColor(value, warn, crit))}>
                {value.toFixed(1)}{unit}
            </div>

            {/* Detail */}
            {detail && (
                <div className="font-mono text-[10px] text-gray-600 w-28 text-right truncate">
                    {detail}
                </div>
            )}
        </div>
    )
}