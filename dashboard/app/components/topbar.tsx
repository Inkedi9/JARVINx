'use client'

import { useStatus } from '@/lib/hooks'
import { Bell, Wifi, WifiOff } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useEffect, useState } from 'react'

export default function Topbar() {
    const { data: status, error } = useStatus()
    const online = !error && status?.online

    const [time, setTime] = useState('')
    const [date, setDate] = useState('')

    useEffect(() => {
        const update = () => {
            const now = new Date()
            setTime(now.toLocaleTimeString('fr-FR', {
                hour: '2-digit', minute: '2-digit', second: '2-digit'
            }))
            setDate(now.toLocaleDateString('fr-FR', {
                day: '2-digit', month: 'short', year: 'numeric'
            }))
        }
        update()
        const id = setInterval(update, 1000)
        return () => clearInterval(id)
    }, [])

    return (
        <header className="h-14 border-b border-border bg-bg-secondary flex items-center justify-between px-6">

            <div className="flex items-center gap-6">
                <div className="flex items-center gap-2">
                    <div className={cn(
                        'w-2 h-2 rounded-full',
                        online ? 'bg-emerald-400 animate-pulse' : 'bg-red-400'
                    )} />
                    <span className="font-mono text-xs text-gray-400">RUNTIME STATUS</span>
                    <span className={cn(
                        'font-mono text-xs font-semibold',
                        online ? 'text-emerald-400' : 'text-red-400'
                    )}>
                        {online ? 'ONLINE' : 'OFFLINE'}
                    </span>
                </div>

                {status && (
                    <>
                        <div className="text-gray-600">·</div>
                        <div className="font-mono text-xs text-gray-400">
                            <span className="text-gray-600">MODEL </span>
                            <span className="text-white">{status.model}</span>
                        </div>
                        <div className="text-gray-600">·</div>
                        <div className="font-mono text-xs text-gray-400">
                            <span className="text-gray-600">CYCLE </span>
                            <span className="text-accent-blue">#{status.cycle_num}</span>
                        </div>
                        <div className="text-gray-600">·</div>
                        <div className="font-mono text-xs text-gray-400">
                            <span className="text-gray-600">UPTIME </span>
                            <span className="text-white">{status.uptime}</span>
                        </div>
                    </>
                )}
            </div>

            <div className="flex items-center gap-4">
                {online ? (
                    <Wifi size={14} className="text-emerald-400" />
                ) : (
                    <WifiOff size={14} className="text-red-400" />
                )}
                <div className="font-mono text-xs text-gray-400">
                    <span className="text-white">{time}</span>
                    <span className="text-gray-600 ml-2">{date}</span>
                </div>
                <button className="relative p-2 rounded-lg hover:bg-bg-tertiary transition-all">
                    <Bell size={16} className="text-gray-400" />
                </button>
            </div>
        </header>
    )
}