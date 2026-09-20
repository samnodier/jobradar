<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import AppSidebar from '@/components/AppSidebar.vue'
interface SystemStatus {
  database: boolean
  redis: boolean
  pending_matches: number
  sources: { name: string; status: string; last_success: string | null }[]
}
const status = ref<SystemStatus | null>(null)
const error = ref('')
let timer: ReturnType<typeof setInterval> | undefined
async function refresh() {
  try {
    const response = await fetch('/api/system/status', { credentials: 'include' })
    if (!response.ok) throw new Error('Service status unavailable')
    status.value = await response.json()
    error.value = ''
  } catch {
    error.value = 'Could not reach the status endpoint'
    status.value = null
  }
}
onMounted(() => {
  refresh()
  timer = setInterval(refresh, 30000)
})
onUnmounted(() => clearInterval(timer))
</script>
<template>
  <div class="flex h-full">
    <AppSidebar />
    <main class="flex-1 p-8 overflow-y-auto space-y-4">
      <h1 class="text-2xl font-bold">System status</h1>
      <p v-if="error" role="alert">{{ error }}</p>
      <template v-if="status"
        ><p>PostgreSQL: {{ status.database ? 'connected' : 'unavailable' }}</p>
        <p>Redis: {{ status.redis ? 'connected' : 'unavailable' }}</p>
        <p>Matching tasks waiting for completion: {{ status.pending_matches }}</p>
        <h2 class="font-semibold">Job sources</h2>
        <p v-if="!status.sources.length">No source has completed a fetch yet.</p>
        <div v-for="source in status.sources" :key="source.name" class="border p-4">
          <strong>{{ source.name }}: {{ source.status }}</strong>
          <p>
            Last successful fetch:
            {{ source.last_success ? new Date(source.last_success).toLocaleString() : 'not yet' }}
          </p>
        </div>
      </template>
    </main>
  </div>
</template>
