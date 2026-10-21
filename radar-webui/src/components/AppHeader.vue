<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Activity, LogIn, LogOut, Moon, Monitor, Sun } from '@lucide/vue'
import { clearToken, isAdmin, isAuthenticated } from '@/lib/auth'
import { cycleTheme, theme } from '@/lib/theme'
import { visibleHostCount } from '@/lib/hostCount'

const route = useRoute()
const router = useRouter()

const settingsTarget = computed(() => {
  if (isAdmin.value) return '/admin/hosts'
  if (isAuthenticated.value) return '/probes'
  return '/login'
})

const navItems = computed(() => {
  const items = [
    { to: '/hosts', label: 'Dashboard', match: (p: string) => p === '/hosts' },
    { to: '/probes', label: 'Probes', match: (p: string) => p === '/probes', auth: true },
    { to: settingsTarget.value, label: 'Settings', match: (p: string) => p.startsWith('/admin') || p === '/request-host' },
    { to: '/about-me', label: 'About', match: (p: string) => p === '/about-me' },
  ]
  return items.filter((item) => !item.auth || isAuthenticated.value)
})

function isActive(match: (p: string) => boolean) {
  return match(route.path)
}

function logout() {
  clearToken()
  void router.push('/login')
}
</script>

<template>
  <header class="glass-panel sticky top-0 z-40 w-full border-b border-white/10">
    <div class="mx-auto flex h-16 max-w-[1700px] items-center justify-between gap-4 px-4 sm:px-6 lg:px-8">
      <div class="flex min-w-0 items-center gap-4 md:gap-6">
        <RouterLink to="/hosts" class="flex items-center gap-3">
          <span class="relative flex h-9 w-9 items-center justify-center rounded-xl bg-gradient-to-tr from-indigo-600 to-blue-500 text-white shadow-lg shadow-indigo-500/20">
            <Activity class="h-5 w-5" aria-hidden="true" />
          </span>
          <span class="hidden text-xl font-bold tracking-tight text-white sm:inline">Radar</span>
        </RouterLink>
        <div
          v-if="visibleHostCount != null"
          class="hidden items-center gap-2 rounded-full border border-emerald-500/30 bg-emerald-950/40 px-3 py-1 text-xs font-medium text-emerald-400 md:flex"
        >
          <span class="relative flex h-2 w-2">
            <span class="absolute inline-flex h-full w-full animate-ping rounded-full bg-emerald-400 opacity-75" />
            <span class="relative inline-flex h-2 w-2 rounded-full bg-emerald-500" />
          </span>
          <span>{{ visibleHostCount }} hosts monitored</span>
        </div>
      </div>

      <nav class="hidden items-center gap-1 lg:flex" aria-label="Main">
        <RouterLink
          v-for="item in navItems"
          :key="item.label"
          :to="item.to"
          class="rounded-lg px-3 py-1.5 text-sm font-medium transition-colors"
          :class="isActive(item.match)
            ? 'bg-white/10 text-white'
            : 'text-slate-400 hover:bg-white/5 hover:text-slate-200'"
        >{{ item.label }}</RouterLink>
      </nav>

      <div class="flex items-center gap-2">
        <button
          type="button"
          class="rounded-lg border border-white/10 bg-slate-800/60 p-2 text-slate-300 transition-colors hover:bg-slate-700/60 hover:text-white"
          :title="`Theme: ${theme}. Click to cycle.`"
          :aria-label="`Theme: ${theme}`"
          @click="cycleTheme"
        >
          <Moon v-if="theme === 'dark'" class="h-4 w-4" />
          <Sun v-else-if="theme === 'light'" class="h-4 w-4" />
          <Monitor v-else class="h-4 w-4" />
        </button>
        <div class="mx-1 hidden h-4 w-px bg-white/10 sm:block" aria-hidden="true" />
        <button
          v-if="isAuthenticated"
          type="button"
          class="flex items-center gap-2 rounded-lg border border-white/10 bg-slate-800/60 px-3 py-1.5 text-sm font-medium text-slate-200 transition-colors hover:bg-slate-700/60"
          @click="logout"
        >
          <LogOut class="h-4 w-4" aria-hidden="true" />
          <span class="hidden sm:inline">Logout</span>
        </button>
        <button
          v-else
          type="button"
          class="flex items-center gap-2 rounded-lg bg-gradient-to-r from-indigo-600 to-indigo-500 px-3.5 py-1.5 text-sm font-medium text-white shadow-lg shadow-indigo-500/20 transition-all hover:from-indigo-500 hover:to-indigo-400"
          @click="router.push('/login')"
        >
          <LogIn class="h-4 w-4" aria-hidden="true" />
          <span>Account</span>
        </button>
      </div>
    </div>
  </header>
</template>
