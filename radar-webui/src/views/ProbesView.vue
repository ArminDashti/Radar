<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
import FilterBar from '@/components/FilterBar.vue'
import LatencyGrid from '@/components/LatencyGrid.vue'
import LatencyLegend from '@/components/LatencyLegend.vue'
import { api, type GridResponse } from '@/lib/api'
import type { Interval, Protocol } from '@/lib/latency'

const interval = ref<Interval>('minutes')
const protocol = ref<Protocol>('http')
const data = ref<GridResponse | null>(null)
const bucketWindow = ref(0)
const loading = ref(false)
const error = ref('')
let timer: ReturnType<typeof setInterval> | undefined
let loadGeneration = 0

function gridUrl() {
  const params = new URLSearchParams({
    interval: interval.value,
    protocol: protocol.value,
  })
  if (bucketWindow.value > 0) params.set('window', String(bucketWindow.value))
  return `/api/grid/probes?${params}`
}

async function load() {
  const generation = ++loadGeneration
  loading.value = true
  error.value = ''
  try {
    const next = await api<GridResponse>(gridUrl())
    if (generation !== loadGeneration) return
    data.value = next
  } catch (reason) {
    if (generation !== loadGeneration) return
    error.value = reason instanceof Error ? reason.message : 'Could not load probe grid'
  } finally {
    if (generation === loadGeneration) loading.value = false
  }
}

function onCapacity(count: number) {
  if (count <= 0 || count === bucketWindow.value) return
  bucketWindow.value = count
  void load()
}

watch([interval, protocol], load)
onMounted(async () => {
  await load()
  timer = setInterval(load, 30_000)
})
onBeforeUnmount(() => clearInterval(timer))
</script>

<template>
  <div class="flex min-h-[calc(100dvh-12rem)] flex-col gap-6">
    <FilterBar
      v-model:interval="interval"
      v-model:protocol="protocol"
      title="Probes"
      subtitle="Latency by probe"
    />
    <LatencyLegend />
    <div v-if="error && data" class="rounded-lg border border-red-500/30 bg-red-500/10 px-4 py-2 text-sm text-red-400">{{ error }}</div>
    <section class="glass-panel flex min-h-0 flex-1 flex-col overflow-hidden rounded-xl border border-white/10 shadow-2xl shadow-black/40">
      <LatencyGrid
        :data="data"
        :interval="interval"
        :loading="loading"
        :error="error"
        boxed-rows
        modern-rows
        @capacity="onCapacity"
      />
    </section>
  </div>
</template>
