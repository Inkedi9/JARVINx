'use client'

import { useAgents, useHistory } from '@/lib/hooks'
import { formatTime, cn } from '@/lib/utils'
import {
    Bot, CheckCircle, AlertCircle, Clock,
    Activity, Zap, XCircle
} from 'lucide-react'

function AgentCard({ agent, cycles }: {
    agent: {
        name: string
        enabled: boolean
        last_run: string
        last_error?: string
        run_count: number
        error_count: number
        schedule_ns: number
    }
    cycles: number
}) {
    const health = agent.run_count === 0
        ? 100
        : Math.round((1 - agent.error_count / agent.run_count) * 100)

    const scheduleS = Math.round(agent.schedule_ns / 1_000_000_000)

    const healthColor = health >= 95 ? 'text-emerald-400'
        : health >= 80 ? 'text-amber-400'
            : 'text-red-400'

    const healthBg = health >= 95 ? 'bg-emerald-500'
        : health >= 80 ? 'bg-amber-500'
            : 'bg-red-500'

    return (
        <div className={cn(
            'bg-bg-secondary border rounded-xl p-5 flex flex-col gap-5 transition-all',
            agent.enabled ? 'border-border hover:border-border-subtle' : 'border-border opacity-60'
        )}>

            {/* Header */}
            <div className="flex items-start justify-between">
                <div className="flex items-center gap-3">
                    <div className={cn(
                        'w-10 h-10 rounded-xl border flex items-center justify-center',
                        agent.enabled
                            ? 'bg-accent-blue/10 border-accent-blue/20'
                            : 'bg-bg-tertiary border-border'
                    )}>
                        <Bot size={18} className={agent.enabled ? 'text-accent-blue' : 'text-gray-600'} />
                    </div>
                    <div>
                        <div className="font-mono text-sm font-semibold text-white capitalize">
                            {agent.name}Agent
                        </div>
                        <div className="flex items-center gap-1.5 mt-0.5">
                            <div className={cn(
                                'w-1.5 h-1.5 rounded-full',
                                agent.enabled ? 'bg-emerald-400 animate-pulse' : 'bg-gray-600'
                            )} />
                            <span className="font-mono text-[10px] text-gray-500">
                                {agent.enabled ? 'Running' : 'Disabled'}
                            </span>
                        </div>
                    </div>
                </div>

                {/* Health badge */}
                <div className="text-right">
                    <div className={cn('font-mono text-xl font-bold', healthColor)}>
                        {health}%
                    </div>
                    <div className="font-mono text-[10px] text-gray-600">health</div>
                </div>
            </div>

            {/* Health bar */}
            <div className="h-1 bg-bg-tertiary rounded-full overflow-hidden">
                <div
                    className={cn('h-full rounded-full transition-all duration-700', healthBg)}
                    style={{ width: `${health}%` }}
                />
            </div>

            {/* Stats grid */}
            <div className="grid grid-cols-3 gap-3">
                <div className="bg-bg-tertiary rounded-lg p-3 text-center">
                    <div className="flex items-center justify-center gap-1 mb-1">
                        <Activity size={11} className="text-gray-500" />
                    </div>
                    <div className="font-mono text-sm font-semibold text-white">
                        {agent.run_count}
                    </div>
                    <div className="font-mono text-[9px] text-gray-600 uppercase tracking-wider">
                        Runs
                    </div>
                </div>

                <div className="bg-bg-tertiary rounded-lg p-3 text-center">
                    <div className="flex items-center justify-center gap-1 mb-1">
                        <XCircle size={11} className="text-gray-500" />
                    </div>
                    <div className={cn(
                        'font-mono text-sm font-semibold',
                        agent.error_count > 0 ? 'text-red-400' : 'text-white'
                    )}>
                        {agent.error_count}
                    </div>
                    <div className="font-mono text-[9px] text-gray-600 uppercase tracking-wider">
                        Errors
                    </div>
                </div>

                <div className="bg-bg-tertiary rounded-lg p-3 text-center">
                    <div className="flex items-center justify-center gap-1 mb-1">
                        <Zap size={11} className="text-gray-500" />
                    </div>
                    <div className="font-mono text-sm font-semibold text-white">
                        {scheduleS}s
                    </div>
                    <div className="font-mono text-[9px] text-gray-600 uppercase tracking-wider">
                        Schedule
                    </div>
                </div>
            </div>

            {/* Last run */}
            <div className="flex items-center gap-2 pt-1 border-t border-border">
                <Clock size={11} className="text-gray-600" />
                <span className="font-mono text-[10px] text-gray-600">Last run</span>
                <span className="font-mono text-[10px] text-gray-400 ml-auto">
                    {agent.last_run ? formatTime(agent.last_run) : '—'}
                </span>
            </div>

            {/* Error message */}
            {agent.last_error && (
                <div className="flex items-start gap-2 bg-red-500/5 border border-red-500/20 rounded-lg p-3">
                    <AlertCircle size={12} className="text-red-400 flex-shrink-0 mt-0.5" />
                    <span className="font-mono text-[10px] text-red-400 leading-relaxed">
                        {agent.last_error}
                    </span>
                </div>
            )}
        </div>
    )
}

