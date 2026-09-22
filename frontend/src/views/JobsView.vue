<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Search, X, PlusIcon } from '@lucide/vue'
import { useToast } from '@/composables/useToast'
import JobRow from '@/components/JobRow.vue'
import JobDetail from '@/components/JobDetail.vue'
import type { Job } from '@/types/job'
import AppSidebar from '@/components/AppSidebar.vue'
import ImportJobPanel from '@/components/ImportJobPanel.vue'
import JobPagination from '@/components/JobPagination.vue'

const route = useRoute()
const router = useRouter()
const toast = useToast()

const jobs = ref<Job[]>([])
const loading = ref(false)
const hasMore = ref(false)
let requestController: AbortController | undefined
let refreshTimer: ReturnType<typeof setInterval> | undefined
let searchTimer: ReturnType<typeof setTimeout> | undefined
const allowedPageSizes = [50, 100, 200, 500]
const pageSize = computed(() => {
  const value = Number(route.query.size)
  return allowedPageSizes.includes(value) ? value : 50
})
const page = computed(() => {
  const value = Number(route.query.page)
  return Number.isSafeInteger(value) && value > 0
    ? Math.min(value, Math.floor(100000 / pageSize.value) + 1)
    : 1
})
const canGoNext = computed(() => hasMore.value && page.value * pageSize.value <= 100000)
function updateQuery(query: Record<string, string | undefined>, push = false) {
  const destination = { query: { ...route.query, ...query } }
  return (push ? router.push(destination) : router.replace(destination)).catch(() => {})
}
function changePage(next: number) {
  if (loading.value || next < 1 || (next > page.value && !canGoNext.value)) return
  void updateQuery({ page: next === 1 ? undefined : String(next) }, true)
}
function changePageSize(size: number) {
  if (!allowedPageSizes.includes(size)) return
  void updateQuery({ size: String(size), page: undefined })
}
const isJobImportOpen = ref(false)
const error = ref('')
const selectedJobId = ref<string | null>(null)
const sortDirection = computed(() => {
  const value = String(route.query.order ?? '')
  return ['newest', 'oldest', 'score_desc', 'score_asc'].includes(value)
    ? value
    : activeFilter.value === 'matched'
      ? 'score_desc'
      : 'newest'
})
const searchTerm = ref<string>(typeof route.query.q === 'string' ? route.query.q : '')
const activeSource = computed(() =>
  typeof route.query.source === 'string' ? route.query.source : '',
)
const activeFilter = computed({
  get: () => (typeof route.query.filter === 'string' ? route.query.filter : 'all'),
  set: (value: string) => {
    void updateQuery({ filter: value === 'all' ? undefined : value, page: undefined })
  },
})

const filters = [
  { label: 'All', value: 'all' },
  { label: 'Best matches', value: 'matched' },
  { label: 'Saved', value: 'saved' },
  { label: 'Applied', value: 'applied' },
  { label: 'Remote', value: 'remote' },
  { label: 'Today', value: 'today' },
]

watch(
  () => route.query.q,
  (value) => {
    searchTerm.value = typeof value === 'string' ? value : ''
  },
)
watch(searchTerm, () => {
  clearTimeout(searchTimer)
  if (searchTerm.value === (route.query.q ?? '')) return
  searchTimer = setTimeout(() => {
    void updateQuery({ q: searchTerm.value.trim() || undefined, page: undefined })
  }, 300)
})
function submitSearch() {
  clearTimeout(searchTimer)
  const query = searchTerm.value.trim()
  if (query === (route.query.q ?? '') && page.value === 1) void fetchJobs()
  else void updateQuery({ q: query || undefined, page: undefined })
}

