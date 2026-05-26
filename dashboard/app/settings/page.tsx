'use client'

import { useStatus } from '@/lib/hooks'
import { cn } from '@/lib/utils'
import { CheckCircle, AlertCircle, ExternalLink, Cpu, HardDrive } from 'lucide-react'

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="bg-bg-secondary border border-border rounded-xl overflow-hidden">
      <div className="px-5 py-3 border-b border-border bg-bg-tertiary">
        <span className="font-mono text-[10px] text-gray-500 tracking-widest uppercase">
          {title}
        </span>
      </div>
      <div className="p-5 space-y-3">{children}</div>
    </div>
  )
}

function ConfigRow({
  label,
  value,
  sub,
  status,
}: {
  label: string
  value: string | number
  sub?: string
  status?: 'ok' | 'warn' | 'error'
}) {
  return (
    <div className="flex items-center justify-between py-2 border-b border-border last:border-0">
      <div>
        <div className="font-mono text-xs text-gray-400">{label}</div>
        {sub && <div className="font-mono text-[10px] text-gray-600 mt-0.5">{sub}</div>}
      </div>
      <div className="flex items-center gap-2">
        {status === 'ok' && <CheckCircle size={12} className="text-emerald-400" />}
        {status === 'warn' && <AlertCircle size={12} className="text-amber-400" />}
        {status === 'error' && <AlertCircle size={12} className="text-red-400" />}
        <span className="font-mono text-xs text-white">{value}</span>
      </div>
    </div>
  )
}

const workspaceLines = [
  '# workspace.yml — JARVINx runtime declaration',
  '# observed → reasoned → acted',
  '',
  'workspace:',
  '  name: jarvinx-local',
  '  version: "1.1"',
  '',
  'runtime:',
  '  loop_interval: 15s',
  '  max_concurrent_agents: 2',
  '  graceful_shutdown: 30s',
  '',
  'llm:',
  '  provider: ollama',
  '  endpoint: http://localhost:11434',
  '  model: llama3.1:8b',
  '  context_window: 8192',
  '',
  'thresholds:',
  '  cpu_warn: 70',
  '  cpu_crit: 85',
  '  ram_warn: 70',
  '  ram_crit: 90',
  '  disk_warn: 70',
  '  disk_crit: 85',
  '',
  'agents:',
  '  - SystemAgent:',
  '    every: 15s',
  '    enabled: true',
  '  - AlertAgent:',
  '    every: 15s',
  '    enabled: true',
  '',
  'notifications:',
  '  discord: ${DISCORD_WEBHOOK}',
]

function getValueColor(val: string): string {
  if (val === 'true') return 'text-emerald-400'
  if (val === 'false') return 'text-red-400'
  if (/^\d/.test(val)) return 'text-blue-400'
  if (val.startsWith('"')) return 'text-amber-400'
  if (val.startsWith('http')) return 'text-blue-400'
  if (val.startsWith('$')) return 'text-purple-400'
  if (val !== '') return 'text-amber-300'
  return ''
}

function WorkspaceLine({ line }: { line: string }) {
  if (line.trim() === '') return <pre className="font-mono text-[11px]">&nbsp;</pre>

  if (line.trim().startsWith('#')) {
    return <pre className="font-mono text-[11px] text-gray-600">{line}</pre>
  }

  if (line.trim().startsWith('-')) {
    return <pre className="font-mono text-[11px] text-gray-400">{line}</pre>
  }

  const colonIdx = line.indexOf(':')
  if (colonIdx > 0) {
    const indent = line.match(/^(\s*)/)?.[1] ?? ''
    const key = line.slice(0, colonIdx).trim()
    const val = line.slice(colonIdx + 1).trim()
    const valColor = getValueColor(val)

    return (
      <pre className="font-mono text-[11px] leading-relaxed">
        <span className="text-gray-500">{indent}</span>
        <span className="text-gray-300">{key}</span>
        <span className="text-gray-500">:</span>
        {val && <span className={cn('ml-1', valColor)}>{val}</span>}
      </pre>
    )
  }

  return <pre className="font-mono text-[11px] text-gray-400">{line}</pre>
}

