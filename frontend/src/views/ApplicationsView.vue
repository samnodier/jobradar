<script setup lang="ts">
import { ArrowLeft, Clipboard, Plus } from '@lucide/vue'
import { computed, onMounted, ref } from 'vue'
import { statusLabels } from '@/constants/applicationStatus'
import { statusOrder } from '@/constants/applicationStatus'
import type { Application } from '@/types/application'
import ApplicationDetail from '@/components/ApplicationDetail.vue'
import AppSidebar from '@/components/AppSidebar.vue'

const applications = ref<Application[]>([])
const selectedApplicationID = ref<string | null>(null)
const loading = ref(false)
const error = ref('')

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
    if (!res.ok) error.value = 'Failed to fetch applications'
    applications.value = (await res.json()) ?? []
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Something went wrong.'
  } finally {
    loading.value = false
  }
})

const hasApplications = computed(() => applications.value.length > 0)

function handleApplicationUpdate(updatedApp: Application) {
  const index = applications.value.findIndex((a) => a.id === updatedApp.id)
  if (index !== -1) {
    applications.value[index] = { ...applications.value[index], ...updatedApp }
  }
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
          <p class="text-sm text-gray-500 mt-1">Track every job you've applied to</p>
        </div>
        <RouterLink to="/jobs" class="button button-primary">
          <Plus />
          Track Application
        </RouterLink>
      </header>

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
        <button class="button button-primary">Track your first application</button>
      </div>

      <!-- Kanban board -->
      <div class="flex gap-4 overflow-x-auto overflow-y-hidden items-start flex-1 min-h-0">
        <div
          v-for="status in statusOrder"
          :key="status"
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
              class="bg-white border border-gray-200 p-3 shadow-[0_1px_3px_rgba(0,0,0,0.08)] cursor-pointer shrink-0 transition-all hover:shadow-[0_4px_12px_rgba(0,0,0,0.12)] hover:-translate-y-px"
              @click="selectedApplicationID = app.id"
            >
              <h3 class="text-sm font-semibold text-gray-900 mb-1">{{ app.job_title }}</h3>
              <p class="text-xs text-gray-400 m-0">{{ app.company_name }}</p>
            </div>
          </div>
        </div>
      </div>
    </main>

    <!-- Detail overlay -->
    <Teleport to="body">
      <Transition name="detail-slide">
        <div
          v-if="detailOpen"
          class="fixed inset-0 z-100 flex justify-end pointer-events-none"
          style="top: var(--topbar-height)"
          @click.self="closeDetail"
        >
          <div
            class="w-[50vw] min-w-140 max-w-full h-full bg-white flex flex-col overflow-y-auto pointer-events-auto shadow-[-4px_0_12px_rgba(0,0,0,0.1)]"
          >
            <!-- Back: hidden on desktop, shown on mobile -->
            <div
              class="hidden items-center gap-2 p-3 border-b border-gray-200 text-gray-400 cursor-pointer text-sm shrink-0"
              @click="closeDetail"
            >
              <ArrowLeft class="w-4 h-4" />
              Back to jobs
            </div>
            <ApplicationDetail
              v-if="selectedApplication"
              :app="selectedApplication"
              @close="closeDetail"
              @updated="handleApplicationUpdate"
            />
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>
