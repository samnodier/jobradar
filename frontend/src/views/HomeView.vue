<template>
  <div class="flex-1 overflow-y-auto h-full">
    <section class="text-center px-8 py-10 mx-auto">
      <div class="max-w-[40rem] mx-auto text-center">
        <div class="inline-flex items-center gap-2.5 px-4 py-2 bg-[rgba(94,106,210,0.12)] text-accent text-[0.8rem] mb-5">
          <span class="w-[7px] h-[7px] bg-accent"></span>
          {{ eyebrowText }}
        </div>

        <h1 class="m-0 text-[clamp(2.8rem,4vw,3.8rem)] leading-[1.05] font-serif">
          Find jobs<br />
          <span class="text-accent">that match your skills</span>
        </h1>

        <p class="mt-6 mb-8 max-w-[42rem] text-gray-700 text-[1.05rem] leading-[1.75] mx-auto">
          JobRadar scrapes the best job boards and surfaces roles matched to your profile — no
          noise, just signal.
        </p>

        <!-- Search Bar -->
        <div class="flex items-center max-w-[520px] mx-auto border border-ui-border pl-2.5 h-11 gap-2 whitespace-nowrap cursor-pointer shrink-0">
          <Search class="w-4 h-4 text-gray-400" />
          <input
            v-model="searchTerm"
            @keyup.enter="goToJobs"
            type="text"
            placeholder="Search by title, company, or skill"
            class="flex-1 border-none outline-none h-full p-3 text-base text-gray-900 placeholder:text-gray-400"
          />
          <button @click="goToJobs" class="h-full bg-accent text-white px-6 font-semibold hover:-translate-y-px transition-transform">
            Search
          </button>
        </div>

        <div class="flex justify-center gap-4 mt-8">
          <RouterLink to="/jobs" class="px-8 py-3 bg-accent text-white font-semibold hover:-translate-y-px transition-transform text-[0.95rem]">
            Browse All Jobs
          </RouterLink>
          <button @click="goToJobs()" class="px-8 py-3 bg-bg-secondary text-gray-900 border border-ui-border font-semibold hover:-translate-y-px transition-transform text-[0.95rem]">
            Find Remote Roles
          </button>
        </div>

        <div class="flex flex-wrap gap-3 mt-7 items-center justify-center">
          <span v-for="tag in ['Go Engineer', 'Vue Developer', 'Full Stack', 'Remote']" :key="tag" 
            class="inline-flex items-center justify-center px-4 py-1 bg-[#f8faff] text-gray-700 border border-ui-border text-[0.9rem]">
            {{ tag }}
          </span>
        </div>
      </div>
    </section>

    <section class="flex flex-col bg-bg-secondary px-8 pb-12">
      <div class="flex flex-col mt-10 w-full max-w-[1200px] mx-auto gap-6">
        <!-- Stats Row -->
        <div class="grid grid-cols-1 md:grid-cols-3 gap-4 w-full">
          <div class="bg-white p-6 border border-ui-border">
            <div class="text-[0.75rem] uppercase tracking-widest text-accent mb-3">Jobs tracked</div>
            <div class="text-4xl font-bold text-gray-900">{{ jobsTracked }}</div>
            <div class="mt-2.5 text-gray-600 text-[0.92rem]">+{{ jobsToday }} today</div>
          </div>
          <div class="bg-white p-6 border border-ui-border">
            <div class="text-[0.75rem] uppercase tracking-widest text-accent mb-3">Sources</div>
            <div class="text-4xl font-bold text-gray-900">2</div>
            <div class="mt-2.5 text-gray-600 text-[0.92rem]">RemoteOK · Adzuna</div>
          </div>
          <div class="bg-white p-6 border border-ui-border">
            <div class="text-[0.75rem] uppercase tracking-widest text-accent mb-3">Latest scrape</div>
            <div class="text-4xl font-bold text-gray-900">{{ latestScrape }}</div>
            <div class="mt-2.5 text-gray-600 text-[0.92rem]">Fresh data from the feed</div>
          </div>
        </div>

        <!-- Recommended Section -->
        <div v-if="authStore.user" class="mt-8">
          <div class="flex justify-between items-end mb-4 gap-4">
            <div>
              <p class="m-0 mb-1.5 uppercase tracking-widest text-[0.78rem] text-gray-600">Matched to your skills</p>
              <h2 class="m-0 text-xl text-gray-900">Recommended for you</h2>
            </div>
            <RouterLink to="/jobs" class="text-accent text-[0.94rem] font-semibold">View all matches</RouterLink>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
            <article v-for="(job, index) in recommendedJobs" :key="job.id" class="bg-white p-5 grid gap-4 border border-ui-border-soft hover:border-ui-border transition-colors">
              <div class="flex justify-between items-start gap-4">
                <h3 class="m-0 text-base text-gray-900 leading-tight font-semibold">{{ job.title }}</h3>
                <Heart :fill="job.is_saved ? 'currentColor' : 'none'" class="w-6 h-6 text-gray-400 cursor-pointer shrink-0" />
              </div>
              <div class="text-[0.82rem] text-accent font-bold">{{ 92 - index * 4 }}% match</div>

              <p class="m-0 mt-1.5 text-gray-700 text-[0.92rem]">{{ job.company_name }}</p>
              <div class="flex flex-wrap gap-2">
                <span class="px-2 py-1 text-[0.75rem] bg-bg-secondary text-gray-700 border border-ui-border">
                  {{ job.is_remote ? 'Remote' : job.job_location || 'On-site' }}
                </span>
                <span class="px-2 py-1 text-[0.75rem] bg-bg-secondary text-gray-700 border border-ui-border">
                  {{ job.employment_type || 'Full time' }}
                </span>
              </div>
              <div class="flex items-center justify-between text-gray-500 text-[0.75rem] pt-2 border-t border-ui-border-soft">
                <span>{{ formatSalary(job) }}</span>
                <span>{{ formatWhen(job) }}</span>
              </div>
            </article>
          </div>
        </div>

        <!-- Latest Jobs Table -->
        <div class="mt-8">
          <div class="flex justify-between items-end mb-4 gap-4">
            <p class="m-0 uppercase tracking-widest text-[0.78rem] text-gray-600">Latest jobs</p>
            <RouterLink to="/jobs" class="text-accent text-[0.94rem] font-semibold">View all jobs</RouterLink>
          </div>

          <div class="bg-white overflow-hidden border border-ui-border">
            <div v-for="job in latestJobs" :key="job.id" 
              class="grid grid-cols-[1fr_auto] md:grid-cols-[2.4fr_1fr_0.9fr_0.8fr] gap-4 items-center p-4 border-b border-ui-border last:border-b-0">
              <div class="flex items-center gap-3.5">
                <div class="w-2 h-2 shrink-0" :class="job.is_remote ? 'bg-[#5e6ad2]' : 'bg-[#22c55e]'"></div>
                <div>
                  <a :href="job.source_url" target="_blank" class="text-[0.95rem] font-semibold text-gray-900 hover:text-accent transition-colors">
                    {{ job.title }}
                  </a>
                </div>
              </div>
              <span class="hidden md:block text-[0.95rem] text-gray-600">{{ job.company_name }}</span>
              <span class="hidden md:block text-[0.95rem] text-gray-500">{{ job.employment_type || 'Full time' }}</span>
              <span class="text-[0.95rem] text-gray-400 text-right md:text-left">{{ formatWhen(job) }}</span>
            </div>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { Heart, Search } from '@lucide/vue'

