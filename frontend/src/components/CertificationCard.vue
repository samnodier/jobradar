<script setup lang="ts">
import type { Certification } from '@/types/certification'
import { Pencil, Trash2, ExternalLink } from '@lucide/vue'
import { computed } from 'vue'
import { ensureAbsoluteUrl } from '@/utils/url'

const props = defineProps<{ cert: Certification }>()
const emit = defineEmits<{
  (e: 'edit'): void
  (e: 'delete'): void
}>()

const c = computed(() => props.cert)

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
          {{ c.certification_name }}
        </h3>
        <p class="text-sm text-gray-500 font-medium">
          {{ c.issuing_organization }}
        </p>
      </div>

      <div class="flex gap-2">
        <button
          class="border border-gray-200 text-gray-500 p-1 cursor-pointer grid place-items-center transition-all hover:text-accent hover:border-accent"
          @click="emit('edit')"
          aria-label="Edit certification"
        >
          <Pencil :size="14" />
        </button>
        <button
          class="border border-gray-200 text-gray-500 p-1 cursor-pointer grid place-items-center transition-all hover:text-red-600 hover:border-red-600"
          @click="emit('delete')"
          aria-label="Delete certification"
        >
          <Trash2 :size="14" />
        </button>
      </div>
    </div>

    <!-- Meta / Dates -->
    <div class="flex flex-wrap items-center gap-3 text-xs text-gray-400">
      <span v-if="c.is_in_progress" class="text-blue-600 font-semibold">In Progress</span>
      <span v-else>
        Issued {{ formatDate(c.issue_date) }}
        <span v-if="!c.does_not_expire && c.expiration_date">
          · Expires {{ formatDate(c.expiration_date) }}</span
        >
        <span v-else-if="c.does_not_expire"> · No Expiration Date</span>
      </span>

      <span v-if="c.credential_id" class="text-gray-300">•</span>
      <span v-if="c.credential_id">ID: {{ c.credential_id }}</span>

      <span v-if="c.credential_url" class="text-gray-300">•</span>
      <a
        v-if="c.credential_url"
        :href="ensureAbsoluteUrl(c.credential_url)"
        target="_blank"
        rel="noreferrer"
        class="inline-flex items-center gap-1 text-accent no-underline hover:underline font-medium"
      >
        <ExternalLink :size="12" /> View Certificate
      </a>
    </div>
  </div>
</template>
