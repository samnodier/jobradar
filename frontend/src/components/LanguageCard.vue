<script setup lang="ts">
import type { Language } from '@/types/language'
import { Pencil, Trash2 } from '@lucide/vue'
import { computed } from 'vue'

const props = defineProps<{ lang: Language }>()
const emit = defineEmits<{
  (e: 'edit'): void
  (e: 'delete'): void
}>()

const l = computed(() => props.lang)

const proficiencyLabels: Record<string, string> = {
  native: 'Native / Bilingual',
  fluent: 'Fluent / Full Professional',
  professional: 'Professional Working',
  elementary: 'Elementary / Basic',
}
</script>

<template>
  <div
    class="flex items-center justify-between p-4 bg-white border border-gray-200 transition-all group"
    @mouseenter="($el as HTMLElement).style.borderColor = 'var(--color-accent)'"
    @mouseleave="($el as HTMLElement).style.borderColor = ''"
  >
    <div>
      <h4 class="text-sm font-semibold text-gray-900 m-0 leading-tight">
        {{ l.user_language }}
      </h4>
      <p class="text-xs text-gray-400 mt-1 m-0">
        {{ proficiencyLabels[l.proficiency || 'professional'] }}
      </p>
    </div>

    <div class="flex gap-1.5 opacity-0 group-hover:opacity-100 transition-opacity">
      <button
        class="border border-gray-100 text-gray-400 p-1 cursor-pointer grid place-items-center transition-all hover:text-accent hover:border-accent"
        @click="emit('edit')"
        aria-label="Edit language"
      >
        <Pencil :size="12" />
      </button>
      <button
        class="border border-gray-100 text-gray-400 p-1 cursor-pointer grid place-items-center transition-all hover:text-red-600 hover:border-red-600"
        @click="emit('delete')"
        aria-label="Delete language"
      >
        <Trash2 :size="12" />
      </button>
    </div>
  </div>
</template>
