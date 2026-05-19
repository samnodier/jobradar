<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useToast } from '@/composables/useToast'
import AppSidebar from '@/components/AppSidebar.vue'
import JobRow from '@/components/JobRow.vue'
import { ArrowLeft, Search, ListFilter } from '@lucide/vue'
import JobDetail from '@/components/JobDetail.vue'
import type { Job } from '@/types/job'

const route = useRoute()
const router = useRouter()
const toast = useToast()

const jobs = ref<Job[]>([])
const loading = ref(false)
const error = ref('')
const selectedJobId = ref<string | null>(null)
const sortDirection = ref<'newest' | 'oldest'>('newest')
const searchTerm = ref<string>('')
const activeFilter = ref<string>((route.query.filter as string) || 'all')

const filters = [
  { label: 'All', value: 'all' },
  { label: 'Saved', value: 'saved' },
  { label: 'Applied', value: 'applied' },
  { label: 'Remote', value: 'remote' },
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
    router.replace({ query: {} }).catch(() => {})
  } else {
    router.replace({
      query: {
        filter: value === 'all' ? undefined : value,
        q: route.query.q,
      },
    }).catch(() => {})
  }
})

const filteredJobs = computed(() => {
  let result = [...jobs.value]
  const q = searchTerm.value.trim().toLowerCase()
  if (q) {
    result = result.filter((job) =>
      job.title.toLowerCase().includes(q) ||
      job.company_name.toLowerCase().includes(q) ||
      job.skills?.some((skill) => skill.toLowerCase().includes(q))
    )
  }
  if (activeFilter.value === 'remote') result = result.filter((job) => job.is_remote)
  else if (activeFilter.value === 'saved') result = result.filter((job) => job.is_saved)
  else if (activeFilter.value === 'applied') result = result.filter((job) => job.is_applied)
  else if (activeFilter.value === 'today') {
    const today = new Date().toDateString()
    result = result.filter((job) => new Date(job.posted_at ?? '').toDateString() === today)
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
    const label = job.posted_at
      ? new Date(job.posted_at).toLocaleDateString(undefined, { month: 'short', day: 'numeric' })
      : 'Unknown'
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
  fetch('/api/jobs', { credentials: 'include' })
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
      job.is_saved = true
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

function toggleSort() {
  sortDirection.value = sortDirection.value === 'newest' ? 'oldest' : 'newest'
}

onMounted(() => fetchJobs())
onUnmounted(() => {})
</script>

<template>
  <div class="grid h-full overflow-hidden" style="grid-template-columns: 220px minmax(0, 1fr)">
    <AppSidebar />

    <main class="flex flex-col gap-6 p-6 min-h-0" style="padding: 24px 28px">
      <!-- Topbar -->
      <header class="flex flex-col gap-2 min-w-0 shrink-0 p-3">
        <!-- Search bar -->
        <div class="flex items-center gap-2 border border-gray-200 bg-white px-3 h-9">
          <Search class="w-4 h-4 text-gray-400 shrink-0" />
          <input
            v-model="searchTerm"
            type="text"
            placeholder="Search by title, company, or skill"
            class="flex-1 border-none outline-none text-sm bg-transparent text-gray-900 placeholder:text-gray-400"
          />
          <span
            v-if="searchTerm"
            class="text-xs text-gray-400 cursor-pointer hover:text-gray-600"
            @click="searchTerm = ''"
          >✕</span>
          <button
            class="button h-7 px-3 text-sm text-white"
            :style="{ background: 'var(--color-accent)' }"
          >
            Search
          </button>
        </div>

        <!-- Filters + sort -->
        <div class="flex items-center gap-2">
          <div class="flex gap-2 flex-wrap flex-1">
            <button
              v-for="f in filters"
              :key="f.value"
              class="border cursor-pointer transition-all h-8 px-2 text-sm"
              :class="activeFilter === f.value
                ? 'border-gray-900 text-accent'
                : 'border-gray-200 bg-white text-gray-900'"
              :style="activeFilter === f.value ? { background: 'var(--color-accent-soft)' } : {}"
              @click="activeFilter = f.value"
            >
              {{ f.label }}
            </button>
          </div>
          <button
            class="flex items-center gap-1.5 h-8 px-2 border border-gray-200 bg-white text-sm text-gray-700 cursor-pointer hover:bg-gray-50 transition-all shrink-0"
            @click="toggleSort"
          >
            <ListFilter class="w-4 h-4" />
            {{ sortDirection === 'newest' ? 'Newest' : 'Oldest' }}
          </button>
        </div>
      </header>

      <!-- Count row -->
      <div class="flex items-center gap-1 text-sm shrink-0">
        <span class="font-semibold text-gray-900">{{ formatCount(filteredJobs.length) }}</span>
        <span v-if="searchTerm" class="italic text-gray-400">for "{{ searchTerm }}"</span>
      </div>

      <!-- Jobs body -->
      <section class="flex-1 bg-white border border-gray-200 flex flex-col min-h-0">
        <div v-if="loading" class="m-4.5 bg-white p-6 border border-gray-200">
          Loading jobs…
        </div>
        <div
          v-else-if="error"
          class="m-4.5 p-6 border border-red-200 bg-red-50 text-red-700"
        >
          {{ error }}
        </div>
        <div v-else-if="filteredJobs.length === 0" class="m-4.5 bg-white p-6 border border-gray-200">
          No jobs match this filter.
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
    </main>

    <!-- Detail panel -->
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
            <!-- Back button: hidden on desktop, visible on mobile -->
            <div
              class="hidden items-center gap-2 p-3 border-b border-gray-200 text-gray-400 cursor-pointer text-sm shrink-0 md:hidden"
              @click="closeDetail"
            >
              <ArrowLeft class="w-4 h-4" />
              Back to jobs
            </div>
            <JobDetail
              v-if="selectedJob"
              :job="selectedJob"
              @close="closeDetail"
              @save="handleSaveJob"
            />
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<style scoped>
/* Mobile overrides — kept in <style> because Tailwind can't conditionally swap grid-template-columns easily */
@media (max-width: 768px) {
  div[style*="grid-template-columns"] {
    grid-template-columns: 1fr !important;
  }

  .detail-back-mobile {
    display: flex !important;
  }

  div[class*="overflow-y-auto"][class*="min-h-0"] {
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
