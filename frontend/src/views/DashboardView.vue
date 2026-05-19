<template>
  <div class="flex h-full bg-black/4">
    <AppSidebar />
    <main class="flex-1 p-8 flex flex-col gap-6 overflow-y-auto">
      <header class="border-b border-gray-200 pb-4">
        <h1 class="text-2xl font-bold">Dashboard</h1>
        <p class="text-sm text-gray-500 mt-1">Your personalized overview is coming soon.</p>
      </header>

      <div class="grid gap-4" style="grid-template-columns: repeat(auto-fit, minmax(300px, 1fr))">
        <div v-for="i in 3" :key="i" class="h-50 bg-white border border-gray-200 rounded-md" />
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import AppSidebar from '@/components/AppSidebar.vue'
</script>
