<template>
  <div class="flex flex-wrap items-center justify-between gap-4 px-1 text-xs text-slate-400">
    <span class="font-medium">Latency spectrum:</span>
    <div class="flex flex-wrap items-center gap-4">
      <span class="flex items-center gap-1.5">
        <span class="inline-block h-3 w-3 rounded-sm bg-emerald-500 shadow-sm shadow-emerald-500/30" />
        ≤50 ms
      </span>
      <span class="flex items-center gap-1.5">
        <span class="inline-block h-3 w-3 rounded-sm bg-blue-500" />
        51–100 ms
      </span>
      <span class="flex items-center gap-1.5">
        <span class="inline-block h-3 w-3 rounded-sm bg-cyan-500" />
        101–200 ms
      </span>
      <span class="flex items-center gap-1.5">
        <span class="inline-block h-3 w-3 rounded-sm bg-amber-500" />
        201–500 ms
      </span>
      <span class="flex items-center gap-1.5">
        <span class="inline-block h-3 w-3 rounded-sm bg-rose-500" />
        &gt;500 ms
      </span>
      <span class="flex items-center gap-1.5">
        <span class="inline-flex h-3 w-3 items-center justify-center rounded-sm border border-rose-500/50 bg-rose-950 text-[9px] font-bold text-rose-300">✕</span>
        Timeout / outage
      </span>
    </div>
  </div>
</template>
