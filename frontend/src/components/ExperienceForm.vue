<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { X, Plus, Trash2 } from '@lucide/vue'
import type { Experience, Skill } from '@/types/experience'

const props = defineProps<{
  experience?: Experience
  open: boolean
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'save', payload: Omit<Experience, 'id' | 'user_id' | 'created_at' | 'updated_at'>): void
}>()

const employmentTypes = [
  'Full-time',
  'Part-time',
  'Contract',
  'Freelance',
  'Internship',
  'Apprenticeship',
  'Volunteer',
]

const industries = [
  'Software / Technology',
  'FinTech',
  'HealthTech',
  'EdTech',
  'SaaS',
  'AI / ML',
  'E-commerce',
  'Media & Entertainment',
  'Consulting',
  'Government',
  'Non-profit',
  'Other',
]

function blankForm() {
  return {
    company_name: '',
    company_url: '' as string | null,
    role_title: '',
    exp_location: '' as string | null,
    industry: '' as string | null,
    employment_type: '' as string | null,
    description: '' as string | null,
    achievements: [] as string[],
    start_date: '',
    end_date: '' as string | null,
    is_current: false,
    skills: [] as Skill[],
  }
}

const form = ref(blankForm())
const newAchievement = ref('')
const newSkillName = ref('')
const errors = ref<Record<string, string>>({})

watch(
  () => props.open,
  (isOpen) => {
    if (isOpen) {
      if (props.experience) {
        form.value = {
          company_name: props.experience.company_name,
          company_url: props.experience.company_url ?? '',
          role_title: props.experience.role_title,
          exp_location: props.experience.exp_location ?? '',
          industry: props.experience.industry ?? '',
          employment_type: props.experience.employment_type ?? '',
          description: props.experience.description ?? '',
          achievements: props.experience.achievements ? [...props.experience.achievements] : [],
          start_date: props.experience.start_date,
          end_date: props.experience.end_date ?? '',
          is_current: props.experience.is_current,
          skills: props.experience.skills ? [...props.experience.skills] : [],
        }
      } else {
        form.value = blankForm()
      }
      errors.value = {}
      newAchievement.value = ''
      newSkillName.value = ''
    }
  },
  { immediate: true },
)

const isEditing = computed(() => !!props.experience)
const endDateDisabled = computed(() => form.value.is_current)

function addAchievement() {
  const text = newAchievement.value.trim()
  if (!text) return
  form.value.achievements.push(text)
  newAchievement.value = ''
}

function removeAchievement(index: number) {
  form.value.achievements.splice(index, 1)
}

function addSkill() {
  const name = newSkillName.value.trim()
  if (!name) return
  const already = form.value.skills.some((s) => s.name.toLowerCase() === name.toLowerCase())
  if (already) return
  form.value.skills.push({ id: crypto.randomUUID(), name })
  newSkillName.value = ''
}

function removeSkill(id: string) {
  form.value.skills = form.value.skills.filter((s) => s.id !== id)
}

function validate(): boolean {
  errors.value = {}
  if (!form.value.company_name.trim()) errors.value.company_name = 'Company name is required.'
  if (!form.value.role_title.trim()) errors.value.role_title = 'Role title is required.'
  if (!form.value.start_date) errors.value.start_date = 'Start date is required.'
  if (!form.value.is_current && !form.value.end_date)
    errors.value.end_date = 'End date is required unless this is your current role.'
  if (
    form.value.start_date &&
    form.value.end_date &&
    !form.value.is_current &&
    form.value.end_date < form.value.start_date
  )
    errors.value.end_date = 'End date must be after start date.'
  return Object.keys(errors.value).length === 0
}

function handleSubmit() {
  if (!validate()) return
  emit('save', {
    company_name: form.value.company_name.trim(),
    company_url: form.value.company_url?.trim() || null,
    role_title: form.value.role_title.trim(),
    exp_location: form.value.exp_location?.trim() || null,
    industry: form.value.industry || null,
    employment_type: form.value.employment_type || null,
    description: form.value.description?.trim() || null,
    achievements: form.value.achievements.length ? form.value.achievements : null,
    start_date: form.value.start_date,
    end_date: form.value.is_current ? null : form.value.end_date || null,
    is_current: form.value.is_current,
    skills: form.value.skills,
  })
}

