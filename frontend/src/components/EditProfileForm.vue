<script setup lang="ts">
import { ref, watch } from 'vue'
import { X } from '@lucide/vue'
import type { User } from '@/types/user'

const props = defineProps<{ user: User; open: boolean; isSaving: boolean }>()
const emit = defineEmits<{
  (e: 'close'): void
  (e: 'save', payload: Partial<User>): void
}>()

const availabilityOptions = ['immediate', 'two_weeks', 'one_month', 'three_months', 'not_looking']

const availabilityLabels: Record<string, string> = {
  immediate: 'Immediately available',
  two_weeks: 'Available in 2 weeks',
  one_month: 'Available in 1 month',
  three_months: 'Available in 3 months',
  not_looking: 'Not actively looking',
}

function blankForm() {
  const u = props.user
  return {
    username: u?.username ?? '',
    full_name: u?.full_name ?? '',
    phone: u?.phone ?? '',
    headline: u?.headline ?? '',
    user_summary: u?.user_summary ?? '',
    user_location: u?.user_location ?? '',
    website_url: u?.website_url ?? '',
    linkedin_url: u?.linkedin_url ?? '',
    github_url: u?.github_url ?? '',
    availability: u?.availability ?? 'immediate',
    min_salary: u?.min_salary ?? null,
    max_salary: u?.max_salary ?? null,
    salary_currency: u?.salary_currency ?? 'USD',
    years_of_experience: u?.years_of_experience ?? null,
  }
}

const form = ref(blankForm())
const errors = ref<Record<string, string>>({})
const activeSection = ref<'identity' | 'career'>('identity')

watch(
  () => props.open,
  (isOpen) => {
    if (isOpen) {
      form.value = blankForm()
      errors.value = {}
      activeSection.value = 'identity'
    }
  },
)

function validate(): boolean {
  errors.value = {}
  if (!form.value.username.trim()) {
    errors.value.username = 'Username is required.'
  }
  if (
    form.value.min_salary &&
    form.value.max_salary &&
    form.value.min_salary > form.value.max_salary
  ) {
    errors.value.max_salary = 'Max salary must be greater than min salary.'
  }
  return Object.keys(errors.value).length === 0
}

async function handleSubmit() {
  if (!validate()) return
  const body = {
    ...form.value,
    phone: form.value.phone || null,
    website_url: form.value.website_url || null,
    linkedin_url: form.value.linkedin_url || null,
    github_url: form.value.github_url || null,
    user_location: form.value.user_location || null,
    headline: form.value.headline || null,
    user_summary: form.value.user_summary || null,
    full_name: form.value.full_name || null,
    years_of_experience: form.value.years_of_experience
      ? Number(form.value.years_of_experience)
      : null,
    min_salary: form.value.min_salary ? Number(form.value.min_salary) : null,
    max_salary: form.value.max_salary ? Number(form.value.max_salary) : null,
  }

  emit('save', body)
}
</script>

