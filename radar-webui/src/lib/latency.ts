export type Interval = 'minutes' | 'hours' | 'days' | 'months'
export type Protocol = 'http' | 'icmp'

/** Latency cell size limits; user preference is shared across intervals. */
export const minSquareSizePx = 9
export const maxSquareSizePx = 32
export const defaultSquareSizePx = 12
export const squareSizeStepPx = 2

export function clampSquareSizePx(value: number): number {
  if (!Number.isFinite(value)) return defaultSquareSizePx
  const stepped = Math.round(value / squareSizeStepPx) * squareSizeStepPx
  return Math.min(maxSquareSizePx, Math.max(minSquareSizePx, stepped))
}

/** How many fixed-size squares fit in a strip (1px gaps), matching API window max. */
export const maxBucketWindow = 720

export function bucketCapacity(stripWidthPx: number, squareSizePx: number, gapPx = 1): number {
  if (stripWidthPx <= 0 || squareSizePx <= 0) return 0
  const count = Math.floor((stripWidthPx + gapPx) / (squareSizePx + gapPx))
  return Math.min(maxBucketWindow, Math.max(0, count))
}

export function formatLatencyMs(ms: number): string {
  if (ms >= 0 && ms < 1) return '<1 ms'
  return `${Math.round(ms)} ms`
}

export function latencyColor(latency: number) {
  if (latency <= 50) return '#10b981'
  if (latency <= 100) return '#3b82f6'
  if (latency <= 200) return '#06b6d4'
  if (latency <= 500) return '#f59e0b'
  return '#ef4444'
}

const clock: Intl.DateTimeFormatOptions = { hour: '2-digit', minute: '2-digit', hour12: false, hourCycle: 'h23' }

export function formatBucketTime(value: string, interval: Interval): string {
  const date = new Date(value)
  if (interval === 'minutes') return date.toLocaleTimeString([], clock)
  if (interval === 'hours') return date.toLocaleString([], { weekday: 'short', ...clock })
  if (interval === 'days') return date.toLocaleDateString([], { day: 'numeric', month: 'short' })
  return date.toLocaleDateString([], { month: 'short', year: '2-digit' })
}