const groupedJobs = computed(() => {
  return jobs.value.reduce<Record<string, Job[]>>((groups, job) => {
    const label = sortDirection.value.startsWith('score_')
      ? sortDirection.value === 'score_desc'
        ? 'Highest score first'
        : 'Lowest score first'
      : job.created_at
        ? new Date(job.created_at).toLocaleDateString(undefined, {
            month: 'short',
            day: 'numeric',
          })
        : 'Unknown'
    groups[label] = groups[label] || []
    groups[label].push(job)
    return groups
  }, {})
})

const linkedJob = ref<Job | null>(null)
const selectedJob = computed(() => {
  if (!selectedJobId.value) return null
  return (
    jobs.value.find((job) => job.id === selectedJobId.value) ??
    (linkedJob.value?.id === selectedJobId.value ? linkedJob.value : null)
  )
})

const detailOpen = computed(() => !!selectedJob.value)

function handleImportJobCreated(job: Job) {
  isJobImportOpen.value = false
  linkedJob.value = job
  selectedJobId.value = job.id
  // Keep the server's page membership intact; an imported job may not match this feed.
  void fetchJobs()
  let attempts = 0
  if (refreshTimer) clearInterval(refreshTimer)
  refreshTimer = setInterval(async () => {
    attempts++
    try {
      const response = await fetch(`/api/jobs/${job.id}`, { credentials: 'include' })
      if (response.ok) {
        const updated: Job = await response.json()
        if (linkedJob.value?.id === updated.id) linkedJob.value = updated
        jobs.value = jobs.value.map((existing) => (existing.id === updated.id ? updated : existing))
        if (updated.is_enriched || attempts >= 12) clearInterval(refreshTimer)
      }
    } catch {
      /* The normal refresh button remains available after a network failure. */
    }
    if (attempts >= 12) clearInterval(refreshTimer)
  }, 5000)
}

function closeDetail() {
  selectedJobId.value = null
  linkedJob.value = null
  const query = { ...route.query }
  delete query.job
  router.replace({ query })
}

async function fetchJobs() {
  requestController?.abort()
  const controller = new AbortController()
  requestController = controller
  loading.value = true
  error.value = ''
  jobs.value = []
  hasMore.value = false
  const params = new URLSearchParams({
    limit: String(pageSize.value),
    offset: String((page.value - 1) * pageSize.value),
    q: typeof route.query.q === 'string' ? route.query.q : '',
    filter: activeFilter.value,
    source: activeSource.value,
    order: sortDirection.value,
  })
  try {
    const response = await fetch(`/api/jobs?${params}`, {
      credentials: 'include',
      signal: controller.signal,
    })
    if (!response.ok) throw new Error('Failed to load jobs')
    const hasMoreHeader = response.headers?.get('X-Has-More')
    if (hasMoreHeader !== 'true' && hasMoreHeader !== 'false')
      throw new Error(
        'Backend update required. Restart the Go server with make run, then refresh this page.',
      )
    const batch: Job[] = (await response.json()) ?? []
    if (controller.signal.aborted || requestController !== controller) return
    hasMore.value = hasMoreHeader === 'true'
    jobs.value = batch
  } catch (err) {
    if (!controller.signal.aborted)
      error.value = err instanceof Error ? err.message : 'Could not load jobs'
  } finally {
    if (!controller.signal.aborted) loading.value = false
  }
}
watch([() => route.query.q, activeFilter, sortDirection, activeSource, page, pageSize], () => {
  void fetchJobs()
})

async function handleSaveJob(job: Job) {
  if (job.is_saved) {
    try {
      const response = await fetch('/api/saved_jobs', {
        method: 'DELETE',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({ job_id: job.id }),
      })
      if (response.ok) job.is_saved = false
      else toast.error('Failed to unsave job')
    } catch (err) {
      toast.error('Something went wrong:' + err)
    }
  } else {
    try {
      const response = await fetch('/api/saved_jobs', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({ job_id: job.id }),
      })
      if (response.ok) job.is_saved = true
      else toast.error('Failed to save job')
    } catch (err) {
      toast.error('Something went wrong:' + err)
    }
  }
}

