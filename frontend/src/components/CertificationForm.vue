<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { X } from '@lucide/vue'
import type { Certification } from '@/types/certification'

const props = defineProps<{
  certification?: Certification
  open: boolean
  isSaving: boolean
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'save', payload: Omit<Certification, 'id' | 'user_id' | 'created_at' | 'updated_at'>): void
}>()

function blankForm() {
  return {
    certification_name: '',
    issuing_organization: '',
    issue_date: '',
    expiration_date: '',
    does_not_expire: false,
    credential_id: '',
    credential_url: '',
    is_in_progress: false,
  }
}

const form = ref(blankForm())
const errors = ref<Record<string, string>>({})

watch(
  () => props.open,
  (isOpen) => {
    if (isOpen) {
      if (props.certification) {
        form.value = {
          certification_name: props.certification.certification_name,
          issuing_organization: props.certification.issuing_organization,
          issue_date: props.certification.issue_date
            ? props.certification.issue_date.substring(0, 7)
            : '',
          expiration_date: props.certification.expiration_date
            ? props.certification.expiration_date.substring(0, 7)
            : '',
          does_not_expire: props.certification.does_not_expire ?? false,
          credential_id: props.certification.credential_id ?? '',
          credential_url: props.certification.credential_url ?? '',
          is_in_progress: props.certification.is_in_progress ?? false,
        }
      } else {
        form.value = blankForm()
      }
      errors.value = {}
    }
  },
  { immediate: true },
)

const isEditing = computed(() => !!props.certification)
const expDateDisabled = computed(() => form.value.does_not_expire || form.value.is_in_progress)

function validate(): boolean {
  errors.value = {}
  if (!form.value.certification_name.trim()) {
    errors.value.certification_name = 'Certification name is required.'
  }
  if (!form.value.issuing_organization.trim()) {
    errors.value.issuing_organization = 'Issuing organization is required.'
  }
  if (!form.value.is_in_progress && !form.value.issue_date) {
    errors.value.issue_date = 'Issue date is required if not in progress.'
  }
  if (
    form.value.issue_date &&
    form.value.expiration_date &&
    !form.value.does_not_expire &&
    !form.value.is_in_progress &&
    form.value.expiration_date < form.value.issue_date
  ) {
    errors.value.expiration_date = 'Expiration date must be after issue date.'
  }
  return Object.keys(errors.value).length === 0
}

function handleSubmit() {
  if (!validate()) return
  emit('save', {
    certification_name: form.value.certification_name.trim(),
    issuing_organization: form.value.issuing_organization.trim(),
    issue_date: form.value.is_in_progress ? null : form.value.issue_date || null,
    expiration_date:
      form.value.does_not_expire || form.value.is_in_progress
        ? null
        : form.value.expiration_date || null,
    does_not_expire: form.value.does_not_expire,
    credential_id: form.value.credential_id.trim() || null,
    credential_url: form.value.credential_url.trim() || null,
    is_in_progress: form.value.is_in_progress,
  })
}
</script>

