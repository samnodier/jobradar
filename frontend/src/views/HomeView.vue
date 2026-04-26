<template>
  <section class="home-hero">
    <div class="hero-copy">
      <div class="hero-eyebrow">
        <span class="hero-dot"></span>
        {{ eyebrowText }}
      </div>

      <h1>
        Find remote jobs<br />
        <span>that match your skills</span>
      </h1>

      <p class="hero-text">
        JobRadar scrapes the best remote job boards and surfaces roles matched to your profile — no noise, just signal.
      </p>

      <div class="search-bar">
        <Search class="search-icon" />
        <input v-model="searchTerm" @keyup.enter="goToJobs" type="text" placeholder="Search by title, company, or skill"
          class="search-input" />
        <button @click="goToJobs" class="button search-button button-primary">Browse jobs</button>
      </div>

      <div class="hero-tags">
        <span class="hero-tag">Go Engineer</span>
        <span class="hero-tag">Vue Developer</span>
        <span class="hero-tag">Full Stack</span>
        <span class="hero-tag">Remote</span>
      </div>
    </div>
  </section>

  <section class="stats-sec">
    <div class="stats-sub-sec">
      <div class="stats-row">
        <div class="stat-card">
          <div class="stat-label">Jobs tracked</div>
          <div class="stat-value">{{ jobsTracked }}</div>
          <div class="stat-sub">+{{ jobsToday }} today</div>
        </div>
        <div class="stat-card">
          <div class="stat-label">Sources</div>
          <div class="stat-value">2</div>
          <div class="stat-sub">RemoteOK · Adzuna</div>
        </div>
        <div class="stat-card">
          <div class="stat-label">Latest scrape</div>
          <div class="stat-value">{{ latestScrape }}</div>
          <div class="stat-sub">Fresh data from the feed</div>
        </div>
      </div>
      <div v-if="authStore.user">
        <div class="section-heading">
          <div>
            <p class="section-eyebrow">Matched to your skills</p>
            <h2 class="section-title">Recommended for you</h2>
          </div>
          <RouterLink to="/jobs" class="section-link">View all matches</RouterLink>
        </div>

        <div class="recommend-grid">
          <article v-for="(job, index) in recommendedJobs" :key="job.id" class="recommend-card">
            <div class="card-top">
              <h3 class="card-title">{{ job.title }}</h3>
              <Heart class="icon-button" />

            </div>
            <div class="card-match">{{ 92 - index * 4 }}% match</div>

            <p class="card-company">{{ job.company }}</p>
            <div class="card-meta">
              <span class="meta-pill">{{ job.is_remote ? 'Remote' : job.location || 'On-site' }}</span>
              <span class="meta-pill">{{ job.employment_type || 'Full time' }}</span>
            </div>
            <div class="card-footer">
              <span class="card-salary">{{ formatSalary(job) }}</span>
              <span class="card-time">{{ formatWhen(job) }}</span>
            </div>
          </article>
        </div>
      </div>

      <div class="home-section">
        <div class="section-heading">
          <div>
            <p class="section-eyebrow">Latest jobs</p>
          </div>
          <RouterLink to="/jobs" class="section-link">View all jobs</RouterLink>
        </div>

        <div class="jobs-table">
          <div v-for="job in latestJobs" :key="job.id" class="jobs-table-row">
            <div class="position-cell">
              <div class="status-dot" :class="statusClass(job)"></div>
              <div>
                <p class="job-title">{{ job.title }}</p>
              </div>
            </div>
            <span class="company-cell">{{ job.company }}</span>
            <span class="type-cell">{{ job.employment_type || 'Full time' }}</span>
            <span class="when-cell">{{ formatWhen(job) }}</span>
          </div>
        </div>
      </div>
    </div>
  </section>

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
  company: string
  location?: string | null
  is_remote?: boolean | null
  employment_type?: string | null
  salary_min?: number | null
  salary_max?: number | null
  posted_at?: string | null
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
      return (new Date(b.posted_at ?? '').getTime() || 0) - (new Date(a.posted_at ?? '').getTime() || 0)
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

function statusClass(job: Job) {
  return job.is_remote ? 'dot-remote' : 'dot-onsite'
}

