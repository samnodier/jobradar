<script setup lang="ts">
import type { Education } from '@/types/education'
import { Pencil, Trash2 } from '@lucide/vue'
import { computed } from 'vue'

const props = defineProps<{ edu: Education }>()
const emit = defineEmits<{
  (e: 'edit'): void
  (e: 'delete'): void
}>()

const education = computed(() => props.edu)

const formatDate = (dateStr: string | null) => {
  if (!dateStr) return ''
  return dateStr.length > 7 ? dateStr.substring(0, 7) : dateStr
}

// Convert degree type (e.g., 'bachelor') to display label (e.g., "Bachelor's Degree")
const degreeTypeLabel = computed(() => {
  const type = education.value.degree_type?.toLowerCase()
  if (!type) return ''
  switch (type) {
    case 'bachelor':
      return "Bachelor's Degree"
    case 'master':
      return "Master's Degree"
    case 'doctorate':
      return 'Doctoral Degree'
    case 'associate':
      return "Associate's Degree"
    case 'bootcamp':
      return 'Bootcamp Certificate'
    case 'diploma':
      return 'Diploma'
    case 'high_school':
      return 'High School Diploma'
    default:
      return type.charAt(0).toUpperCase() + type.slice(1)
  }
})
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
          {{ education.institution_name }}
        </h3>
        <p class="flex items-center gap-2 text-sm text-gray-500">
          <span class="text-gray-900 font-medium">
            {{ degreeTypeLabel
            }}<span v-if="education.degree_name">: {{ education.degree_name }}</span>
          </span>
          <span v-if="education.field_of_study" class="text-gray-300">•</span>
          <span v-if="education.field_of_study">{{ education.field_of_study }}</span>
        </p>
      </div>

      <div class="flex gap-2">
        <button
          class="border border-gray-200 text-gray-500 p-1 cursor-pointer grid place-items-center transition-all hover:text-accent hover:border-accent"
          @click="emit('edit')"
          aria-label="Edit education"
        >
          <Pencil :size="14" />
        </button>
        <button
          class="border border-gray-200 text-gray-500 p-1 cursor-pointer grid place-items-center transition-all hover:text-red-600 hover:border-red-600"
          @click="emit('delete')"
          aria-label="Delete education"
        >
          <Trash2 :size="14" />
        </button>
      </div>
    </div>

    <!-- Meta / Dates -->
    <div class="flex items-center gap-2 text-xs text-gray-400 mb-3">
      <span>
        {{ formatDate(education.start_date) }} —
        {{ education.is_current ? 'Present' : formatDate(education.end_date) }}
      </span>
      <span v-if="education.is_highlighted" class="text-gray-300">•</span>
      <span
        v-if="education.is_highlighted"
        class="text-[10px] px-1.5 py-0.5 bg-green-50 border border-green-200 text-green-700 font-semibold uppercase tracking-wider"
      >
        Highlighted
      </span>
    </div>

    <!-- Description -->
    <p
      v-if="education.description"
      class="text-sm leading-relaxed text-gray-500 mb-0 whitespace-pre-wrap"
    >
      {{ education.description }}
    </p>
  </div>
</template>