<template>
  <div
    v-if="open"
    class="fixed inset-0 bg-black/45 flex items-start justify-end z-100"
    role="dialog"
    aria-modal="true"
    :aria-label="isEditing ? 'Edit certification' : 'Add certification'"
    @click.self="emit('close')"
  >
    <!-- Panel -->
    <div
      class="w-full max-w-140 h-dvh bg-white border-l border-gray-200 flex flex-col overflow-hidden"
    >
      <!-- Header -->
      <div class="flex items-center justify-between px-6 py-5 border-b border-gray-200 shrink-0">
        <h2 class="text-lg font-semibold text-gray-900">
          {{ isEditing ? 'Edit certification' : 'Add certification' }}
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
        <!-- Cert Name -->
        <div class="flex flex-col gap-2">
          <label class="text-sm font-semibold text-gray-900" for="cert_name">
            Certification name <span class="text-red-600 ml-0. 5">*</span>
          </label>
          <input
            id="cert_name"
            v-model="form.certification_name"
            type="text"
            placeholder="e.g. AWS Certified Solutions Architect"
            class="px-3 py-2 border text-sm bg-white text-gray-900 w-full transition-all focus:outline-none focus:border-accent focus:shadow-[0_0_0_1px_var(--color-accent)]"
            :class="errors.certification_name ? 'border-red-600' : 'border-gray-200'"
          />
          <p v-if="errors.certification_name" class="text-xs text-red-600">
            {{ errors.certification_name }}
          </p>
        </div>

        <!-- Issuer -->
        <div class="flex flex-col gap-2">
          <label class="text-sm font-semibold text-gray-900" for="issuer">
            Issuing organization <span class="text-red-600 ml-0. 5">*</span>
          </label>
          <input
            id="issuer"
            v-model="form.issuing_organization"
            type="text"
            placeholder="e.g. Amazon Web Services"
            class="px-3 py-2 border text-sm bg-white text-gray-900 w-full transition-all focus:outline-none focus:border-accent focus:shadow-[0_0_0_1px_var(--color-accent)]"
            :class="errors.issuing_organization ? 'border-red-600' : 'border-gray-200'"
          />
          <p v-if="errors.issuing_organization" class="text-xs text-red-600">
            {{ errors.issuing_organization }}
          </p>
        </div>

        <!-- Start + End Date -->
        <div class="grid grid-cols-2 gap-4">
          <div class="flex flex-col gap-2">
            <label class="text-sm font-semibold text-gray-900" for="issue_date">Issue date</label>
            <input
              id="issue_date"
              v-model="form.issue_date"
              type="month"
              :disabled="form.is_in_progress"
              :max="new Date().toISOString().substring(0, 7)"
              class="px-3 py-2 border text-sm bg-white text-gray-900 w-full transition-all focus:outline-none focus:border-accent focus:shadow-[0_0_0_1px_var(--color-accent)] disabled:opacity-45 disabled:cursor-not-allowed"
              :class="errors.issue_date ? 'border-red-600' : 'border-gray-200'"
            />
            <p v-if="errors.issue_date" class="text-xs text-red- 600">{{ errors.issue_date }}</p>
          </div>

          <div class="flex flex-col gap-2">
            <label class="text-sm font-semibold text-gray-900" for="expiration_date"
              >Expiration date</label
            >
            <input
              id="expiration_date"
              v-model="form.expiration_date"
              type="month"
              :disabled="expDateDisabled"
              :min="form.issue_date || undefined"
              class="px-3 py-2 border text-sm bg-white text-gray-900 w-full transition-all focus:outline-none focus:border-accent focus:shadow-[0_0_0_1px_var(--color-accent)] disabled:opacity-45 disabled:cursor-not-allowed disabled:bg-black/4"
              :class="errors.expiration_date ? 'border-red-600' : 'border-gray-200'"
            />
            <p v-if="errors.expiration_date" class="text-xs text-red-600">
              {{ errors.expiration_date }}
            </p>
          </div>
        </div>

        <!-- Toggles -->
        <div class="flex flex-col gap-3">
          <div class="flex flex-row items-center gap-2">
            <input
              id="is_in_progress"
              v-model="form.is_in_progress"
              type="checkbox"
              class="w-4 h-4 cursor-pointer shrink-0"
              :style="{ accentColor: 'var(--color-accent)' }"
            />
            <label for="is_in_progress" class="text-sm text-gray-900 cursor-pointer">
              I am currently working toward this certification
            </label>
          </div>

          <div class="flex flex-row items-center gap-2">
            <input
              id="does_not_expire"
              v-model="form.does_not_expire"
              type="checkbox"
              :disabled="form.is_in_progress"
              class="w-4 h-4 cursor-pointer shrink-0"
              :style="{ accentColor: 'var(--color-accent)' }"
            />
            <label for="does_not_expire" class="text-sm text-gray-900 cursor-pointer">
              This certification does not expire
            </label>
          </div>
        </div>

        <!-- IDs and Links -->
        <div class="grid grid-cols-2 gap-4">
          <div class="flex flex-col gap-2">
            <label class="text-sm font-semibold text-gray-900" for="credential_id"
              >Credential ID</label
            >
            <input
              id="credential_id"
              v-model="form.credential_id"
              type="text"
              placeholder="e.g. AWS-ASA-12345"
              class="px-3 py-2 border border-gray-200 text-sm bg-white text-gray-900 w-full transition-all focus:outline-none focus:border-accent focus:shadow-[0_0_0_1px_var(--color-accent)]"
            />
          </div>

          <div class="flex flex-col gap-2">
            <label class="text-sm font-semibold text-gray-900" for="credential_url"
              >Credential URL</label
            >
            <input
              id="credential_url"
              v-model="form.credential_url"
              type="url"
              placeholder="https://credly.com/..."
              class="px-3 py-2 border border-gray-200 text-sm bg-white text-gray-900 w-full transition-all focus:outline-none focus:border-accent focus:shadow-[0_0_0_1px_var(--color-accent)]"
            />
          </div>
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
            {{ isSaving ? 'Saving…' : isEditing ? 'Save changes' : 'Add certification' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>
