<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { X } from '@lucide/vue'
import type { Education } from '@/types/education'

const props = defineProps<{
  education?: Education
  open: boolean
  isSaving: boolean
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'save', payload: Omit<Education, 'id' | 'user_id' | 'created_at' | 'updated_at'>): void
}>()

const degreeTypes = [
  { value: 'bachelor', label: "Bachelor's Degree" },
  { value: 'master', label: "Master's Degree" },
  { value: 'doctorate', label: 'Doctoral Degree' },
  { value: 'associate', label: "Associate's Degree" },
  { value: 'bootcamp', label: 'Bootcamp Certificate' },
  { value: 'diploma', label: 'Diploma' },
  { value: 'high_school', label: 'High School Diploma' },
  { value: 'other', label: 'Other' },
]

function blankForm() {
  return {
    institution_name: '',
    degree_type: 'bachelor',
    degree_name: '',
    field_of_study: '',
    start_date: '',
    end_date: '',
    is_current: false,
    description: '',
    is_highlighted: true,
  }
}

const form = ref(blankForm())
const errors = ref<Record<string, string>>({})

watch(
  () => props.open,
  (isOpen) => {
    if (isOpen) {
      if (props.education) {
        form.value = {
          institution_name: props.education.institution_name,
          degree_type: props.education.degree_type || 'bachelor',
          degree_name: props.education.degree_name ?? '',
          field_of_study: props.education.field_of_study ?? '',
          start_date: props.education.start_date ? props.education.start_date.substring(0, 7) : '',
          end_date: props.education.end_date ? props.education.end_date.substring(0, 7) : '',
          is_current: props.education.is_current,
          description: props.education.description ?? '',
          is_highlighted: props.education.is_highlighted ?? true,
        }
      } else {
        form.value = blankForm()
      }
      errors.value = {}
    }
  },
  { immediate: true },
)

const isEditing = computed(() => !!props.education)
const endDateDisabled = computed(() => form.value.is_current)

function validate(): boolean {
  errors.value = {}
  if (!form.value.institution_name.trim()) {
    errors.value.institution_name = 'Institution name is required.'
  }
  if (!form.value.start_date) {
    errors.value.start_date = 'Start date is required.'
  }
  if (!form.value.is_current && !form.value.end_date) {
    errors.value.end_date = 'End date is required unless still enrolled.'
  }
  if (
    form.value.start_date &&
    form.value.end_date &&
    !form.value.is_current &&
    form.value.end_date < form.value.start_date
  ) {
    errors.value.end_date = 'End date must be after start date.'
  }
  return Object.keys(errors.value).length === 0
}

function handleSubmit() {
  if (!validate()) return
  emit('save', {
    institution_name: form.value.institution_name.trim(),
    degree_type: form.value.degree_type,
    degree_name: form.value.degree_name.trim() || null,
    field_of_study: form.value.field_of_study.trim() || null,
    start_date: form.value.start_date || null,
    end_date: form.value.is_current ? null : form.value.end_date || null,
    is_current: form.value.is_current,
    description: form.value.description.trim() || null,
    is_highlighted: form.value.is_highlighted,
  })
}
</script>

