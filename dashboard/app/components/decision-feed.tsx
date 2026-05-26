'use client'

import { useHistory } from '@/lib/hooks'
import { formatTime, actionColor, cn } from '@/lib/utils'
import { Terminal } from 'lucide-react'

export default function DecisionFeed() {
    const { data } = useHistory()

    return (
        <div className="bg-bg-secondary border border-border rounded-xl p-5 flex flex-col gap-4">
            <div className="flex items-center justify-between">
                <span className="font-mono text-[10px] text-gray-500 tracking-widest uppercase">
                    Recent Decisions
                </span>
                <span className="font-mono text-[10px] text-gray-600">
                    {data.total} total
                </span>
            </div>

            <div className="space-y-2">
                {data.cycles.slice(0, 5).map((cycle) => (
                    <div
                        key={cycle.cycle_num}
                        className="flex gap-3 p-3 rounded-lg bg-bg-tertiary border border-border group hover:border-border-subtle transition-all"
                    >
                        {/* Time */}
                        <div className="font-mono text-[10px] text-gray-600 flex-shrink-0 pt-0.5 w-14">
                            {formatTime(cycle.timestamp)}
                        </div>

                        {/* Content */}
                        <div className="flex-1 min-w-0">
                            <div className="flex items-center gap-2 mb-1">
                                <span className="font-mono text-[10px] text-gray-400 capitalize">
                                    SystemAgent
                                </span>
                                <span className={cn(
                                    'font-mono text-[9px] px-1.5 py-0.5 rounded border',
                                    actionColor(cycle.action)
                                )}>
                                    {cycle.action}
                                </span>
                                {cycle.command && (
                                    <span className="flex items-center gap-1 font-mono text-[9px] text-amber-400">
                                        <Terminal size={9} />
                                        {cycle.command}
                                    </span>
                                )}
                            </div>
                            <div className="font-mono text-[10px] text-gray-500 truncate">
                                {cycle.analysis}
                            </div>
                        </div>

                        {/* Cycle num */}
                        <div className="font-mono text-[10px] text-gray-700 flex-shrink-0">
                            #{cycle.cycle_num}
                        </div>
                    </div>
                ))}

                {data.cycles.length === 0 && (
                    <div className="font-mono text-[11px] text-gray-600 text-center py-4">
                        En attente du premier cycle...
                    </div>
                )}
            </div>
        </div>
    )
}