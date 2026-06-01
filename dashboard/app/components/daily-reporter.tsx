'use client'

import { useState } from 'react'
import { usePolling } from '@/lib/hooks'
import { api, DailyReportResponse } from '@/lib/api'
import { Clock, Send, CheckCircle } from 'lucide-react'
import { cn } from '@/lib/utils'

function useDailyReport() {
  return usePolling<DailyReportResponse>(
    api.dailyReport,
    30_000,
    { enabled: false, scheduled_at: '08:00' }
  )
}

export default function DailyReporter() {
  const { data }          = useDailyReport()
  const [sending, setSending] = useState(false)
  const [sent, setSent]       = useState(false)

  async function handleSend() {
    setSending(true)
    try {
      await api.sendDailyReport()
      setSent(true)
      setTimeout(() => setSent(false), 3000)
    } catch (e) {
      console.error(e)
    } finally {
      setSending(false)
    }
  }

  if (!data.enabled) return null

  return (
    <div className="bg-bg-secondary border border-border rounded-xl p-4">
      <div className="flex items-center justify-between mb-3">
        <div className="flex items-center gap-2">
          <Clock size={13} className="text-accent-blue" />
          <span className="font-mono text-[10px] text-gray-500 uppercase tracking-widest">
            Daily Report
          </span>
        </div>
        <span className="font-mono text-[10px] text-gray-600">
          {data.scheduled_at}
        </span>
      </div>

      <div className="space-y-1.5 mb-4">
        <div className="flex justify-between">
          <span className="font-mono text-[10px] text-gray-600">Dernier envoi</span>
          <span className="font-mono text-[10px] text-gray-400">
            {data.last_sent
              ? new Date(data.last_sent).toLocaleString('fr-FR')
              : '—'}
          </span>
        </div>
        <div className="flex justify-between">
          <span className="font-mono text-[10px] text-gray-600">Prochain envoi</span>
          <span className="font-mono text-[10px] text-accent-blue">
            {data.next_send
              ? new Date(data.next_send).toLocaleTimeString('fr-FR', {
                  hour: '2-digit', minute: '2-digit'
                })
              : '—'}
          </span>
        </div>
      </div>

      <button
        onClick={handleSend}
        disabled={sending || sent}
        className={cn(
          'w-full flex items-center justify-center gap-2 py-2 rounded-lg',
          'font-mono text-[10px] border transition-all',
          sent
            ? 'bg-emerald-500/10 border-emerald-500/20 text-emerald-400'
            : 'bg-accent-blue/10 border-accent-blue/20 text-accent-blue hover:bg-accent-blue/20 disabled:opacity-50'
        )}
      >
        {sent
          ? <><CheckCircle size={11} /> Envoyé</>
          : sending
          ? <><Send size={11} className="animate-pulse" /> Envoi...</>
          : <><Send size={11} /> Envoyer maintenant</>
        }
      </button>
    </div>
  )
}