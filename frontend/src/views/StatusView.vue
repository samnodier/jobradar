<script setup lang="ts">
import { computed, ref, onMounted, onUnmounted } from 'vue'
import { RefreshCw } from '@lucide/vue'
import AppSidebar from '@/components/AppSidebar.vue'
interface Day {
  day: string
  successes: number
  failures: number
}
interface Source {
  name: string
  status: string
  detail?: string
  jobs_last_run?: number | null
  last_success: string | null
  last_checked?: string | null
  history?: Day[]
}
interface SystemStatus {
  deployment?: {
    revision: string
    started_at: string
    update?: { status: string; target_revision: string; checked_at: string }
    backup?: { completed_at: string }
  }
  alerts?: { code: string; detail: string }[]
  monitor?: { checked_at: string; email_status: string; email_detail: string }
  history_available?: boolean
  database: boolean
  redis: boolean
  pending_matches: number
  sources: Source[]
}
const status = ref<SystemStatus | null>(null)
const error = ref('')
const loading = ref(false)
const lastChecked = ref<Date | null>(null)
const now = ref(Date.now())
let timer: ReturnType<typeof setInterval> | undefined
let request: AbortController | undefined
const stale = computed(() => !!lastChecked.value && now.value - lastChecked.value.getTime() > 90000)
function sourceState(source: Source) {
  const value = source.status.toLowerCase()
  if (['ok', 'healthy', 'success', 'operational'].includes(value)) return 'success'
  if (['error', 'failed', 'failing', 'failure', 'unhealthy', 'degraded', 'down'].includes(value))
    return 'failure'
  return 'unknown'
}
const banner = computed(() => {
  if (!status.value) return { text: 'Waiting for a health check', tone: 'unknown' }
  if (error.value || stale.value) return { text: 'Status may be out of date', tone: 'unknown' }
  if (
    status.value.monitor?.email_status === 'unconfirmed' ||
    status.value.alerts?.some((alert) => alert.code !== 'publication_pending') ||
    !status.value.database ||
    !status.value.redis ||
    status.value.sources.some((source) => sourceState(source) === 'failure')
  )
    return { text: 'Some services need attention', tone: 'failure' }
  if (
    status.value.history_available === false ||
    !status.value.sources.length ||
    status.value.sources.some((source) => sourceState(source) === 'unknown')
  )
    return { text: 'Core services connected · source status incomplete', tone: 'unknown' }
  return { text: 'Core services connected · latest fetches successful', tone: 'success' }
})
const days = computed(() => {
  const today = new Date(lastChecked.value ?? now.value)
  return Array.from({ length: 30 }, (_, index) => {
    const day = new Date(today)
    day.setUTCDate(day.getUTCDate() - 29 + index)
    return day.toISOString().slice(0, 10)
  })
})
function history(source: Source) {
  return days.value.map((day) => {
    const entry = source.history?.find((row) => row.day === day)
    const successes = entry?.successes ?? 0
    const failures = entry?.failures ?? 0
    return {
      day,
      successes,
      failures,
      tone: failures ? (successes ? 'mixed' : 'failure') : successes ? 'success' : 'unknown',
      label: `${day} UTC: ${successes + failures ? `${successes} successful, ${failures} failed fetches` : 'No recorded fetches'}`,
    }
  })
}
function rate(source: Source) {
  const counts = history(source).reduce(
    (sum, day) => ({
      successes: sum.successes + day.successes,
      total: sum.total + day.successes + day.failures,
    }),
    { successes: 0, total: 0 },
  )
  return counts.total
    ? `${((counts.successes / counts.total) * 100).toFixed(1)}% fetch success · ${counts.total} recorded attempts`
    : 'No recorded attempts in this period'
}
function time(value?: string | null) {
  return value && !value.startsWith('0001-') ? new Date(value).toLocaleString() : 'Not recorded'
}
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
    now.value = Date.now()
    error.value = ''
  } catch {
    if (!controller.signal.aborted) error.value = 'Could not refresh system status. Try again.'
  } finally {
    if (!controller.signal.aborted) loading.value = false
  }
}
onMounted(() => {
  void refresh()
  timer = setInterval(() => {
    now.value = Date.now()
    void refresh()
  }, 30000)
})
onUnmounted(() => {
  clearInterval(timer)
  request?.abort()
})
</script>
<template>
  <div class="flex h-full">
    <AppSidebar />
    <main class="flex-1 min-w-0 overflow-y-auto p-5 md:p-8" :aria-busy="loading">
      <div class="mx-auto max-w-5xl space-y-7">
        <header class="flex flex-wrap items-start justify-between gap-4">
          <div>
            <h1 class="text-2xl font-bold">System health</h1>
            <p class="mt-1 text-sm text-gray-500">
              Job collection and matching · refreshes every 30 seconds
            </p>
          </div>
          <button
            type="button"
            class="button button-secondary gap-2"
            :disabled="loading"
            @click="refresh"
          >
            <RefreshCw :size="16" aria-hidden="true" />{{ loading ? 'Refreshing…' : 'Refresh' }}
          </button>
        </header>
        <section
          class="border p-5 md:p-7"
          :class="
            banner.tone === 'success'
              ? 'border-green-200 bg-green-50'
              : banner.tone === 'failure'
                ? 'border-red-200 bg-red-50'
                : 'border-gray-200 bg-gray-50'
          "
          aria-label="Health overview"
        >
          <h2 role="status" class="text-lg font-semibold">{{ banner.text }}</h2>
          <p class="mt-2 text-sm text-gray-600">
            {{
              lastChecked
                ? `Snapshot checked ${lastChecked.toLocaleString()}`
                : 'No successful check yet.'
            }}
          </p>
          <p v-if="error" role="alert" class="mt-2 text-sm text-red-700">
            {{ error }} <span v-if="status">Showing the last successful snapshot.</span>
          </p>
        </section>
        <template v-if="status">
          <section
            v-if="status.deployment"
            class="border border-gray-200 p-5 space-y-2"
            aria-label="Deployment"
          >
            <h2 class="font-semibold">Running deployment</h2>
            <p class="text-sm break-all">
              Version: <code>{{ status.deployment.revision }}</code>
            </p>
            <p class="text-sm text-gray-500">
              Application started: {{ time(status.deployment.started_at) }}
            </p>
            <template v-if="status.deployment.update?.status">
              <p class="text-sm">
                Latest update check: {{ status.deployment.update.status.replace(/_/g, ' ') }} ·
                {{ time(status.deployment.update.checked_at) }}
              </p>
              <p class="text-xs break-all">
                Selected release: <code>{{ status.deployment.update.target_revision }}</code>
              </p>
            </template>
            <p
              v-if="status.monitor?.checked_at && !status.monitor.checked_at.startsWith('0001')"
              class="text-sm"
            >
              Operational monitor: {{ time(status.monitor.checked_at) }} · Email
              {{ (status.monitor.email_status || 'not recorded').replace(/_/g, ' ') }}
            </p>
            <p v-if="status.monitor?.email_detail" class="text-sm text-amber-800">
              {{ status.monitor.email_detail }}
            </p>
            <p class="text-sm">
              Last server backup: {{ time(status.deployment.backup?.completed_at) }}
            </p>
            <p class="text-xs text-gray-500">
              This version comes from the running application. A successful health check alone does
              not prove the latest release is running.
            </p>
          </section>
          <section
            v-if="status.alerts?.length"
            class="border border-amber-200 bg-amber-50 p-5 space-y-2"
            aria-label="Operational alerts"
          >
            <h2 class="font-semibold">Needs attention</h2>
            <ul class="space-y-2 text-sm">
              <li v-for="alert in status.alerts" :key="alert.code">{{ alert.detail }}</li>
            </ul>
            <p class="text-xs text-gray-500">
              The server monitor can email these alerts when configured. An outage of the whole
              server or its email provider needs an independent monitor.
            </p>
          </section>
          <section
            class="border border-gray-200 divide-y divide-gray-200"
            aria-label="Core services"
          >
            <div class="service-row">
              <div>
                <h2 class="font-semibold">Database</h2>
                <p class="text-sm text-gray-500">Jobs, profiles and applications</p>
              </div>
              <span :class="status.database ? 'text-green-700' : 'text-red-700'">{{
                status.database ? 'Connected' : 'Unavailable'
              }}</span>
            </div>
            <div class="service-row">
              <div>
                <h2 class="font-semibold">Queue and sessions</h2>
                <p class="text-sm text-gray-500">Redis sessions and background work</p>
              </div>
              <span :class="status.redis ? 'text-green-700' : 'text-red-700'">{{
                status.redis ? 'Connected' : 'Unavailable'
              }}</span>
            </div>
            <div class="service-row">
              <div>
                <h2 class="font-semibold">Matching backlog</h2>
                <p class="text-sm text-gray-500">Profile changes can add work to this queue</p>
              </div>
              <span>{{ status.pending_matches }} pending</span>
            </div>
          </section>
          <section class="space-y-4" aria-label="Job source history">
            <p
              v-if="status.history_available === false"
              role="alert"
              class="text-sm text-amber-700"
            >
              Source history is temporarily unavailable. Current service checks are shown below.
            </p>
            <div>
              <h2 class="text-lg font-semibold">Job sources</h2>
              <p class="text-sm text-gray-500">
                30 days of recorded fetch attempts, grouped by UTC day. This measures fetch results,
                not provider uptime. Refresh does not fetch new jobs.
              </p>
            </div>
            <div
              class="flex flex-wrap gap-x-5 gap-y-2 text-xs text-gray-600"
              aria-label="History legend"
            >
              <span
                v-for="item in [
                  { tone: 'success', label: 'Successful' },
                  { tone: 'mixed', label: 'Mixed results' },
                  { tone: 'failure', label: 'Failed' },
                  { tone: 'unknown', label: 'No data' },
                ]"
                :key="item.tone"
                class="inline-flex items-center gap-2"
                ><span class="history-key" :class="`history-${item.tone}`" aria-hidden="true" />{{
                  item.label
                }}</span
              >
            </div>
            <p v-if="!status.sources.length" class="border border-gray-200 p-5 text-gray-500">
              No source has recorded a fetch yet.
            </p>
            <div v-else class="border border-gray-200 divide-y divide-gray-200">
              <article v-for="source in status.sources" :key="source.name" class="p-5 space-y-3">
                <div class="flex flex-wrap justify-between items-baseline gap-2">
                  <h3 class="font-semibold break-words">{{ source.name }}</h3>
                  <span
                    class="text-sm"
                    :class="
                      sourceState(source) === 'success'
                        ? 'text-green-700'
                        : sourceState(source) === 'failure'
                          ? 'text-red-700'
                          : 'text-gray-500'
                    "
                    >Latest result: {{ source.status || 'Unknown' }}</span
                  >
                </div>
                <p v-if="source.detail" class="text-sm text-amber-800">{{ source.detail }}</p>
                <div class="history-bars" :aria-label="`${source.name} daily fetch results`">
                  <span
                    v-for="day in history(source)"
                    :key="day.day"
                    tabindex="0"
                    role="img"
                    class="history-cell"
                    :class="`history-${day.tone}`"
                    :aria-label="day.label"
                    :title="day.label"
                  />
                </div>
                <div class="flex flex-wrap justify-between gap-2 text-xs text-gray-500">
                  <span>{{ days[0] }} → today (UTC)</span><span>{{ rate(source) }}</span>
                </div>
                <details class="text-sm text-gray-600">
                  <summary class="cursor-pointer text-accent">Fetch details</summary>
                  <div class="mt-2 space-y-1">
                    <p>Last attempt: {{ time(source.last_checked) }}</p>
                    <p v-if="source.jobs_last_run != null">
                      Jobs saved or updated on that attempt: {{ source.jobs_last_run }}
                    </p>
                    <p>Last success: {{ time(source.last_success) }}</p>
                    <p>
                      Gray days have no recorded attempts. Historical results begin when collection
                      history was enabled.
                    </p>
                  </div>
                </details>
              </article>
            </div>
          </section>
        </template>
      </div>
    </main>
  </div>
</template>
<style scoped>
.service-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 1.25rem;
  font-size: 0.875rem;
}
.history-bars {
  display: grid;
  grid-template-columns: repeat(30, minmax(0, 1fr));
  gap: 3px;
}
.history-cell {
  display: block;
  height: 28px;
}
.history-cell:focus-visible {
  outline: 2px solid var(--color-accent);
  outline-offset: 2px;
}
.history-key {
  display: inline-block;
  width: 9px;
  height: 9px;
}
.history-success {
  background: #16a34a;
}
.history-mixed {
  background: #d97706;
}
.history-failure {
  background: #dc2626;
}
.history-unknown {
  background: #d1d5db;
}
</style>