<template>
  <div
    v-if="open"
    class="fixed inset-0 bg-black/45 flex items-start justify-end z-100"
    role="dialog"
    aria-modal="true"
    aria-label="Edit profile"
    @click.self="emit('close')"
  >
    <!-- Panel -->
    <div
      class="w-full max-w-140 h-dvh bg-white border-l border-gray-200 flex flex-col overflow-hidden"
    >
      <!-- Header -->
      <div class="flex items-center justify-between px-6 py-5 border-b border-gray-200 shrink-0">
        <h2 class="text-lg font-semibold text-gray-900">Edit Profile</h2>
        <button
          type="button"
          class="grid place-items-center w-8 h-8 border border-gray-200 text-gray-500 cursor-pointer transition-all hover:text-gray-900 hover:border-gray-400"
          @click="emit('close')"
          aria-label="Close"
        >
          <X :size="18" />
        </button>
      </div>

      <!-- Section tabs -->
      <div class="flex border-b border-gray-200 shrink-0">
        <button
          v-for="tab in [
            { key: 'identity', label: 'Identity' },
            { key: 'career', label: 'Career' },
          ]"
          :key="tab.key"
          type="button"
          class="px-6 py-3 text-sm font-medium border-b-2 -mb-px transition-all cursor-pointer"
          :class="
            activeSection === tab.key
              ? 'text-accent'
              : 'border-transparent text-gray-500 hover:text-gray-900'
          "
          :style="activeSection === tab.key ? { borderBottomColor: 'var(--color-accent)' } : {}"
          @click="activeSection = tab.key as 'identity' | 'career'"
        >
          {{ tab.label }}
        </button>
      </div>

      <!-- Scrollable form -->
      <form
        class="flex-1 overflow-y-auto px-6 py-6 flex flex-col gap-5"
        @submit.prevent="handleSubmit"
        novalidate
      >
        <!-- ── IDENTITY ── -->
        <template v-if="activeSection === 'identity'">
          <!-- Username + Full name -->
          <div class="grid grid-cols-2 gap-4">
            <div class="flex flex-col gap-2">
              <label class="text-sm font-semibold text-gray-900" for="ep_username">
                Username <span class="text-red-600 ml-0.5">*</span>
              </label>
              <input
                id="ep_username"
                v-model="form.username"
                type="text"
                placeholder="e.g. samnodier"
                class="px-3 py-2 border text-sm bg-white text-gray-900 w-full transition-all focus:outline-none focus:border-accent focus:shadow-[0_0_0_1px_var(--color-accent)]"
                :class="errors.username ? 'border-red-600' : 'border-gray-200'"
              />
              <p v-if="errors.username" class="text-xs text-red-600">{{ errors.username }}</p>
            </div>

            <div class="flex flex-col gap-2">
              <label class="text-sm font-semibold text-gray-900" for="ep_full_name"
                >Full name</label
              >
              <input
                id="ep_full_name"
                v-model="form.full_name"
                type="text"
                placeholder="e.g. Sam Nodier"
                class="px-3 py-2 border border-gray-200 text-sm bg-white text-gray-900 w-full transition-all focus:outline-none focus:border-accent focus:shadow-[0_0_0_1px_var(--color-accent)]"
              />
            </div>
          </div>

          <!-- Headline -->
          <div class="flex flex-col gap-2">
            <label class="text-sm font-semibold text-gray-900" for="ep_headline">Headline</label>
            <input
              id="ep_headline"
              v-model="form.headline"
              type="text"
              placeholder="e.g. Full-stack Engineer · Go · Vue.js"
              class="px-3 py-2 border border-gray-200 text-sm bg-white text-gray-900 w-full transition-all focus:outline-none focus:border-accent focus:shadow-[0_0_0_1px_var(--color-accent)]"
            />
          </div>

          <!-- Summary -->
          <div class="flex flex-col gap-2">
            <label class="text-sm font-semibold text-gray-900" for="ep_summary">Summary</label>
            <textarea
              id="ep_summary"
              v-model="form.user_summary"
              placeholder="Brief professional summary shown on your profile..."
              rows="4"
              class="px-3 py-2 border border-gray-200 text-sm bg-white text-gray-900 w-full transition-all focus:outline-none focus:border-accent focus:shadow-[0_0_0_1px_var(--color-accent)] resize-y min-h-24 leading-relaxed font-[inherit]"
            />
          </div>

          <!-- Location + Phone -->
          <div class="grid grid-cols-2 gap-4">
            <div class="flex flex-col gap-2">
              <label class="text-sm font-semibold text-gray-900" for="ep_location">Location</label>
              <input
                id="ep_location"
                v-model="form.user_location"
                type="text"
                placeholder="e.g. Kigali, Rwanda"
                class="px-3 py-2 border border-gray-200 text-sm bg-white text-gray-900 w-full transition-all focus:outline-none focus:border-accent focus:shadow-[0_0_0_1px_var(--color-accent)]"
              />
            </div>

            <div class="flex flex-col gap-2">
              <label class="text-sm font-semibold text-gray-900" for="ep_phone">Phone</label>
              <input
                id="ep_phone"
                v-model="form.phone"
                type="tel"
                placeholder="+250 ..."
                class="px-3 py-2 border border-gray-200 text-sm bg-white text-gray-900 w-full transition-all focus:outline-none focus:border-accent focus:shadow-[0_0_0_1px_var(--color-accent)]"
              />
            </div>
          </div>

          <!-- Website + LinkedIn -->
          <div class="grid grid-cols-2 gap-4">
            <div class="flex flex-col gap-2">
              <label class="text-sm font-semibold text-gray-900" for="ep_website">Website</label>
              <input
                id="ep_website"
                v-model="form.website_url"
                type="url"
                placeholder="https://yoursite.com"
                class="px-3 py-2 border border-gray-200 text-sm bg-white text-gray-900 w-full transition-all focus:outline-none focus:border-accent focus:shadow-[0_0_0_1px_var(--color-accent)]"
              />
            </div>

            <div class="flex flex-col gap-2">
              <label class="text-sm font-semibold text-gray-900" for="ep_linkedin">LinkedIn</label>
              <input
                id="ep_linkedin"
                v-model="form.linkedin_url"
                type="url"
                placeholder="https://linkedin.com/in/..."
                class="px-3 py-2 border border-gray-200 text-sm bg-white text-gray-900 w-full transition-all focus:outline-none focus:border-accent focus:shadow-[0_0_0_1px_var(--color-accent)]"
              />
            </div>
          </div>

          <!-- GitHub -->
          <div class="flex flex-col gap-2">
            <label class="text-sm font-semibold text-gray-900" for="ep_github">GitHub</label>
            <input
              id="ep_github"
              v-model="form.github_url"
              type="url"
              placeholder="https://github.com/..."
              class="px-3 py-2 border border-gray-200 text-sm bg-white text-gray-900 w-full transition-all focus:outline-none focus:border-accent focus:shadow-[0_0_0_1px_var(--color-accent)]"
            />
          </div>
        </template>

        <!-- ── CAREER ── -->
        <template v-else>
          <!-- Years of experience + Availability -->
          <div class="grid grid-cols-2 gap-4">
            <div class="flex flex-col gap-2">
              <label class="text-sm font-semibold text-gray-900" for="ep_years"
                >Years of experience</label
              >
              <input
                id="ep_years"
                v-model="form.years_of_experience"
                type="number"
                min="0"
                max="50"
                placeholder="e.g. 5"
                class="px-3 py-2 border border-gray-200 text-sm bg-white text-gray-900 w-full transition-all focus:outline-none focus:border-accent focus:shadow-[0_0_0_1px_var(--color-accent)]"
              />
            </div>

            <div class="flex flex-col gap-2">
              <label class="text-sm font-semibold text-gray-900" for="ep_availability"
                >Availability</label
              >
              <select
                id="ep_availability"
                v-model="form.availability"
                class="px-3 py-2 border border-gray-200 text-sm bg-white text-gray-900 w-full transition-all focus:outline-none focus:border-accent focus:shadow-[0_0_0_1px_var(--color-accent)]"
              >
                <option v-for="opt in availabilityOptions" :key="opt" :value="opt">
                  {{ availabilityLabels[opt] }}
                </option>
              </select>
            </div>
          </div>

          <!-- Min + Max salary -->
          <div class="grid grid-cols-2 gap-4">
            <div class="flex flex-col gap-2">
              <label class="text-sm font-semibold text-gray-900" for="ep_min_salary"
                >Min salary</label
              >
              <input
                id="ep_min_salary"
                v-model="form.min_salary"
                type="number"
                min="0"
                placeholder="e.g. 80000"
                class="px-3 py-2 border border-gray-200 text-sm bg-white text-gray-900 w-full transition-all focus:outline-none focus:border-accent focus:shadow-[0_0_0_1px_var(--color-accent)]"
              />
            </div>

            <div class="flex flex-col gap-2">
              <label class="text-sm font-semibold text-gray-900" for="ep_max_salary"
                >Max salary</label
              >
              <input
                id="ep_max_salary"
                v-model="form.max_salary"
                type="number"
                min="0"
                placeholder="e.g. 150000"
                class="px-3 py-2 border text-sm bg-white text-gray-900 w-full transition-all focus:outline-none focus:border-accent focus:shadow-[0_0_0_1px_var(--color-accent)]"
                :class="errors.max_salary ? 'border-red-600' : 'border-gray-200'"
              />
              <p v-if="errors.max_salary" class="text-xs text-red-600">{{ errors.max_salary }}</p>
            </div>
          </div>

          <!-- Salary currency -->
          <div class="flex flex-col gap-2">
            <label class="text-sm font-semibold text-gray-900" for="ep_currency"
              >Salary currency</label
            >
            <select
              id="ep_currency"
              v-model="form.salary_currency"
              class="px-3 py-2 border border-gray-200 text-sm bg-white text-gray-900 w-full transition-all focus:outline-none focus:border-accent focus:shadow-[0_0_0_1px_var(--color-accent)]"
            >
              <option value="USD">USD</option>
              <option value="EUR">EUR</option>
              <option value="GBP">GBP</option>
              <option value="RWF">RWF</option>
            </select>
          </div>
        </template>

        <!-- Footer -->
        <div class="flex justify-end gap-3 pt-4 border-t border-gray-200 mt-auto">
          <button
            type="button"
            class="bg-gray-200 border-gray-500 button button-secondary"
            @click="emit('close')"
          >
            Cancel
          </button>
          <button
            type="submit"
            class="bg-accent px-4 py-2 disabled:opacity-50"
            :disabled="isSaving"
          >
            {{ isSaving ? 'Saving…' : 'Save changes' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>
