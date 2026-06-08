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
    confidence?: number
    trigger_cpu?: number
    trigger_ram?: number
    trigger_disk?: number
}

export interface ExecLastResult {
    command: string
    output?: string
    error?: string
    success: boolean
    duration_ms: number
    timed_out?: boolean
}

export interface StatusResponse {
    online: boolean
    model: string
    interval: string
    cycle_num: number
    uptime: string
    dry_run: boolean
    circuit_state?: 'closed' | 'open' | 'half-open'
    last_cycle?: CycleRecord
    exec_guard?: {
        last_cmd: string
        cooldown_remaining_seconds: number
    }
    last_exec_result?: ExecLastResult
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
    cpu_forecast?: string
    ram_forecast?: string
    disk_forecast?: string
    recent_alerts: string[]
}

export interface SnapshotBucket {
    timestamp: string
    cpu_avg: number
    cpu_max: number
    mem_avg: number
    mem_max: number
    disk_avg: number
    disk_max: number
    count: number
}

export interface AlertEntry {
    timestamp: string
    level: 'warning' | 'critical'
    metric: string
    value: number
    threshold: number
    message: string
    cycles_above: number
}

export interface AlertsResponse {
    alerts: AlertEntry[]
    total: number
}

export interface HistoryFullResponse {
    range: string
    from: string
    to: string
    bucket_hours: number
    buckets: SnapshotBucket[]
    total_snapshots: number
    available: boolean
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
    llmContext: async () => {
        const res = await fetchAPI<LLMContextResponse>('/api/llm-context')
        return {
            ...res,
            recent_alerts: res.recent_alerts ?? [],
        }
    },
    alerts: () => fetchAPI<AlertsResponse>('/api/alerts?limit=200'),
    historyFull: (range: '7d' | '30d' | '90d') =>
        fetchAPI<HistoryFullResponse>(`/api/history/full?range=${range}`),
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