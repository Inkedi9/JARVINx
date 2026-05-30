'use client'

import { useAgents } from '@/lib/hooks'
import { formatTime, cn } from '@/lib/utils'
import { Bot, AlertCircle, CheckCircle } from 'lucide-react'

export default function AgentList() {
    const { data, error } = useAgents()

    return (
        <div className="bg-bg-secondary border border-border rounded-xl p-5 flex flex-col gap-4">
            <div className="flex items-center justify-between">
                <span className="font-mono text-[10px] text-gray-500 tracking-widest uppercase">
                    Active Agents
                </span>
                <span className="font-mono text-[10px] text-emerald-400">
                    {data.agents.filter(a => a.enabled).length} RUNNING
                </span>
            </div>

            {error && (
                <div className="font-mono text-[11px] text-red-400">
                    Runtime inaccessible
                </div>
            )}

            <div className="space-y-3">
                {data.agents.map((agent) => {
                    // Calcul health — alerts ne comptent plus comme erreurs
                    const health = agent.run_count === 0
                        ? 100
                        : Math.round((1 - agent.error_count / Math.max(agent.run_count, 1)) * 100)

                    const healthColor = health >= 95
                        ? 'text-emerald-400'
                        : health >= 80
                            ? 'text-amber-400'
                            : 'text-red-400'

                    return (
                        <div
                            key={agent.name}
                            className="flex items-center gap-3 p-3 rounded-lg bg-bg-tertiary border border-border hover:border-border-subtle transition-all"
                        >
                            {/* Icon */}
                            <div className="w-8 h-8 rounded-lg bg-bg-primary border border-border flex items-center justify-center flex-shrink-0">
                                <Bot size={14} className="text-gray-400" />
                            </div>

                            {/* Info */}
                            <div className="flex-1 min-w-0">
                                <div className="flex items-center gap-2">
                                    <span className="font-mono text-xs text-white capitalize">
                                        {agent.name}Agent
                                    </span>
                                    {agent.error_count > 0 ? (
                                        <AlertCircle size={10} className="text-red-400" />
                                    ) : (
                                        <CheckCircle size={10} className="text-emerald-400" />
                                    )}
                                </div>
                                <div className="font-mono text-[10px] text-gray-600 mt-0.5 truncate">
                                    {agent.enabled ? (
                                        <>
                                            <span className="text-emerald-400">● Running</span>
                                            {agent.last_run && (
                                                <span className="ml-2">
                                                    Last {formatTime(agent.last_run)}
                                                </span>
                                            )}
                                        </>
                                    ) : (
                                        <span className="text-gray-600">● Disabled</span>
                                    )}
                                </div>
                            </div>

                            {/* Health */}
                            <div className="text-right flex-shrink-0">
                                <div className={cn('font-mono text-xs font-semibold', healthColor)}>
                                    {health}%
                                </div>
                                <div className="font-mono text-[9px] text-gray-600">
                                    {agent.run_count} runs
                                </div>
                            </div>

                            <div className="font-mono text-[10px] text-gray-600 mt-0.5">
                                {agent.alert_count > 0 && (
                                    <span className="text-amber-400 mr-2">
                                        ⚡ {agent.alert_count} alerts
                                    </span>
                                )}
                                {agent.run_count} runs
                            </div>
                        </div>
                    )
                })}

                {data.agents.length === 0 && !error && (
                    <div className="font-mono text-[11px] text-gray-600 text-center py-4">
                        Aucun agent enregistré
                    </div>
                )}
            </div>
        </div>
    )
}