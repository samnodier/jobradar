<script setup lang="ts">
import ApplicationDetail from '@/components/ApplicationDetail.vue'
import AppSidebar from '@/components/AppSidebar.vue'
import type { Application } from '@/types/application'
import { ArrowLeft, Clipboard, Plus } from '@lucide/vue'
import { computed, onMounted, ref } from 'vue'
import { statusLabels } from '@/constants/applicationStatus'
import { statusOrder } from '@/constants/applicationStatus'

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
    const res = await fetch('/api/applications', {
      credentials: 'include',
    })

    if (!res.ok) {
      error.value = 'Failed to fetch applications'
    }

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
    applications.value[index] = {
      ...applications.value[index],
      ...updatedApp,
    }
  }
}

const applicationByStatus = computed(() => {
  const seed: StatusGroups = Object.fromEntries(
    statusOrder.map((status) => [status, [] as Application[]]),
  )

  return applications.value.reduce((groups, app) => {
    const status = app.application_status
    if (status in groups) {
      groups[status]!.push(app)
    } else {
      groups['other']!.push(app)
    }
    return groups
  }, seed)
})
</script>

<template>
  <div class="applications-page">
    <AppSidebar />

    <main class="applications-main">
      <header class="page-header">
        <div>
          <h1 class="page-title">Applications</h1>
          <p class="page-subtitle">Track every job you've applied to</p>
        </div>

        <button class="button button-primary"><Plus /> Track Application</button>
      </header>

      <!-- Loading -->
      <div v-if="loading" class="state-container">
        <div class="skeleton-list">
          <div v-for="n in 4" :key="n" class="skeleton-row">
            <div class="skeleton skeleton-text-lg"></div>
            <div class="skeleton skeleton-text-sm"></div>
            <div class="skeleton skeleton-badge"></div>
          </div>
        </div>
      </div>

      <!-- Error -->
      <div v-else-if="error" class="state-container">
        <p class="error-text">{{ error }}</p>
      </div>

      <!-- Empty state -->
      <div v-else-if="!hasApplications" class="empty-state">
        <div class="empty-icon"><Clipboard /></div>
        <h2 class="empty-title">No applications yet</h2>
        <p class="empty-subtitle">
          Start tracking jobs you've applied to and follow up at the right time.
        </p>
        <button class="button-primary">Track your first application</button>
      </div>

      <!-- Kanban board -->
      <div class="kanban-board">
        <!-- Outer Loop: Create the columns -->
        <div v-for="status in statusOrder" :key="status" class="kanban-column">
          <div class="column-header">
            <span class="column-title">{{ statusLabels[status] }}</span>
            <span class="column-count">{{ applicationByStatus[status]?.length }}</span>
          </div>
          <div class="column-body">
            <div
              v-for="app in applicationByStatus[status]"
              :key="app.id"
              class="kanban-card"
              @click="selectedApplicationID = app.id"
            >
              <h3 class="card-title">{{ app.job_title }}</h3>
              <p class="card-company">{{ app.company_name }}</p>
            </div>
          </div>
        </div>
      </div>
    </main>

    <!-- Details panel - slides in from the right as overlay -->
    <Teleport to="body">
      <Transition name="detail-slide">
        <div v-if="detailOpen" class="detail-overlay" @click.self="closeDetail">
          <div class="detail-drawer">
            <div class="detail-back" @click="closeDetail">
              <ArrowLeft />
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

<style scoped>
.applications-page {
  display: flex;
  height: 100%;
  background: var(--color-bg-secondary);
  overflow: hidden;
}

.applications-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: var(--spacing-6);
  padding: var(--spacing-8);
  overflow: hidden;
  min-height: 0;
}

/* Header */
.page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--spacing-4);
  padding-bottom: var(--spacing-6);
  border-bottom: 1px solid var(--color-border);
  flex-shrink: 0;
}

.page-title {
  font-size: var(--text-2xl);
  font-weight: var(--font-bold);
  color: var(--color-text-primary);
  line-height: 1.2;
}

.page-subtitle {
  font-size: var(--text-sm);
  color: var(--color-text-secondary);
  margin-top: var(--spacing-1);
}

/* Button */
.button-primary {
  padding: var(--spacing-2) var(--spacing-4);
  background: var(--color-accent);
  color: #fff;
  border: none;
  border-radius: var(--radius-md);
  font-size: var(--text-sm);
  font-weight: var(--font-semibold);
  cursor: pointer;
  white-space: nowrap;
  transition: opacity 0.15s ease;
  display: flex;
  align-items: center;
  justify-content: center;
}

