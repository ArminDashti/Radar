<script setup lang="ts">
import { computed } from 'vue'
import { ArrowDownRight, CheckCircle2, Globe, Layers, Server, ShieldCheck, Zap } from '@lucide/vue'
import type { GridResponse } from '@/lib/api'
import { formatLatencyMs } from '@/lib/latency'
import { rowCurrentLatencyMs, rowUptimePercent } from '@/lib/rowStats'

const props = defineProps<{
  data: GridResponse | null
  probeCount: number
}>()

const hostCount = computed(() => props.data?.rows.length ?? 0)

const avgUptime = computed(() => {
  const rows = props.data?.rows ?? []
  if (!rows.length) return null
  const values = rows
    .map((row) => rowUptimePercent(row.cells))
    .filter((v): v is number => v != null)
  if (!values.length) return null
  return values.reduce((sum, v) => sum + v, 0) / values.length
})

const avgLatency = computed(() => {
  const rows = props.data?.rows ?? []
  const values = rows
    .map((row) => rowCurrentLatencyMs(row.cells))
    .filter((v): v is number => v != null)
  if (!values.length) return null
  return values.reduce((sum, v) => sum + v, 0) / values.length
})

const operationalCount = computed(() => {
  const rows = props.data?.rows ?? []
  return rows.filter((row) => {
    const ms = rowCurrentLatencyMs(row.cells)
    return ms != null
  }).length
})
</script>

<template>
  <div class="grid grid-cols-2 gap-4 md:grid-cols-4">
    <div class="glass-card flex items-center justify-between rounded-xl p-4">
      <div>
        <p class="text-xs font-medium text-slate-400">Uptime (window)</p>
        <p class="mt-1 text-2xl font-semibold text-white">
          {{ avgUptime != null ? `${avgUptime.toFixed(2)}%` : '—' }}
        </p>
        <span class="mt-1 flex items-center gap-1 text-xs text-emerald-400">
          <ShieldCheck class="h-3 w-3" aria-hidden="true" />
          Sampled buckets
        </span>
      </div>
      <div class="rounded-xl border border-emerald-500/20 bg-emerald-500/10 p-3 text-emerald-400">
        <ShieldCheck class="h-6 w-6" aria-hidden="true" />
      </div>
    </div>

    <div class="glass-card flex items-center justify-between rounded-xl p-4">
      <div>
        <p class="text-xs font-medium text-slate-400">Avg latency</p>
        <p class="mt-1 text-2xl font-semibold text-white">
          {{ avgLatency != null ? formatLatencyMs(avgLatency).replace(' ms', '') : '—' }}
          <span v-if="avgLatency != null" class="text-sm font-normal text-slate-400">ms</span>
        </p>
        <span class="mt-1 flex items-center gap-1 text-xs text-emerald-400">
          <ArrowDownRight class="h-3 w-3" aria-hidden="true" />
          Current window
        </span>
      </div>
      <div class="rounded-xl border border-blue-500/20 bg-blue-500/10 p-3 text-blue-400">
        <Zap class="h-6 w-6" aria-hidden="true" />
      </div>
    </div>

    <div class="glass-card flex items-center justify-between rounded-xl p-4">
      <div>
        <p class="text-xs font-medium text-slate-400">Probes</p>
        <p class="mt-1 text-2xl font-semibold text-white">{{ probeCount }}</p>
        <span class="mt-1 flex items-center gap-1 text-xs text-slate-400">
          <Globe class="h-3 w-3" aria-hidden="true" />
          Edge locations
        </span>
      </div>
      <div class="rounded-xl border border-purple-500/20 bg-purple-500/10 p-3 text-purple-400">
        <Server class="h-6 w-6" aria-hidden="true" />
      </div>
    </div>

    <div class="glass-card flex items-center justify-between rounded-xl p-4">
      <div>
        <p class="text-xs font-medium text-slate-400">Monitored hosts</p>
        <p class="mt-1 text-2xl font-semibold text-white">{{ hostCount }}</p>
        <span class="mt-1 flex items-center gap-1 text-xs text-emerald-400">
          <CheckCircle2 class="h-3 w-3" aria-hidden="true" />
          {{ operationalCount }} responding
        </span>
      </div>
      <div class="rounded-xl border border-indigo-500/20 bg-indigo-500/10 p-3 text-indigo-400">
        <Layers class="h-6 w-6" aria-hidden="true" />
      </div>
    </div>
  </div>
</template>
