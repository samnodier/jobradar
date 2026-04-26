<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import AppSidebar from '@/components/AppSidebar.vue'
import JobRow from '@/components/JobRow.vue'
import JobDetail from '@/components/JobDetail.vue'
import type { Job } from '@/types/job'

const route = useRoute()
const router = useRouter()
const jobs = ref<Job[]>([])
const loading = ref(false)
const error = ref('')
const selectedJobId = ref<string | null>(null)
const sortDirection = ref<'newest' | 'oldest'>('newest')

const activeFilter = ref<string>((route.query.filter as string) || 'all')

watch(
  () => route.query.filter,
  (value) => {
    activeFilter.value = (Array.isArray(value) ? value[0] : value) || 'all'
  },
)

watch(activeFilter, (value) => {
  if (value === 'all') {
    router.replace({ query: {} }).catch(() => { })
  } else {
    router.replace({
      query: {
        filter: value === 'all' ? undefined : value,
        q: route.query.q, 
      } 
    }).catch(() => { })
  }
})

const filteredJobs = computed(() => {
  if (activeFilter.value === 'saved') {
    return jobs.value.filter((job) => job.status === 'saved')
  }
  if (activeFilter.value === 'applied') {
    return jobs.value.filter((job) => job.status === 'applied')
  }
  return jobs.value
})

const sortedJobs = computed(() => {
  return [...filteredJobs.value].sort((a, b) => {
    const left = new Date(a.posted_at ?? '').getTime()
    const right = new Date(b.posted_at ?? '').getTime()
    return sortDirection.value === 'newest' ? right - left : left - right
  })
})

const groupedJobs = computed(() => {
  return sortedJobs.value.reduce<Record<string, Job[]>>((groups, job) => {
    const label = job.posted_at ? new Date(job.posted_at).toLocaleDateString(undefined, { month: 'short', day: 'numeric' }) : 'Unknown'
    groups[label] = groups[label] || []
    groups[label].push(job)
    return groups
  }, {})
})

const selectedJob = computed(() => {
  if (!selectedJobId.value) return null
  return jobs.value.find((job) => job.id === selectedJobId.value) ?? null
})

const searchQuery = computed(() =>
  String(route.query.search || '').toLowerCase()
)



function fetchJobs() {
  loading.value = true
  error.value = ''
  fetch('/api/jobs')
    .then(async (response) => {
      if (!response.ok) throw new Error('Failed to load jobs')
      jobs.value = await response.json()
      if (!selectedJobId.value && jobs.value.length) {
        selectedJobId.value = jobs.value[0]?.id || null
      }
    })
    .catch((err) => {
      error.value = err instanceof Error ? err.message : 'Unknown error'
    })
    .finally(() => {
      loading.value = false
    })
}

function formatCount(count: number) {
  return count === 1 ? '1 role' : `${count} roles`
}

function toggleSort() {
  sortDirection.value = sortDirection.value === 'newest' ? 'oldest' : 'newest'
}

function statusLabel(job: Job) {
  if (job.status === 'saved') return 'Saved'
  if (job.status === 'applied') return 'Applied'
  return 'Open'
}

function formatLocation(job: Job) {
  const parts: string[] = []
  if (job.location) parts.push(job.location)
  if (job.is_remote) parts.push('Remote')
  return parts.join(' • ')
}

onMounted(fetchJobs)
</script>

