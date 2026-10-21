import type { GridCell } from '@/lib/api'
import { formatLatencyMs } from '@/lib/latency'

export function rowUptimePercent(cells: Array<GridCell | null>): number | null {
  const sampled = cells.filter((cell) => cell != null)
  if (!sampled.length) return null
  const ok = sampled.filter((cell) => cell!.ok).length
  return (ok / sampled.length) * 100
}

export function rowCurrentLatencyMs(cells: Array<GridCell | null>): number | null {
  for (let i = cells.length - 1; i >= 0; i--) {
    const cell = cells[i]
    if (cell?.ok && cell.latency_ms != null) return cell.latency_ms
  }
  return null
}

export function formatUptimePercent(value: number | null): string {
  if (value === null) return '—'
  return `${value.toFixed(1)}%`
}

export function formatCurrentLatency(value: number | null): string {
  if (value === null) return '—'
  return formatLatencyMs(value)
}
