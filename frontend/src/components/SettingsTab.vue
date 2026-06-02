<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import { usePreferencesStore } from '@/stores/preferences'
import { useAuthStore } from '@/stores/auth'
import { useToast } from '@/composables/useToast'

import type { User } from '@/types/user'

const preferencesStore = usePreferencesStore()
const authStore = useAuthStore()
const router = useRouter()
const toast = useToast()

const isDeleteConfirmVisible = ref(false)
const typedEmail = ref('')
const deleteError = ref('')

const geminiInput = ref('')
const isReplacingKey = ref(false)

const { preferences } = storeToRefs(preferencesStore)

const props = defineProps<{
  user: User | null
}>()

async function toggleNotifyJobs() {
  try {
    await preferencesStore.toggleNotifyJobs()
    toast.success(
      preferences.value?.notify_jobs ? 'Job notifications enabled' : 'Job notifications disabled',
    )
  } catch (err) {
    toast.error('Failed to update notification setting' + err)
  }
}

function cancelDelete() {
  isDeleteConfirmVisible.value = false
  typedEmail.value = ''
  deleteError.value = ''
}

async function confirmDelete() {
  // Local validation
  if (typedEmail.value !== props.user?.email) {
    deleteError.value = 'Please enter your email address to confirm account deletion.'
    return
  }
  deleteError.value = ''
  isDeleteConfirmVisible.value = false

  // Trigger store action
  try {
    await authStore.deleteAccount()

    toast.success('Your account has been successfully deleted.')
    router.push({ name: 'login' })
  } catch (err) {
    toast.error(authStore.error || 'Failed to delete account' + err)
    deleteError.value = 'An error occured while deleting your account.'
  }
}

async function saveGeminiKey() {
  if (!geminiInput.value) return
  const key = geminiInput.value

  await authStore.setGeminiKey(key)
  if (!authStore.error) {
    toast.success('API Key securely saved')
    // Clear the plaintext from the ref immediately
    geminiInput.value = ''
    isReplacingKey.value = false
  } else {
    toast.error(authStore.error)
  }
}

function cancelReplaceKey() {
  isReplacingKey.value = false
  geminiInput.value = ''
}
</script>
<template>
  <section class="flex flex-col gap-6">
    <div class="grid gap-6" style="grid-template-columns: repeat(auto-fit, minmax(400px, 1fr))">
      <!-- Notifications -->
      <section class="p-6 bg-white border border-gray-200">
        <h2 class="text-lg font-semibold text-black mb-4">Notifications</h2>
        <div class="flex flex-row justify-between items-center gap-2 mb-4">
          <div class="flex flex-col gap-0.5">
            <label class="text-sm font-semibold text-black">Job Recommendations</label>
            <span class="text-xs text-gray-400">Receive notifications about matching jobs</span>
          </div>
          <button
            class="toggle-switch"
            role="switch"
            :aria-checked="preferences?.notify_jobs"
            @click="toggleNotifyJobs"
          />
        </div>
      </section>

      <section class="p-6 bg-white border border-gray-200">
        <h2 class="text-lg font-semibold text-black mb-4">Integrations</h2>
        <div class="flex flex-col gap-2">
          <label class="text-sm font-semibold text-black">Google Gemini API Key</label>
          <p class="text-xs text-black mb-4">
            Required for AI job matching. Get your free key from
            <a
              href="https://aistudio.google.com/apikey"
              target="_blank"
              rel="noopener noreferrer"
              class="text-accent hover:underline"
              >Google AI Studio</a
            >.
          </p>

          <div
            v-if="props.user?.has_gemini_key && !isReplacingKey"
            class="flex items-center justify-between bg-green-50 p-4 border border-green-200 rounded"
          >
            <span class="text-sm text-green-700 flex items-center gap-2">
              <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                <path
                  fill-rule="evenodd"
                  d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
                  clip-rule="evenodd"
                />
              </svg>
              Key Configured
            </span>
            <button class="button button-secondary text-sm" @click="isReplacingKey = true">
              Replace
            </button>
          </div>

          <div v-else class="flex flex-col gap-2">
            <input
              v-model="geminiInput"
              type="password"
              placeholder="AIzaSy..."
              class="px-4 py-2 border border-gray-200 text-sm bg-white text-black w-full focus:outline-none focus:border-accent transition-all"
              :disabled="authStore.isSaving"
            />
            <div class="flex gap-2 mt-1">
              <button
                class="button button-primary"
                @click="saveGeminiKey"
                :disabled="!geminiInput || authStore.isSaving"
              >
                {{ authStore.isSaving ? 'Saving...' : 'Save Key' }}
              </button>
              <button
                v-if="props.user?.has_gemini_key && isReplacingKey"
                class="button button-secondary"
                @click="cancelReplaceKey"
                :disabled="authStore.isSaving"
              >
                Cancel
              </button>
            </div>
          </div>
        </div>
      </section>

      <!-- Privacy & Account -->
      <section class="p-6 bg-white border border-gray-200">
        <h2 class="text-lg font-semibold text-black mb-4">Privacy & Account</h2>

        <div class="flex flex-col gap-2 mb-4">
          <button class="button button-secondary">Download My Data</button>
        </div>

        <div v-if="!isDeleteConfirmVisible" class="flex flex-col gap-2">
          <button class="button button-danger" @click="isDeleteConfirmVisible = true">
            Delete My Account
          </button>
        </div>

        <div v-else class="flex flex-col gap-2">
          <p class="text-xs text-accent">This action is permanent and cannot be undone.</p>
          <label class="text-sm font-semibold text-black">Type your email to confirm</label>
          <input
            v-model="typedEmail"
            type="email"
            placeholder="your@email.com"
            class="px-4 py-2 border border-gray-200 text-sm bg-white text-black w-full focus:outline-none focus:border-accent focus:shadow-[0_0_0_1px_var(--color-accent)] transition-all"
          />
          <p v-if="deleteError" class="text-xs text-red-600">{{ deleteError }}</p>
          <div class="flex gap-2 mt-1">
            <button class="button button-secondary" @click="cancelDelete">Cancel</button>
            <button
              class="button button-primary"
              :disabled="typedEmail !== props.user?.email"
              @click="confirmDelete"
            >
              Confirm Delete
            </button>
          </div>
        </div>
      </section>
    </div>
  </section>
</template>