<template>
  <div class="jobs-shell">
    <AppSidebar />

    <main class="jobs-main">
      <header class="jobs-topbar">
        <div>
          <p class="eyebrow">Talent board</p>
          <h1>Job search with focus.</h1>
          <p class="page-copy">Filter, sort, and preview roles in one clean workspace.</p>
        </div>

        <div class="topbar-actions">
          <button class="button button-secondary" @click="toggleSort">
            Sort: {{ sortDirection === 'newest' ? 'Newest' : 'Oldest' }}
          </button>
          <span class="total-count">{{ formatCount(filteredJobs.length) }}</span>
        </div>
      </header>

      <section class="jobs-body">
        <div class="jobs-list-panel">
          <div class="jobs-filter-bar">
            <button :class="['filter-pill', { active: activeFilter === 'all' }]"
              @click="activeFilter = 'all'">All</button>
            <button :class="['filter-pill', { active: activeFilter === 'saved' }]"
              @click="activeFilter = 'saved'">Saved</button>
            <button :class="['filter-pill', { active: activeFilter === 'applied' }]"
              @click="activeFilter = 'applied'">Applied</button>
          </div>

          <div class="job-list-scroll">
            <div v-if="loading" class="status-card">Loading jobs…</div>
            <div v-else-if="error" class="status-card status-error">{{ error }}</div>
            <div v-else-if="filteredJobs.length === 0" class="status-card">No jobs match this filter.</div>

            <div v-else>
              <div v-for="(group, label) in groupedJobs" :key="label" class="job-group">
                <div class="group-label">{{ label }}</div>
                <div>
                  <JobRow v-for="job in group" :key="job.id" :job="job" :selected="job.id === selectedJobId"
                    @click="selectedJobId = job.id" />
                </div>
              </div>
            </div>
          </div>
        </div>

        <div class="jobs-detail-panel">
          <JobDetail v-if="selectedJob" :job="selectedJob" @close="selectedJobId = null" />
          <div v-else class="status-card">Select a job to preview.</div>
        </div>
      </section>
    </main>
  </div>
</template>

<style scoped>
.jobs-shell {
  width: min(1200px, 100%);
  margin: 0 auto;
  display: grid;
  grid-template-columns: 220px minmax(0, 1fr);
  min-height: calc(100vh - 96px);
  gap: 1.5rem;
}

.jobs-main {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  padding: 24px 28px;
}

.jobs-topbar {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 1rem;
}

.eyebrow {
  margin: 0 0 0.4rem;
  text-transform: uppercase;
  font-size: 0.78rem;
  letter-spacing: 0.12em;
  color: #64748b;
}

h1 {
  margin: 0;
  font-size: clamp(2rem, 2.6vw, 3rem);
}

.page-copy {
  max-width: 560px;
  color: #475569;
  margin-top: 0.75rem;
  line-height: 1.75;
}

.topbar-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.75rem;
}

.total-count {
  font-weight: 600;
  color: #1f2937;
}

.jobs-body {
  display: grid;
  grid-template-columns: 0.95fr 0.6fr;
  gap: 1.25rem;
  min-height: 0;
}

.jobs-list-panel {
  display: flex;
  flex-direction: column;
  background: #fff;
  border: 1px solid #e8e8e4;
  border-radius: 28px;
  overflow: hidden;
  min-height: 0;
}

.jobs-filter-bar {
  display: flex;
  gap: 0.75rem;
  padding: 18px 20px;
  border-bottom: 1px solid #f0f0ec;
  background: #fafaf8;
}

.filter-pill {
  border: 1px solid transparent;
  border-radius: 999px;
  padding: 0.75rem 1rem;
  background: #fff;
  color: #475569;
  cursor: pointer;
  transition: all 0.12s;
}

.filter-pill.active {
  background: #eef2ff;
  border-color: #c7d2fe;
  color: #1d4ed8;
}

.job-list-scroll {
  overflow-y: auto;
  min-height: 0;
  max-height: calc(100vh - 240px);
}

.job-group {
  padding: 16px 20px;
}

.group-label {
  margin-bottom: 10px;
  font-size: 0.85rem;
  color: #64748b;
  text-transform: uppercase;
  letter-spacing: 0.08em;
}

.jobs-detail-panel {
  min-height: 0;
}

.status-card {
  margin: 18px;
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 22px;
  padding: 1.5rem;
}

.status-error {
  border-color: #fecaca;
  background: #fef2f2;
  color: #b91c1c;
}

.button-secondary {
  border: 1px solid #d1d5db;
  background: #fff;
  color: #111827;
  border-radius: 999px;
  padding: 0.9rem 1.1rem;
  font-weight: 600;
  cursor: pointer;
}

.button-secondary:hover {
  background: #f8fafc;
}

@media (max-width: 1060px) {
  .jobs-shell {
    grid-template-columns: 1fr;
  }

  .jobs-body {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 720px) {
  .jobs-main {
    padding: 18px 16px;
  }

  .jobs-topbar {
    flex-direction: column;
    align-items: stretch;
  }

  .jobs-filter-bar {
    flex-wrap: wrap;
  }

  .job-list-scroll {
    max-height: calc(100vh - 320px);
  }
}
</style>
