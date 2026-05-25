<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { X } from '@lucide/vue'
import type { Project } from '@/types/project'

const props = defineProps<{
  project?: Project
  open: boolean
  isSaving: boolean
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'save', payload: Omit<Project, 'id' | 'user_id' | 'created_at' | 'updated_at'>): void
}>()

function blankForm() {
  return {
    title: '',
    role_title: '',
    description: '',
    impact: '',
    project_url: '',
    repository_url: '',
    start_date: '',
    end_date: '',
    is_current: false,
    is_featured: false,
  }
}

const form = ref(blankForm())
const errors = ref<Record<string, string>>({})

watch(
  () => props.open,
  (isOpen) => {
    if (isOpen) {
      if (props.project) {
        form.value = {
          title: props.project.title,
          role_title: props.project.role_title ?? '',
          description: props.project.description ?? '',
          impact: props.project.impact ?? '',
          project_url: props.project.project_url ?? '',
          repository_url: props.project.repository_url ?? '',
          start_date: props.project.start_date ? props.project.start_date.substring(0, 7) : '',
          end_date: props.project.end_date ? props.project.end_date.substring(0, 7) : '',
          is_current: props.project.is_current,
          is_featured: props.project.is_featured ?? false,
        }
      } else {
        form.value = blankForm()
      }
      errors.value = {}
    }
  },
  { immediate: true },
)

const isEditing = computed(() => !!props.project)
const endDateDisabled = computed(() => form.value.is_current)

function validate(): boolean {
  errors.value = {}
  if (!form.value.title.trim()) {
    errors.value.title = 'Project title is required.'
  }
  if (!form.value.start_date) {
    errors.value.start_date = 'Start date is required.'
  }
  if (!form.value.is_current && !form.value.end_date) {
    errors.value.end_date = 'End date is required unless still working on this project.'
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
    title: form.value.title.trim(),
    role_title: form.value.role_title.trim() || null,
    description: form.value.description.trim() || null,
    impact: form.value.impact.trim() || null,
    project_url: form.value.project_url.trim() || null,
    repository_url: form.value.repository_url.trim() || null,
    start_date: form.value.start_date,
    end_date: form.value.is_current ? null : form.value.end_date || null,
    is_current: form.value.is_current,
    is_featured: form.value.is_featured,
  })
}
</script>

<template>
  <div
    v-if="open"
    class="fixed inset-0 bg-black/45 flex items-start justify-end z-100"
    role="dialog"
    aria-modal="true"
    :aria-label="isEditing ? 'Edit project' : 'Add project'"
    @click.self="emit('close')"
  >
    <!-- Panel -->
    <div
      class="w-full max-w-140 h-dvh bg-white border-l border-gray-200 flex flex-col overflow-hidden"
    >
      <!-- Header -->
      <div class="flex items-center justify-between px-6 py-5 border-b border-gray-200 shrink-0">
        <h2 class="text-lg font-semibold text-gray-900">
          {{ isEditing ? 'Edit project' : 'Add project' }}
        </h2>
        <button
          type="button"
          class="grid place-items-center w-8 h-8 border border-gray-200 text-gray-500 cursor-pointer transition-all hover:text- gray-900 hover:border-gray-400"
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
        <!-- Project Title -->
        <div class="flex flex-col gap-2">
          <label class="text-sm font-semibold text-gray-900" for="title">
            Project title <span class="text-red-600 ml-0. 5">*</span>
          </label>
          <input
            id="title"
            v-model="form.title"
            type="text"
            placeholder="e.g. Jobradar"
            class="px-3 py-2 border text-sm bg-white text-gray-900 w-full transition-all focus:outline-none focus:border-accent focus:shadow-[0_0_0_1px_var(--color-accent)]"
            :class="errors.title ? 'border-red-600' : 'border-gray-200'"
          />
          <p v-if="errors.title" class="text-xs text-red-600">{{ errors.title }}</p>
        </div>

        <!-- Role title -->
        <div class="flex flex-col gap-2">
          <label class="text-sm font-semibold text-gray-900" for="role_title"
            >Your role / contribution</label
          >
          <input
            id="role_title"
            v-model="form.role_title"
            type="text"
            placeholder="e.g. Sole Creator, Lead Developer"
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

        <!-- Current / Featured Toggles -->
        <div class="flex flex-col gap-3">
          <div class="flex flex-row items-center gap-2">
            <input
              id="is_current"
              v-model="form.is_current"
              type="checkbox"
              class="w-4 h-4 cursor-pointer shrink-0"
              :style="{ accentColor: 'var(--color-accent)' }"
            />
            <label for="is_current" class="text-sm text-gray-900 cursor-pointer">
              I am currently working on this project
            </label>
          </div>

          <div class="flex flex-row items-center gap-2">
            <input
              id="is_featured"
              v-model="form.is_featured"
              type="checkbox"
              class="w-4 h-4 cursor-pointer shrink-0"
              :style="{ accentColor: 'var(--color-accent)' }"
            />
            <label for="is_featured" class="text-sm text-gray-900 cursor-pointer">
              Feature this project on my profile highlight reel
            </label>
          </div>
        </div>

        <!-- URLs -->
        <div class="grid grid-cols-2 gap-4">
          <div class="flex flex-col gap-2">
            <label class="text-sm font-semibold text-gray-900" for="project_url">Project URL</label>
            <input
              id="project_url"
              v-model="form.project_url"
              type="url"
              placeholder="https://yoursite.com"
              class="px-3 py-2 border border-gray-200 text-sm bg-white text-gray-900 w-full transition-all focus:outline-none focus:border-accent focus:shadow-[0_0_0_1px_var(--color-accent)]"
            />
          </div>

          <div class="flex flex-col gap-2">
            <label class="text-sm font-semibold text-gray-900" for="repository_url"
              >Repository URL (Git)</label
            >
            <input
              id="repository_url"
              v-model="form.repository_url"
              type="url"
              placeholder="https://github.com/..."
              class="px-3 py-2 border border-gray-200 text-sm bg-white text-gray-900 w-full transition-all focus:outline-none focus:border-accent focus:shadow-[0_0_0_1px_var(--color-accent)]"
            />
          </div>
        </div>

        <!-- Description -->
        <div class="flex flex-col gap-2">
          <label class="text-sm font-semibold text-gray-900" for="description">Description</label>
          <textarea
            id="description"
            v-model="form.description"
            placeholder="What technologies did you use? What problem does this solve?..."
            rows="4"
            class="px-3 py-2 border border-gray-200 text-sm bg-white text-gray-900 w-full transition-all focus:outline-none focus:border-accent focus:shadow-[0_0_0_1px_var(--color-accent)] resize-y min-h-24 leading-relaxed font-[inherit]"
          />
        </div>

        <!-- Impact -->
        <div class="flex flex-col gap-2">
          <label class="text-sm font-semibold text-gray-900" for="impact">Impact & outcomes</label>
          <textarea
            id="impact"
            v-model="form.impact"
            placeholder="e.g. Handled 10k requests/min, reduced bundle size by 30%, gained 100 stars..."
            rows="3"
            class="px-3 py-2 border border-gray-200 text-sm bg-white text-gray-900 w-full transition-all focus:outline-none focus:border-accent focus:shadow-[0_0_0_1px_var(--color-accent)] resize-y min-h-16 leading-relaxed font-[inherit]"
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
            {{ isSaving ? 'Saving…' : isEditing ? 'Save changes' : 'Add project' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>
