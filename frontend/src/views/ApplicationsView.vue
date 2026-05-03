<script setup lang="ts">
import AppSidebar from '@/components/AppSidebar.vue'
import type { Application } from '@/types/application'
// import { useAuthStore } from '@/stores/auth'
import { Clipboard, Plus } from '@lucide/vue'
import { computed, onMounted, ref } from 'vue'

// const authStore = useAuthStore()

const applications = ref<Application[]>([])
const loading = ref(false)
const error = ref('')

const statusLabels: Record<string, string> = {
  saved: 'Saved',
  applied: 'Applied',
  interview: 'Interview',
  offer: 'Offer',
  rejected: 'Rejected',
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

      <!-- Applications table -->
      <div v-else class="applications-table-wrapper">
        <table class="applications-table">
          <thead>
            <tr>
              <th>Company</th>
              <th>Role</th>
              <th>Status</th>
              <th>Applied</th>
              <th>Follow up</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="app in applications" :key="app.id" class="application-row">
              <td class="td-company">{{ app.job_company }}</td>
              <td class="td-role">{{ app.job_title }}</td>
              <td></td>
              <td>
                <span class="status-badge">
                  {{ statusLabels[app.status] ?? app.status }}
                </span>
              </td>
              <td class="td-date">
                {{ app.applied_at ? new Date(app.applied_at).toLocaleDateString() : '—' }}
              </td>
              <td class="td-date">
                {{ app.follow_up_at ? new Date(app.follow_up_at).toLocaleDateString() : '—' }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </main>
  </div>
</template>

<style scoped>
.applications-page {
  display: flex;
  min-height: 100vh;
  background: var(--color-bg-secondary);
}

.applications-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: var(--spacing-6);
  padding: var(--spacing-8);
  overflow-y: auto;
}

/* Header */
.page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--spacing-4);
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

/* Table */
.applications-table-wrapper {
  background: var(--color-bg-primary);
  border-radius: var(--radius-lg);
  overflow: hidden;
  border: 1px solid var(--color-border);
}

.applications-table {
  width: 100%;
  border-collapse: collapse;
}

.applications-table thead tr {
  background: var(--color-bg-secondary);
  border-bottom: 1px solid var(--color-border);
}

.applications-table th {
  padding: var(--spacing-3) var(--spacing-4);
  text-align: left;
  font-size: var(--text-xs);
  font-weight: var(--font-semibold);
  color: var(--color-text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.application-row {
  border-bottom: 1px solid var(--color-border);
  transition: background 0.12s ease;
  cursor: pointer;
}

.application-row:last-child {
  border-bottom: none;
}

.application-row:hover {
  background: var(--color-bg-secondary);
}

.application-row td {
  padding: var(--spacing-3) var(--spacing-4);
  font-size: var(--text-sm);
  color: var(--color-text-primary);
}

.td-company {
  font-weight: var(--font-semibold);
}

.td-date {
  color: var(--color-text-secondary);
  font-size: var(--text-xs);
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
</style>
