<script setup lang="ts">
import { onMounted, ref } from 'vue'

const mailto = ref('')
const photoAvailable = ref(true)
const portraitSrc = '/about-me/armin.png'
onMounted(() => {
  const user = 'arminonline71'; const at = '@'; const domain = 'gmail'; const dot = '.'; const tld = 'com'
  mailto.value = `mailto:${user}${at}${domain}${dot}${tld}`
})
</script>

<template>
  <article class="h-full min-h-0 overflow-auto rounded-2xl border border-border/50 bg-card/30 px-5 py-10 shadow-xl shadow-black/20"><div class="mx-auto max-w-3xl"><header><h1 class="text-4xl font-bold tracking-tight">About Me</h1><p class="mt-2 text-lg text-muted-foreground">Armin Dashti — vibe coder, conductor of craft.</p></header>
    <section class="mt-10 grid gap-8 sm:grid-cols-[220px_1fr]"><div><img v-if="photoAvailable" :src="portraitSrc" alt="Portrait of Armin Dashti" class="aspect-[4/5] w-full rounded-lg border object-cover shadow-sm" @error="photoAvailable = false" /><div v-else class="grid aspect-[4/5] place-items-center rounded-lg border bg-muted p-5 text-center text-sm text-muted-foreground">Photo placeholder — add public/about-me/armin.png</div></div><div class="space-y-5 leading-7 text-muted-foreground"><p>Armin Dashti is a software engineer and vibe coder who shapes intent with clarity and constraint, then lets Cursor’s agents compose the implementation. He reviews, refines, and ships only what remains worthy.</p><p>This application was written by AI agents—specifically Cursor—guided by Armin’s intent, then tempered by his review until fit to ship.</p></div></section>
    <section class="mt-10"><h2 class="text-2xl font-semibold">Interests</h2><p class="mt-3 leading-7 text-muted-foreground">Cars, movies, vibe coding, coding, politics, military history and technology, jet fighters, and aircraft.</p></section>
    <p class="mt-8 text-sm"><span class="font-semibold">Contact:</span> <a :href="mailto" class="text-primary underline-offset-4 hover:underline">arminonline71 [at] gmail [dot] com</a></p>
    <section class="mt-12 space-y-5"><blockquote class="border-l-4 border-primary pl-5 text-lg italic leading-8">"Beware the quiet man. For while others speak, he watches. While others act, he plans. And when they finally rest, he strikes." — Anonymous</blockquote><blockquote dir="rtl" lang="fa" class="border-r-4 border-primary pr-5 text-lg leading-9">«از مرد خاموش برحذر باش. زیرا در حالی که دیگران سخن می‌گویند، او نظاره می‌کند. در حالی که دیگران عمل می‌کنند، او برنامه می‌ریزد. و هنگامی که سرانجام آرام می‌گیرند، او ضربه می‌زند.» — ناشناس</blockquote></section>
  </div></article>
</template>
