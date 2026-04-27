<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
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
const searchTerm = ref<string>('')
const activeFilter = ref<string>((route.query.filter as string) || 'all')
const headerHeight = ref(53)
let resizeObserver: ResizeObserver | null = null

const filters = [
  { label: 'All', value: 'all' },
  { label: 'Remote', value: 'remote' },
  { label: 'Saved', value: 'saved' },
  { label: 'Applied', value: 'applied' },
  { label: 'Today', value: 'today' },
]

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
  let result = [...jobs.value]

  const q = searchTerm.value.trim().toLowerCase()

  if (q) {
    result = result.filter((job) => {
      return (
        job.title.toLowerCase().includes(q) ||
        job.company.toLowerCase().includes(q) ||
        job.skills?.some((skill) => skill.toLowerCase().includes(q))
      )
    })
  }
  if (activeFilter.value === 'remote') return result.filter((job) => job.is_remote)
  if (activeFilter.value === 'saved') return jobs.value.filter((job) => job.status === 'saved')
  if (activeFilter.value === 'applied') return jobs.value.filter((job) => job.status === 'applied')
  if (activeFilter.value === 'today') {
    const today = new Date().toDateString()
    result = result.filter((job) => {
      const postedDate = new Date(job.posted_at ?? '').toDateString()
      return postedDate === today
    })
  }

  return result
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

const detailOpen = computed(() => !!selectedJob.value)
function closeDetail() {
  selectedJobId.value = null
}


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

onMounted(() => {
  fetchJobs()

  // Find the topbar element dynamically
  const header = document.querySelector('.topbar')
  if (header) {
    // Update height initially and whenever the header resizes
    resizeObserver = new ResizeObserver(() => {
      headerHeight.value = header.getBoundingClientRect().height
    })
    resizeObserver.observe(header)
  }
})

onUnmounted(() => {
  // Clean up the observer to prevent memory leaks when navigating away
  if (resizeObserver) {
    resizeObserver.disconnect()
  }
})
</script>

<template>
  <div class="jobs-shell">
    <AppSidebar />

    <main class="jobs-main">
      <header class="jobs-topbar">

        <div class="search-bar">
          <Search class="search-icon" />
          <input v-model="searchTerm" type="text" placeholder="Search by title, company, or skill"
            class="search-input" />
          <span v-if="searchTerm" class="search-clear" @click="searchTerm = ''">✕</span>
          <button class="button search-button button-primary">Search</button>

        </div>

        <div class="topbar-actions">
          <div class="jobs-filter-bar">
            <button v-for="f in filters" :key="f.value" :class="['filter-pill', { active: activeFilter === f.value }]"
              @click="activeFilter = f.value">{{ f.label }}</button>
          </div>
          <button class="button-sort" @click="toggleSort">
            <Sort />
            {{ sortDirection == 'newest' ? 'Newest' : 'Oldest' }}
          </button>
        </div>
      </header>

      <div class="count-row">
        <span class="count-label">{{ formatCount(filteredJobs.length) }}</span>
        <span v-if="searchTerm" class="count-query">for "{{ searchTerm }}"</span>
      </div>

      <section class="jobs-body">
        <div v-if="loading" class="status-card">Loading jobs…</div>
        <div v-else-if="error" class="status-card status-error">{{ error }}</div>
        <div v-else-if="filteredJobs.length === 0" class="status-card">No jobs match this filter.</div>
        <div v-else class="job-list-scroll">
            <div v-for="(group, label) in groupedJobs" :key="label" class="job-group">
              <div class="group-label">{{ label }}</div>
              <JobRow v-for="job in group" :key="job.id" :job="job" :selected="job.id === selectedJobId"
                @click="selectedJobId = job.id" />
            </div>
        </div>
      </section>
    </main>

    <!-- Details panel - slides in from the right as overlay -->
    <Teleport to="body">
      <Transition name="detail-slide">
        <div v-if="detailOpen" class="detail-overlay" @click.self="closeDetail" :style="{ top: `${headerHeight}px` }">
          <div class="detail-drawer">
            <div class="detail-back" @click="closeDetail">
              <ArrowLeft />
              Back to jobs
            </div>
            <JobDetail v-if="selectedJob" :job="selectedJob" @close="closeDetail" />
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<style scoped>
.jobs-shell {
  display: grid;
  grid-template-columns: 220px minmax(0, 1fr);
  min-height: calc(100vh - 96px);
}

.jobs-main {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  padding: 24px 28px;
}

/* Search */

.jobs-topbar {
  display: flex;
  flex-direction: column;
  padding: var(--spacing-3);
  min-width: 0;
  gap: var(--spacing-2);
}


.topbar-actions {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
}

.total-count {
  font-weight: 600;
  color: #1f2937;
}

.jobs-body {
  flex: 1;
  background: var(--color-bg-primary);
  border: 1px solid var(--color-border);
  overflow: hidden;
}

/* Filters */

.jobs-filter-bar {
  display: flex;
  gap: var(--spacing-2);
  flex-wrap: wrap;
  flex: 1;
}

.filter-pill {
  border: 1px solid var(--color-border);
  background: var(--bg);
  color: var(--text);
  cursor: pointer;
  transition: all 0.12s;
  height: 2rem;
  padding: 0 var(--spacing-2);
  font-size: var(--text-sm);

}

.filter-pill.active {
  background: var(--accent-soft);
  border-color: var(--color-text);
  color: var(--accent);
}

/* Count */
.count-row {
  display: flex;
  align-items: center;
  gap: var(--spacing-1);
  font-size: var(--text-sm);

}

.count-label {
  font-weight: 600;
  color: var(--text);
}

.count-query {
  font-style: italic;
  color: var(--muted);
}

.count-query {
  font-style: italic;
  color: var(--muted);
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
  color: var(--muted);
  text-transform: uppercase;
  letter-spacing: 0.08em;
}

/* Detail Overlay */

.detail-overlay {
  position: fixed;
  inset: 0;
  z-index: 100;
  display: flex;
  justify-content: flex-end;
}

.detail-drawer {
  width: 50vw;
  min-width: 560px;
  max-width: 100%;
  height: 100%;
  background: var(--surface);
  box-shadow: -4px 0 12px rgba(0, 0, 0, 0.1);
  display: flex;
  flex-direction: column;
  overflow-y: auto;
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

.status-card {
  margin: 18px;
  background: var(--color-bg-primary);
  border: var(--color-border);
  padding: 1.5rem;
}

.status-error {
  border-color: #fecaca;
  background: #fef2f2;
  color: #b91c1c;
}
/* Mobile */
@media (max-width: 768px) {
  .jobs-shell {
    grid-template-columns: 1fr;
  }

  .jobs-main {
    padding: var(--spacing-4);
  }

  .detail-drawer {
    width: 100%;
  }

  .detail-back {
    visibility: visible;
    height: auto;
    padding: var(--spacing-3);
    overflow: visible;
  }

  .job-list-scroll {
    max-height: calc(100vh - 320px);
  }
}

@media (max-width: 480px) {
.filter-pill {
  font-size: var(--text-xs);
  padding: 0 var(--spacing-1);
}
}
</style>
