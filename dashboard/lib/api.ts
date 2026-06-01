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
    dry_run: boolean
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
    alert_count: number   // nouveau
    schedule_ns: number   // renommé depuis schedule_ms
}

export interface AgentsResponse {
    agents: AgentStatus[]
    total: number
}

export interface ContainerState {
    id: string
    name: string
    image: string
    status: string
    running: boolean
    exited: boolean
}

export interface DockerResponse {
    available: boolean
    containers: ContainerState[]
    total: number
    running: number
    exited: number
}

export interface LogStatus {
    filepath: string
    size_bytes: number
    size_mb: number
    max_bytes: number
    max_mb: number
    used_percent: number
    backups: string[]
    backup_count: number
}

export interface LogsStatusResponse {
    main_log: LogStatus
    alert_log: LogStatus
}

export interface FileAgentResponse {
    enabled: boolean
    watch_paths: string[]
    max_size_mb: number
    last_run?: string
    run_count: number
    alert_count: number
    last_error?: string
}

export interface DailyReportResponse {
    enabled: boolean
    scheduled_at: string
    last_sent?: string
    next_send?: string
}

export interface SendReportResponse {
    sent: boolean
    message: string
}

export interface LLMContextResponse {
    cycle_count: number
    dominant_action: string
    alert_rate: number
    cpu_trend: string
    ram_trend: string
    disk_trend: string
    recent_alerts: string[]
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
    docker: () => fetchAPI<DockerResponse>('/api/docker'),
    logsStatus: () => fetchAPI<LogsStatusResponse>('/api/logs/status'),
    file: () => fetchAPI<FileAgentResponse>('/api/file'),
    dailyReport: () => fetchAPI<DailyReportResponse>('/api/daily-report'),
    llmContext: () => fetchAPI<LLMContextResponse>('/api/llm-context'),
    sendDailyReport: async (): Promise<SendReportResponse> => {
        const res = await fetch(`${RUNTIME_URL}/api/daily-report/send`, {
            method: 'POST',
            cache: 'no-store',
        })
        if (!res.ok) throw new Error(`API ${res.status}`)
        return res.json()
    },
}

// ── Toggle  ───────────────────────────────────────────────────────────────────

export async function toggleAgent(name: string): Promise<ToggleResponse> {
    const res = await fetch(`${RUNTIME_URL}/api/agents/toggle`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name }),
    })
    if (!res.ok) throw new Error(`Toggle failed: ${res.status}`)
    return res.json()
}

export interface ToggleResponse {
    name: string
    enabled: boolean
    message: string
}