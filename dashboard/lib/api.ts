const RUNTIME_URL = process.env.NEXT_PUBLIC_RUNTIME_URL ?? 'http://localhost:8080'

// ── Types ────────────────────────────────────────────────────────────────────

export interface Snapshot {
    timestamp: string
    cpu_percent: number
    mem_used_mb: number
    mem_total_mb: number
    mem_percent: number
    disk_used_gb: number
    disk_total_gb: number
    disk_percent: number
}

export interface CycleRecord {
    cycle_num: number
    timestamp: string
    snapshot: Snapshot
    action: 'log' | 'alert' | 'suggest' | 'execute'
    analysis: string
    reason: string
    command?: string
}

export interface StatusResponse {
    online: boolean
    model: string
    interval: string
    cycle_num: number
    uptime: string
    last_cycle?: CycleRecord
}

export interface HistoryResponse {
    cycles: CycleRecord[]
    total: number
}

export interface AgentStatus {
    name: string
    enabled: boolean
    last_run: string
    last_error?: string
    run_count: number
    error_count: number
    schedule_ms: number
}

export interface AgentsResponse {
    agents: AgentStatus[]
    total: number
}

// ── Client ───────────────────────────────────────────────────────────────────

async function fetchAPI<T>(endpoint: string): Promise<T> {
    const res = await fetch(`${RUNTIME_URL}${endpoint}`, {
        cache: 'no-store',
    })
    if (!res.ok) throw new Error(`API ${res.status} on ${endpoint}`)
    return res.json()
}

export const api = {
    status: () => fetchAPI<StatusResponse>('/api/status'),
    history: () => fetchAPI<HistoryResponse>('/api/history'),
    agents: () => fetchAPI<AgentsResponse>('/api/agents'),
}