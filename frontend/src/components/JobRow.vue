<script setup lang="ts">
import type { Job } from '@/types/job'
import { useJobFormatting } from '@/composables/useJobFormatting'
import { Bookmark, ListCheck, ListPlus, ListX, Sparkles } from '@lucide/vue'
import { computed, ref } from 'vue'

const { scoreBadgeClass, timeAgo } = useJobFormatting()
const props = defineProps<{ job: Job; selected: boolean }>()
const emit = defineEmits(['click', 'save', 'apply'])

const isHovered = ref(false)
const sourceName = computed(() => {
  const names: Record<string, string> = {
    adzuna: 'Adzuna',
    remoteok: 'RemoteOK',
    greenhouse: 'Greenhouse',
    jobicy: 'Jobicy',
    myjobmag: 'MyJobMag',
    manual: 'Manual',
  }
  return names[props.job.job_source?.toLowerCase()] ?? props.job.job_source ?? 'Unknown source'
})

function addToApplications(event: MouseEvent) {
  event.stopPropagation()
  emit('apply', props.job)
}

function saveJob(event: MouseEvent) {
  event.stopPropagation()
  emit('save', props.job)
}
</script>

<template>
  <div
    class="job-row border-b border-zinc-200 transition-colors cursor-pointer"
    :class="selected ? 'bg-indigo-50' : 'bg-white hover:bg-zinc-50'"
    @click="$emit('click')"
  >
    <div class="min-w-0">
      <div class="text-sm font-medium text-gray-900 truncate job-title" :title="job.title">{{ job.title }}</div>
      <div class="text-xs text-gray-500 truncate mt-1" :title="job.company_name">
        {{ job.company_name }}
      </div>
    </div>
    <span class="text-xs text-gray-600 truncate source" :title="sourceName">{{ sourceName }}</span>

    <span
      class="h-5 px-1.75 text-xs bg-zinc-100 text-gray-600 border border-zinc-200 flex items-center shrink-0"
    >
      {{ job.is_remote == null ? 'Unknown' : job.is_remote ? 'Remote' : 'On-site' }}
    </span>

    <span v-if="job.is_matched" class="score text-xs font-medium flex items-center justify-center">
      <p
        :class="[
          'w-10 h-5 px-1.75 mr-1  shrink-0 text-center',
          scoreBadgeClass(job.match_score ?? 0),
        ]"
      >
        {{ job.is_matched ? `${job.match_score}%` : '' }}
      </p>
      <span class="text-accent h-5 w-5" v-if="job.is_enriched">
        <Sparkles :size="20" color="currentColor" />
      </span>
      <span v-else class="h-5 w-5"></span>
    </span>
    <span v-else class="score"> </span>

    <span class="text-xs text-gray-500 text-right whitespace-nowrap age">
      {{ timeAgo(job.posted_at) }}
    </span>

    <span class="actions flex gap-3 text-gray-500">
      <button
        class="cursor-pointer border-none bg-transparent transition-colors hover:text-accent"
        :class="{ 'text-accent': job.is_saved }"
        :aria-label="job.is_saved ? 'Unsave job' : 'Save job'"
        @click="saveJob"
      >
        <Bookmark
          :fill="job.is_saved ? 'var(--color-accent)' : 'none'"
          :color="job.is_saved ? 'var(--color-accent)' : 'currentColor'"
        />
      </button>
      <button
        class="cursor-pointer border-none bg-transparent transition-colors hover:text-accent disabled:opacity-50 disabled:cursor-not-allowed"
        :class="{ 'text-accent': job.is_applied }"
        aria-label="Track application"
        :disabled="job.is_applied"
        @mouseenter="isHovered = true"
        @mouseleave="isHovered = false"
        @click="addToApplications"
      >
        <ListPlus v-if="!job.is_applied" />
        <ListX v-else-if="isHovered && job.is_applied" />
        <ListCheck v-if="!isHovered && job.is_applied" color="currentColor" />
      </button>
    </span>
  </div>
</template>

<style scoped>
.job-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 130px 65px 70px 70px 60px;
  align-items: center;
  gap: 12px;
  min-height: 52px;
  padding: 8px 12px;
}
@media (max-width: 1150px) {
  .job-row {
    grid-template-columns: minmax(0, 1fr) 70px 60px;
    gap: 8px 12px;
  }
  .job-row > :first-child {
    grid-column: 1 / 3;
    white-space: normal;
  }
  .job-title {
    white-space: normal;
  }
  .actions {
    grid-column: 3;
    grid-row: 1;
  }
  .source {
    grid-column: 1;
  }
  .age {
    grid-column: 1;
    grid-row: 3;
    text-align: left;
  }
  .score {
    grid-column: 3;
    grid-row: 2;
  }
}
</style>
