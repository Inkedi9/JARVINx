'use client'

import { useEffect, useState, useRef } from 'react'
import { api, StatusResponse, HistoryResponse, AgentsResponse } from './api'

const MAX_BACKOFF_MS = 30_000  // 30s max
const BASE_BACKOFF_MS = 1_000   // 1s initial

function usePolling<T>(
    fetcher: () => Promise<T>,
    interval: number,
    initial: T
) {
    const [data, setData] = useState<T>(initial)
    const [error, setError] = useState<string | null>(null)
    const backoffRef = useRef(BASE_BACKOFF_MS)
    const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

    useEffect(() => {
        let cancelled = false
        backoffRef.current = BASE_BACKOFF_MS

        async function poll() {
            try {
                const result = await fetcher()
                if (cancelled) return
                setData(result)
                setError(null)
                backoffRef.current = BASE_BACKOFF_MS
                timerRef.current = setTimeout(poll, interval)
            } catch (e) {
                if (cancelled) return
                const msg = e instanceof Error ? e.message : 'Erreur inconnue'
                setError(msg)
                const nextBackoff = Math.min(backoffRef.current * 2, MAX_BACKOFF_MS)
                backoffRef.current = nextBackoff
                timerRef.current = setTimeout(poll, nextBackoff)
            }
        }

        poll()

        return () => {
            cancelled = true
            if (timerRef.current) clearTimeout(timerRef.current)
        }
    }, [fetcher, interval])

    return { data, error }
}

export function useStatus() {
    return usePolling<StatusResponse | null>(api.status, 5_000, null)
}

export function useHistory() {
    return usePolling<HistoryResponse>(api.history, 15_000, { cycles: [], total: 0 })
}

export function useAgents() {
    return usePolling<AgentsResponse>(api.agents, 10_000, { agents: [], total: 0 })
}
