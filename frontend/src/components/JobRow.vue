<template>
  <div
    class="flex items-center px-5 h-11 border-b border-[#f0f0ec] gap-3 transition-colors cursor-pointer"
    :class="selected ? 'bg-[#f0f1fb]' : 'bg-white hover:bg-[#fafaf8]'"
    @click="$emit('click')"
  >
    <span
      v-if="job.is_saved || job.is_applied"
      class="w-1.75 h-1.75 rounded-full shrink-0"
      :style="{ background: 'var(--color-accent)' }"
    />
    <div v-else class="w-1.75 shrink-0" />

    <span class="text-[13.5px] text-[#1a1a1a] flex-1 truncate min-w-0 cursor-pointer">
      {{ job.title }}
    </span>

    <span class="text-xs text-[#888] w-32.5 shrink-0 truncate">
      {{ job.company_name }}
    </span>

    <span
      class="h-5 px-1.75 rounded text-[11px] bg-[#f5f5f3] text-[#888] border border-[#e8e8e4] flex items-center shrink-0"
    >
      {{ job.is_remote ? 'Remote' : 'On-site' }}
    </span>

    <span class="text-xs text-[#bbb] w-17.5 text-right shrink-0">
      {{ timeAgo(job.posted_at) }}
    </span>

    <span class="flex gap-4 text-gray-400">
      <button
        class="cursor-pointer border-none bg-transparent transition-colors hover:text-accent"
        :class="{ 'text-accent': job.is_saved }"
        @click="saveJob"
      >
        <HeartIcon
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
import { HeartIcon, ListCheck, ListPlus, ListX } from '@lucide/vue'
import { ref } from 'vue'

const props = defineProps<{ job: Job; selected: boolean }>()
const emit = defineEmits(['click', 'save', 'apply'])

const isHovered = ref(false)

function timeAgo(dateStr: string | null | undefined): string {
  if (!dateStr) return ''
  const diff = Date.now() - new Date(dateStr).getTime()
  const hours = Math.floor(diff / 3600000)
  if (hours < 1) return 'just now'
  if (hours < 24) return `${hours}h ago`
  const days = Math.floor(hours / 24)
  return `${days}d ago`
}

function addToApplications(event: MouseEvent) {
  event.stopPropagation()
  emit('apply', props.job)
}

function saveJob(event: MouseEvent) {
  event.stopPropagation()
  emit('save', props.job)
}
</script>
