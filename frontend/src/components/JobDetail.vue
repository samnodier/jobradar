<template>
  <aside class="job-detail">
    <div class="detail-header">
      <div>
        <p class="detail-eyebrow">Job details</p>
        <h2>{{ job.title }}</h2>
        <p class="company-name">{{ job.company }}</p>
      </div>
      <button class="close-button" @click="$emit('close')">×</button>
    </div>

    <div class="detail-meta">
      <span>{{ job.location || 'Remote' }}</span>
      <span>•</span>
      <span>{{ job.employment_type || 'Full time' }}</span>
      <span>•</span>
      <span>{{ job.salary_min && job.salary_max ? formatSalary(job) : 'Salary not listed' }}</span>
    </div>

    <div class="detail-block">
      <p class="block-title">Description</p>
      <p class="block-text">{{ job.description || 'No description available.' }}</p>
    </div>

    <div class="detail-block">
      <p class="block-title">Skills</p>
      <div class="skill-list">
        <span v-for="skill in job.skills" :key="skill" class="skill-tag">{{ skill }}</span>
      </div>
    </div>

    <a :href="job.url" target="_blank" rel="noreferrer" class="button button-primary">View job posting</a>
  </aside>
</template>

<script setup lang="ts">
import type { Job } from '@/types/job'

defineProps<{ job: Job }>()
defineEmits(['close'])

function formatSalary(job: Job) {
  const currency = job.currency || 'USD'
  return `${currency} ${job.salary_min ?? 0} - ${job.salary_max ?? 0}`
}
</script>

<style scoped>
.job-detail {
  width: 380px;
  padding: 20px;
  background: #fff;
  border-left: 1px solid #e8e8e4;
  box-shadow: -12px 0 35px rgba(15, 23, 42, 0.08);
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.detail-header {
  display: flex;
  gap: 1rem;
  align-items: flex-start;
}

.detail-eyebrow {
  margin: 0 0 0.4rem;
  text-transform: uppercase;
  font-size: 0.75rem;
  color: #64748b;
  letter-spacing: 0.14em;
}

h2 {
  margin: 0;
  font-size: 1.3rem;
}

.company-name {
  margin: 0.5rem 0 0;
  color: #64748b;
}

.close-button {
  margin-left: auto;
  border: none;
  background: transparent;
  color: #64748b;
  font-size: 1.55rem;
  line-height: 1;
  cursor: pointer;
}

.detail-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  color: #475569;
  font-size: 0.9rem;
}

.detail-block {
  display: grid;
  gap: 0.75rem;
}

.block-title {
  margin: 0;
  font-weight: 600;
}

.block-text {
  margin: 0;
  color: #475569;
  line-height: 1.7;
}

.skill-list {
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
