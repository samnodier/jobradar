<template>
  <div
    class="flex items-center px-5 h-11 border-b border-zinc-200 gap-3 transition-colors cursor-pointer"
    :class="selected ? 'bg-indigo-50' : 'bg-white hover:bg-zinc-50'"
    @click="$emit('click')"
  >
    <span class="text-sm text-[#1a1a1a] flex-1 truncate min-w-0 cursor-pointer">
      {{ job.title }}
    </span>

    <span class="text-xs text-zinc-400 w-32.5 shrink-0 truncate">
      {{ job.company_name }}
    </span>

    <span
      class="h-5 px-1.75 text-xs bg-zinc-100 text-zinc-400 border border-zinc-200 flex items-center shrink-0"
    >
      {{ job.is_remote ? 'Remote' : 'On-site' }}
    </span>

    <span
      v-if="job.is_matched"
      :class="[
        'min-w-2 text-xs font-medium  h-5 px-1.75  flex items-center shrink-0',
        scoreBadgeClass,
      ]"
    >
      {{ job.is_matched ? `${job.match_score}%` : '' }}
    </span>
    <span v-else class="min-w-2 h-5 px-1.75 flex items-center shrink-0"> </span>

    <span class="text-xs text-zinc-200 w-17.5 text-right shrink-0">
      {{ timeAgo(job.posted_at) }}
    </span>

    <span class="flex gap-4 text-gray-400">
      <button
        class="cursor-pointer border-none bg-transparent transition-colors hover:text-accent"
        :class="{ 'text-accent': job.is_saved }"
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

<script setup lang="ts">
import type { Job } from '@/types/job'
import { useJobFormatting } from '@/composables/useJobFormatting'
import { Bookmark, ListCheck, ListPlus, ListX } from '@lucide/vue'
import { ref } from 'vue'

const { scoreBadgeClass, timeAgo } = useJobFormatting()
const props = defineProps<{ job: Job; selected: boolean }>()
const emit = defineEmits(['click', 'save', 'apply'])

const isHovered = ref(false)

function addToApplications(event: MouseEvent) {
  event.stopPropagation()
  emit('apply', props.job)
}

function saveJob(event: MouseEvent) {
  event.stopPropagation()
  emit('save', props.job)
}
</script>
