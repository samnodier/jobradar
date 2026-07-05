<script setup lang="ts">
import { reactive, ref } from 'vue'
import type { Job } from '@/types/job'
import { X, Loader2 } from '@lucide/vue'
import { useToast } from '@/composables/useToast'

const toast = useToast()
const emit = defineEmits<{ close: []; created: [job: Job] }>()

type Step = 'url' | 'loading' | 'review'
const step = ref<Step>('url')
const skillInput = ref('')

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

function addSkill() {
  const trimmed = skillInput.value.trim()
  if (!trimmed) return
  // avoid duplicates (case-insensitive)
  const exists = form.skills.some((s) => s.toLowerCase() === trimmed.toLowerCase())
  if (!exists) {
    form.skills.push(trimmed)
  }
  skillInput.value = ''
}

function removeSkill(index: number) {
  form.skills.splice(index, 1)
}

async function handleExtract() {
  const trimmedURL = url.value.trim()
  if (!trimmedURL) return
  step.value = 'loading'
  try {
    const res = await fetch('/api/jobs/import', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify({ url: trimmedURL }),
    })

    // Success - merge whatever came back (partial or full) and review
    if (res.ok) {
      const draft = await res.json()
      Object.assign(form, emptyDraft(), draft, { source_url: trimmedURL })
      step.value = 'review'
      return
    }

    const data = await res.json().catch(() => null)
    const message = data?.error ?? 'Could not import that URL'

    // 502: transiet LLM failure - Gemini congested/503
    // 422: couldn't not fetch that url (dead host, SSRF block, timeout)
    // 422: no gemini API key configured
    // 422: ErrPermanent - bad/revoked key, extraction wont ever proceed
    if (res.status === 502 || res.status === 422) {
      toast.error(message)
      enterManually()
      return
    }

    // 400: bad body/ unparseable URL / bad scheme / no host
    // 500: server's fault (parse/decrypt/extractor build)
    // 401, 413, 429 and other errors.
    // And any other errors as well
    toast.error(message)
    step.value = 'url'
  } catch (err) {
    // Network/parse failure - couldn't reach the API
    toast.error(err instanceof Error ? err.message : 'Import failed')
    step.value = 'url' // hard failure (bad URL, blocked host) → back to entry
  }
}

function enterManually() {
  Object.assign(form, emptyDraft(), { source_url: url.value.trim() })
  step.value = 'review'
}

async function handleConfirm() {
  if (!validate()) {
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
          id="url"
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
          <input id="is-remote" name="remote" v-model="form.is_remote" type="checkbox" />
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
          <label for="job-skills" class="text-sm font-semibold text-gray-900">Skills</label>
          <div class="flex flex-wrap gap-1.5 mb-1">
            <span
              v-for="(skill, index) in form.skills"
              :key="skill"
              class="inline-flex items-center gap-1 px-2 py-0.5 text-xs font-medium bg-black/4 border border-gray-200 text-gray-700"
            >
              {{ skill }}
              <button
                type="button"
                class="text-gray-400 hover:text-gray-900 cursor-pointer"
                @click="removeSkill(index)"
              >
                <X :size="12" />
              </button>
            </span>
          </div>
          <input
            id="job-skills"
            name="skills"
            v-model="skillInput"
            type="text"
            placeholder="Type a skill and press Enter"
            class="px-4 py-2 border border-gray-200 text-sm w-full focus:outline-none focus:border-accent"
            @keydown.enter.prevent="addSkill"
          />
        </div>

        <div class="flex flex-col gap-1">
          <label for="job-description" class="text-sm font-semibold text-gray-900"
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
