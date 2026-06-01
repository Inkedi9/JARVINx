'use client'

import { useDocker } from '@/lib/hooks'
import { cn, formatTime } from '@/lib/utils'
import { Container, CheckCircle, XCircle, RefreshCw } from 'lucide-react'
import { useState } from 'react'
import { api } from '@/lib/api'

function StatusBadge({ running, exited }: { running: boolean; exited: boolean }) {
    if (running) {
        return (
            <span className="flex items-center gap-1.5 font-mono text-[10px] text-emerald-400">
                <div className="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse" />
                running
            </span>
        )
    }
    if (exited) {
        return (
            <span className="flex items-center gap-1.5 font-mono text-[10px] text-red-400">
                <div className="w-1.5 h-1.5 rounded-full bg-red-400" />
                exited
            </span>
        )
    }
    return (
        <span className="flex items-center gap-1.5 font-mono text-[10px] text-gray-500">
            <div className="w-1.5 h-1.5 rounded-full bg-gray-500" />
            unknown
        </span>
    )
}

export default function ContainersPage() {
    const { data: docker, error } = useDocker()
    const [filter, setFilter] = useState<'all' | 'running' | 'exited'>('all')

    const filtered = docker.containers.filter(c => {
        if (filter === 'running') return c.running
        if (filter === 'exited') return c.exited
        return true
    })

    const filterBtn = (f: typeof filter, label: string, count: number, color: string) => (
        <button
            onClick={() => setFilter(f)}
            className={cn(
                'flex items-center gap-2 px-3 py-1.5 rounded-lg font-mono text-[10px] border transition-all',
                filter === f
                    ? `bg-${color}-500/10 border-${color}-500/20 text-${color}-400`
                    : 'bg-bg-secondary border-border text-gray-500 hover:text-gray-300'
            )}
        >
            {label}
            <span className={cn(
                'px-1.5 py-0.5 rounded text-[9px] font-semibold',
                filter === f ? `bg-${color}-500/20` : 'bg-bg-tertiary'
            )}>
                {count}
            </span>
        </button>
    )

    return (
        <div className="space-y-6 max-w-[1400px]">

            {/* Header */}
            <div className="flex items-center justify-between">
                <div>
                    <div className="font-mono text-[10px] text-gray-600 tracking-widest uppercase mb-1">
                        Containers · Docker
                    </div>
                    <h1 className="text-xl font-semibold text-white">Container Registry</h1>
                </div>

                {/* Stats pills */}
                <div className="flex items-center gap-2">
                    {filterBtn('all', 'All', docker.total, 'gray')}
                    {filterBtn('running', 'Running', docker.running, 'emerald')}
                    {filterBtn('exited', 'Exited', docker.exited, 'red')}
                </div>
            </div>

            {/* Docker unavailable */}
            {!docker.available && (
                <div className="bg-amber-500/10 border border-amber-500/20 rounded-xl p-4">
                    <div className="font-mono text-xs text-amber-400">
                        Docker non accessible — vérifie que Docker Desktop tourne et expose le port TCP 2375
                    </div>
                </div>
            )}

            {error && (
                <div className="bg-red-500/10 border border-red-500/20 rounded-xl p-4">
                    <div className="font-mono text-xs text-red-400">
                        Runtime inaccessible — vérifie que JARVINx tourne sur le port 8080
                    </div>
                </div>
            )}

            {/* Table */}
            {docker.available && (
                <div className="bg-bg-secondary border border-border rounded-xl overflow-hidden">

                    {/* Header table */}
                    <div className="grid grid-cols-[1fr_180px_80px] gap-4 px-5 py-3 border-b border-border bg-bg-tertiary">
                        {['Container', 'Image', 'Status'].map(h => (
                            <div key={h} className="font-mono text-[9px] text-gray-600 uppercase tracking-widest">
                                {h}
                            </div>
                        ))}
                    </div>

                    {/* Rows */}
                    <div className="divide-y divide-border">
                        {filtered.length === 0 && (
                            <div className="flex flex-col items-center justify-center py-12 gap-3">
                                <Container size={28} className="text-gray-700" />
                                <div className="font-mono text-sm text-gray-600">
                                    Aucun container {filter !== 'all' ? filter : ''}
                                </div>
                            </div>
                        )}

                        {filtered.map(c => (
                            <div
                                key={c.id}
                                className={cn(
                                    'grid grid-cols-[1fr_180px_80px] gap-4 px-5 py-3.5 hover:bg-bg-tertiary transition-all',
                                    c.exited && 'opacity-60'
                                )}
                            >
                                {/* Name + ID */}
                                <div className="flex flex-col justify-center min-w-0">
                                    <div className="font-mono text-xs text-white truncate">
                                        {c.name}
                                    </div>
                                    <div className="font-mono text-[10px] text-gray-600 mt-0.5">
                                        {c.id}
                                    </div>
                                </div>

                                {/* Image */}
                                <div className="flex items-center min-w-0">
                                    <span className="font-mono text-[10px] text-gray-400 truncate">
                                        {(c.image ?? '').length > 40
                                            ? (c.image ?? '').slice(0, 40) + '...'
                                            : (c.image ?? '—')}
                                    </span>
                                </div>

                                {/* Status */}
                                <div className="flex items-center">
                                    <StatusBadge running={c.running} exited={c.exited} />
                                </div>
                            </div>
                        ))}
                    </div>
                </div>
            )}

            {/* Footer note */}
            <div className="font-mono text-[10px] text-gray-600">
                Rafraîchissement toutes les 15s · {docker.total} containers total
            </div>
        </div>
    )
}