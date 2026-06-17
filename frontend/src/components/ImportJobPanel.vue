<script setup lang="ts">
import { reactive, ref, computed } from 'vue'
import type { Job } from '@/types/job'
import { X, Loader2 } from '@lucide/vue'
import { useToast } from '@/composables/useToast'

const toast = useToast()
const emit = defineEmits<{ close: []; created: [job: Job] }>()

type Step = 'url' | 'loading' | 'review'
const step = ref<Step>('url')

const url = ref('')
const error = ref('')
const isSaving = ref(false)
const errors = reactive<Record<string, string>>({})

interface JobDraft {
  title: string
  company_name: string
  description: string
  source_url: string
  job_location: string
  is_remote: boolean
  salary_min: number | null
  salary_max: number | null
  currency: string
  skills: string[]
  logo_url: string
}

function emptyDraft(): JobDraft {
  return {
    title: '',
    company_name: '',
    description: '',
    source_url: '',
    job_location: '',
    is_remote: false,
    salary_min: null,
    salary_max: null,
    currency: 'USD',
    skills: [],
    logo_url: '',
  }
}

const form = reactive<JobDraft>(emptyDraft())

function validate(): boolean {
  // clear the previous error
  for (const key in errors) {
    delete errors[key]
  }

  if (!form.title.trim()) errors.title = 'Title is required'
  if (!form.company_name.trim()) errors.company_name = 'Company is required'
  if (!form.description.trim()) errors.description = 'Job description is required'

  return Object.keys(errors).length === 0
}

// skills as a comma-separated string for the input, synced to the array
const skillsText = computed({
  get: () => form.skills.join(', '),
  set: (val: string) => {
    form.skills = val
      .split(',')
      .map((s) => s.trim())
      .filter(Boolean)
  },
})

async function handleExtract() {
  if (!validate()) return
  step.value = 'loading'
  error.value = ''
  try {
    const res = await fetch('/api/jobs/import', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify({ url: url.value.trim() }),
    })

    if (res.status === 422) {
      error.value =
        'Add a Gemini API key in Settings to auto-fill job details — or enter them manually below.'
      // still let them proceed manually
      Object.assign(form, emptyDraft(), { source_url: url.value.trim() })
      step.value = 'review'
      return
    }
    if (!res.ok) {
      const data = await res.json().catch(() => null)
      throw new Error(data?.error ?? 'Could not import that URL')
    }

    const draft = await res.json()
    // Merge whatever came back over an empty draft — partial or full,
    // both land here; missing fields just stay empty.
    Object.assign(form, emptyDraft(), draft, { source_url: trimmed })
    step.value = 'review'
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Import failed'
    step.value = 'url' // hard failure (bad URL, blocked host) → back to entry
  }
}

function enterManually() {
  Object.assign(form, emptyDraft(), { source_url: url.value.trim() })
  step.value = 'review'
}

async function handleConfirm() {
  if (!form.title.trim() || !form.company_name.trim() || !form.description.trim()) {
    error.value = 'Title company are required.'
    return
  }
  isSaving.value = true
  error.value = ''
  try {
    const res = await fetch('/api/jobs/import/confirm', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify(form),
    })
    if (!res.ok) {
      const data = await res.json().catch(() => null)
      throw new Error(data?.error ?? 'Failed to save job')
    }
    const job = await res.json()
    toast.success('Job saved')
    emit('created', job)
    emit('close')
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Save failed'
  } finally {
    isSaving.value = false
  }
}
</script>

