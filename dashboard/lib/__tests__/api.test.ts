import { api } from '../api'

// Mock fetch global
global.fetch = jest.fn()

const mockFetch = global.fetch as jest.MockedFunction<typeof fetch>

describe('api.status', () => {
    beforeEach(() => {
        mockFetch.mockClear()
    })

    it('calls /api/status endpoint', async () => {
        mockFetch.mockResolvedValueOnce({
            ok: true,
            json: async () => ({
                online: true,
                model: 'llama3.1:8b',
                interval: '15s',
                cycle_num: 42,
                uptime: '1h 0m 0s',
            }),
        } as Response)

        const result = await api.status()

        expect(mockFetch).toHaveBeenCalledWith(
            expect.stringContaining('/api/status'),
            expect.any(Object)
        )
        expect(result.online).toBe(true)
        expect(result.model).toBe('llama3.1:8b')
        expect(result.cycle_num).toBe(42)
    })

    it('throws on non-ok response', async () => {
        mockFetch.mockResolvedValueOnce({
            ok: false,
            status: 500,
        } as Response)

        await expect(api.status()).rejects.toThrow('API 500')
    })
})

describe('api.history', () => {
    it('calls /api/history endpoint', async () => {
        mockFetch.mockResolvedValueOnce({
            ok: true,
            json: async () => ({ cycles: [], total: 0 }),
        } as Response)

        const result = await api.history()

        expect(mockFetch).toHaveBeenCalledWith(
            expect.stringContaining('/api/history'),
            expect.any(Object)
        )
        expect(result.cycles).toEqual([])
        expect(result.total).toBe(0)
    })
})

describe('api.agents', () => {
    it('calls /api/agents endpoint', async () => {
        mockFetch.mockResolvedValueOnce({
            ok: true,
            json: async () => ({
                agents: [
                    {
                        name: 'system',
                        enabled: true,
                        run_count: 5,
                        error_count: 0,
                        alert_count: 0,
                        schedule_ns: 15_000_000_000,
                        last_run: new Date().toISOString(),
                    }
                ],
                total: 1,
            }),
        } as Response)

        const result = await api.agents()

        expect(result.agents).toHaveLength(1)
        expect(result.agents[0].name).toBe('system')
        expect(result.total).toBe(1)
    })
})