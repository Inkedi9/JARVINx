import { cn } from '@/lib/utils'
import { TrendingUp } from 'lucide-react'

interface StatCardProps {
    label: string
    value: string | number
    sub?: string
    trend?: string
    accent?: boolean
    children?: React.ReactNode
}

export default function StatCard({
    label,
    value,
    sub,
    trend,
    accent,
    children,
}: StatCardProps) {
    return (
        <div className="bg-bg-secondary border border-border rounded-xl p-5 flex flex-col gap-3">
            <div className="font-mono text-[10px] text-gray-500 tracking-widest uppercase">
                {label}
            </div>
            <div className={cn(
                'font-mono text-3xl font-semibold leading-none',
                accent ? 'text-accent-blue' : 'text-white'
            )}>
                {value}
            </div>
            {children}
            {(sub || trend) && (
                <div className="flex items-center gap-2 mt-auto">
                    {trend && (
                        <div className="flex items-center gap-1 text-emerald-400">
                            <TrendingUp size={12} />
                            <span className="font-mono text-[10px]">{trend}</span>
                        </div>
                    )}
                    {sub && (
                        <span className="font-mono text-[10px] text-gray-600">{sub}</span>
                    )}
                </div>
            )}
        </div>
    )
}