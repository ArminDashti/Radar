<script setup lang="ts">
import { computed, ref } from 'vue'
import type { GridCell } from '@/lib/api'
import type { Interval } from '@/lib/latency'
import { formatBucketTime, formatLatencyMs, latencyColor } from '@/lib/latency'

const props = defineProps<{ cell: GridCell | null; bucket: string; interval: Interval; tall?: boolean }>()
const hovering = ref(false)
const tipX = ref(0)
const tipY = ref(0)

const latency = computed(() => props.cell?.ok && props.cell.latency_ms != null ? props.cell.latency_ms : null)
const timedOut = computed(() => props.cell != null && !props.cell.ok)
const latencyHex = computed(() => (latency.value != null ? latencyColor(latency.value) : ''))

const tooltip = computed(() => {
  const time = formatBucketTime(props.bucket, props.interval)
  if (latency.value !== null) return { time, main: formatLatencyMs(latency.value), status: 'OK' }
  if (timedOut.value) return { time, main: 'Timeout', status: 'Error' }
  return null
})

function showTip(event: MouseEvent) {
  if (!tooltip.value) return
  hovering.value = true
  tipX.value = event.clientX + 12
  tipY.value = event.clientY + 12
}

function hideTip() { hovering.value = false }
</script>

<template>
  <span
    class="latency-block relative block shrink-0 overflow-visible"
    :style="{
      width: 'var(--sq-size)',
      height: tall ? 'calc(var(--sq-size) * 1.8)' : 'var(--sq-size)',
      minWidth: 'var(--sq-size)',
      minHeight: tall ? 'calc(var(--sq-size) * 1.8)' : 'var(--sq-size)',
      flexShrink: 0,
    }"
    :aria-label="tooltip ? `${tooltip.time} · ${tooltip.main}` : 'No sample'"
    @mouseenter="showTip"
    @mouseleave="hideTip"
  >
    <span
      v-if="latency !== null"
      class="block h-full w-full rounded-sm ring-1 ring-black/10"
      :style="{ backgroundColor: latencyHex }"
    />
    <span
      v-else-if="timedOut"
      class="grid h-full w-full place-items-center rounded-sm border border-rose-500/60 bg-rose-950 text-[8px] font-bold text-rose-300"
      aria-hidden="true"
    >✕</span>
    <span
      v-else
      class="block h-full w-full rounded-sm bg-white/[0.04] ring-1 ring-white/10"
      aria-hidden="true"
    />
    <Teleport to="body">
      <div
        v-if="hovering && tooltip"
        class="glass-card pointer-events-none fixed z-50 min-w-[10rem] rounded-lg border border-white/20 p-3 text-xs shadow-2xl"
        :style="{ left: `${tipX}px`, top: `${tipY}px` }"
      >
        <div class="mb-1.5 flex items-center justify-between gap-4 border-b border-white/10 pb-1.5">
          <span class="font-mono text-slate-400">{{ tooltip.time }}</span>
        </div>
        <div class="flex items-center justify-between gap-4">
          <span class="text-slate-400">Latency</span>
          <span class="font-mono font-semibold" :style="{ color: latencyHex || '#f43f5e' }">{{ tooltip.main }}</span>
        </div>
        <div class="mt-1 flex items-center justify-between gap-4">
          <span class="text-slate-400">Status</span>
          <span class="font-mono text-slate-300">{{ tooltip.status }}</span>
        </div>
      </div>
    </Teleport>
  </span>
</template>
