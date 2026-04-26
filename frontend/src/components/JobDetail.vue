<template>
  <aside class="job-detail">
    <div class="detail-header">
      <div class="company-row">
        <div class="company-info">
          <span class="company-name">{{ job.company }}</span>
          <span class="source-badge">{{ job.source }}</span>
        </div>
        <button class="close-button" @click="$emit('close')">
          <X :size="16" />
      </button>
      </div>
      <h2 class="job-title">{{ job.title }}</h2>

      <div class="meta-row">
        <span class="meta-pill">
          <MapPin :size="12" />
          {{ job.location || (job.is_remote ? 'Remote' : 'On-site') }}
        </span>
        <span class="meta-pill">
          <Briefcase :size="12" />
          {{ job.employment_type || 'Full time' }}
        </span>
        <span class="meta-pill meta-pill--salary">
          <DollarSign :size="12" />
          {{ job.salary_min && job.salary_max ? formatSalary(job) : 'Salary not listed' }}
        </span>
        <span class="meta-pill meta-pill--muted">
          <Clock :size="12" />
          {{ timeAgo(job.posted_at || job.created_at) }}
        </span>
      </div>
    </div>

    <!-- Actions -->
    <div class="detail-actions">
      <a :href="job.url" target="_blank" rel="noreferrer" class="button-apply">
        Apply now
        <ExternalLink :size="13" />
      </a>
      <button class="button-save" :class="{ saved: isSaved }" @click="isSaved = !isSaved">
        {{ isSaved ? 'Saved' : 'Save' }}
      </button>
      <button class="button-icon" @click="copyLink" :title="copied ? 'Copied!' : 'Copy link'">
        <Check v-if="copied" :size="14" />
        <Link2 v-else :size="14" />
      </button>
    </div>

    <div v-if="job.skills?.length" class="detail-section">
      <p class="section-label">Skills</p>
      <div class="skills-wrap">
        <span v-for="skill in job.skills" :key="skill" class="skill-tag">{{ skill }}</span>
      </div>
    </div>

    <div class="detail-block">
      <p class="block-title">Description</p>
      <p class="block-text">{{ job.description || 'No description available.' }}</p>
    </div>


    <a :href="job.url" target="_blank" rel="noreferrer" class="button button-primary">View job posting</a>
  </aside>
</template>

<script setup lang="ts">
import type { Job } from '@/types/job'
import { X } from '@lucide/vue';
import { ref } from 'vue';

const props = defineProps<{ job: Job }>()
defineEmits(['close'])

const isSaved = ref(props.job.status === 'saved')
const copied = ref(false)

function formatSalary(job: Job) {
  const currency = job.currency || 'USD'
  return `${currency} ${job.salary_min ?? 0} - ${job.salary_max ?? 0}`
}

function timeAgo(date: string | null | undefined): string {
  if (!date) return 'Recently posted'
  const diff = Date.now() - new Date(date).getTime()
  const hours = Math.floor(diff / 3600000)
  if (hours < 1) return 'Just now'
  if (hours < 24) return `${hours}h ago`
  const days = Math.floor(hours / 24)
  if (days === 1) return 'Yesterday'
  if (days < 7) return `${days}d ago`
  return new Date(date).toLocaleDateString(undefined, { month: 'short', day: 'numeric' })
}

async function copyLink() {
  try {
    await navigator.clipboard.writeText(props.job.url)
    copied.value = true
    setTimeout(() => copied.value = false, 2000)
  } catch {

  }
}
</script>

<style scoped>
.job-detail {
  background: var(--color-bg-primary);
  display: flex;
  flex-direction: column;
  height: 100%;
  font-family: var(--font-base);
  gap: var(--spacing-2);
}

.detail-header {
  padding: var(--spacing-4);
  border-bottom: 1px solid var(--color-border);
  flex-shrink: 0;
}

.company-row {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  margin-bottom: var(--spacing-1);
}

.company-info {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-1);
  flex: 1;
  min-width: 0;
}

.company-name {
  font-size: var(--text-sm);
  color: var(--muted);
  font-weight: 600;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.source-badge {
  color: var(--text);
  font-size: var(--text-xs);
  text-transform: uppercase;
}

.job-title {
  margin: 0 0 var(--spacing-1);
  font-size: var(--text-3xl);
  font-weight: var(--font-weight-bold);
  line-height: 1.4;
}

.company-name {
  margin: 0.5rem 0 0;
  color: #64748b;
}

.close-button {
  width: 25px;
  height: 25px;
  border: none;
  background: transparent;
  color: var(--color-muted);
  font-size: 1.55rem;
  line-height: 1;
  cursor: pointer;
  flex-shrink: 0;
  transition: var(--transition-fast);
  display: flex;
  align-items: center;
  justify-content: center;
}

.meta-row {
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-2);
}

.meta-pill {
  display: inline-flex;
  align-items: center;
  gap: var(--spacing-1);
  font-size: var(--text-xs);
  color: var(--color-muted);
  background: var(--color-bg-secondary);
  padding: 0 var(--spacing-2);
  border: 1px solid var(--color-border);
  height: 22px;
}

.meta-pill--salary {
  color: var(--color-accent);
  border-color: var(--color-accent);
  font-weight: var(--font-weight-medium);
}

.meta-pill--muted {
  background: var(--color-bg-primary);
  border-color: var(--color-border);
}

.detail-actions {
  display: flex;
  gap: 8px;
  padding: var(--spacing-4);
  border-bottom: 1px solid var(--color-border);
  flex-shrink: 0;
}

.button-apply {
  flex: 1;
  height: 32px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-1);
  background: var(--color-accent);
  color: #FFF;
  border: none;
  font-size: 13px;
  font-weight: var(--font-weight-medium);
  cursor: pointer;
  text-decoration: none;
  transition: var(--transition-fast);
}

.button-apply--full {
  height: 36px;
  font-size: var(--text-base);
}

.detail-section {
  display: grid;
  gap: 0.75rem;
}

.section-label {
  margin: 0;
  color: #475569;
  line-height: 1.7;
}

.skills-wrap {
  display: flex;
  flex-wrap: wrap;
  gap: 0.65rem;
}

.skill-tag {
  background: #f5f5f3;
  color: #1a1a1a;
  padding: 0.45rem 0.8rem;
  border-radius: 999px;
  font-size: 0.85rem;
}

.button-primary {
  width: 100%;
  justify-content: center;
  display: inline-flex;
}
</style>
