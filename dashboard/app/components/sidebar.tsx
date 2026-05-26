'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import {
    LayoutDashboard,
    Bot,
    Database,
    Wrench,
    History,
    Settings,
    Terminal,
} from 'lucide-react'
import { cn } from '@/lib/utils'

const nav = [
    { href: '/', label: 'Overview', icon: LayoutDashboard },
    { href: '/agents', label: 'Agents', icon: Bot },
    { href: '/memory', label: 'Memory', icon: Database },
    { href: '/tools', label: 'Tools', icon: Wrench },
    { href: '/history', label: 'History', icon: History },
    { href: '/settings', label: 'Settings', icon: Settings },
]

export default function Sidebar() {
    const path = usePathname()

    return (
        <aside className="fixed left-0 top-0 h-full w-[220px] bg-bg-secondary border-r border-border flex flex-col z-50">

            {/* Logo */}
            <div className="p-5 border-b border-border">
                <div className="flex items-center gap-3">
                    <div className="w-8 h-8 bg-accent-blue rounded-lg flex items-center justify-center">
                        <span className="font-mono text-[11px] font-bold text-bg-primary">JX</span>
                    </div>
                    <div>
                        <div className="font-mono text-sm font-semibold tracking-[2px] text-white">
                            JARVINX
                        </div>
                        <div className="text-[10px] text-accent-blue tracking-[1px] uppercase">
                            Autonomous Runtime
                        </div>
                    </div>
                </div>
            </div>

            {/* Nav */}
            <nav className="flex-1 p-3 space-y-1">
                {nav.map(({ href, label, icon: Icon }) => {
                    const active = path === href
                    return (
                        <Link
                            key={href}
                            href={href}
                            className={cn(
                                'flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm transition-all',
                                active
                                    ? 'bg-accent-blue/10 text-accent-blue border border-accent-blue/20'
                                    : 'text-gray-400 hover:text-white hover:bg-bg-tertiary'
                            )}
                        >
                            <Icon size={16} />
                            {label}
                        </Link>
                    )
                })}
            </nav>

            {/* Footer */}
            <div className="p-4 border-t border-border space-y-2">
                <div className="flex items-center gap-2">
                    <div className="w-2 h-2 rounded-full bg-emerald-400 animate-pulse" />
                    <span className="font-mono text-[11px] text-gray-400">SYSTEM ONLINE</span>
                </div>
                <div className="font-mono text-[10px] text-gray-600">v1.1 · runtime</div>
                <Link
                    href="/settings"
                    className="flex items-center gap-2 w-full px-3 py-2 rounded-lg bg-bg-tertiary border border-border text-gray-400 hover:text-white text-xs transition-all"
                >
                    <Terminal size={14} />
                    SYSTEM CONSOLE
                </Link>
            </div>
        </aside>
    )
}