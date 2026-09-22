<script setup lang="ts">
defineProps<{
  page: number
  pageSize: number
  hasNext: boolean
  loading: boolean
  position: 'top' | 'bottom'
}>()
const emit = defineEmits<{ page: [value: number]; pageSize: [value: number] }>()
</script>
<template>
  <nav
    :aria-label="`Job pages (${position})`"
    class="flex flex-wrap items-center justify-between gap-3 shrink-0 border border-gray-200 bg-white p-3"
  >
    <label class="flex items-center gap-2 text-sm">
      Jobs per page
      <select
        :aria-label="`Jobs per page (${position})`"
        :value="pageSize"
        :disabled="loading"
        class="form-input w-auto"
        @change="emit('pageSize', Number(($event.target as HTMLSelectElement).value))"
      >
        <option v-for="size in [50, 100, 200, 500]" :key="size" :value="size">{{ size }}</option>
      </select>
    </label>
    <div class="flex items-center gap-3">
      <button
        type="button"
        class="button button-secondary"
        :disabled="loading || page <= 1"
        @click="emit('page', page - 1)"
      >
        Previous
      </button>
      <span class="text-sm" aria-live="polite">Page {{ page }}</span>
      <button
        type="button"
        class="button button-secondary"
        :disabled="loading || !hasNext"
        @click="emit('page', page + 1)"
      >
        Next
      </button>
    </div>
  </nav>
</template>