function WorkspaceViewer() {
  return (
    <div className="bg-bg-secondary border border-border rounded-xl overflow-hidden">
      <div className="flex items-center justify-between px-4 py-2.5 border-b border-border bg-bg-tertiary">
        <div className="flex items-center gap-3">
          <div className="flex gap-1.5">
            <div className="w-3 h-3 rounded-full bg-red-500/70" />
            <div className="w-3 h-3 rounded-full bg-amber-500/70" />
            <div className="w-3 h-3 rounded-full bg-emerald-500/70" />
          </div>
          <span className="font-mono text-[11px] text-gray-500">WORKSPACE.YML</span>
        </div>
        <div className="flex items-center gap-3">
          <div className="flex items-center gap-1.5 bg-emerald-500/10 border border-emerald-500/20 rounded px-2 py-0.5">
            <div className="w-1.5 h-1.5 rounded-full bg-emerald-400" />
            <span className="font-mono text-[9px] text-emerald-400">VALID</span>
          </div>
          <span className="font-mono text-[10px] text-gray-600">
            {workspaceLines.length} lines
          </span>
        </div>
      </div>

      <div className="p-4 overflow-auto max-h-96">
        <table className="w-full border-collapse">
          <tbody>
            {workspaceLines.map((line, i) => (
              <tr key={i} className="hover:bg-bg-tertiary/50 group">
                <td className="pr-4 py-0.5 text-right select-none w-8 align-top">
                  <span className="font-mono text-[10px] text-gray-700 group-hover:text-gray-600">
                    {i + 1}
                  </span>
                </td>
                <td className="py-0.5">
                  <WorkspaceLine line={line} />
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}

const thresholds = [
  { label: 'CPU', icon: Cpu, warn: 70, crit: 85 },
  { label: 'RAM', icon: HardDrive, warn: 70, crit: 90 },
  { label: 'Disk', icon: HardDrive, warn: 70, crit: 85 },
]

const apiEndpoints = [
  { path: '/api/status', desc: 'Dernier cycle + métriques' },
  { path: '/api/history', desc: '10 derniers cycles' },
  { path: '/api/agents', desc: 'Registry agents' },
]

export default function SettingsPage() {
  const { data: status, error } = useStatus()
  const online = !error && status?.online

  const statusClass = online
    ? 'bg-emerald-500/10 border-emerald-500/20 text-emerald-400'
    : 'bg-red-500/10 border-red-500/20 text-red-400'

  const dotClass = online
    ? 'bg-emerald-400 animate-pulse'
    : 'bg-red-400'

  return (
    <div className="space-y-6 max-w-[1400px]">

      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <div className="font-mono text-[10px] text-gray-600 tracking-widest uppercase mb-1">
            Settings · Configuration
          </div>
          <h1 className="text-xl font-semibold text-white">Runtime Configuration</h1>
        </div>
        <div className={cn('flex items-center gap-2 px-3 py-2 rounded-lg border font-mono text-xs', statusClass)}>
          <div className={cn('w-1.5 h-1.5 rounded-full', dotClass)} />
          {online ? 'Runtime connecté' : 'Runtime inaccessible'}
        </div>
      </div>

      <div className="grid grid-cols-2 gap-6">

        {/* Colonne gauche */}
        <div className="space-y-4">
          <Section title="Runtime">
            <ConfigRow label="Statut" value={online ? 'Online' : 'Offline'} status={online ? 'ok' : 'error'} />
            <ConfigRow label="Modèle" value={status?.model ?? '—'} sub="Ollama local" status="ok" />
            <ConfigRow label="Intervalle" value={status?.interval ?? '—'} sub="Fréquence d'observation" />
            <ConfigRow label="Cycle" value={status ? `#${status.cycle_num}` : '—'} sub="Depuis le dernier démarrage" />
            <ConfigRow label="Uptime" value={status?.uptime ?? '—'} />
          </Section>

          <Section title="Alert Thresholds">
            <div className="grid grid-cols-3 gap-3">
              {thresholds.map(({ label, icon: Icon, warn, crit }) => (
                <div key={label} className="bg-bg-tertiary border border-border rounded-lg p-3">
                  <div className="flex items-center gap-2 mb-3">
                    <Icon size={12} className="text-gray-500" />
                    <span className="font-mono text-[10px] text-gray-500 uppercase">{label}</span>
                  </div>
                  <div className="space-y-1.5">
                    <div className="flex justify-between">
                      <span className="font-mono text-[9px] text-gray-600">warn</span>
                      <span className="font-mono text-[11px] text-amber-400">{warn}%</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="font-mono text-[9px] text-gray-600">crit</span>
                      <span className="font-mono text-[11px] text-red-400">{crit}%</span>
                    </div>
                  </div>
                </div>
              ))}
            </div>
            <div className="font-mono text-[10px] text-gray-600 pt-1">
              Modifiable dans{' '}
              <span className="text-accent-blue">runtime/config/config.go</span>
            </div>
          </Section>

          <Section title="Notifications">
            <ConfigRow label="Discord Webhook" value="Configuré" sub="Via DISCORD_WEBHOOK dans .env" status="ok" />
            <ConfigRow label="Alert cooldown" value="5 cycles" sub="Entre deux alertes identiques" />
            <ConfigRow label="Min cycles" value="2 cycles" sub="Consécutifs avant alerte CPU/RAM" />
          </Section>

          <Section title="API Endpoints">
            {apiEndpoints.map(({ path, desc }) => {
              const url = 'http://localhost:8080' + path
              return (
                <div key={path} className="flex items-center justify-between py-1.5 border-b border-border last:border-0">
                  <div>
                    <span className="font-mono text-[11px] text-accent-blue">{path}</span>
                    <span className="font-mono text-[10px] text-gray-600 ml-3">{desc}</span>
                  </div>
                  <a
                    href={url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="p-1 text-gray-600 hover:text-white transition-colors"
                  >
                    <ExternalLink size={12} />
                  </a>
                </div>
              )
            })}
          </Section>
        </div>

        {/* Colonne droite */}
        <div className="space-y-4">
          <div>
            <div className="font-mono text-[10px] text-gray-600 tracking-widest uppercase mb-3">
              Workspace Declaration
            </div>
            <WorkspaceViewer />
          </div>

          <Section title="Stack">
            <ConfigRow label="Runtime" value="Go 1.21+" status="ok" />
            <ConfigRow label="LLM" value="Ollama (local)" status="ok" />
            <ConfigRow label="Dashboard" value="Next.js 14" status="ok" />
            <ConfigRow label="Tests" value="58 tests unitaires" status="ok" />
            <ConfigRow label="Version" value="v1.1 stable" status="ok" />
          </Section>
        </div>
      </div >
    </div >
  )
}