export default function AgentsPage() {
    const { data: agents, error } = useAgents()
    const { data: history } = useHistory()

    const running = agents.agents.filter(a => a.enabled).length
    const errors = agents.agents.reduce((acc, a) => acc + a.error_count, 0)

    return (
        <div className="space-y-6 max-w-[1400px]">

            {/* Header */}
            <div className="flex items-center justify-between">
                <div>
                    <div className="font-mono text-[10px] text-gray-600 tracking-widest uppercase mb-1">
                        Agents · Registry
                    </div>
                    <h1 className="text-xl font-semibold text-white">Agent Registry</h1>
                </div>

                {/* Summary pills */}
                <div className="flex items-center gap-3">
                    <div className="flex items-center gap-2 bg-bg-secondary border border-border rounded-lg px-3 py-2">
                        <CheckCircle size={12} className="text-emerald-400" />
                        <span className="font-mono text-xs text-gray-400">
                            {running} running
                        </span>
                    </div>
                    <div className="flex items-center gap-2 bg-bg-secondary border border-border rounded-lg px-3 py-2">
                        <Bot size={12} className="text-gray-400" />
                        <span className="font-mono text-xs text-gray-400">
                            {agents.total} total
                        </span>
                    </div>
                    {errors > 0 && (
                        <div className="flex items-center gap-2 bg-red-500/10 border border-red-500/20 rounded-lg px-3 py-2">
                            <AlertCircle size={12} className="text-red-400" />
                            <span className="font-mono text-xs text-red-400">
                                {errors} errors
                            </span>
                        </div>
                    )}
                </div>
            </div>

            {error && (
                <div className="bg-red-500/10 border border-red-500/20 rounded-xl p-4">
                    <div className="font-mono text-xs text-red-400">
                        Runtime inaccessible — vérifie que JARVINx tourne sur le port 8080
                    </div>
                </div>
            )}

            {/* Agent cards */}
            <div className="grid grid-cols-3 gap-4">
                {agents.agents.map((agent) => (
                    <AgentCard
                        key={agent.name}
                        agent={agent}
                        cycles={history.total}
                    />
                ))}

                {agents.agents.length === 0 && !error && (
                    <div className="col-span-3 bg-bg-secondary border border-border rounded-xl p-12 text-center">
                        <Bot size={32} className="text-gray-700 mx-auto mb-3" />
                        <div className="font-mono text-sm text-gray-600">
                            Aucun agent enregistré
                        </div>
                    </div>
                )}
            </div>

            {/* Interface note */}
            <div className="bg-bg-secondary border border-border rounded-xl p-4">
                <div className="font-mono text-[10px] text-gray-600 tracking-widest uppercase mb-2">
                    Enable / Disable
                </div>
                <div className="font-mono text-[11px] text-gray-500 leading-relaxed">
                    Pour activer ou désactiver un agent à chaud, utilise la CLI interactive :{' '}
                    <span className="text-accent-blue">interval</span>,{' '}
                    <span className="text-accent-blue">status</span>,{' '}
                    <span className="text-accent-blue">history</span>.
                    L'API d'administration REST est prévue en v2.0.
                </div>
            </div>

        </div>
    )
}