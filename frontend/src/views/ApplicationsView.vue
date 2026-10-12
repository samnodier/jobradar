<script setup lang="ts">
import { Clipboard, Plus } from '@lucide/vue'
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { statusLabels } from '@/constants/applicationStatus'
import { statusOrder } from '@/constants/applicationStatus'
import type { Application } from '@/types/application'
import ApplicationDetail from '@/components/ApplicationDetail.vue'
import AppSidebar from '@/components/AppSidebar.vue'

const applications = ref<Application[]>([])
const selectedApplicationID = ref<string | null>(null)
const loading = ref(false)
const error = ref('')
const moveError = ref('')
const announcement = ref('')
const movingIDs = ref(new Set<string>())
const draggedID = ref<string | null>(null)
const dropStatus = ref<string | null>(null)
const editableStatuses = statusOrder.filter((status) => status !== 'other')

type StatusGroups = Record<string, Application[]>

const selectedApplication = computed(() => {
  if (!selectedApplicationID.value) return null
  return (
    applications.value.find((application) => application.id === selectedApplicationID.value) ?? null
  )
})

const detailOpen = computed(() => !!selectedApplicationID.value)

function closeDetail() {
  selectedApplicationID.value = null
}

onMounted(async () => {
  loading.value = true
  error.value = ''
  try {
    const res = await fetch('/api/applications', { credentials: 'include' })
    if (!res.ok) throw new Error('Failed to fetch applications')
    applications.value = (await res.json()) ?? []
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Something went wrong.'
  } finally {
    loading.value = false
  }
})

const clock = ref(Date.now())
let reminderTimer: ReturnType<typeof setInterval> | undefined
onMounted(() => { reminderTimer = setInterval(() => { clock.value = Date.now() }, 60000) })
onUnmounted(() => clearInterval(reminderTimer))
const dueFollowUps = computed(() => applications.value.filter(app =>
  app.follow_up_at && new Date(app.follow_up_at).getTime() <= clock.value &&
  !['accepted', 'rejected', 'withdrawn'].includes(app.application_status),
).sort((a, b) => new Date(a.follow_up_at!).getTime() - new Date(b.follow_up_at!).getTime()))
const outcomeRows = computed(() => {
  const rows = new Map<string, { source: string; band: string; total: number; interviews: number; offers: number; rejected: number }>()
  for (const app of applications.value) {
    const source = app.job_source ?? 'Unknown source'
    const band = app.match_score == null ? 'Unscored' : app.match_score < 60 ? '0–59' : app.match_score < 80 ? '60–79' : '80–100'
    const key = `${source}:${band}`
    const row = rows.get(key) ?? { source, band, total: 0, interviews: 0, offers: 0, rejected: 0 }
    row.total++
    if (app.application_status === 'interview') row.interviews++
    if (['offer', 'accepted'].includes(app.application_status)) row.offers++
    if (app.application_status === 'rejected') row.rejected++
    rows.set(key, row)
  }
  return [...rows.values()].sort((a, b) => a.source.localeCompare(b.source) || a.band.localeCompare(b.band))
})
const hasApplications = computed(() => applications.value.length > 0)

function handleApplicationUpdate(updatedApp: Partial<Application> & { id: string }) {
  const existing = applications.value.find((a) => a.id === updatedApp.id)
  if (existing) Object.assign(existing, updatedApp)
}

async function moveApplication(id: string, status: string) {
  const app = applications.value.find((item) => item.id === id)
  if (
    !app ||
    movingIDs.value.has(id) ||
    !editableStatuses.includes(status) ||
    app.application_status === status
  )
    return
  const previousStatus = app.application_status
  movingIDs.value.add(id)
  moveError.value = ''
  app.application_status = status
  announcement.value = `Moving ${app.job_title} to ${statusLabels[status]}…`
  try {
    const response = await fetch(`/api/applications/${id}`, {
      method: 'PATCH',
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ application_status: status }),
    })
    if (!response.ok) throw new Error('Could not move application. Please try again.')
    const saved = await response.json()
    handleApplicationUpdate({
      id,
      application_status: saved.application_status,
      last_status_changed_at: saved.last_status_changed_at,
      updated_at: saved.updated_at,
    })
    announcement.value = `${app.job_title} moved to ${statusLabels[status]}.`
  } catch (err) {
    app.application_status = previousStatus
    announcement.value = ''
    moveError.value = err instanceof Error ? err.message : 'Could not move application.'
  } finally {
    movingIDs.value.delete(id)
  }
}