function goToJobs() {
  const q = searchTerm.value.trim()
  router.push({
    path: `/jobs`,
    query: q ? { q } : {}
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

<style scoped>
.home-hero {
  text-align: center;
  border-bottom: var(--color-border);
  padding: 2.5rem 2rem;
  margin: 0 auto;
}

.hero-copy {
  max-width: 40rem;
  margin: 0 auto;

  text-align: center;
}

.hero-eyebrow {
  display: inline-flex;
  align-items: center;
  gap: 0.65rem;
  padding: 0.6rem 1rem;
  background: rgba(94, 106, 210, 0.12);
  color: var(--color-accent);
  font-size: 0.8rem;
  margin-bottom: 1.25rem;
}

.hero-dot {
  width: 7px;
  height: 7px;
  background: var(--color-accent);
}

h1 {
  margin: 0;
  font-size: clamp(2.8rem, 4vw, 3.8rem);
  line-height: 1.05;
  font-family: Newsreader, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Inter', sans-serif;
}

h1 span {
  color: var(--color-accent);
}

.hero-text {
  margin: 1.5rem 0 2rem;
  max-width: 42rem;
  color: var(--color-text);
  font-size: 1.05rem;
  line-height: 1.75;
}


.hero-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  margin-top: 1.75rem;
  display: flex;
  align-items: center;
  justify-content: center;
}

.hero-tag {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0.3rem 1rem;
  background: #f8faff;
  color: var(--color-text);
  border: 1px solid var(--color-border);
  font-size: 0.9rem;
}

.hero-card {
  background: #fff;
  padding: 1.75rem;
}

.hero-card-header {
  font-weight: 700;
  margin-bottom: 1rem;
  color: var(--color-text);
}

.hero-card ul {
  list-style: none;
  padding: 0;
  margin: 0;
  display: grid;
  gap: 1rem;
}

.hero-card li {
  position: relative;
  padding-left: 1.5rem;
  color: var(--color-text);
  font-size: 0.97rem;
}

.hero-card li::before {
  content: '';
  position: absolute;
  width: 6px;
  height: 6px;
  left: 0;
  top: 0.55rem;
  background: var(--color-accent);
}

.stats-sec {
  display: flex;
  flex-direction: column;
  background: var(--color-bg-secondary);
  margin: 0 auto;
  padding: 0 2rem;
  gap: 1.5rem;
}

.stats-sub-sec {
  display: flex;
  flex-direction: column;
  margin-top: 2.5rem;
  width: min(1200px, 100%);
  margin: 2rem auto;
  padding: 0 2rem;
  gap: 1.5rem;
}


.stats-row {
  display: grid;
  width: 100%;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 1rem;
}

.stat-card {
  background: #fff;
  padding: 1.4rem 1.5rem;
  border: 1px solid var(--color-border);
}

.stat-label {
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.12em;
  color: var(--color-accent);
  margin-bottom: 0.8rem;
}

.stat-value {
  font-size: 2rem;
  font-weight: 700;
  color: var(--color-text);
}

.stat-sub {
  margin-top: 0.65rem;
  color: var(--color-text);
  font-size: 0.92rem;
}

.section-heading {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  align-items: flex-end;
  margin-bottom: 1rem;
}

.section-eyebrow {
  margin: 0 0 0.35rem;
  text-transform: uppercase;
  letter-spacing: 0.12em;
  font-size: 0.78rem;
  color: var(--color-text);
}

.section-title {
  margin: 0;
  font-size: 1.25rem;
  color: var(--color-text);
}

.section-link {
  color: var(--color-accent);
  font-size: 0.94rem;
  font-weight: 600;
}

.recommend-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 1rem;
}

.recommend-card {
  background: #fff;
  padding: 1.3rem;
  display: grid;
  gap: 1rem;
}

.card-top {
  display: flex;
  justify-content: space-between;
  align-items: start;
  gap: 1rem;
}

.company-badge {
  width: 40px;
  height: 40px;
  background: var(--color-bg-secondary);
  color: var(--color-text);
  display: grid;
  place-items: center;
  font-weight: 700;
}

.icon-button {
  width: 25px;
  height: 25px;
  color: var(--color-text);
  cursor: pointer;
  flex-shrink: 0;
}

.card-match {
  font-size: 0.82rem;
  color: var(--color-accent);
  font-weight: 700;
}

.card-title {
  margin: 0;
  font-size: 1rem;
  color: var(--color-text);
}

.card-company {
  margin: 0.35rem 0 0;
  color: var(--color-text);
  font-size: 0.92rem;
}

.card-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.meta-pill {
  padding: var(--spacing-1) var(--spacing-2);
  font-size: var(--text-xs);
  background: var(--color-bg-secondary);
  color: var(--color-text);
  border: 1px solid var(--color-border)
}

.card-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  color: var(--color-text);
  font-size: var(--text-xs);
}

.jobs-table {
  background: #fff;
  overflow: hidden;
  border: 1px solid var(--color-border);
}

.jobs-table-row {
  display: grid;
  grid-template-columns: minmax(0, 2.4fr) minmax(0, 1fr) minmax(120px, 0.9fr) minmax(90px, 0.8fr);
  gap: 1rem;
  align-items: center;
  padding: 1rem 1.25rem;
}

.jobs-table-head {
  background: var(--color-bg-secondary);
  color: var(--color-text);
  font-size: var(--text-xs);
  text-transform: uppercase;
  letter-spacing: 0.1em;
}

.jobs-table-row {
  border-bottom: 1px solid var(--color-border);
  color: var(--color-text);
  font-size: var(--text-xs);
  font-weight: var(--font-weight-regular);
}

.position-cell {
  display: flex;
  align-items: center;
  gap: 0.9rem;
}

.job-title {
  margin: 0;
  font-size: 0.95rem;
  font-weight: 600;
}

.job-subtitle {
  margin: 0.35rem 0 0;
  font-size: 0.86rem;
  color: var(--color-text);
}

.company-cell,
.type-cell,
.when-cell {
  font-size: 0.95rem;
  color: var(--color-text);
}

.status-dot {
  width: 8px;
  height: 8px;
  flex-shrink: 0;
}

.dot-remote {
  background: #5e6ad2;
}

.dot-onsite {
  background: #22c55e;
}
</style>
