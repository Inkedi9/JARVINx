'use client'

import { cn } from '@/lib/utils'

const steps = [
    { id: 1, key: 'observe', label: 'OBSERVE', sub: 'Collect data' },
    { id: 2, key: 'think', label: 'THINK', sub: 'Process context' },
    { id: 3, key: 'decide', label: 'DECIDE', sub: 'Evaluate' },
    { id: 4, key: 'act', label: 'ACT', sub: 'Execute' },
    { id: 5, key: 'learn', label: 'LEARN', sub: 'Improve' },
]

interface RuntimeCycleProps {
    activeStep?: number
}

export default function RuntimeCycle({ activeStep = 1 }: RuntimeCycleProps) {
    return (
        <div className="bg-bg-secondary border border-border rounded-xl p-6">
            <div className="flex items-center justify-between mb-8">
                <span className="font-mono text-[10px] text-gray-500 tracking-widest uppercase">
                    Runtime Cycle
                </span>
                <span className="font-mono text-[10px] text-accent-blue">
                    STEP {activeStep} OF {steps.length}
                </span>
            </div>

            <div className="flex items-center justify-between relative">
                {/* Ligne de connexion */}
                <div className="absolute top-6 left-[10%] right-[10%] h-px bg-border z-0" />

                {steps.map((step) => {
                    const isDone = step.id < activeStep
                    const isActive = step.id === activeStep

                    return (
                        <div key={step.key} className="flex flex-col items-center gap-3 z-10">

                            {/* Numéro étape */}
                            <div className="font-mono text-[9px] text-gray-600 mb-1">
                                0{step.id}
                            </div>

                            {/* Cercle */}
                            <div className={cn(
                                'w-12 h-12 rounded-full border-2 flex items-center justify-center transition-all',
                                isActive
                                    ? 'border-accent-blue bg-accent-blue/10 shadow-[0_0_20px_rgba(59,130,246,0.3)]'
                                    : isDone
                                        ? 'border-emerald-500 bg-emerald-500/10'
                                        : 'border-border bg-bg-tertiary'
                            )}>
                                {isDone ? (
                                    <span className="text-emerald-400 text-sm">✓</span>
                                ) : isActive ? (
                                    <div className="w-2 h-2 rounded-full bg-accent-blue animate-pulse" />
                                ) : (
                                    <div className="w-2 h-2 rounded-full bg-gray-600" />
                                )}
                            </div>

                            {/* Label */}
                            <div className="text-center">
                                <div className={cn(
                                    'font-mono text-[11px] font-semibold',
                                    isActive ? 'text-accent-blue' : isDone ? 'text-emerald-400' : 'text-gray-600'
                                )}>
                                    {step.label}
                                </div>
                                <div className="font-mono text-[9px] text-gray-600 mt-0.5">
                                    {step.sub}
                                </div>
                            </div>
                        </div>
                    )
                })}
            </div>
        </div>
    )
}