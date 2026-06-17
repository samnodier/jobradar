<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import type { Application } from '@/types/application'
import { VueDatePicker } from '@vuepic/vue-datepicker'
import '@vuepic/vue-datepicker/dist/main.css'
import { ArrowUpRight, Building2, Calendar, Clock, MapPin, RefreshCw, Wifi, X } from '@lucide/vue'
import { useToast } from '@/composables/useToast'
import { statusLabels, statusOrder } from '@/constants/applicationStatus'

const toast = useToast()

const formatDateTime = new Intl.DateTimeFormat('en-US', {
  dateStyle: 'medium',
  timeStyle: 'short',
})

const props = defineProps<{ app: Application }>()
const emit = defineEmits(['close', 'updated'])

const currentText = ref(props.app.notes || '')
const isSaving = ref(false)
const lastSaved = ref<string | null>(null)
let debounceTimer: ReturnType<typeof setTimeout> | null = null

const companyInitials = computed(() => {
  const words = props.app.company_name?.split(/\s+/) ?? []
  if (words.length === 0) return '?'
  if (words.length === 1) return words[0]!.slice(0, 2).toUpperCase()
  return (words[0]![0]! + words[1]![0]!).toUpperCase()
})

async function updateApplication(
  payload: Partial<Application> & {
    clear_applied_at?: boolean
    clear_follow_up_at?: boolean
  },
) {
  isSaving.value = true
  const body = {
    notes: payload.notes,
    application_status: payload.application_status,
    applied_at: payload.applied_at,
    clear_applied_at: payload.clear_applied_at ?? false,
    follow_up_at: payload.follow_up_at,
    clear_follow_up_at: payload.clear_follow_up_at ?? false,
  }
  try {
    const response = await fetch(`/api/applications/${props.app.id}`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify(body),
    })
    if (!response.ok) {
      const data = await response.json().catch(() => null)
      throw new Error(data?.error ?? 'Update failed')
    }
    const data = await response.json()
    emit('updated', data)
    lastSaved.value = formatDateTime.format(new Date(data.updated_at))
  } catch (err) {
    toast.error(err instanceof Error ? err.message : 'Failed to save changes')
  } finally {
    isSaving.value = false
  }
}

watch(
  () => props.app.id,
  () => {
    currentText.value = props.app.notes ?? ''
    if (props.app.updated_at) {
      lastSaved.value = formatDateTime.format(new Date(props.app.updated_at))
    } else {
      lastSaved.value = null
    }
  },
  { immediate: true },
)

function handleNotesInput() {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    updateApplication({ notes: currentText.value })
  }, 2000)
}

function handleStatusChange(event: Event) {
  const val = (event.target as HTMLSelectElement).value
  updateApplication({ application_status: val })
}

function handleDateChange(field: 'applied_at' | 'follow_up_at', date: Date | null) {
  if (date) {
    // The component emits a native Date object -> convert directly
    updateApplication({ [field]: date.toISOString() })
  } else {
    // Handles when the user clears the input
    updateApplication({ [field]: null, [`clear_${field}`]: true })
  }
}

function formatDate(date: string | null | undefined): string {
  if (!date) return '—'
  return new Date(date).toLocaleDateString(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  })
}

function formatDateShort(date: string | null | undefined): string {
  if (!date) return '—'
  return new Date(date).toLocaleDateString(undefined, {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  })
}
</script>

