import { formatTime, metricColor, metricBg, actionColor } from '../utils'

describe('formatTime', () => {
    it('formats ISO string to HH:MM:SS', () => {
        const iso = '2026-05-26T14:32:17.000Z'
        const result = formatTime(iso)
        // Vérifie le format HH:MM:SS
        expect(result).toMatch(/^\d{2}:\d{2}:\d{2}$/)
    })
})

describe('metricColor', () => {
    it('returns green below warn threshold', () => {
        expect(metricColor(50, 70, 85)).toBe('text-emerald-400')
    })

    it('returns amber between warn and crit', () => {
        expect(metricColor(75, 70, 85)).toBe('text-amber-400')
    })

    it('returns red above crit threshold', () => {
        expect(metricColor(90, 70, 85)).toBe('text-red-400')
    })

    it('returns red at exact crit threshold', () => {
        expect(metricColor(85, 70, 85)).toBe('text-red-400')
    })

    it('returns amber at exact warn threshold', () => {
        expect(metricColor(70, 70, 85)).toBe('text-amber-400')
    })
})

describe('metricBg', () => {
    it('returns green bg below warn', () => {
        expect(metricBg(50, 70, 85)).toBe('bg-emerald-500')
    })

    it('returns amber bg between thresholds', () => {
        expect(metricBg(75, 70, 85)).toBe('bg-amber-500')
    })

    it('returns red bg above crit', () => {
        expect(metricBg(90, 70, 85)).toBe('bg-red-500')
    })
})

describe('actionColor', () => {
    it('returns green classes for log', () => {
        const result = actionColor('log')
        expect(result).toContain('emerald')
    })

    it('returns amber classes for suggest', () => {
        const result = actionColor('suggest')
        expect(result).toContain('amber')
    })

    it('returns red classes for alert', () => {
        const result = actionColor('alert')
        expect(result).toContain('red')
    })

    it('returns blue classes for execute', () => {
        const result = actionColor('execute')
        expect(result).toContain('blue')
    })

    it('returns log classes for unknown action', () => {
        const result = actionColor('unknown')
        expect(result).toContain('emerald')
    })
})