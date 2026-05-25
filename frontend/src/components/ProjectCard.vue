<script setup lang="ts">
import type { Project } from '@/types/project'
import { Pencil, Trash2, ExternalLink, Code } from '@lucide/vue'
import { computed } from 'vue'

const props = defineProps<{ project: Project }>()
const emit = defineEmits<{
  (e: 'edit'): void
  (e: 'delete'): void
}>()

const proj = computed(() => props.project)

const formatDate = (dateStr: string | null) => {
  if (!dateStr) return ''
  return dateStr.length > 7 ? dateStr.substring(0, 7) : dateStr
}
</script>

<template>
  <div
    class="p-5 bg-white border border-gray-200 transition-all group"
    :style="{ '--hover-accent': 'var(--color-accent)' }"
    @mouseenter="($el as HTMLElement).style.borderColor = 'var(--color-accent)'"
    @mouseleave="($el as HTMLElement).style.borderColor = ''"
  >
    <!-- Header -->
    <div class="flex justify-between items-start mb-2">
      <div>
        <h3 class="text-base font-semibold text-gray-900 tracking-tight">
          {{ proj.title }}
        </h3>
        <p v-if="proj.role_title" class="text-sm text-gray-500 font-medium">
          {{ proj.role_title }}
        </p>
      </div>

      <div class="flex gap-2">
        <button
          class="border border-gray-200 text-gray-500 p-1 cursor-pointer grid place-items-center transition-all hover:text-accent hover:border-accent"
          @click="emit('edit')"
          aria-label="Edit project"
        >
          <Pencil :size="14" />
        </button>
        <button
          class="border border-gray-200 text-gray-500 p-1 cursor-pointer grid place-items-center transition-all hover:text-red-600 hover:border-red-600"
          @click="emit('delete')"
          aria-label="Delete project"
        >
          <Trash2 :size="14" />
        </button>
      </div>
    </div>

    <!-- Meta / Links / Dates -->
    <div class="flex flex-wrap items-center gap-3 text-xs text-gray-400 mb-3">
      <span>
        {{ formatDate(proj.start_date) }} —
        {{ proj.is_current ? 'Present' : formatDate(proj.end_date) }}
      </span>

      <span v-if="proj.is_featured" class="text-gray-300">•</span>
      <span
        v-if="proj.is_featured"
        class="text-[10px] px-1.5 py-0.5 bg-yellow-50 border border-yellow-200 text-yellow-700 font-semibold uppercase tracking-wider"
      >
        Featured
      </span>

      <!-- Project URL -->
      <span v-if="proj.project_url" class="text-gray-300">•</span>
      <a
        v-if="proj.project_url"
        :href="proj.project_url"
        target="_blank"
        rel="noreferrer"
        class="inline-flex items-center gap-1 text-accent no-underline hover:underline font-medium"
      >
        <ExternalLink :size="12" /> Live Demo
      </a>

      <!-- Repository URL -->
      <span v-if="proj.repository_url" class="text-gray-300">•</span>
      <a
        v-if="proj.repository_url"
        :href="proj.repository_url"
        target="_blank"
        rel="noreferrer"
        class="inline-flex items-center gap-1 text-accent no-underline hover:underline font-medium"
      >
        <Code :size="12" /> Source Code
      </a>
    </div>

    <!-- Description -->
    <p
      v-if="proj.description"
      class="text-sm leading-relaxed text-gray-500 mb-3 whitespace-pre-wrap"
    >
      {{ proj.description }}
    </p>

    <!-- Impact -->
    <div
      v-if="proj.impact"
      class="mt-3 p-3 bg-black/4 border-l-2 border-accent text-sm text-gray-600 italic"
    >
      <strong
        class="text-gray-900 block text-xs not-italic uppercase font-semibold tracking-wider mb-1"
        >Impact & Results</strong
      >
      {{ proj.impact }}
    </div>
  </div>
</template>
