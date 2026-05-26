'use client'

import { useEffect, useState, useCallback } from 'react'
import { api, StatusResponse, HistoryResponse, AgentsResponse } from './api'

function usePolling<T>(
    fetcher: () => Promise<T>,
    interval: number,
    initial: T
) {
    const [data, setData] = useState<T>(initial)
    const [error, setError] = useState<string | null>(null)

    const fetch = useCallback(async () => {
        try {
            const result = await fetcher()
            setData(result)
            setError(null)
        } catch (e) {
            setError(e instanceof Error ? e.message : 'Erreur inconnue')
        }
    }, [fetcher])

    useEffect(() => {
        fetch()
        const id = setInterval(fetch, interval)
        return () => clearInterval(id)
    }, [fetch, interval])

    return { data, error }
}

export function useStatus() {
    return usePolling<StatusResponse | null>(
        api.status,
        5000,
        null
    )
}

export function useHistory() {
    return usePolling<HistoryResponse>(
        api.history,
        15000,
        { cycles: [], total: 0 }
    )
}

export function useAgents() {
    return usePolling<AgentsResponse>(
        api.agents,
        10000,
        { agents: [], total: 0 }
    )
}