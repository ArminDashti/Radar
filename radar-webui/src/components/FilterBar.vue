<script setup lang="ts">
import { Clock, Globe, Minus, Plus, Search, Wifi } from '@lucide/vue'
import type { Probe } from '@/lib/api'
import type { Interval, Protocol } from '@/lib/latency'
import { useSquareSize } from '@/lib/squareSize'
import MultiSelect from './MultiSelect.vue'

defineProps<{
  interval: Interval
  protocol: Protocol
  probes?: Probe[]
  selectedProbes?: string[]
  hostOptions?: Array<{ value: string; label: string }>
  selectedHosts?: string[]
  hostSearch?: string
  showProbe?: boolean
  showHosts?: boolean
  showHostSearch?: boolean
  title?: string
  subtitle?: string
}>()

const emit = defineEmits<{
  'update:interval': [value: Interval]
  'update:protocol': [value: Protocol]
  'update:selectedProbes': [value: string[]]
  'update:selectedHosts': [value: string[]]
  'update:hostSearch': [value: string]
}>()

const { squareSizePx, canDecrease, canIncrease, decreaseSquareSize, increaseSquareSize } = useSquareSize()

const chipClass =
  'flex items-center gap-2 rounded-lg border border-white/10 bg-slate-900/80 px-3 py-1.5 text-sm'
const selectClass = 'bg-transparent font-medium text-white focus:outline-none cursor-pointer'
</script>

<template>
  <div class="glass-panel rounded-xl p-4">
    <div v-if="title" class="mb-4 min-w-0">
      <h2 class="text-lg font-semibold tracking-tight text-white">{{ title }}</h2>
      <p v-if="subtitle" class="mt-0.5 text-sm text-slate-400">{{ subtitle }}</p>
    </div>

    <div class="flex flex-wrap items-center justify-between gap-4 lg:flex-nowrap">
      <div class="flex w-full flex-wrap items-center gap-3 lg:w-auto">
        <div v-if="showProbe" :class="chipClass">
          <Globe class="h-4 w-4 shrink-0 text-slate-400" aria-hidden="true" />
          <span class="text-xs font-medium text-slate-400">Probes:</span>
          <MultiSelect
            :model-value="selectedProbes ?? []"
            :options="(probes ?? []).map((probe) => ({ value: probe.code, label: probe.code }))"
            label=""
            placeholder="All"
            class="min-w-[6rem]"
            @update:model-value="emit('update:selectedProbes', $event)"
          />
        </div>

        <div :class="chipClass">
          <Wifi class="h-4 w-4 shrink-0 text-slate-400" aria-hidden="true" />
          <span class="text-xs font-medium text-slate-400">Protocol:</span>
          <select
            aria-label="Protocol"
            :class="selectClass"
            :value="protocol"
            @change="emit('update:protocol', ($event.target as HTMLSelectElement).value as Protocol)"
          >
            <option value="http" class="bg-slate-900 text-slate-200">HTTP</option>
            <option value="icmp" class="bg-slate-900 text-slate-200">ICMP</option>
          </select>
        </div>

        <div :class="chipClass">
          <Clock class="h-4 w-4 shrink-0 text-slate-400" aria-hidden="true" />
          <select
            aria-label="Interval"
            :class="selectClass"
            :value="interval"
            @change="emit('update:interval', ($event.target as HTMLSelectElement).value as Interval)"
          >
            <option value="minutes" class="bg-slate-900 text-slate-200">Minutes</option>
            <option value="hours" class="bg-slate-900 text-slate-200">Hours</option>
            <option value="days" class="bg-slate-900 text-slate-200">Days</option>
            <option value="months" class="bg-slate-900 text-slate-200">Months</option>
          </select>
        </div>

        <div v-if="showHostSearch" class="relative min-w-[12rem] flex-1 sm:max-w-xs">
          <Search class="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400" aria-hidden="true" />
          <input
            type="search"
            :value="hostSearch ?? ''"
            placeholder="Filter hosts…"
            class="w-full rounded-lg border border-white/10 bg-slate-900/80 py-1.5 pl-9 pr-3 text-sm text-white placeholder-slate-500 transition-colors focus:border-indigo-500/60 focus:outline-none"
            @input="emit('update:hostSearch', ($event.target as HTMLInputElement).value)"
          />
        </div>

        <MultiSelect
          v-if="showHosts"
          :model-value="selectedHosts ?? []"
          :options="hostOptions ?? []"
          label="Hosts"
          placeholder="All Hosts"
          @update:model-value="emit('update:selectedHosts', $event)"
        />
      </div>

      <div class="flex w-full items-center justify-between gap-3 lg:w-auto lg:justify-end">
        <div class="flex items-center gap-1 rounded-lg border border-white/10 bg-slate-900/80 p-1 text-xs" role="group" aria-label="Cell size">
          <span class="px-2 text-slate-400">Cell size:</span>
          <button
            type="button"
            class="rounded p-1 text-slate-300 transition-colors hover:bg-slate-800 disabled:opacity-40"
            :disabled="!canDecrease"
            aria-label="Decrease cell size"
            @click="decreaseSquareSize"
          >
            <Minus class="h-3.5 w-3.5" />
          </button>
          <span class="min-w-[2.5rem] px-1 text-center font-mono font-semibold tabular-nums text-indigo-400" aria-live="polite">{{ squareSizePx }}px</span>
          <button
            type="button"
            class="rounded p-1 text-slate-300 transition-colors hover:bg-slate-800 disabled:opacity-40"
            :disabled="!canIncrease"
            aria-label="Increase cell size"
            @click="increaseSquareSize"
          >
            <Plus class="h-3.5 w-3.5" />
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
