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

      <div class="hero-actions">
        <input v-model="searchTerm" @keyup.enter="goToJobs" type="text" placeholder="Search by title, company, or skill"
          class="search-input" />
        <button @click="goToJobs" class="button button-primary">Browse jobs</button>
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
            <div class="company-badge">{{ job.company?.slice(0, 2).toUpperCase() }}</div>
            <button class="icon-button">♡</button>
          </div>
          <div class="card-match">{{ 92 - index * 4 }}% match</div>
          <h3 class="card-title">{{ job.title }}</h3>
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
  </section>

</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

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
  display: flex;
  justify-content: center;
  gap: 2rem;
  align-items: center;
  padding: 2.5rem 0;
}

.hero-copy {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  max-width: 42rem;
  text-align: center;
}

.hero-eyebrow {
  display: inline-flex;
  align-items: center;
  gap: 0.65rem;
  padding: 0.6rem 1rem;
  background: rgba(94, 106, 210, 0.12);
  color: #5e6ad2;
  font-size: 0.82rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.12em;
  margin-bottom: 1.25rem;
}

.hero-dot {
  width: 7px;
  height: 7px;
  background: #5e6ad2;
}

h1 {
  margin: 0;
  font-size: clamp(2.8rem, 4vw, 3.8rem);
  line-height: 1.05;
  font-family: Newsreader, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Inter', sans-serif;
}

h1 span {
  color: var(--accent);
}

.hero-text {
  margin: 1.5rem 0 2rem;
  max-width: 42rem;
  color: #475569;
  font-size: 1.05rem;
  line-height: 1.75;
}

.hero-actions {
  display: flex;
  gap: 1rem;
  flex-wrap: wrap;
}

.hero-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  margin-top: 1.75rem;
}

.hero-tag {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0.3rem 1rem;
  background: #f8faff;
  color: #475569;
  border: 1px solid rgba(94, 106, 210, 0.18);
  font-size: 0.9rem;
}

.hero-card {
  background: #fff;
  padding: 1.75rem;
  box-shadow: 0 24px 60px rgba(15, 23, 42, 0.08);
}

.hero-card-header {
  font-weight: 700;
  margin-bottom: 1rem;
  color: #111827;
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
  color: #475569;
  font-size: 0.97rem;
}

.hero-card li::before {
  content: '';
  position: absolute;
  width: 6px;
  height: 6px;
  left: 0;
  top: 0.55rem;
  background: var(--accent);
}

.stats-sec {
  display: flex;
  flex-direction: column;
  margin-top: 2.5rem;
}

.stats-row {
  display: flex;
  flex-direction: row;
  width: 100%;
}

.stat-card {
  background: #fff;
  padding: 1.4rem 1.5rem;
  box-shadow: 0 16px 40px rgba(15, 23, 42, 0.06);
}

.stat-label {
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.12em;
  color: #6b7280;
  margin-bottom: 0.8rem;
}

.stat-value {
  font-size: 2rem;
  font-weight: 700;
  color: #111827;
}

.stat-sub {
  margin-top: 0.65rem;
  color: #6b7280;
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
  color: #6b7280;
}

.section-title {
  margin: 0;
  font-size: 1.25rem;
  color: #111827;
}

.section-link {
  color: var(--accent);
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
  box-shadow: 0 16px 40px rgba(15, 23, 42, 0.05);
  display: grid;
  gap: 1rem;
}

.card-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
}

.company-badge {
  width: 40px;
  height: 40px;
  background: #eef2ff;
  color: #374151;
  display: grid;
  place-items: center;
  font-weight: 700;
}

.icon-button {
  width: 34px;
  height: 34px;
  border: 1px solid rgba(148, 163, 184, 0.25);
  background: transparent;
  color: #6b7280;
  cursor: pointer;
}

.card-match {
  font-size: 0.82rem;
  color: #4f46e5;
  font-weight: 700;
}

.card-title {
  margin: 0;
  font-size: 1rem;
  color: #111827;
}

.card-company {
  margin: 0.35rem 0 0;
  color: #6b7280;
  font-size: 0.92rem;
}

.card-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.meta-pill {
  padding: 0.55rem 0.75rem;
  font-size: 0.78rem;
  background: #f8faff;
  color: #475569;
}

.card-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  color: #6b7280;
  font-size: 0.9rem;
}

.jobs-table {
  background: #fff;
  box-shadow: 0 16px 40px rgba(15, 23, 42, 0.05);
  overflow: hidden;
}

.jobs-table-row {
  display: grid;
  grid-template-columns: minmax(0, 2.4fr) minmax(0, 1fr) minmax(120px, 0.9fr) minmax(90px, 0.8fr);
  gap: 1rem;
  align-items: center;
  padding: 1rem 1.25rem;
}

.jobs-table-head {
  background: #f8fafc;
  color: #6b7280;
  font-size: 0.82rem;
  text-transform: uppercase;
  letter-spacing: 0.1em;
}

.jobs-table-row {
  border-top: 1px solid #f0f0ec;
  color: #111827;
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
  color: #6b7280;
}

.company-cell,
.type-cell,
.when-cell {
  font-size: 0.95rem;
  color: #475569;
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
