<script setup lang="ts">
import AppHeader from '@/components/AppHeader.vue'
import AppFooter from '@/components/AppFooter.vue'
</script>

<template>
  <div class="flex min-h-dvh flex-col bg-[#080c14] text-slate-100">
    <AppHeader />
    <main class="mx-auto w-full max-w-[1700px] flex-1 space-y-6 px-4 py-6 sm:px-6 lg:px-8">
      <RouterView />
    </main>
    <AppFooter />
  </div>
</template>
