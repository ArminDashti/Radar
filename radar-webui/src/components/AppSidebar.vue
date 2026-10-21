<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Activity, ChevronLeft, ChevronRight, Info, Server, Settings } from '@lucide/vue'
import { isAdmin, isAuthenticated } from '@/lib/auth'
import { visibleHostCount } from '@/lib/hostCount'

const props = defineProps<{ collapsed: boolean }>()
const emit = defineEmits<{ 'update:collapsed': [value: boolean] }>()

const route = useRoute()
const router = useRouter()

const settingsTarget = computed(() => {
  if (isAdmin.value) return '/admin/hosts'
  if (isAuthenticated.value) return '/probes'
  return '/login'
})

const navItems = computed(() => [
  {
    to: '/hosts',
    label: 'Hosts',
    icon: Server,
    badge: visibleHostCount.value,
    active: route.path === '/hosts',
  },
  {
    to: settingsTarget.value,
    label: 'Settings',
    icon: Settings,
    badge: null as number | null,
    active: route.path.startsWith('/admin') || route.path === '/probes' || route.path === '/request-host',
  },
  {
    to: '/about-me',
    label: 'About',
    icon: Info,
    badge: null,
    active: route.path === '/about-me',
  },
])

function toggle() {
  emit('update:collapsed', !props.collapsed)
}

function go(to: string) {
  void router.push(to)
}
</script>

<template>
  <aside
    class="flex shrink-0 flex-col border-r border-border/60 bg-[hsl(220,18%,7%)] transition-[width] duration-200"
    :class="collapsed ? 'w-[4.25rem]' : 'w-[13.5rem]'"
  >
    <div class="flex items-center gap-2 px-3 py-4">
      <RouterLink to="/hosts" class="flex min-w-0 flex-1 items-center gap-2.5 overflow-hidden">
        <span class="grid h-9 w-9 shrink-0 place-items-center rounded-xl bg-primary/15 text-primary">
          <Activity class="h-5 w-5" aria-hidden="true" />
        </span>
        <span v-if="!collapsed" class="truncate text-lg font-semibold tracking-tight">Radar</span>
      </RouterLink>
      <button
        type="button"
        class="grid h-8 w-8 shrink-0 place-items-center rounded-lg text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
        :aria-label="collapsed ? 'Expand sidebar' : 'Collapse sidebar'"
        @click="toggle"
      >
        <ChevronRight v-if="collapsed" class="h-4 w-4" />
        <ChevronLeft v-else class="h-4 w-4" />
      </button>
    </div>

    <nav class="flex flex-1 flex-col gap-1 px-2" aria-label="Sidebar">
      <button
        v-for="item in navItems"
        :key="item.label"
        type="button"
        class="flex w-full items-center gap-3 rounded-xl px-3 py-2.5 text-sm font-medium transition-colors"
        :class="item.active
          ? 'bg-primary/15 text-primary'
          : 'text-muted-foreground hover:bg-accent/80 hover:text-foreground'"
        :title="collapsed ? item.label : undefined"
        @click="go(item.to)"
      >
        <component :is="item.icon" class="h-[1.125rem] w-[1.125rem] shrink-0" aria-hidden="true" />
        <span v-if="!collapsed" class="truncate">{{ item.label }}</span>
        <span
          v-if="!collapsed && item.badge != null"
          class="ml-auto rounded-md bg-primary/20 px-2 py-0.5 text-xs font-semibold tabular-nums text-primary"
        >{{ item.badge }}</span>
      </button>
    </nav>

    <div v-if="!collapsed" class="mt-auto px-3 pb-5 pt-6">
      <div
        class="relative overflow-hidden rounded-2xl border border-border/40 bg-gradient-to-b from-slate-900/80 to-[hsl(220,22%,5%)] px-3 py-6"
        aria-hidden="true"
      >
        <div class="absolute inset-x-0 bottom-0 h-16 bg-gradient-to-t from-indigo-950/40 to-transparent" />
        <div class="relative h-14">
          <div class="absolute bottom-2 left-1 w-20 h-6 rounded-t-full bg-slate-800/90" />
          <div class="absolute bottom-3 left-8 w-24 h-8 rounded-t-full bg-slate-700/80" />
          <div class="absolute bottom-1 right-2 w-16 h-5 rounded-t-full bg-slate-800/70" />
          <div class="absolute right-4 top-1 h-3 w-3 rotate-12 rounded-full bg-amber-200/30 blur-[1px]" />
        </div>
        <p class="relative mt-2 text-center text-[11px] leading-snug text-muted-foreground">
          Keep an eye on what matters.
        </p>
      </div>
    </div>
  </aside>
</template>