<template>
  <aside
    class="bg-white flex flex-col h-full w-full max-w-xl"
    style="font-family: var(--font-base)"
  >
    <!-- Header -->
    <div class="p-4 border-b border-gray-200 flex items-center justify-between shrink-0">
      <h2 class="text-lg font-bold text-gray-900">Import a job</h2>
      <button class="text-gray-400 hover:text-gray-900" @click="emit('close')">
        <X :size="18" />
      </button>
    </div>

    <div class="flex-1 overflow-y-auto p-4 flex flex-col gap-4">
      <!-- STEP 1: URL entry -->
      <template v-if="step === 'url'">
        <label for="url" class="text-sm font-semibold text-gray-900">Job posting URL</label>
        <input
          v-model="url"
          name="url"
          type="url"
          placeholder="https://..."
          class="px-4 py-2 border border-gray-200 text-sm bg-white text-gray-900 w-full focus:outline-none focus:border-accent"
          @keyup.enter="handleExtract"
        />
        <p v-if="error" class="text-xs text-red-600">{{ error }}</p>
        <div class="flex gap-2">
          <button class="button button-primary" :disabled="!url.trim()" @click="handleExtract">
            Import
          </button>
          <button class="button button-secondary" @click="enterManually">Enter manually</button>
        </div>
      </template>

      <!-- STEP 2: loading -->
      <template v-else-if="step === 'loading'">
        <div class="flex flex-col items-center justify-center py-12 gap-3 text-gray-400">
          <Loader2 :size="24" class="animate-spin" />
          <span class="text-sm">Reading the posting…</span>
        </div>
      </template>

      <!-- STEP 3: review form -->
      <template v-else>
        <div class="flex flex-col gap-1">
          <label for="job-title" class="text-sm font-semibold text-gray-900">Title</label>
          <input
            id="job-title"
            name="title"
            type="text"
            v-model="form.title"
            class="px-4 py-2 border border-gray-200 text-sm w-full focus:outline-none focus:border-accent"
          />
          <p v-if="errors.title" class="text-xs text-red-600">{{ errors.title }}</p>
        </div>

        <div class="flex flex-col gap-1">
          <label for="company-name" class="text-sm font-semibold text-gray-900">Company</label>
          <input
            id="company-name"
            name="company"
            v-model="form.company_name"
            type="text"
            class="px-4 py-2 border border-gray-200 text-sm w-full focus:outline-none focus:border-accent"
          />
          <p v-if="errors.company" class="text-xs text-red-600">{{ errors.company }}</p>
        </div>

        <div class="flex flex-col gap-1">
          <label for="job-location" class="text-sm font-semibold text-gray-900">Location</label>
          <input
            id="job-location"
            name="location"
            v-model="form.job_location"
            type="text"
            class="px-4 py-2 border border-gray-200 text-sm w-full focus:outline-none focus:border-accent"
          />
        </div>

        <label for="is-remote" class="flex items-center gap-2 text-sm text-gray-900">
          <input name="remote" v-model="form.is_remote" type="checkbox" />
          Remote
        </label>

        <div class="grid grid-cols-2 gap-2">
          <div class="flex flex-col gap-1">
            <label for="job-salary-min" class="text-sm font-semibold text-gray-900"
              >Min salary</label
            >
            <input
              id="job-salary-min"
              name="salary-min"
              v-model.number="form.salary_min"
              type="number"
              class="px-4 py-2 border border-gray-200 text-sm w-full focus:outline-none focus:border-accent"
            />
          </div>
          <div class="flex flex-col gap-1">
            <label for="job-salary-max" class="text-sm font-semibold text-gray-900"
              >Max salary</label
            >
            <input
              id="job-salary-max"
              name="salary-max"
              v-model.number="form.salary_max"
              type="number"
              class="px-4 py-2 border border-gray-200 text-sm w-full focus:outline-none focus:border-accent"
            />
          </div>
        </div>

        <div class="flex flex-col gap-1">
          <label class="text-sm font-semibold text-gray-900">Skills</label>
          <input
            v-model="skillsText"
            type="text"
            placeholder="Go, PostgreSQL, Vue"
            class="px-4 py-2 border border-gray-200 text-sm w-full focus:outline-none focus:border-accent"
          />
          <span class="text-xs text-gray-400">Comma-separated</span>
        </div>

        <div class="flex flex-col gap-1">
          <label id="job-description" class="text-sm font-semibold text-gray-900"
            >Description</label
          >
          <textarea
            id="job-description"
            name="description"
            v-model="form.description"
            rows="6"
            class="px-4 py-2 border border-gray-200 text-sm w-full focus:outline-none focus:border-accent resize-y"
          />
        </div>
        <p v-if="errors.title" class="text-xs text-red-600">{{ errors.description }}</p>
      </template>
    </div>

    <!-- Footer (review only) -->
    <div
      v-if="step === 'review'"
      class="p-4 border-t border-gray-200 flex justify-end gap-2 shrink-0"
    >
      <button class="button button-secondary" @click="emit('close')">Cancel</button>
      <button class="button button-primary" :disabled="isSaving" @click="handleConfirm">
        {{ isSaving ? 'Saving…' : 'Save job' }}
      </button>
    </div>
  </aside>
</template>
