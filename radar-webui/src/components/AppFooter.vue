<template>
  <footer class="glass-panel mt-auto border-t border-white/10 py-5">
    <div class="mx-auto flex max-w-[1700px] flex-col items-center justify-between gap-3 px-4 text-xs text-slate-400 sm:flex-row sm:px-6 lg:px-8">
      <div class="flex items-center gap-2">
        <span class="font-semibold text-slate-200">Radar</span>
        <span aria-hidden="true">•</span>
        <span>Real-time latency &amp; availability</span>
      </div>
      <div class="flex items-center gap-6">
        <RouterLink to="/hosts" class="transition-colors hover:text-slate-200">Hosts</RouterLink>
        <RouterLink to="/about-me" class="transition-colors hover:text-slate-200">About</RouterLink>
      </div>
    </div>
  </footer>
</template>
