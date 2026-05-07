<script setup lang="ts">
import type { Application } from '@/types/application'
import { ArrowUpRight, Building2, Calendar, Clock, MapPin, RefreshCw, Wifi, X } from '@lucide/vue'
import { statusLabels, statusOrder } from '@/constants/applicationStatus'
import { computed, ref, watch } from 'vue'
import { useToast } from '@/composables/useToast'

const toast = useToast()

const formatDateTime = new Intl.DateTimeFormat('en-US', {
  dateStyle: 'medium',
  timeStyle: 'short',
})

const props = defineProps<{ app: Application }>()
const emit = defineEmits(['close', 'updated'])

const currentText = ref(props.app.notes || '')
const isSaving = ref(false)
const lastSaved = ref<string | null>(null)
let debounceTimer: ReturnType<typeof setTimeout> | null = null

const companyInitials = computed(() => {
  const words = props.app.company_name?.split(/\s+/) ?? []
  if (words.length === 0) return '?'
  if (words.length === 1) return words[0]!.slice(0, 2).toUpperCase()
  return (words[0]![0]! + words[1]![0]!).toUpperCase()
})

const formatDateForInput = (date: string | null | undefined) => {
  if (!date) return ''
  return new Date(date).toISOString().split('T')[0]
}

async function updateApplication(payload: Partial<Application>) {
  isSaving.value = true

  // Map frontend field names to backend JSON names (pointers)
  const body = {
    application_notes: payload.notes,
    application_status: payload.application_status,
    applied_at: payload.applied_at,
    follow_up_at: payload.follow_up_at,
  }

  try {
    const response = await fetch(`/api/applications/${props.app.id}`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify(body),
    })

    if (!response.ok) {
      const data = await response.json().catch(() => null)
      throw new Error(data?.error ?? 'Update failed')
    }

    const data = await response.json()
    emit('updated', data)
    lastSaved.value = formatDateTime.format(new Date(data.updated_at))
  } catch (err) {
    toast.error(err instanceof Error ? err.message : 'Failed to save changes')
  } finally {
    isSaving.value = false
  }
}

watch(
  () => props.app.id,
  () => {
    currentText.value = props.app.notes ?? ''
    if (props.app.updated_at) {
      lastSaved.value = formatDateTime.format(new Date(props.app.updated_at))
    } else {
      lastSaved.value = null
    }
  },
  { immediate: true },
)

function handleNotesInput() {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    updateApplication({ notes: currentText.value })
  }, 2000)
}

function handleStatusChange(event: Event) {
  const val = (event.target as HTMLSelectElement).value
  updateApplication({ application_status: val })
}

function handleDateChange(field: 'applied_at' | 'follow_up_at', event: Event) {
  const val = (event.target as HTMLInputElement).value
  updateApplication({ [field]: val ? new Date(val).toISOString() : null })
}

function formatDate(date: string | null | undefined): string {
  if (!date) return '—'
  return new Date(date).toLocaleDateString(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  })
}

function formatDateShort(date: string | null | undefined): string {
  if (!date) return '—'
  return new Date(date).toLocaleDateString(undefined, {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  })
}
</script>

