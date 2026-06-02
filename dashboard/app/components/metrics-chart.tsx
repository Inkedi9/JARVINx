'use client'

import {
    AreaChart,
    Area,
    XAxis,
    YAxis,
    CartesianGrid,
    Tooltip,
    ResponsiveContainer,
    Legend,
} from 'recharts'
import type { SnapshotBucket } from '@/lib/api'

interface MetricsChartProps {
    buckets: SnapshotBucket[]
    bucketHours: number
}

function formatBucketTime(ts: string, bucketHours: number): string {
    const d = new Date(ts)
    if (bucketHours === 1) {
        return d.toLocaleString('fr-FR', { weekday: 'short', hour: '2-digit', minute: '2-digit' })
    }
    return d.toLocaleString('fr-FR', { month: 'short', day: 'numeric' })
}

export function MetricsChart({ buckets, bucketHours }: MetricsChartProps) {
    if (buckets.length === 0) {
        return (
            <div className="flex items-center justify-center h-56 text-gray-600 font-mono text-xs">
                Pas encore assez de données historiques
            </div>
        )
    }

    const data = buckets.map(b => ({
        time: formatBucketTime(b.timestamp, bucketHours),
        cpu: Number(b.cpu_avg.toFixed(1)),
        ram: Number(b.mem_avg.toFixed(1)),
        disk: Number(b.disk_avg.toFixed(1)),
    }))

    const tickStep = Math.max(1, Math.floor(data.length / (bucketHours === 1 ? 12 : 10)))

    return (
        <ResponsiveContainer width="100%" height={240}>
            <AreaChart data={data} margin={{ top: 4, right: 12, left: 0, bottom: 0 }}>
                <defs>
                    <linearGradient id="gcpu" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="5%" stopColor="#3b82f6" stopOpacity={0.18} />
                        <stop offset="95%" stopColor="#3b82f6" stopOpacity={0} />
                    </linearGradient>
                    <linearGradient id="gram" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="5%" stopColor="#10b981" stopOpacity={0.18} />
                        <stop offset="95%" stopColor="#10b981" stopOpacity={0} />
                    </linearGradient>
                    <linearGradient id="gdisk" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="5%" stopColor="#f59e0b" stopOpacity={0.18} />
                        <stop offset="95%" stopColor="#f59e0b" stopOpacity={0} />
                    </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.05)" />
                <XAxis
                    dataKey="time"
                    stroke="#374151"
                    tick={{ fill: '#6b7280', fontSize: 10, fontFamily: 'monospace' }}
                    interval={tickStep - 1}
                    tickLine={false}
                />
                <YAxis
                    domain={[0, 100]}
                    stroke="#374151"
                    tick={{ fill: '#6b7280', fontSize: 10, fontFamily: 'monospace' }}
                    tickLine={false}
                    tickFormatter={(v: number) => `${v}%`}
                    width={38}
                />
                <Tooltip
                    contentStyle={{
                        backgroundColor: '#0f0f1a',
                        border: '1px solid rgba(255,255,255,0.08)',
                        borderRadius: 8,
                        fontFamily: 'monospace',
                        fontSize: 11,
                    }}
                    labelStyle={{ color: '#9ca3af', marginBottom: 4 }}
                    formatter={(value) => [`${Number(value).toFixed(1)}%`]}
                />
                <Legend
                    wrapperStyle={{ fontFamily: 'monospace', fontSize: 10, paddingTop: 10 }}
                    formatter={(v) => String(v).toUpperCase()}
                />
                <Area type="monotone" dataKey="cpu" name="cpu" stroke="#3b82f6" fill="url(#gcpu)" strokeWidth={1.5} dot={false} activeDot={{ r: 3 }} />
                <Area type="monotone" dataKey="ram" name="ram" stroke="#10b981" fill="url(#gram)" strokeWidth={1.5} dot={false} activeDot={{ r: 3 }} />
                <Area type="monotone" dataKey="disk" name="disk" stroke="#f59e0b" fill="url(#gdisk)" strokeWidth={1.5} dot={false} activeDot={{ r: 3 }} />
            </AreaChart>
        </ResponsiveContainer>
    )
}
