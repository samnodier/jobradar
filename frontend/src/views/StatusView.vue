<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { RefreshCw } from '@lucide/vue'
import AppSidebar from '@/components/AppSidebar.vue'
interface SystemStatus {
  database: boolean
  redis: boolean
  pending_matches: number
  sources: { name: string; status: string; last_success: string | null }[]
}
const status = ref<SystemStatus | null>(null)
const error = ref('')
const loading = ref(false)
const lastChecked = ref<Date | null>(null)
let timer: ReturnType<typeof setInterval> | undefined
let request: AbortController | undefined
async function refresh() {
  if (loading.value) return
  const controller = new AbortController()
  request = controller
  loading.value = true
  try {
    const response = await fetch('/api/system/status', {
      credentials: 'include',
      signal: controller.signal,
    })
    if (!response.ok) throw new Error('Service status unavailable')
    const result: SystemStatus = await response.json()
    if (controller.signal.aborted) return
    status.value = result
    lastChecked.value = new Date()
    error.value = ''
  } catch {
    if (!controller.signal.aborted) error.value = 'Could not refresh system status. Try again.'
  } finally {
    if (!controller.signal.aborted) loading.value = false
  }
}
onMounted(() => {
  void refresh()
  timer = setInterval(refresh, 30000)
})
onUnmounted(() => {
  clearInterval(timer)
  request?.abort()
})
</script>
<template>
  <div class="flex h-full">
    <AppSidebar />
    <main class="flex-1 min-w-0 p-8 overflow-y-auto space-y-6" :aria-busy="loading">
      <header class="flex flex-wrap items-start justify-between gap-4">
        <div>
          <h1 class="text-2xl font-bold">System health</h1>
          <p class="mt-1 text-sm text-gray-500">
            Check job collection and matching. Updates every 30 seconds.
          </p>
        </div>
        <button type="button" class="button button-secondary" :disabled="loading" @click="refresh">
          <RefreshCw :size="16" aria-hidden="true" />
          {{ loading ? 'Refreshing…' : 'Refresh' }}
        </button>
      </header>
      <p v-if="error" role="alert" class="border border-red-200 bg-red-50 p-4 text-red-700">
        {{ error }} <span v-if="status">Showing the last successful check below.</span>
      </p>
      <p role="status" aria-live="polite" class="text-sm text-gray-500">
        <template v-if="lastChecked">Last checked: {{ lastChecked.toLocaleString() }}</template>
        <template v-else-if="loading">Checking system health…</template>
        <template v-else>No successful check yet.</template>
      </p>
      <template v-if="status">
        <div class="grid gap-6 md:grid-cols-2 xl:grid-cols-3">
          <section class="settings-card">
            <h2 class="font-semibold">Database</h2>
            <p
              class="text-lg font-semibold"
              :class="status.database ? 'text-green-700' : 'text-red-700'"
            >
              {{ status.database ? 'Connected' : 'Unavailable' }}
            </p>
            <p class="text-sm text-gray-500">
              PostgreSQL stores jobs, profiles, applications and source history.
            </p>
          </section>
          <section class="settings-card">
            <h2 class="font-semibold">Queue and sessions</h2>
            <p
              class="text-lg font-semibold"
              :class="status.redis ? 'text-green-700' : 'text-red-700'"
            >
              {{ status.redis ? 'Connected' : 'Unavailable' }}
            </p>
            <p class="text-sm text-gray-500">
              Redis supports sign-in sessions and background work delivery.
            </p>
          </section>
          <section class="settings-card">
            <h2 class="font-semibold">Pending matches</h2>
            <p class="text-2xl font-semibold">{{ status.pending_matches }}</p>
            <p class="text-sm text-gray-500">
              Jobs waiting for matching to complete. Profile changes can increase this count.
            </p>
          </section>
        </div>
        <section class="space-y-4">
          <div>
            <h2 class="text-lg font-semibold">Job sources</h2>
            <p class="text-sm text-gray-500">
              Each row shows the last recorded fetch result, not a live uptime check of the
              provider. Refresh checks this stored history; it does not fetch new jobs.
            </p>
          </div>
          <p v-if="!status.sources.length" class="settings-card">
            No source has recorded a fetch yet.
          </p>
          <div v-else class="grid gap-6 md:grid-cols-2 xl:grid-cols-3">
            <article v-for="source in status.sources" :key="source.name" class="settings-card">
              <h3 class="font-semibold break-words">{{ source.name }}</h3>
              <p class="text-sm">
                Last recorded result: <strong>{{ source.status }}</strong>
              </p>
              <p class="text-sm text-gray-500">
                Last successful fetch:
                {{
                  source.last_success ? new Date(source.last_success).toLocaleString() : 'Not yet'
                }}
              </p>
            </article>
          </div>
        </section>
      </template>
    </main>
  </div>
</template>