function handleClose() {
  emit('close')
}
</script>

<template>
  <div
    v-if="open"
    class="fixed inset-0 bg-black/45 flex items-start justify-end z-100"
    role="dialog"
    aria-modal="true"
    :aria-label="isEditing ? 'Edit experience' : 'Add experience'"
    @click.self="handleClose"
  >
    <!-- Panel -->
    <div
      class="w-full max-w-140 h-dvh bg-white border-l border-gray-200 flex flex-col overflow-hidden"
    >
      <!-- Header -->
      <div class="flex items-center justify-between px-6 py-5 border-b border-gray-200 shrink-0">
        <h2 class="text-lg font-semibold text-gray-900">
          {{ isEditing ? 'Edit experience' : 'Add experience' }}
        </h2>
        <button
          type="button"
          class="grid place-items-center w-8 h-8 border border-gray-200 text-gray-500 cursor-pointer transition-all hover:text-gray-900 hover:border-gray-400"
          @click="handleClose"
          aria-label="Close"
        >
          <X :size="18" />
        </button>
      </div>

      <!-- Scrollable body -->
      <form
        class="flex-1 overflow-y-auto px-6 py-6 flex flex-col gap-5"
        @submit.prevent="handleSubmit"
        novalidate
      >
        <!-- Company + URL -->
        <div class="grid grid-cols-2 gap-4">
          <div class="flex flex-col gap-2">
            <label class="text-sm font-semibold text-gray-900" for="company_name">
              Company name <span class="text-red-600 ml-0.5">*</span>
            </label>
            <input
              id="company_name"
              v-model="form.company_name"
              type="text"
              placeholder="e.g. Andela"
              autocomplete="organization"
              class="px-3 py-2 border text-sm bg-white text-gray-900 w-full transition-all focus:outline-none focus:border-accent focus:shadow-[0_0_0_1px_var(--color-accent)]"
              :class="errors.company_name ? 'border-red-600' : 'border-gray-200'"
            />
            <p v-if="errors.company_name" class="text-xs text-red-600">{{ errors.company_name }}</p>
          </div>

          <div class="flex flex-col gap-2">
            <label class="text-sm font-semibold text-gray-900" for="company_url"
              >Company website</label
            >
            <input
              id="company_url"
              v-model="form.company_url"
              type="url"
              placeholder="https://company.com"
              class="px-3 py-2 border border-gray-200 text-sm bg-white text-gray-900 w-full transition-all focus:outline-none focus:border-accent focus:shadow-[0_0_0_1px_var(--color-accent)]"
            />
          </div>
        </div>

        <!-- Role + Employment type -->
        <div class="grid grid-cols-2 gap-4">
          <div class="flex flex-col gap-2">
            <label class="text-sm font-semibold text-gray-900" for="role_title">
              Role title <span class="text-red-600 ml-0.5">*</span>
            </label>
            <input
              id="role_title"
              v-model="form.role_title"
              type="text"
              placeholder="e.g. Senior Engineer"
              class="px-3 py-2 border text-sm bg-white text-gray-900 w-full transition-all focus:outline-none focus:border-accent focus:shadow-[0_0_0_1px_var(--color-accent)]"
              :class="errors.role_title ? 'border-red-600' : 'border-gray-200'"
            />
            <p v-if="errors.role_title" class="text-xs text-red-600">{{ errors.role_title }}</p>
          </div>

          <div class="flex flex-col gap-2">
            <label class="text-sm font-semibold text-gray-900" for="employment_type"
              >Employment type</label
            >
            <select
              id="employment_type"
              v-model="form.employment_type"
              class="px-3 py-2 border border-gray-200 text-sm bg-white text-gray-900 w-full transition-all focus:outline-none focus:border-accent focus:shadow-[0_0_0_1px_var(--color-accent)]"
            >
              <option value="">Select type</option>
              <option v-for="type in employmentTypes" :key="type" :value="type">{{ type }}</option>
            </select>
          </div>
        </div>

        <!-- Location + Industry -->
        <div class="grid grid-cols-2 gap-4">
          <div class="flex flex-col gap-2">
            <label class="text-sm font-semibold text-gray-900" for="exp_location">Location</label>
            <input
              id="exp_location"
              v-model="form.exp_location"
              type="text"
              placeholder="e.g. Kigali, Rwanda or Remote"
              class="px-3 py-2 border border-gray-200 text-sm bg-white text-gray-900 w-full transition-all focus:outline-none focus:border-accent focus:shadow-[0_0_0_1px_var(--color-accent)]"
            />
          </div>

          <div class="flex flex-col gap-2">
            <label class="text-sm font-semibold text-gray-900" for="industry">Industry</label>
            <select
              id="industry"
              v-model="form.industry"
              class="px-3 py-2 border border-gray-200 text-sm bg-white text-gray-900 w-full transition-all focus:outline-none focus:border-accent focus:shadow-[0_0_0_1px_var(--color-accent)]"
            >
              <option value="">Select industry</option>
              <option v-for="ind in industries" :key="ind" :value="ind">{{ ind }}</option>
            </select>
          </div>
        </div>

        <!-- Start + End dates -->
        <div class="grid grid-cols-2 gap-4">
          <div class="flex flex-col gap-2">
            <label class="text-sm font-semibold text-gray-900" for="start_date">
              Start date <span class="text-red-600 ml-0.5">*</span>
            </label>
            <input
              id="start_date"
              v-model="form.start_date"
              type="month"
              :max="new Date().toISOString().substring(0, 7)"
              class="px-3 py-2 border text-sm bg-white text-gray-900 w-full transition-all focus:outline-none focus:border-accent focus:shadow-[0_0_0_1px_var(--color-accent)]"
              :class="errors.start_date ? 'border-red-600' : 'border-gray-200'"
            />
            <p v-if="errors.start_date" class="text-xs text-red-600">{{ errors.start_date }}</p>
          </div>

          <div class="flex flex-col gap-2">
            <label class="text-sm font-semibold text-gray-900" for="end_date">End date</label>
            <input
              id="end_date"
              v-model="form.end_date"
              type="month"
              :disabled="endDateDisabled"
              :min="form.start_date || undefined"
              :max="form.is_current ? undefined : new Date().toISOString().substring(0, 7)"
              class="px-3 py-2 border text-sm bg-white text-gray-900 w-full transition-all focus:outline-none focus:border-accent focus:shadow-[0_0_0_1px_var(--color-accent)] disabled:opacity-45 disabled:cursor-not-allowed disabled:bg-black/[0.04]"
              :class="errors.end_date ? 'border-red-600' : 'border-gray-200'"
            />
            <p v-if="errors.end_date" class="text-xs text-red-600">{{ errors.end_date }}</p>
          </div>
        </div>

        <!-- Current role checkbox -->
        <div class="flex flex-row items-center gap-2">
          <input
            id="is_current"
            v-model="form.is_current"
            type="checkbox"
            class="w-4 h-4 cursor-pointer shrink-0"
            :style="{ accentColor: 'var(--color-accent)' }"
          />
          <label for="is_current" class="text-sm text-gray-900 cursor-pointer">
            I currently work here
          </label>
        </div>

        <!-- Description -->
        <div class="flex flex-col gap-2">
          <label class="text-sm font-semibold text-gray-900" for="description">Description</label>
          <textarea
            id="description"
            v-model="form.description"
            placeholder="Describe your role, responsibilities, and impact..."
            rows="4"
            class="px-3 py-2 border border-gray-200 text-sm bg-white text-gray-900 w-full transition-all focus:outline-none focus:border-accent focus:shadow-[0_0_0_1px_var(--color-accent)] resize-y min-h-24 leading-relaxed font-[inherit]"
          />
        </div>

        <!-- Achievements -->
        <div class="flex flex-col gap-2">
          <label class="text-sm font-semibold text-gray-900">Key achievements</label>
          <p class="text-xs text-gray-500 -mt-1">
            Add bullet points that highlight impact — think metrics, outcomes, and deliverables.
          </p>

          <ul
            v-if="form.achievements.length"
            class="flex flex-col gap-2 list-none p-0 mb-2"
            role="list"
          >
            <li
              v-for="(item, idx) in form.achievements"
              :key="idx"
              class="flex items-start gap-2 px-3 py-2 bg-black/4 border border-gray-200"
            >
              <span
                class="font-bold shrink-0 leading-snug"
                :style="{ color: 'var(--color-accent)' }"
                >•</span
              >
              <span class="flex-1 text-sm text-gray-900 leading-snug">{{ item }}</span>
              <button
                type="button"
                class="text-gray-500 cursor-pointer grid place-items-center p-1 shrink-0 transition-all hover:text-red-600"
                @click="removeAchievement(idx)"
                :aria-label="`Remove achievement: ${item}`"
              >
                <Trash2 :size="14" />
              </button>
            </li>
          </ul>

          <div class="flex gap-2">
            <input
              v-model="newAchievement"
              type="text"
              placeholder="e.g. Reduced API latency by 40% by optimising query plans"
              class="px-3 py-2 border border-gray-200 text-sm bg-white text-gray-900 w-full transition-all focus:outline-none focus:border-accent focus:shadow-[0_0_0_1px_var(--color-accent)]"
              @keydown.enter.prevent="addAchievement"
            />
            <button
              type="button"
              class="inline-flex items-center gap-1 px-3 py-2 bg-black/4 border border-gray-200 text-sm font-semibold text-gray-900 cursor-pointer whitespace-nowrap shrink-0 transition-all hover:border-accent hover:text-accent disabled:opacity-40 disabled:cursor-not-allowed"
              @click="addAchievement"
              :disabled="!newAchievement.trim()"
            >
              <Plus :size="16" />
              Add
            </button>
          </div>
        </div>

        <!-- Skills -->
        <div class="flex flex-col gap-2">
          <label class="text-sm font-semibold text-gray-900">Skills used</label>
          <p class="text-xs text-gray-500 -mt-1">
            Tag the technologies and skills relevant to this role.
          </p>

          <div v-if="form.skills.length" class="flex flex-wrap gap-2 mb-2">
            <span
              v-for="skill in form.skills"
              :key="skill.id"
              class="inline-flex items-center gap-1 px-2 py-1 bg-white border text-xs font-medium"
              :style="{ borderColor: 'var(--color-accent)', color: 'var(--color-accent)' }"
            >
              {{ skill.name }}
              <button
                type="button"
                class="border-none text-sm leading-none p-0 cursor-pointer"
                :style="{ color: 'var(--color-accent)' }"
                @click="removeSkill(skill.id)"
                :aria-label="`Remove skill ${skill.name}`"
              >
                ×
              </button>
            </span>
          </div>

          <div class="flex gap-2">
            <input
              v-model="newSkillName"
              type="text"
              placeholder="e.g. Go, PostgreSQL, Docker"
              class="px-3 py-2 border border-gray-200 text-sm bg-white text-gray-900 w-full transition-all focus:outline-none focus:border-accent focus:shadow-[0_0_0_1px_var(--color-accent)]"
              @keydown.enter.prevent="addSkill"
            />
            <button
              type="button"
              class="inline-flex items-center gap-1 px-4 py-2 bg-black/4 border border-gray-200 text-sm font-semibold text-gray-900 cursor-pointer whitespace-nowrap shrink-0 transition-all hover:border-accent hover:text-accent disabled:opacity-40 disabled:cursor-not-allowed"
              @click="addSkill"
              :disabled="!newSkillName.trim()"
            >
              <Plus :size="16" />
              Add
            </button>
          </div>
        </div>

        <!-- Footer -->
        <div class="flex justify-end gap-3 pt-4 border-t border-gray-200 mt-auto">
          <button
            type="button"
            class="px-4 py-2 bg-gray-100 border border-gray-300"
            @click="handleClose"
          >
            Cancel
          </button>
          <button type="submit" class="bg-accent text-accent-foreground px-4 py-2">
            {{ isEditing ? 'Save changes' : 'Add experience' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<style scoped>
/* Mobile: bottom sheet layout */
@media (max-width: 640px) {
  .fixed {
    align-items: flex-end !important;
    justify-content: stretch !important;
  }

  .fixed > div {
    max-width: 100% !important;
    height: 92dvh !important;
    border-left: none !important;
    border-top: 1px solid #e5e7eb;
  }

  .grid.grid-cols-2 {
    grid-template-columns: 1fr !important;
  }
}
</style>
