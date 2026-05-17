<script setup lang="ts">
// NEW FILE: src/components/EditProfileForm.vue
import { ref, watch } from 'vue'
import { X } from '@lucide/vue'
import { useAuthStore } from '@/stores/auth'
import { useToast } from '@/composables/useToast'

const props = defineProps<{ open: boolean }>()
const emit = defineEmits<{
  (e: 'close'): void
}>()

const authStore = useAuthStore()
const toast = useToast()

const availabilityOptions = ['immediate', 'two_weeks', 'one_month', 'three_months', 'not_looking']

const availabilityLabels: Record<string, string> = {
  immediate: 'Immediately available',
  two_weeks: 'Available in 2 weeks',
  one_month: 'Available in 1 month',
  three_months: 'Available in 3 months',
  not_looking: 'Not actively looking',
}

function blankForm() {
  const u = authStore.user
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
const saving = ref(false)
// ADDED: internal tab state for the two sections
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
  saving.value = true
  try {
    const res = await fetch('/api/users/me', {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify({
        ...form.value,
        // Send null for empty strings on optional fields
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
      }),
    })

    if (!res.ok) {
      const data = await res.json().catch(() => null)
      if (res.status === 409) {
        errors.value.username = 'Username already taken.'
        activeSection.value = 'identity' // Switch back to show the error
      } else {
        toast.error(data?.error ?? 'Failed to save profile.')
      }
      return
    }

    const updated = await res.json()
    // ADDED: update the auth store so ProfileView re-renders immediately
    authStore.user = updated
    toast.success('Profile updated.')
    emit('close')
  } catch {
    toast.error('Something went wrong.')
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div
    v-if="open"
    class="overlay"
    @click.self="emit('close')"
    role="dialog"
    aria-modal="true"
    aria-label="Edit profile"
  >
    <div class="panel">
      <!-- Header -->
      <div class="panel-header">
        <h2 class="panel-title">Edit Profile</h2>
        <button class="close-btn" type="button" @click="emit('close')" aria-label="Close">
          <X :size="18" />
        </button>
      </div>

      <!-- ADDED: Internal section tabs -->
      <div class="panel-tabs">
        <button
          type="button"
          class="panel-tab"
          :class="{ 'panel-tab--active': activeSection === 'identity' }"
          @click="activeSection = 'identity'"
        >
          Identity
        </button>
        <button
          type="button"
          class="panel-tab"
          :class="{ 'panel-tab--active': activeSection === 'career' }"
          @click="activeSection = 'career'"
        >
          Career
        </button>
      </div>

      <form class="panel-body" @submit.prevent="handleSubmit" novalidate>
        <!-- ── IDENTITY SECTION ── -->
        <template v-if="activeSection === 'identity'">
          <div class="form-row">
            <div class="field" :class="{ 'field--error': errors.username }">
              <label class="field-label" for="ep_username">
                Username <span class="required">*</span>
              </label>
              <input
                id="ep_username"
                v-model="form.username"
                type="text"
                class="form-input"
                placeholder="e.g. samnodier"
              />
              <p v-if="errors.username" class="field-error">{{ errors.username }}</p>
            </div>

            <div class="field">
              <label class="field-label" for="ep_full_name">Full name</label>
              <input
                id="ep_full_name"
                v-model="form.full_name"
                type="text"
                class="form-input"
                placeholder="e.g. Sam Nodier"
              />
            </div>
          </div>

          <div class="field">
            <label class="field-label" for="ep_headline">Headline</label>
            <input
              id="ep_headline"
              v-model="form.headline"
              type="text"
              class="form-input"
              placeholder="e.g. Full-stack Engineer · Go · Vue.js"
            />
          </div>

          <div class="field">
            <label class="field-label" for="ep_summary">Summary</label>
            <textarea
              id="ep_summary"
              v-model="form.user_summary"
              class="form-input form-textarea"
              placeholder="Brief professional summary shown on your profile..."
              rows="4"
            />
          </div>

          <div class="form-row">
            <div class="field">
              <label class="field-label" for="ep_location">Location</label>
              <input
                id="ep_location"
                v-model="form.user_location"
                type="text"
                class="form-input"
                placeholder="e.g. Kigali, Rwanda"
              />
            </div>

            <div class="field">
              <label class="field-label" for="ep_phone">Phone</label>
              <input
                id="ep_phone"
                v-model="form.phone"
                type="tel"
                class="form-input"
                placeholder="+250 ..."
              />
            </div>
          </div>

          <div class="form-row">
            <div class="field">
              <label class="field-label" for="ep_website">Website</label>
              <input
                id="ep_website"
                v-model="form.website_url"
                type="url"
                class="form-input"
                placeholder="https://yoursite.com"
              />
            </div>

            <div class="field">
              <label class="field-label" for="ep_linkedin">LinkedIn</label>
              <input
                id="ep_linkedin"
                v-model="form.linkedin_url"
                type="url"
                class="form-input"
                placeholder="https://linkedin.com/in/..."
              />
            </div>
          </div>

          <div class="field">
            <label class="field-label" for="ep_github">GitHub</label>
            <input
              id="ep_github"
              v-model="form.github_url"
              type="url"
              class="form-input"
              placeholder="https://github.com/..."
            />
          </div>
        </template>

        <!-- ── CAREER SECTION ── -->
        <template v-else>
          <div class="form-row">
            <div class="field">
              <label class="field-label" for="ep_years">Years of experience</label>
              <input
                id="ep_years"
                v-model="form.years_of_experience"
                type="number"
                min="0"
                max="50"
                class="form-input"
                placeholder="e.g. 5"
              />
            </div>

            <div class="field">
              <label class="field-label" for="ep_availability">Availability</label>
              <select id="ep_availability" v-model="form.availability" class="form-input">
                <option v-for="opt in availabilityOptions" :key="opt" :value="opt">
                  {{ availabilityLabels[opt] }}
                </option>
              </select>
            </div>
          </div>

          <div class="form-row">
            <div class="field">
              <label class="field-label" for="ep_min_salary">Min salary</label>
              <input
                id="ep_min_salary"
                v-model="form.min_salary"
                type="number"
                min="0"
                class="form-input"
                placeholder="e.g. 80000"
              />
            </div>

            <div class="field" :class="{ 'field--error': errors.max_salary }">
              <label class="field-label" for="ep_max_salary">Max salary</label>
              <input
                id="ep_max_salary"
                v-model="form.max_salary"
                type="number"
                min="0"
                class="form-input"
                placeholder="e.g. 150000"
              />
              <p v-if="errors.max_salary" class="field-error">{{ errors.max_salary }}</p>
            </div>
          </div>

          <div class="field">
            <label class="field-label" for="ep_currency">Salary currency</label>
            <select id="ep_currency" v-model="form.salary_currency" class="form-input">
              <option value="USD">USD</option>
              <option value="EUR">EUR</option>
              <option value="GBP">GBP</option>
              <option value="RWF">RWF</option>
            </select>
          </div>
        </template>

        <!-- Footer -->
        <div class="panel-footer">
          <button type="button" class="button button-secondary" @click="emit('close')">
            Cancel
          </button>
          <button type="submit" class="button button-primary" :disabled="saving">
            {{ saving ? 'Saving…' : 'Save changes' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>
