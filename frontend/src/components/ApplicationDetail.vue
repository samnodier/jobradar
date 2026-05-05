<script setup lang="ts">
import type { Application } from '@/types/application'
import { ArrowUpRight, X } from '@lucide/vue'
import { statusLabels } from '@/constants/applicationStatus'

const props = defineProps<{ app: Application }>()
defineEmits(['close'])

function formatDate(date: string | null | undefined): string {
  if (!date) return '—'
  return new Date(date).toLocaleDateString(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  })
}
</script>
<template>
  <aside class="application-detail">
    <!-- Header -->
    <div class="detail-header">
      <div class="company-row">
        <div class="company-info">
          <span class="company-name">{{ props.app.job_company }}</span>
          <span class="status-badge" :data-status="app.status">
            {{ statusLabels[app.status] ?? app.status }}
          </span>
        </div>
        <button class="close-button" @click="$emit('close')">
          <X :size="16" />
        </button>
      </div>
      <h2 class="job-title">{{ app.job_title }}</h2>
    </div>

    <!-- Timeline -->
    <div class="detail-section">
      <p class="section-label">Timeline</p>
      <div class="timeline">
        <div class="timeline-row">
          <span class="timeline-label">Applied</span>
          <span class="timeline-value">{{ formatDate(app.applied_at) }}</span>
        </div>
        <div class="timeline-row">
          <span class="timeline-label">Follow up</span>
          <span class="timeline-value">{{ formatDate(app.follow_up_at) }}</span>
        </div>
      </div>
    </div>

    <!-- Notes -->
    <div class="detail-section">
      <p class="section-label">Notes</p>
      <p class="notes-text">{{ app.notes || 'No notes added yet.' }}</p>
    </div>

    <!-- Actions -->
    <div class="detail-actions" v-if="app.job_url">
      <a :href="app.job_url" target="_blank" rel="noreferrer" class="button-apply">
        View job posting
        <ArrowUpRight :size="13" />
      </a>
    </div>
  </aside>
</template>

<style scoped>
.application-detail {
  background: var(--color-bg-primary);
  display: flex;
  flex-direction: column;
  height: 100%;
  font-family: var(--font-base);
}

/* Header */
.detail-header {
  padding: var(--spacing-4);
  border-bottom: 1px solid var(--color-border);
  flex-shrink: 0;
}

.company-row {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--spacing-2);
  margin-bottom: var(--spacing-2);
}

.company-info {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  flex: 1;
  min-width: 0;
  flex-wrap: wrap;
}

.company-name {
  font-size: var(--text-sm);
  font-weight: var(--font-semibold);
  color: var(--color-text-secondary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.job-title {
  margin: 0;
  font-size: var(--text-xl);
  font-weight: var(--font-bold);
  color: var(--color-text-primary);
  line-height: 1.25;
}

.close-button {
  width: 25px;
  height: 25px;
  border: none;
  background: transparent;
  color: var(--color-text-secondary);
  cursor: pointer;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: color 0.12s ease;
}

.close-button:hover {
  color: var(--color-text-primary);
}

/* Status badge */
.status-badge {
  display: inline-flex;
  align-items: center;
  padding: 2px var(--spacing-2);
  border-radius: 999px;
  font-size: var(--text-xs);
  font-weight: var(--font-semibold);
  background: var(--color-bg-secondary);
  color: var(--color-text-secondary);
}

[data-status='applied'] {
  background: #dbeafe;
  color: #1d4ed8;
}
[data-status='interview'] {
  background: #fef9c3;
  color: #854d0e;
}
[data-status='offer'] {
  background: #dcfce7;
  color: #15803d;
}
[data-status='rejected'] {
  background: #fee2e2;
  color: #b91c1c;
}
[data-status='saved'] {
  background: var(--color-bg-secondary);
  color: var(--color-text-secondary);
}

/* Sections */
.detail-section {
  padding: var(--spacing-4);
  border-bottom: 1px solid var(--color-border);
}

.section-label {
  font-size: var(--text-xs);
  font-weight: var(--font-semibold);
  color: var(--color-text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin-bottom: var(--spacing-3);
}

/* Timeline */
.timeline {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-2);
}

.timeline-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.timeline-label {
  font-size: var(--text-sm);
  color: var(--color-text-secondary);
}

.timeline-value {
  font-size: var(--text-sm);
  color: var(--color-text-primary);
  font-weight: var(--font-semibold);
}

/* Notes */
.notes-text {
  font-size: var(--text-sm);
  color: var(--color-text-secondary);
  line-height: 1.6;
}

/* Actions */
.detail-actions {
  padding: var(--spacing-4);
  margin-top: auto;
  border-top: 1px solid var(--color-border);
  flex-shrink: 0;
}

.button-apply {
  width: 100%;
  height: 32px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-1);
  background: var(--color-accent);
  color: #fff;
  border: none;
  font-size: var(--text-sm);
  font-weight: var(--font-semibold);
  cursor: pointer;
  text-decoration: none;
  transition: opacity 0.15s ease;
  border-radius: var(--radius-md);
}

.button-apply:hover {
  opacity: 0.88;
}
</style>
