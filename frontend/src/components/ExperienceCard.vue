<script setup lang="ts">
import type { Experience } from '@/types/experience'
import { Pencil, Trash2 } from '@lucide/vue'
import { computed } from 'vue'

const props = defineProps<{ exp: Experience }>()
const emit = defineEmits<{
  (e: 'edit'): void
  (e: 'delete'): void
}>()

const experience = computed(() => props.exp)

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
          {{ experience.role_title }}
        </h3>
        <p class="flex items-center gap-2 text-sm text-gray-500">
          <span class="text-gray-900 font-medium">{{ experience.company_name }}</span>
          <span v-if="experience.employment_type" class="text-gray-300">•</span>
          <span v-if="experience.employment_type">{{ experience.employment_type }}</span>
        </p>
      </div>

      <div class="flex gap-2">
        <button
          class="border border-gray-200 text-gray-500 p-1 cursor-pointer grid place-items-center transition-all hover:text-accent hover:border-accent"
          @click="emit('edit')"
          aria-label="Edit experience"
        >
          <Pencil :size="14" />
        </button>
        <button
          class="border border-gray-200 text-gray-500 p-1 cursor-pointer grid place-items-center transition-all hover:text-red-600 hover:border-red-600"
          @click="emit('delete')"
          aria-label="Delete experience"
        >
          <Trash2 :size="14" />
        </button>
      </div>
    </div>

    <!-- Meta -->
    <div class="flex items-center gap-2 text-xs text-gray-400 mb-4">
      <span>
        {{ formatDate(experience.start_date) }} —
        {{ experience.is_current ? 'Present' : formatDate(experience.end_date) }}
      </span>
      <span v-if="experience.exp_location" class="text-gray-300">•</span>
      <span v-if="experience.exp_location">{{ experience.exp_location }}</span>
    </div>

    <!-- Description -->
    <p
      v-if="experience.description"
      class="text-sm leading-relaxed text-gray-500 mb-4 whitespace-pre-wrap"
    >
      {{ experience.description }}
    </p>

    <!-- Achievements -->
    <ul v-if="experience.achievements?.length" class="pl-4 mb-4">
      <li
        v-for="(achievement, index) in experience.achievements"
        :key="index"
        class="text-sm text-gray-500 mb-1 leading-snug"
      >
        {{ achievement }}
      </li>
    </ul>

    <!-- Skills -->
    <div v-if="experience.skills?.length" class="flex flex-wrap gap-2">
      <span
        v-for="skill in experience.skills"
        :key="skill.id"
        class="text-[11px] px-2 py-0.5 bg-black/4 border border-gray-200 text-gray-500 font-medium"
      >
        {{ skill.name }}
      </span>
    </div>
  </div>
</template>
