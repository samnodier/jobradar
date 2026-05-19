<template>
  <aside class="bg-white flex flex-col h-full gap-2" style="font-family: var(--font-base)">
    <!-- Header -->
    <div class="px-4 py-4 border-b border-gray-200 shrink-0">
      <div class="flex items-center gap-2 mb-1">
        <div class="flex flex-col gap-1 flex-1 min-w-0">
          <span class="text-sm text-gray-500 font-semibold truncate">
            {{ job.company_name }}
          </span>
          <span class="text-xs text-gray-900 uppercase">
            {{ job.job_source }}
          </span>
        </div>
        <button
          class="w-25px h-25px bg-transparent border-none text-gray-400 cursor-pointer shrink-0 flex items-center justify-center transition-all"
          @click="$emit('close')"
        >
          <X :size="16" />
        </button>
      </div>

      <h2 class="m-0 mb-1 text-3xl font-bold leading-tight">
        {{ job.title }}
      </h2>

      <div class="flex flex-wrap gap-2">
        <span
          class="inline-flex items-center gap-1 text-xs text-gray-500 bg-black/0.04 px-2 border border-gray-200 h-22px"
        >
          <MapPin :size="12" />
          {{ job.job_location || (job.is_remote ? 'Remote' : 'On-site') }}
        </span>
        <span
          class="inline-flex items-center gap-1 text-xs text-gray-500 bg-black/0.04 px-2 border border-gray-200 h-22px"
        >
          <Briefcase :size="12" />
          {{ job.employment_type || 'Full time' }}
        </span>
        <span
          class="inline-flex items-center gap-1 text-xs px-2 border h-22px font-medium"
          :style="{
            color: 'var(--color-accent)',
            borderColor: 'var(--color-accent)',
            background: 'var(--color-accent-soft)',
          }"
        >
          <DollarSign :size="12" />
          {{ job.salary_min && job.salary_max ? formatSalary(job) : 'Salary not listed' }}
        </span>
        <span
          class="inline-flex items-center gap-1 text-xs text-gray-500 bg-white px-2 border border-gray-200 h-22px"
        >
          <Clock :size="12" />
          {{ timeAgo(job.posted_at || job.created_at) }}
        </span>
      </div>
    </div>

    <!-- Actions -->
    <div class="flex gap-2 px-4 py-4 border-b border-gray-200 shrink-0">
      <a
        :href="job.source_url"
        target="_blank"
        rel="noreferrer"
        class="flex-1 h-8 inline-flex items-center justify-center gap-1 text-white text-sm font-medium cursor-pointer no-underline transition-all"
        :style="{ background: 'var(--color-accent)' }"
      >
        Apply now
        <ExternalLink :size="13" />
      </a>
      <button
        class="h-8 px-14px border text-sm cursor-pointer transition-all whitespace-nowrap"
        :class="
          job.is_saved
            ? 'text-white border-transparent'
            : 'border-gray-200 bg-black/0.04 text-gray-900'
        "
        :style="
          job.is_saved
            ? { background: 'var(--color-accent)', borderColor: 'var(--color-accent)' }
            : {}
        "
        @click="$emit('save', job)"
      >
        {{ job.is_saved ? 'Saved' : 'Save' }}
      </button>
      <button
        class="w-8 h-8 border border-gray-200 bg-black/0.04 text-gray-900 text-sm cursor-pointer flex items-center justify-center shrink-0 transition-all"
        @click="copyLink"
        :title="copied ? 'Copied!' : 'Copy link'"
      >
        <Check v-if="copied" :size="14" />
        <Link2 v-else :size="14" />
      </button>
    </div>

    <!-- Skills -->
    <div v-if="job.skills?.length" class="px-4 py-4">
      <p class="text-xs font-medium text-gray-500 uppercase mb-2">Skills</p>
      <div class="flex flex-wrap gap-2">
        <span
          v-for="skill in job.skills"
          :key="skill"
          class="h-5 text-gray-900 px-2 text-xs flex items-center capitalize border"
          :style="{ background: 'var(--color-accent-lighter)', borderColor: 'var(--color-accent)' }"
        >
          {{ skill }}
        </span>
      </div>
    </div>

    <!-- Description -->
    <div class="px-4 py-4">
      <p class="text-xs font-medium text-gray-500 uppercase mb-2">About the role</p>
      <div
        v-html="DOMPurify.sanitize(job.description || 'No description available.')"
        class="raw-description text-sm text-gray-900 leading-snug m-0 whitespace-pre-line"
      ></div>
    </div>

    <!-- Footer CTA -->
    <a
      :href="job.source_url"
      target="_blank"
      rel="noreferrer"
      class="w-full inline-flex items-center justify-center h-9 text-base font-semibold text-white no-underline transition-all"
      :style="{ background: 'var(--color-accent)' }"
    >
      View job posting
    </a>
  </aside>
</template>

<script setup lang="ts">
import type { Job } from '@/types/job'
import { Check, ExternalLink, Link2, X } from '@lucide/vue'
import { ref } from 'vue'
import DOMPurify from 'dompurify'

const props = defineProps<{ job: Job }>()
defineEmits(['close', 'save'])

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
    await navigator.clipboard.writeText(props.job.source_url)
    copied.value = true
    setTimeout(() => (copied.value = false), 2000)
  } catch {}
}
</script>

<style scoped>
/* Only :deep() rules remain — Tailwind cannot reach into v-html content */
.raw-description :deep(h1) {
  font-size: 1rem;
  font-weight: 600;
  color: #111827;
  margin: 0 0 12px;
  line-height: 1.5;
}
.raw-description :deep(h2) {
  font-size: 1rem;
  font-weight: 600;
  color: #111827;
  margin: 20px 0 8px;
}
.raw-description :deep(h3) {
  font-size: 0.875rem;
  font-weight: 600;
  color: #111827;
  margin: 14px 0 6px;
}
.raw-description :deep(p) {
  font-size: 0.875rem;
  color: #111827;
  line-height: 1.375;
  margin: 0 0 6px;
}
.raw-description :deep(span) {
  font-size: 0.875rem !important;
  color: #111827;
  line-height: 1.375;
  margin: 0 0 6px;
}
.raw-description :deep(p:empty),
.raw-description :deep(p[style*='min-height']:empty) {
  display: none;
}
.raw-description :deep(ul) {
  padding-left: 16px;
  margin: 6px 0 12px;
}
.raw-description :deep(li) {
  font-size: 0.875rem;
  color: #111827;
  line-height: 1.375;
  margin-bottom: 4px;
}
.raw-description :deep(li p) {
  margin: 0;
  display: inline;
}
.raw-description :deep(strong) {
  font-weight: 600;
  color: #111827;
}
.raw-description :deep(a) {
  color: var(--color-accent);
  text-decoration: none;
}
.raw-description :deep(a:hover) {
  text-decoration: underline;
}
.raw-description :deep(br:last-child),
.raw-description :deep(br + br) {
  display: none;
}
.raw-description :deep(img) {
  display: none;
}
</style>