<template>
  <aside class="application-detail">
    <!-- Header -->
    <div class="detail-header">
      <div class="header-top">
        <div class="company-logo">
          <img
            v-if="app.job_logo_url"
            :src="app.job_logo_url"
            :alt="app.company_name"
            class="logo-img"
          />
          <span v-else class="logo-initials">{{ companyInitials }}</span>
        </div>

        <div class="header-meta">
          <div class="title-row">
            <h2 class="job-title">{{ app.job_title }}</h2>
            <button class="close-button" @click="$emit('close')">
              <X :size="16" />
            </button>
          </div>
          <div class="company-row">
            <Building2 :size="13" class="meta-icon" />
            <span class="company-name">{{ app.company_name }}</span>
          </div>
          <div class="meta-row">
            <div v-if="app.job_location || app.job_is_remote" class="meta-chip">
              <MapPin :size="12" />
              <span>{{
                app.job_is_remote ? 'Remote' : (app.job_location ?? 'Location unknown')
              }}</span>
            </div>
            <div v-if="app.job_is_remote" class="meta-chip meta-chip--remote">
              <Wifi :size="12" />
              <span>Remote</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Status selector -->
      <div class="status-row">
        <label class="status-label-text">Status</label>
        <select
          class="status-select"
          :value="app.application_status"
          :data-status="app.application_status"
          @change="handleStatusChange"
        >
          <option v-for="status in statusOrder" :key="status" :value="status">
            {{ statusLabels[status] }}
          </option>
        </select>
      </div>
    </div>

    <!-- Scrollable body -->
    <div class="detail-body">
      <!-- Timeline -->
      <div class="detail-section">
        <p class="section-label">Timeline</p>
        <div class="timeline">
          <div class="timeline-row">
            <div class="timeline-icon-wrap">
              <Calendar :size="13" />
            </div>
            <span class="timeline-label">Applied</span>
            <input
              type="date"
              class="date-input"
              :value="formatDateForInput(app.applied_at)"
              @change="handleDateChange('applied_at', $event)"
            />
          </div>
          <div class="timeline-row">
            <div class="timeline-icon-wrap">
              <Clock :size="13" />
            </div>
            <span class="timeline-label">Follow up</span>
            <input
              type="date"
              class="date-input"
              :value="formatDateForInput(app.follow_up_at)"
              @change="handleDateChange('follow_up_at', $event)"
            />
          </div>
          <div class="timeline-row">
            <div class="timeline-icon-wrap">
              <RefreshCw :size="13" />
            </div>
            <span class="timeline-label">Last Activity</span>
            <span class="timeline-value">{{ formatDate(app.last_status_changed_at) }}</span>
          </div>
          <div class="timeline-row">
            <div class="timeline-icon-wrap">
              <Clock :size="13" />
            </div>
            <span class="timeline-label">Tracked on</span>
            <span class="timeline-value">{{ formatDateShort(app.created_at) }}</span>
          </div>
        </div>
      </div>

      <!-- Notes -->
      <div class="detail-section detail-section--notes">
        <div class="notes-header">
          <h2 class="section-label">Notes</h2>
          <p class="notes-status" v-if="isSaving">Saving...</p>
        </div>
        <textarea
          v-model="currentText"
          @input="handleNotesInput"
          class="notes-textarea"
          placeholder="Add your research, interview notes, or follow-up plan here..."
        ></textarea>
        <p class="notes-status" v-if="lastSaved && !isSaving">Last saved: {{ lastSaved }}</p>
      </div>
    </div>
    <!-- Actions -->
    <div class="detail-footer" v-if="app.source_url">
      <a :href="app.source_url" target="_blank" rel="noreferrer" class="button-apply">
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
  display: flex;
  flex-direction: column;
  gap: var(--spacing-3);
}

.header-top {
  display: flex;
  align-items: flex-start;
  gap: var(--spacing-3);
}

.company-logo {
  width: 40px;
  height: 40px;
  border-radius: var(--radius-md);
  border: 1px solid var(--color-border);
  background: var(--color-bg-secondary);
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
}

.logo-img {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.logo-initials {
  font-size: var(--text-base);
  font-weight: var(--font-bold);
  color: var(--color-text-muted);
  letter-spacing: 0.02em;
}

/* Title / meta */
.header-meta {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: var(--spacing-1);
}

.title-row {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--spacing-2);
}

.job-title {
  margin: 0;
  font-size: var(--text-xl);
  font-weight: var(--font-bold);
  color: var(--color-text-primary);
  line-height: 1.25;
}

.company-row {
  display: flex;
  align-items: flex-start;
  gap: var(--spacing-2);
}

.company-name {
  font-size: var(--text-sm);
  font-weight: var(--font-semibold);
  color: var(--color-text-muted);
}

.meta-icon {
  color: var(--color-text-muted);
  flex-shrink: 0;
}

.meta-row {
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-2);
  margin-top: var(--spacing-1);
}

