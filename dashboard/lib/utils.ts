import { clsx, type ClassValue } from 'clsx'

export function cn(...inputs: ClassValue[]) {
    return clsx(inputs)
}

export function formatTime(iso: string): string {
    return new Date(iso).toLocaleTimeString('fr-FR', {
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
    })
}

export function formatUptime(uptime: string): string {
    return uptime
}

export function metricColor(pct: number, warn = 70, crit = 85): string {
    if (pct >= crit) return 'text-red-400'
    if (pct >= warn) return 'text-amber-400'
    return 'text-emerald-400'
}

export function metricBg(pct: number, warn = 70, crit = 85): string {
    if (pct >= crit) return 'bg-red-500'
    if (pct >= warn) return 'bg-amber-500'
    return 'bg-emerald-500'
}

export function actionColor(action: string): string {
    switch (action) {
        case 'alert': return 'text-red-400 bg-red-400/10 border-red-400/20'
        case 'suggest': return 'text-amber-400 bg-amber-400/10 border-amber-400/20'
        case 'execute': return 'text-blue-400 bg-blue-400/10 border-blue-400/20'
        default: return 'text-emerald-400 bg-emerald-400/10 border-emerald-400/20'
    }
}