<template>
  <div class="flex h-full bg-black/0.04">
    <AppSidebar />
    <main class="flex-1 p-8 flex flex-col gap-6 overflow-y-auto">
      <header class="border-b border-gray-200 pb-4">
        <h1 class="text-2xl font-bold">System Status</h1>
        <p class="text-sm text-gray-500 mt-1">Real-time health monitoring for JobRadar services.</p>
      </header>

      <div class="flex flex-col gap-3">
        <div
          v-for="service in services"
          :key="service.name"
          class="flex items-center gap-4 bg-white p-4 border border-gray-200"
        >
          <div
            class="w-3 h-3 shrink-0"
            :class="{
              'bg-green-500': service.status === 'up',
              'bg-yellow-400': service.status === 'warning',
              'bg-red-500': service.status === 'down',
            }"
          />
          <div>
            <h3 class="text-sm font-semibold">{{ service.name }}</h3>
            <p class="text-xs text-gray-400">{{ service.statusText }}</p>
          </div>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import AppSidebar from '@/components/AppSidebar.vue'

const services = [
  { name: 'API Server', status: 'up', statusText: 'Healthy' },
  { name: 'Database', status: 'up', statusText: 'Healthy' },
  { name: 'Redis Cache', status: 'up', statusText: 'Healthy' },
  { name: 'Scrapers', status: 'warning', statusText: 'Degraded' },
]
</script>