.meta-chip {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 2px var(--spacing-2);
  border-radius: 999px;
  font-size: var(--text-xs);
  font-weight: var(--font-semibold);
  background: var(--color-bg-secondary);
  color: var(--color-text-secondary);
  border: 1px solid var(--color-border);
}

.meta-chip--remote {
  background: #dcfce7;
  color: #15803d;
  border-color: #bbf7d0;
}

.close-button {
  width: 25px;
  height: 25px;
  border: none;
  background: transparent;
  color: var(--color-text-muted);
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

/* Status row*/
.status-row {
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
}

.status-label-text {
  font-size: var(--text-xs);
  font-weight: var(--font-semibold);
  color: var(--color-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  white-space: nowrap;
}

.status-select {
  flex: 1;
  padding: var(--spacing-1) var(--spacing-2);
  border-radius: var(--radius-md);
  border: 1px solid var(--color-border);
  font-size: var(--text-sm);
  font-weight: var(--font-semibold);
  font-family: var(--font-base);
  background: var(--color-bg-secondary);
  color: var(--color-text-primary);
  cursor: pointer;
  transition: border-color 0.15s ease;
  appearance: auto;
}

.status-select:focus {
  outline: none;
  border-color: var(--color-accent);
}

.status-select[data-status='applied'] {
  background: #dbeafe;
  color: #1d4ed8;
  border-color: #bfdbfe;
}
.status-select[data-status='interview'] {
  background: #fef9c3;
  color: #854d0e;
  border-color: #fef08a;
}
.status-select[data-status='offer'] {
  background: #dcfce7;
  color: #15803d;
  border-color: #bbf7d0;
}
.status-select[data-status='rejected'] {
  background: #fee2e2;
  color: #b91c1c;
  border-color: #fecaca;
}
.status-select[data-status='saved'] {
  background: var(--color-bg-secondary);
  color: var(--color-text-muted);
}

.detail-body {
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
}

/* Sections */
.detail-section {
  padding: var(--spacing-4);
  border-bottom: 1px solid var(--color-border);
}

.detail-section--notes {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.section-label {
  font-size: var(--text-xs);
  font-weight: var(--font-semibold);
  color: var(--color-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin: 0 0 var(--spacing-3) 0;
}

/* Timeline */
.timeline {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-2);
}

.timeline-row {
  display: grid;
  grid-template-columns: 20px 1fr auto;
  align-items: center;
  gap: var(--spacing-2);
}

.timeline-icon-wrap {
  color: var(--color-text-secondary);
  display: flex;
  align-items: center;
  justify-content: center;
}

.timeline-label {
  font-size: var(--text-sm);
  color: var(--color-text-muted);
}

.timeline-value {
  font-size: var(--text-sm);
  color: var(--color-text-primary);
  font-weight: var(--font-semibold);
  text-align: right;
}

.date-input {
  font-size: var(--text-xs);
  font-family: var(--font-base);
  border: 1px solid var(--color-border);
  background: var(--color-bg-secondary);
  color: var(--color-text-primary);
  padding: 2px 4px;
  border-radius: var(--radius-sm);
  cursor: pointer;
  display: block;
  margin-left: auto;
}

/* Notes */
.notes-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--spacing-3);
}

.notes-header .section-label {
  margin: 0;
}

.notes-saving {
  font-size: var(--text-xs);
  color: var(--color-text-muted);
  font-style: italic;
}

.notes-textarea {
  font-size: var(--text-sm);
  color: var(--color-text);
  line-height: 1.6;
  width: 100%;
  font-family: monospace;
  min-height: 140px;
  flex: 1;
  padding: var(--spacing-3);
  resize: vertical;
  border: 1px solid var(--color-border);
  background: var(--color-bg-secondary);
  transition: border-color 0.15s ease;
}

.notes-textarea:focus {
  outline: none;
  border-color: var(--color-accent);
  background: var(--color-bg-primary);
}

.notes-status {
  font-size: var(--text-xs);
  color: var(--color-text-muted);
  margin-top: var(--spacing-2);
}

/* Actions */
.detail-footer {
  padding: var(--spacing-4);
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
}

.button-apply:hover {
  opacity: 0.88;
}
</style>