.button-primary:hover {
  opacity: 0.88;
}

/* Loading skeletons */
.state-container {
  flex: 1;
}

.skeleton-list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-3);
}

.skeleton-row {
  display: flex;
  align-items: center;
  gap: var(--spacing-4);
  padding: var(--spacing-4);
  background: var(--color-bg-primary);
  border-radius: var(--radius-md);
}

@keyframes shimmer {
  0% {
    opacity: 0.5;
  }
  50% {
    opacity: 1;
  }
  100% {
    opacity: 0.5;
  }
}

.skeleton {
  background: var(--color-bg-secondary);
  border-radius: var(--radius-sm);
  animation: shimmer 1.5s ease-in-out infinite;
}

.skeleton-text-lg {
  height: 16px;
  width: 180px;
}
.skeleton-text-sm {
  height: 14px;
  width: 120px;
  flex: 1;
}
.skeleton-badge {
  height: 22px;
  width: 72px;
  border-radius: 999px;
}

/* Empty state */
.empty-state {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
  padding: var(--spacing-16) var(--spacing-8);
  gap: var(--spacing-4);
}

.empty-icon {
  font-size: 2.5rem;
}

.empty-title {
  font-size: var(--text-xl);
  font-weight: var(--font-semibold);
  color: var(--color-text-primary);
}

.empty-subtitle {
  font-size: var(--text-sm);
  color: var(--color-text-secondary);
  max-width: 380px;
}

/* Error */
.error-text {
  color: var(--color-danger, #dc2626);
  font-size: var(--text-sm);
  padding: var(--spacing-4);
}

/* KanBan */

.kanban-board {
  display: flex;
  gap: var(--spacing-4);
  overflow-x: auto;
  overflow-y: hidden;
  align-items: flex-start;
  flex: 1;
  min-height: 0;
}

.kanban-column {
  flex: 0 0 300px;
  background: var(--color-bg-secondary);
  display: flex;
  flex-direction: column;
  gap: var(--spacing-2);
  padding: var(--spacing-2);
  height: 100%;
  min-height: 0;
}

.column-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-bottom: var(--spacing-2);
  flex-shrink: 0;
}

.column-title {
  font-size: var(--text-sm);
  font-weight: var(--font-semibold);
  color: var(--color-text-primary);
}

.column-count {
  font-size: var(--text-xs);
  color: var(--color-text-muted);
  background: var(--color-bg-primary);
  padding: 2px 8px;
}

.column-body {
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: var(--spacing-2);
  padding-right: 4px; /* Space for scrollbar */
}

.kanban-card {
  background: var(--color-bg-primary);
  border: 1px solid var(--color-border);
  padding: var(--spacing-3);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
  cursor: pointer;
  flex-shrink: 0;
  transition:
    box-shadow 0.15s ease,
    transform 0.1s ease;
}

.kanban-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.12);
  transform: translateY(-1px);
}

.card-title {
  font-size: var(--text-sm);
  font-weight: var(--font-semibold);
  color: var(--color-text-primary);
  margin: 0 0 var(--spacing-1);
}

.card-company {
  font-size: var(--text-xs);
  color: var(--color-text-muted);
  margin: 0;
}

/* Status badges */
.status-badge {
  display: inline-flex;
  align-items: center;
  padding: 2px var(--spacing-2);
  border-radius: 999px;
  font-size: var(--text-xs);
  font-weight: var(--font-semibold);
  background: var(--color-bg-secondary);
  color: var(--color-text-secondary);
}

/* Detail Overlay */
.detail-overlay {
  position: fixed;
  top: var(--topbar-height) !important;
  inset: 0;
  z-index: 100;
  display: flex;
  justify-content: flex-end;
  pointer-events: none;
}

.detail-drawer {
  width: 50vw;
  min-width: 560px;
  max-width: 100%;
  height: 100%;
  background: var(--color-bg-primary);
  box-shadow: -4px 0 12px rgba(0, 0, 0, 0.1);
  display: flex;
  flex-direction: column;
  overflow-y: auto;
  pointer-events: auto;
}

/* hide on desktop, show on mobile */
.detail-back {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  padding: var(--spacing-3);
  border-bottom: 1px solid var(--color-border);
  color: var(--muted);
  cursor: pointer;
  height: 0;
  padding: 0;
  overflow: hidden;
  visibility: hidden;
  flex-shrink: 0;
  font-size: var(--text-sm);
}
</style>