type JobStats = {
  total_jobs: number
  new_jobs_today: number
  latest_scrape_at: string | null
}

type Job = {
  id: string
  title: string
  company_name: string
  source_url: string
  job_location?: string | null
  is_remote?: boolean | null
  employment_type?: string | null
  salary_min?: number | null
  salary_max?: number | null
  posted_at?: string | null
  is_saved: boolean
}
const router = useRouter()
const authStore = useAuthStore()

const stats = ref<JobStats | null>(null)
const jobs = ref<Job[]>([])
const searchTerm = ref('')

const eyebrowText = computed(() => {
  if (!stats.value) {
    return 'Discover your next role'
  }
  return stats.value.new_jobs_today > 0
    ? `${stats.value.new_jobs_today} new job${stats.value.new_jobs_today === 1 ? '' : 's'} scraped today`
    : 'Discover your next role'
})

const jobsTracked = computed(() => {
  return stats.value?.total_jobs.toLocaleString() ?? '12,483'
})

const jobsToday = computed(() => {
  return stats.value?.new_jobs_today ?? 241
})

const latestScrape = computed(() => {
  if (!stats.value?.latest_scrape_at) {
    return 'Just now'
  }
  return new Date(stats.value.latest_scrape_at).toLocaleTimeString([], {
    hour: 'numeric',
    minute: '2-digit',
  })
})

const latestJobs = computed(() => {
  return [...jobs.value]
    .sort((a, b) => {
      return (
        (new Date(b.posted_at ?? '').getTime() || 0) - (new Date(a.posted_at ?? '').getTime() || 0)
      )
    })
    .slice(0, 5)
})

const recommendedJobs = computed(() => {
  if (!authStore.user) return []
  return latestJobs.value.slice(0, 3)
})

function formatSalary(job: Job) {
  if (!job.salary_min && !job.salary_max) return 'Salary n/a'
  const currency = 'USD'
  if (job.salary_min && job.salary_max) {
    return `${currency} ${job.salary_min.toLocaleString()} - ${job.salary_max.toLocaleString()}`
  }
  return `${currency} ${job.salary_min ?? job.salary_max}`
}

function formatWhen(job: Job) {
  if (!job.posted_at) return 'Unknown'
  const date = new Date(job.posted_at)
  const diff = Date.now() - date.getTime()
  const hours = Math.floor(diff / 3600000)
  if (hours < 24) return `${hours}h ago`
  const days = Math.floor(hours / 24)
  return `${days}d ago`
}

function goToJobs() {
  const q = searchTerm.value.trim()
  router.push({
    path: `/jobs`,
    query: q ? { q } : {},
  })
}

onMounted(async () => {
  try {
    const response = await fetch('/api/jobs/stats')
    if (response.ok) {
      stats.value = await response.json()
    }
  } catch {
    stats.value = null
  }

  try {
    const response = await fetch('/api/jobs')
    if (response.ok) {
      jobs.value = await response.json()
    }
  } catch {
    jobs.value = []
  }
})
</script>
