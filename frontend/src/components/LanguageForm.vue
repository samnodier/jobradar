<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { X } from '@lucide/vue'
import type { Language, LanguageInput } from '@/types/language'

const props = defineProps<{
  lang?: Language
  open: boolean
  isSaving: boolean
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'save', payload: LanguageInput): void
}>()

const proficiencies = [
  { value: 'native', label: 'Native / Bilingual' },
  { value: 'fluent', label: 'Fluent / Full Professional' },
  { value: 'professional', label: 'Professional Working' },
  { value: 'elementary', label: 'Elementary / Basic' },
]

function blankForm() {
  return {
    user_language: '',
    proficiency: 'professional',
  }
}

const form = ref(blankForm())
const errors = ref<Record<string, string>>({})

watch(
  () => props.open,
  (isOpen) => {
    if (isOpen) {
      if (props.lang) {
        form.value = {
          user_language: props.lang.user_language,
          proficiency: props.lang.proficiency || 'professional',
        }
      } else {
        form.value = blankForm()
      }
      errors.value = {}
    }
  },
  { immediate: true },
)

const isEditing = computed(() => !!props.lang)

function validate(): boolean {
  errors.value = {}
  if (!form.value.user_language.trim()) {
    errors.value.user_language = 'Language name is required.'
  }
  return Object.keys(errors.value).length === 0
}

function handleSubmit() {
  if (!validate()) return
  emit('save', {
    user_language: form.value.user_language.trim(),
    proficiency: form.value.proficiency,
  })
}
</script>

<template>
  <div
    v-if="open"
    class="fixed inset-0 bg-black/45 flex items-start justify-end z-100"
    role="dialog"
    aria-modal="true"
    :aria-label="isEditing ? 'Edit language' : 'Add language'"
    @click.self="emit('close')"
  >
    <!-- Panel -->
    <div
      class="w-full max-w-100 h-dvh bg-white border-l border-gray-200 flex flex-col overflow-hidden"
    >
      <!-- Header -->
      <div class="flex items-center justify-between px-6 py-5 border-b border-gray-200 shrink-0">
        <h2 class="text-lg font-semibold text-gray-900">
          {{ isEditing ? 'Edit language' : 'Add language' }}
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
        <!-- Language Name -->
        <div class="flex flex-col gap-2">
          <label class="text-sm font-semibold text-gray-900" for="lang_name">
            Language <span class="text-red-600 ml-0.5">*</span>
          </label>
          <input
            id="lang_name"
            v-model="form.user_language"
            type="text"
            placeholder="e.g. French, Spanish"
            :disabled="isEditing"
            class="px-3 py-2 border text-sm bg-white text-gray-900 w-full transition-all focus:outline-none focus:border-accent focus:shadow-[0_0_0_1px_var(--color-accent)] disabled:bg-black/4 disabled:opacity-50"
            :class="errors.user_language ? 'border-red-600' : 'border-gray-200'"
          />
          <p v-if="errors.user_language" class="text-xs text-red-600">
            {{ errors.user_language }}
          </p>
        </div>

        <!-- Proficiency -->
        <div class="flex flex-col gap-2">
          <label class="text-sm font-semibold text-gray-900" for="proficiency"
            >Proficiency level</label
          >
          <select
            id="proficiency"
            v-model="form.proficiency"
            class="px-3 py-2 border border-gray-200 text-sm bg-white text-gray-900 w-full transition-all focus:outline-none focus:border-accent focus:shadow-[0_0_0_1px_var(--color-accent)]"
          >
            <option v-for="prof in proficiencies" :key="prof.value" :value="prof.value">
              {{ prof.label }}
            </option>
          </select>
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
            {{ isSaving ? 'Saving…' : isEditing ? 'Save changes' : 'Add language' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>