<template>
  <div
    v-if="open"
    class="fixed inset-0 bg-black/45 flex items-start justify-end z-100"
    role="dialog"
    aria-modal="true"
    :aria-label="isEditing ? 'Edit education' : 'Add education'"
    @click.self="emit('close')"
  >
    <!-- Panel -->
    <div
      class="w-full max-w-140 h-dvh bg-white border-l border-gray-200 flex flex-col overflow-hidden"
    >
      <!-- Header -->
      <div class="flex items-center justify-between px-6 py-5 border-b border-gray-200 shrink-0">
        <h2 class="text-lg font-semibold text-gray-900">
          {{ isEditing ? 'Edit education' : 'Add education' }}
        </h2>
        <button
          type="button"
          class="grid place-items-center w-8 h-8 border border-gray-200 text-gray-500 cursor-pointer transition-all hover:text-gray-900 hover:border-gray-400"
          @click="emit('close')"
          aria-label="Close"
        >
          <X :size="18" />
        </button>
      </div>

      <!-- Form body -->
      <form
        class="flex-1 overflow-y-auto px-6 py-6 flex flex-col gap-5"
        @submit.prevent="handleSubmit"
        novalidate
      >
        <!-- Institution -->
        <div class="flex flex-col gap-2">
          <label class="text-sm font-semibold text-gray-900" for="institution_name">
            School or Institution <span class="text-red-600 ml-0. 5">*</span>
          </label>
          <input
            id="institution_name"
            v-model="form.institution_name"
            type="text"
            placeholder="e.g. University of Waterloo"
            class="px-3 py-2 border text-sm bg-white text-gray-900 w-full transition-all focus:outline-none focus:border-accent focus:shadow-[0_0_0_1px_var(--color-accent)]"
            :class="errors.institution_name ? 'border-red-600' : 'border-gray-200'"
          />
          <p v-if="errors.institution_name" class="text-xs text-red-600">
            {{ errors.institution_name }}
          </p>
        </div>

        <!-- Degree Type + Degree Name -->
        <div class="grid grid-cols-2 gap-4">
          <div class="flex flex-col gap-2">
            <label class="text-sm font-semibold text-gray-900" for="degree_type">Degree type</label>
            <select
              id="degree_type"
              v-model="form.degree_type"
              class="px-3 py-2 border border-gray-200 text-sm bg-white text-gray-900 w-full transition-all focus:outline-none focus:border-accent focus:shadow-[0_0_0_1px_var(--color-accent)]"
            >
              <option v-for="type in degreeTypes" :key="type.value" :value="type.value">
                {{ type.label }}
              </option>
            </select>
          </div>

          <div class="flex flex-col gap-2">
            <label class="text-sm font-semibold text-gray-900" for="degree_name">Degree name</label>
            <input
              id="degree_name"
              v-model="form.degree_name"
              type="text"
              placeholder="e.g. Bachelor of Science"
              class="px-3 py-2 border border-gray-200 text-sm bg-white text-gray-900 w-full transition-all focus:outline-none focus:border-accent focus:shadow-[0_0_0_1px_var(--color-accent)]"
            />
          </div>
        </div>

        <!-- Field of study -->
        <div class="flex flex-col gap-2">
          <label class="text-sm font-semibold text-gray-900" for="field_of_study"
            >Field of study</label
          >
          <input
            id="field_of_study"
            v-model="form.field_of_study"
            type="text"
            placeholder="e.g. Computer Science"
            class="px-3 py-2 border border-gray-200 text-sm bg-white text-gray-900 w-full transition-all focus:outline-none focus:border-accent focus:shadow-[0_0_0_1px_var(--color-accent)]"
          />
        </div>

        <!-- Start + End Date -->
        <div class="grid grid-cols-2 gap-4">
          <div class="flex flex-col gap-2">
            <label class="text-sm font-semibold text-gray-900" for="start_date">
              Start date <span class="text-red-600 ml-0. 5">*</span>
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
              class="px-3 py-2 border text-sm bg-white text-gray-900 w-full transition-all focus:outline-none focus:border-accent focus:shadow-[0_0_0_1px_var(--color-accent)] disabled:opacity-45 disabled:cursor-not-allowed disabled:bg-black/4"
              :class="errors.end_date ? 'border-red-600' : 'border-gray-200'"
            />
            <p v-if="errors.end_date" class="text-xs text-red-600">{{ errors.end_date }}</p>
          </div>
        </div>

        <!-- Current enrolled checkbox -->
        <div class="flex flex-row items-center gap-2">
          <input
            id="is_current"
            v-model="form.is_current"
            type="checkbox"
            class="w-4 h-4 cursor-pointer shrink-0"
            :style="{ accentColor: 'var(--color-accent)' }"
          />
          <label for="is_current" class="text-sm text-gray-900 cursor-pointer">
            I am currently studying here
          </label>
        </div>

        <!-- Description -->
        <div class="flex flex-col gap-2">
          <label class="text-sm font-semibold text-gray-900" for="description">Description</label>
          <textarea
            id="description"
            v-model="form.description"
            placeholder="Focus coursework, honors, or leadership activities..."
            rows="4"
            class="px-3 py-2 border border-gray-200 text-sm bg-white text-gray-900 w-full transition-all focus:outline-none focus:border-accent focus:shadow-[0_0_0_1px_var(--color-accent)] resize-y min-h-24 leading-relaxed font-[inherit]"
          />
        </div>

        <!-- Footer -->
        <div class="flex justify-end gap-3 pt-4 border-t border-gray-200 mt-auto">
          <button
            type="button"
            class="px-4 py-2 bg-gray-100 border border-gray-300 cursor-pointer"
            @click="emit('close')"
          >
            Cancel
          </button>
          <button
            type="submit"
            class="bg-accent text-accent-foreground px-4 py-2 disabled:opacity-50 cursor-pointer"
            :disabled="isSaving"
          >
            {{ isSaving ? 'Saving…' : isEditing ? 'Save changes' : 'Add education' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>