function startDrag(event: DragEvent, id: string) {
  if (movingIDs.value.has(id)) {
    event.preventDefault()
    return
  }
  draggedID.value = id
  if (event.dataTransfer) {
    event.dataTransfer.effectAllowed = 'move'
    event.dataTransfer.setData('text/plain', id)
  }
}

function endDrag() {
  draggedID.value = null
  dropStatus.value = null
}

function allowDrop(event: DragEvent, status: string) {
  if (!draggedID.value || !editableStatuses.includes(status)) return
  event.preventDefault()
  dropStatus.value = status
  if (event.dataTransfer) event.dataTransfer.dropEffect = 'move'
}

function dropApplication(event: DragEvent, status: string) {
  event.preventDefault()
  const id = draggedID.value
  endDrag()
  if (id) void moveApplication(id, status)
}

const applicationByStatus = computed(() => {
  const seed: StatusGroups = Object.fromEntries(
    statusOrder.map((status) => [status, [] as Application[]]),
  )
  return applications.value.reduce((groups, app) => {
    const status = app.application_status
    if (status in groups) groups[status]!.push(app)
    else groups['other']!.push(app)
    return groups
  }, seed)
})
</script>

<template>
  <div class="flex h-full bg-black/4 overflow-hidden">
    <AppSidebar />

    <main class="flex-1 flex flex-col gap-6 p-8 overflow-hidden min-h-0">
      <!-- Header -->
      <header class="flex items-start justify-between gap-4 pb-6 border-b border-gray-200 shrink-0">
        <div>
          <h1 class="text-2xl font-bold text-gray-900 leading-tight">Applications</h1>
          <p class="text-sm text-gray-500 mt-1">
            Drag applications between stages, or use each card’s status selector.
          </p>
        </div>
        <RouterLink to="/jobs" class="button button-primary">
          <Plus />
          Track Application
        </RouterLink>
      </header>

      <section v-if="!loading && !error && dueFollowUps.length" class="border border-amber-200 bg-amber-50 p-4 shrink-0" aria-label="Due follow-ups">
        <h2 class="font-semibold">Follow-ups due · {{ dueFollowUps.length }}</h2>
        <ul class="mt-2 space-y-1 max-h-32 overflow-y-auto">
          <li v-for="app in dueFollowUps" :key="app.id">
            <button class="text-sm underline text-accent" @click="selectedApplicationID = app.id">{{ app.job_title }} · {{ app.company_name }} · {{ new Date(app.follow_up_at!).toLocaleDateString() }}</button>
          </li>
        </ul>
        <p class="text-xs text-gray-500 mt-2">Open the application to reschedule or clear its follow-up date after contacting the employer.</p>
      </section>
      <details v-if="!loading && !error && hasApplications" class="border border-gray-200 p-4 shrink-0">
        <summary class="cursor-pointer font-semibold">Outcomes by source and score</summary>
        <p class="text-xs text-gray-500 my-2">Current stages and current scores, not historical conversion rates. Scores can change when your profile or feedback changes.</p>
        <div class="overflow-auto max-h-48">
          <table class="w-full text-sm text-left"><thead><tr><th>Source</th><th>Score</th><th>Tracked</th><th>Interview now</th><th>Offer / accepted</th><th>Rejected</th></tr></thead>
          <tbody><tr v-for="row in outcomeRows" :key="`${row.source}:${row.band}`"><td>{{ row.source }}</td><td>{{ row.band }}</td><td>{{ row.total }}</td><td>{{ row.interviews }}</td><td>{{ row.offers }}</td><td>{{ row.rejected }}</td></tr></tbody></table>
        </div>
      </details>
      <!-- Loading skeletons -->
      <div v-if="loading" class="flex-1">
        <div class="flex flex-col gap-3">
          <div v-for="n in 4" :key="n" class="flex items-center gap-4 p-4 bg-white">
            <div class="h-4 w-180px bg-black/0.04 animate-pulse"></div>
            <div class="h-14px flex-1 bg-black/0.04 animate-pulse"></div>
            <div class="h-22px w-72px bg-black/0.04 animate-pulse"></div>
          </div>
        </div>
      </div>

      <!-- Error -->
      <div v-else-if="error" class="flex-1">
        <p class="text-sm text-red-600 p-4">{{ error }}</p>
      </div>

      <!-- Empty state -->
      <div
        v-else-if="!hasApplications"
        class="flex-1 flex flex-col items-center justify-center text-center px-8 py-16 gap-4"
      >
        <div class="text-[2.5rem] text-gray-400">
          <Clipboard />
        </div>
        <h2 class="text-xl font-semibold text-gray-900">No applications yet</h2>
        <p class="text-sm text-gray-500 max-w-95">
          Start tracking jobs you've applied to and follow up at the right time.
        </p>
        <RouterLink to="/jobs" class="button button-primary"
          >Track your first application</RouterLink
        >
      </div>

      <p v-if="moveError" role="alert" class="text-sm text-red-600">{{ moveError }}</p>
      <p role="status" aria-live="polite" class="sr-only">{{ announcement }}</p>
      <!-- Kanban board -->
      <div
        v-if="!loading && !error && hasApplications"
        class="flex gap-4 overflow-x-auto overflow-y-hidden items-start flex-1 min-h-0"
      >
        <div
          v-for="status in statusOrder"
          :key="status"
          :data-status="status"
          :class="{ 'outline outline-2 outline-accent': dropStatus === status }"
          @dragover="allowDrop($event, status)"
          @drop="dropApplication($event, status)"
          class="shrink-0 w-75 bg-black/4 flex flex-col gap-2 p-2 h-full min-h-0"
        >
          <!-- Column header -->
          <div class="flex items-center justify-between pb-2 shrink-0">
            <span class="text-sm font-semibold text-gray-900">{{ statusLabels[status] }}</span>
            <span class="text-xs text-gray-400 bg-white px-2 py-0.5">
              {{ applicationByStatus[status]?.length }}
            </span>
          </div>

          <!-- Column cards -->
          <div class="flex-1 overflow-y-auto flex flex-col gap-2 pr-1">
            <div
              v-for="app in applicationByStatus[status]"
              :key="app.id"
              :data-application-id="app.id"
              :draggable="!movingIDs.has(app.id)"
              :aria-busy="movingIDs.has(app.id)"
              @dragstart="startDrag($event, app.id)"
              @dragend="endDrag"
              class="bg-white border border-gray-200 p-3 shadow-[0_1px_3px_rgba(0,0,0,0.08)] cursor-pointer shrink-0 transition-all hover:shadow-[0_4px_12px_rgba(0,0,0,0.12)] hover:-translate-y-px"
            >
              <button
                type="button"
                class="text-sm font-semibold text-gray-900 mb-1 text-left hover:underline"
                @click="selectedApplicationID = app.id"
              >
                {{ app.job_title }}
              </button>
              <p class="text-xs text-gray-400 m-0">{{ app.company_name }}</p>
              <label class="block text-xs mt-3">
                Move to
                <select
                  :aria-label="`Status for ${app.job_title}`"
                  :value="app.application_status"
                  :disabled="movingIDs.has(app.id)"
                  class="block w-full border border-gray-200 p-1 mt-1 bg-white"
                  @change="moveApplication(app.id, ($event.target as HTMLSelectElement).value)"
                >
                  <option v-for="target in editableStatuses" :key="target" :value="target">
                    {{ statusLabels[target] }}
                  </option>
                </select>
              </label>
              <span v-if="movingIDs.has(app.id)" class="text-xs text-gray-500">Saving…</span>
            </div>
          </div>
        </div>
      </div>
    </main>

    <!-- Job Detail overlay -->
    <Teleport to="body">
      <Transition name="detail-slide">
        <div
          v-if="detailOpen"
          class="fixed inset-0 z-100 bg-black/45 flex justify-end top-(--topbar-height)"
          @click.self="closeDetail"
        >
          <ApplicationDetail
            v-if="selectedApplication"
            :app="selectedApplication"
            :status-saving="movingIDs.has(selectedApplication.id)"
            @status-change="moveApplication(selectedApplication.id, $event)"
            @close="closeDetail"
            @updated="handleApplicationUpdate"
          />
        </div>
      </Transition>
    </Teleport>
  </div>
</template>
