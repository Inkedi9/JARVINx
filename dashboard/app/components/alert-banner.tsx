'use client'

import { useStatus } from '@/lib/hooks'
import { formatTime } from '@/lib/utils'
import { AlertTriangle, X } from 'lucide-react'
import { useState, useEffect } from 'react'

export default function AlertBanner() {
  const { data: status } = useStatus()
  const [dismissed, setDismissed] = useState(false)
  const [lastCycleNum, setLastCycleNum] = useState(0)

  const cycle = status?.last_cycle

  // Reset dismissed quand un nouveau cycle alert arrive
  useEffect(() => {
    if (cycle && cycle.cycle_num !== lastCycleNum) {
      setLastCycleNum(cycle.cycle_num)
      if (cycle.action === 'alert') {
        setDismissed(false)
      }
    }
  }, [cycle, lastCycleNum])

  if (!cycle || cycle.action !== 'alert' || dismissed) return null

  return (
    <div className="flex items-center gap-3 px-6 py-2.5 bg-red-500/10 border-b border-red-500/30">
      <AlertTriangle size={13} className="text-red-400 flex-shrink-0" />
      <div className="flex-1 font-mono text-[11px] text-red-300 truncate">
        <span className="text-red-400 font-semibold mr-2">ALERT</span>
        SystemAgent
        <span className="text-red-500 mx-2">·</span>
        {cycle.analysis}
        {cycle.command && (
          <>
            <span className="text-red-500 mx-2">·</span>
            <span className="text-amber-400">POST → {cycle.command}</span>
          </>
        )}
        <span className="text-red-600 ml-3">{formatTime(cycle.timestamp)}</span>
      </div>
      <button
        onClick={() => setDismissed(true)}
        className="text-red-600 hover:text-red-400 transition-colors flex-shrink-0"
      >
        <X size={13} />
      </button>
    </div>
  )
}