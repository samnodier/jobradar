<template>
  <div class="job-row" :class="{ selected }" @click="$emit('click')">
    <span class="status-dot" :class="`status-${job.status ?? 'new'}`" />
    <span class="job-title">{{ job.title }}</span>
    <span class="job-company">{{ job.company }}</span>
    <span class="job-tag">{{ job.is_remote ? 'Remote' : 'On-site' }}</span>
    <span class="job-meta">{{ timeAgo(job.posted_at) }}</span>
    <span class="job-actions">
      <button @click="saveJob"><HeartIcon /></button>
      <button @click="addToApplications"><Plus /></button>
    </span>
  </div>
</template>

<script setup lang="ts">
import type { Job } from '@/types/job'
import { HeartIcon, Plus } from '@lucide/vue'

const props = defineProps<{ job: Job; selected: boolean }>()
const emit = defineEmits(['click', 'save', 'apply'])

function timeAgo(dateStr: string | null | undefined): string {
  if (!dateStr) return ''
  const diff = Date.now() - new Date(dateStr).getTime()
  const hours = Math.floor(diff / 3600000)
  if (hours < 1) return 'just now'
  if (hours < 24) return `${hours}h ago`
  const days = Math.floor(hours / 24)
  return `${days}d ago`
}

// Add job to the application job
function addToApplications(event: MouseEvent) {
  event.stopPropagation()
  emit('apply', props.job)
}

function saveJob(event: MouseEvent) {
  event.stopPropagation()
  emit('save', props.job)
}
</script>

<style scoped>
.job-row {
  display: flex;
  align-items: center;
  padding: 0 20px;
  height: 44px;
  border-bottom: 1px solid #f0f0ec;
  gap: 12px;
  transition: background 0.08s;
  background: #fff;
}

.job-row:hover {
  background: #fafaf8;
}

.job-row.selected {
  background: #f0f1fb;
}

.status-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  flex-shrink: 0;
}

.status-new {
  background: #5e6ad2;
}
.status-saved {
  background: #26c6a6;
}
.status-applied {
  background: #f5a623;
}

.job-title {
  font-size: 13.5px;
  color: #1a1a1a;
  flex: 1;
  white-space: nowrap;
  overflow: hidden;
  cursor: pointer;
  text-overflow: ellipsis;
  min-width: 0;
}

.job-company {
  font-size: 12px;
  color: #888;
  width: 130px;
  flex-shrink: 0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.job-tag {
  height: 20px;
  padding: 0 7px;
  border-radius: 4px;
  font-size: 11px;
  background: #f5f5f3;
  color: #888;
  border: 1px solid #e8e8e4;
  display: flex;
  align-items: center;
  flex-shrink: 0;
}

.job-meta {
  font-size: 12px;
  color: #bbb;
  width: 70px;
  text-align: right;
  flex-shrink: 0;
}

.job-actions {
  display: flex;
  column-gap: 1rem;
  color: var(--color-text-muted);
}

.job-actions button {
  cursor: pointer;
  border: none;
  background: transparent;
}

.job-actions button:hover {
  color: var(--color-accent);
}
</style>