async function handleApplyJob(job: Job) {
  try {
    const response = await fetch('/api/applications', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify({ job_id: job.id, application_status: 'applied' }),
    })
    if (response.ok) {
      job.is_applied = true
    } else {
      const data = await response.json().catch(() => null)
      toast.error(data?.error ?? 'Failed to track application')
    }
  } catch (err) {
    toast.error('Something went wrong:' + err)
  }
}

function formatCount(count: number) {
  return count === 1 ? '1 role' : `${count} roles`
}

async function openLinkedJob() {
  const id = route.query.job
  if (typeof id !== 'string') return
  try {
    const r = await fetch(`/api/jobs/${encodeURIComponent(id)}`, { credentials: 'include' })
    if (!r.ok) throw new Error('Job no longer available')
    linkedJob.value = await r.json()
    selectedJobId.value = id
  } catch (e) {
    toast.error(e instanceof Error ? e.message : 'Could not open job')
  }
}
watch(() => route.query.job, openLinkedJob)
onMounted(() => {
  fetchJobs()
  openLinkedJob()
})
onUnmounted(() => {
  requestController?.abort()
  clearInterval(refreshTimer)
  clearTimeout(searchTimer)
})
</script>

<template>
  <div class="grid h-full overflow-hidden" style="grid-template-columns: 220px minmax(0, 1fr)">
    <AppSidebar />

    <main class="flex flex-col gap-4 p-4 lg:p-6 min-h-0 min-w-0">
      <!-- Topbar -->
      <header class="flex flex-col gap-2 min-w-0 shrink-0">
        <!-- Search bar -->
        <div class="flex items-center gap-2 border border-gray-200 bg-white pl-3 h-9">
          <Search class="w-4 h-4 text-gray-400 shrink-0" />
          <input
            id="search"
            name="search"
            v-model="searchTerm"
            @keydown.enter="submitSearch"
            type="text"
            placeholder="Search by title, company, or skill"
            class="flex-1 min-w-0 border-none outline-none text-sm bg-transparent text-gray-900 placeholder:text-gray-400"
          />
          <span
            v-if="searchTerm"
            class="text-xs text-gray-400 cursor-pointer hover:text-gray-600"
            @click="searchTerm = ''"
          >
            <X />
          </span>
          <button
            @click="submitSearch"
            class="button h-full px-5 text-sm text-white"
            :style="{ background: 'var(--color-accent)' }"
          >
            Search
          </button>
        </div>

        <!-- Filters + sort -->
        <div class="flex flex-wrap items-center gap-2">
          <div class="flex gap-2 flex-wrap basis-full xl:basis-auto flex-1">
            <button
              v-for="f in filters"
              :key="f.value"
              class="border cursor-pointer transition-all h-8 px-2 text-sm"
              :class="
                activeFilter === f.value
                  ? 'border-gray-900 text-accent'
                  : 'border-gray-200 bg-white text-gray-900'
              "
              :style="activeFilter === f.value ? { background: 'var(--color-accent-soft)' } : {}"
              @click="activeFilter = f.value"
            >
              {{ f.label }}
            </button>
          </div>
          <select
            aria-label="Sort jobs"
            class="border border-gray-200 bg-white px-3 py-2 text-sm"
            :value="sortDirection"
            @change="
              updateQuery({ order: ($event.target as HTMLSelectElement).value, page: undefined })
            "
          >
            <option value="newest">Newest</option>
            <option value="oldest">Oldest</option>
            <option value="score_desc">Highest score</option>
            <option value="score_asc">Lowest score</option>
          </select>
          <button class="button button-secondary" @click="fetchJobs()" :disabled="loading">
            Refresh
          </button>
          <button @click="isJobImportOpen = true" class="button button-primary">
            Import job <PlusIcon :size="20" />
          </button>
        </div>
      </header>

      <div v-if="activeSource" class="flex items-center gap-3 text-sm">
        <span
          >Source: <strong>{{ activeSource }}</strong></span
        >
        <RouterLink
          :to="{ query: { ...route.query, source: undefined, page: undefined } }"
          class="button button-secondary"
          >All sources</RouterLink
        >
      </div>
      <p v-if="activeFilter === 'matched'" class="text-sm text-gray-600">
        Jobs at or above your minimum score in Settings. Your work and location preferences also
        apply; jobs being rescored appear when matching finishes.
        <RouterLink to="/profile" class="text-accent underline">Review profile settings</RouterLink>
      </p>
      <!-- Count row -->
      <div class="flex items-center gap-1 text-sm shrink-0">
        <span class="font-semibold text-gray-900">{{
          formatCount(jobs.length) + ' on this page'
        }}</span>
        <span v-if="searchTerm" class="italic text-gray-400">for "{{ searchTerm }}"</span>
      </div>

      <!-- Jobs body -->
      <section class="flex-1 bg-white border border-gray-200 flex flex-col min-h-0">
        <div v-if="loading" class="m-4.5 bg-white p-6 border border-gray-200">Loading jobs…</div>
        <div v-else-if="error" class="m-4.5 p-6 border border-red-200 bg-red-50 text-red-700">
          {{ error }}
        </div>
        <div v-else-if="jobs.length === 0" class="m-4.5 bg-white p-6 border border-gray-200">
          <template v-if="activeFilter === 'matched' && page === 1">
            No jobs meet your matching settings and current filters yet. Review your minimum score
            and preferences, or browse All to see jobs with lower scores.
          </template>
          <template v-else
            >No jobs on this page. Try the previous page or adjust your filters.</template
          >
        </div>
        <div v-else class="flex-1 overflow-y-auto min-h-0">
          <div v-for="(group, label) in groupedJobs" :key="label" class="px-5 py-4">
            <div class="mb-2.5 text-[0.85rem] text-gray-400 uppercase tracking-[0.08em]">
              {{ label }}
            </div>
            <JobRow
              v-for="job in group"
              :key="job.id"
              :job="job"
              :selected="job.id === selectedJobId"
              @click="selectedJobId = job.id"
              @save="handleSaveJob"
              @apply="handleApplyJob"
            />
          </div>
        </div>
      </section>
      <JobPagination
        :page="page"
        :page-size="pageSize"
        :has-next="canGoNext"
        :loading="loading"
        @page="changePage"
        @page-size="changePageSize"
      />
    </main>

    <!-- Detail panel -->
    <Teleport to="body">
      <Transition name="detail-slide">
        <div
          v-if="detailOpen"
          class="fixed inset-0 z-100 bg-black/45 flex justify-end pointer-events-none top-(--topbar-height)"
          @click.self="closeDetail"
        >
          <JobDetail
            v-if="selectedJob"
            :job="selectedJob"
            @close="closeDetail"
            @save="handleSaveJob"
          />
        </div>
      </Transition>
    </Teleport>

    <!-- Job Import -->
    <Teleport to="body">
      <Transition name="detail-slide">
        <div
          v-if="isJobImportOpen"
          class="fixed inset-0 z-100 bg-black/45 flex justify-end top-(--topbar-height)"
          @click.self="isJobImportOpen = false"
        >
          <ImportJobPanel
            v-if="isJobImportOpen"
            @close="isJobImportOpen = false"
            @created="handleImportJobCreated"
          />
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<style scoped>
/* Mobile overrides — kept in <style> because Tailwind can't conditionally swap grid-template-columns easily */
@media (max-width: 768px) {
  div[style*='grid-template-columns'] {
    grid-template-columns: 1fr !important;
  }

  .detail-back-mobile {
    display: flex !important;
  }

  div[class*='overflow-y-auto'][class*='min-h-0'] {
    max-height: calc(100vh - 320px);
  }
}

@media (max-width: 480px) {
  button.filter-pill-sm {
    font-size: 0.75rem;
    padding: 0 4px;
  }
}
</style>