<template>
  <aside
    class="bg-white w-full max-w-xl min-w-xl h-full flex flex-col overflow-y-auto pointer-events-auto font-base"
  >
    <!-- Header -->
    <div class="p-4 border-b border-gray-200 shrink-0 flex flex-col gap-3">
      <!-- Top: logo + meta -->
      <div class="flex items-start gap-3">
        <!-- Company logo -->
        <div
          class="w-10 h-10 border border-gray-200 bg-black/4 shrink-0 flex items-center justify-center overflow-hidden"
        >
          <img
            v-if="app.job_logo_url"
            :src="app.job_logo_url"
            :alt="app.company_name"
            class="w-full h-full object-contain"
          />
          <span v-else class="text-base font-bold text-gray-400 tracking-wide">
            {{ companyInitials }}
          </span>
        </div>

        <!-- Title + company + meta -->
        <div class="flex-1 min-w-0 flex flex-col gap-1">
          <div class="flex items-start justify-between gap-2">
            <h2 class="m-0 text-xl font-bold text-gray-900 leading-tight">{{ app.job_title }}</h2>
            <button
              class="w-6.25 h-6.25 bg-transparent border-none text-gray-400 cursor-pointer shrink-0 flex items-center justify-center transition-colors hover:text-gray-900"
              @click="$emit('close')"
            >
              <X :size="16" />
            </button>
          </div>

          <div class="flex items-center gap-2">
            <Building2 :size="13" class="text-gray-400 shrink-0" />
            <span class="text-sm font-semibold text-gray-400">{{ app.company_name }}</span>
          </div>

          <div class="flex flex-wrap gap-2 mt-1">
            <div
              v-if="app.job_location || app.job_is_remote"
              class="inline-flex items-center gap-1 px-2 py-0.5 text-xs font-semibold bg-black/4 text-gray-500 border border-gray-200"
            >
              <MapPin :size="12" />
              <span>{{
                app.job_is_remote ? 'Remote' : (app.job_location ?? 'Location unknown')
              }}</span>
            </div>
            <div
              v-if="app.job_is_remote"
              class="inline-flex items-center gap-1 px-2 py-0.5 text-xs font-semibold bg-green-100 text-green-700 border border-green-200"
            >
              <Wifi :size="12" />
              <span>Remote</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Status selector -->
      <div class="flex items-center gap-3">
        <label
          class="text-xs font-semibold text-gray-400 uppercase tracking-wide whitespace-nowrap"
        >
          Status
        </label>
        <select
          class="flex-1 px-2 py-1 border text-sm font-semibold cursor-pointer transition-colors focus:outline-none focus:border-accent appearance-auto"
          :class="{
            'bg-blue-100 text-blue-700 border-blue-200': app.application_status === 'applied',
            'bg-yellow-50 text-yellow-800 border-yellow-200':
              app.application_status === 'interview',
            'bg-green-100 text-green-700 border-green-200': app.application_status === 'offer',
            'bg-red-100 text-red-700 border-red-200': app.application_status === 'rejected',
            'bg-black/4 text-gray-400 border-gray-200': ![
              'applied',
              'interview',
              'offer',
              'rejected',
            ].includes(app.application_status),
          }"
          :value="app.application_status"
          @change="handleStatusChange"
        >
          <option v-for="status in statusOrder" :key="status" :value="status">
            {{ statusLabels[status] }}
          </option>
        </select>
      </div>
    </div>

    <!-- Scrollable body -->
    <div class="flex-1 overflow-y-auto flex flex-col">
      <!-- Timeline -->
      <div class="p-4 border-b border-gray-200">
        <p class="text-xs font-semibold text-gray-400 uppercase tracking-wide mb-3">Timeline</p>
        <div class="flex flex-col gap-2">
          <div class="grid items-center gap-2" style="grid-template-columns: 20px 1fr auto">
            <div class="flex items-center justify-center text-gray-500">
              <Calendar :size="13" />
            </div>
            <span class="text-sm text-gray-400">Applied</span>
            <VueDatePicker
              :model-value="app.applied_at ? new Date(app.applied_at) : null"
              :time-config="{ enableTimePicker: false }"
              auto-apply
              :clearable="false"
              input-class-name="text-xs border border-gray-200 bg-black/4 text-gray-900 px-1 py-0.5 cursor-pointer max-w-35 min-w-0 focus:outline-none focus:border-accent"
              @update:model-value="(d) => handleDateChange('applied_at', d)"
            />
          </div>

          <div class="grid items-center gap-2" style="grid-template-columns: 20px 1fr auto">
            <div class="flex items-center justify-center text-gray-500"><Clock :size="13" /></div>
            <span class="text-sm text-gray-400">Follow up</span>
            <VueDatePicker
              :model-value="app.follow_up_at ? new Date(app.follow_up_at) : null"
              :time-config="{ enableTimePicker: false }"
              auto-apply
              :clearable="true"
              input-class-name="text-xs border border-gray-200 bg-black/4 text-gray-900 px-1 py-0.5 cursor-pointer max-w-35 min-w-0 focus:outline-none focus:border-accent"
              @update:model-value="(d) => handleDateChange('follow_up_at', d)"
            />
          </div>

          <div class="grid items-center gap-2" style="grid-template-columns: 20px 1fr auto">
            <div class="flex items-center justify-center text-gray-500">
              <RefreshCw :size="13" />
            </div>
            <span class="text-sm text-gray-400">Last Activity</span>
            <span class="text-sm text-gray-900 font-semibold text-right">{{
              formatDate(app.updated_at)
            }}</span>
          </div>

          <div class="grid items-center gap-2" style="grid-template-columns: 20px 1fr auto">
            <div class="flex items-center justify-center text-gray-500"><Clock :size="13" /></div>
            <span class="text-sm text-gray-400">Tracked on</span>
            <span class="text-sm text-gray-900 font-semibold text-right">{{
              formatDateShort(app.created_at)
            }}</span>
          </div>
        </div>
      </div>

      <!-- Notes -->
      <div class="p-4 flex-1 flex flex-col">
        <div class="flex items-center justify-between mb-3">
          <p class="text-xs font-semibold text-gray-400 uppercase tracking-wide m-0">Notes</p>
          <p v-if="isSaving" class="text-xs text-gray-400 italic">Saving...</p>
        </div>
        <textarea
          v-model="currentText"
          @input="handleNotesInput"
          placeholder="Add your research, interview notes, or follow-up plan here..."
          class="text-sm text-gray-900 leading-relaxed w-full font-mono min-h-35 flex-1 p-3 resize-y border border-gray-200 bg-black/4 transition-colors focus:outline-none focus:border-accent focus:bg-white"
        ></textarea>
        <p v-if="lastSaved && !isSaving" class="text-xs text-gray-400 mt-2">
          Last saved: {{ lastSaved }}
        </p>
      </div>
    </div>

    <!-- Footer -->
    <div v-if="app.source_url" class="p-4 border-t border-gray-200 shrink-0">
      <a
        :href="app.source_url"
        target="_blank"
        rel="noreferrer"
        class="primary button-primary w-full h-8 inline-flex items-center justify-center gap-1 text-white text-sm font-semibold no-underline cursor-pointer transition-opacity hover:opacity-[0.88]"
        :style="{ background: 'var(--color-accent)' }"
      >
        View job posting
        <ArrowUpRight :size="13" />
      </a>
    </div>
  </aside>
</template>
