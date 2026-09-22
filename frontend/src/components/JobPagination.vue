<script setup lang="ts">
import { ChevronLeft, ChevronRight } from '@lucide/vue'
defineProps<{
  page: number
  pageSize: number
  hasNext: boolean
  loading: boolean
}>()
const emit = defineEmits<{ page: [value: number]; pageSize: [value: number] }>()
</script>
<template>
  <nav
    aria-label="Job pages"
    class="flex flex-wrap items-center justify-between gap-3 shrink-0 border-t border-gray-200 bg-white py-3"
  >
    <label class="flex items-center gap-2 text-sm whitespace-nowrap">
      Jobs per page
      <select
        aria-label="Jobs per page"
        :value="pageSize"
        :disabled="loading"
        class="page-size"
        @change="emit('pageSize', Number(($event.target as HTMLSelectElement).value))"
      >
        <option v-for="size in [50, 100, 200, 500]" :key="size" :value="size">{{ size }}</option>
      </select>
    </label>
    <div class="flex items-center gap-3">
      <button
        type="button"
        class="button button-ghost !p-2"
        aria-label="Previous page"
        title="Previous page"
        :disabled="loading || page <= 1"
        @click="emit('page', page - 1)"
      >
        <ChevronLeft :size="20" aria-hidden="true" />
      </button>
      <span class="text-sm" aria-live="polite">Page {{ page }}</span>
      <button
        type="button"
        class="button button-ghost !p-2"
        aria-label="Next page"
        title="Next page"
        :disabled="loading || !hasNext"
        @click="emit('page', page + 1)"
      >
        <ChevronRight :size="20" aria-hidden="true" />
      </button>
    </div>
  </nav>
</template>

<style scoped>
.page-size {
  width: 5rem;
  padding: 0.4rem;
  border: 1px solid #d1d5db;
  background: white;
}
</style>
