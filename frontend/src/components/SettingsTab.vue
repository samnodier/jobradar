<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import { usePreferencesStore } from '@/stores/preferences'
import { useAuthStore } from '@/stores/auth'
import { useToast } from '@/composables/useToast'

import type { User } from '@/types/user'
import { Check } from '@lucide/vue'

const preferencesStore = usePreferencesStore()
const authStore = useAuthStore()
const router = useRouter()
const toast = useToast()

const isDeleteConfirmVisible = ref(false)
const typedEmail = ref('')
const deleteError = ref('')

interface ApiProvider {
  id: string
  label: string
  keyUrl: string
  keyUrlText: string
  placeholder: string
}

// The AI providers a user can bring a key for. Must stay in sync with the
// backend allowlist (internal/llm IsKnownProvider).
const apiProviders: ApiProvider[] = [
  {
    id: 'groq',
    label: 'Groq API Key',
    keyUrl: 'https://console.groq.com/keys',
    keyUrlText: 'Groq Console',
    placeholder: 'gsk_...',
  },
  {
    id: 'gemini',
    label: 'Google Gemini API Key',
    keyUrl: 'https://aistudio.google.com/apikey',
    keyUrlText: 'Google AI Studio',
    placeholder: 'AIzaSy...',
  },
]

const keyInputs = reactive<Record<string, string>>({})
const replacingKey = reactive<Record<string, boolean>>({})

function isConfigured(providerId: string): boolean {
  return props.user?.configured_providers?.includes(providerId) ?? false
}

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

async function saveApiKey(providerId: string) {
  const key = keyInputs[providerId]
  if (!key) return

  await authStore.setApiKey(providerId, key)
  if (!authStore.error) {
    toast.success('API Key securely saved')
    // Clear the plaintext from state immediately
    keyInputs[providerId] = ''
    replacingKey[providerId] = false
  } else {
    toast.error(authStore.error)
  }
}

function cancelReplaceKey(providerId: string) {
  replacingKey[providerId] = false
  keyInputs[providerId] = ''
}

async function removeApiKey(providerId: string) {
  await authStore.deleteApiKey(providerId)
  if (!authStore.error) {
    toast.success('API key removed')
    replacingKey[providerId] = false
    keyInputs[providerId] = ''
  } else {
    toast.error(authStore.error)
  }
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
        <p class="text-xs text-black mb-4">
          Bring your own API key for AI job matching and imports. When more than one key is
          configured, the fastest provider is used first and the others act as fallbacks if it
          fails.
        </p>
        <div
          v-for="provider in apiProviders"
          :key="provider.id"
          class="flex flex-col gap-2 mb-6 last:mb-0"
        >
          <label :for="`api-key-${provider.id}`" class="text-sm font-semibold text-black">{{
            provider.label
          }}</label>
          <p class="text-xs text-black mb-2">
            Get your free key from
            <a
              :href="provider.keyUrl"
              target="_blank"
              rel="noopener noreferrer"
              class="text-accent hover:underline"
              >{{ provider.keyUrlText }}</a
            >.
          </p>

          <div
            v-if="isConfigured(provider.id) && !replacingKey[provider.id]"
            class="flex items-center justify-between bg-green-50 p-1 border border-green-200"
          >
            <span class="text-sm text-green-700 flex items-center gap-2">
              <Check />
              Key Configured
            </span>
            <div class="flex gap-2">
              <button
                class="button button-secondary text-sm"
                @click="replacingKey[provider.id] = true"
                :disabled="authStore.isSaving"
              >
                Replace
              </button>
              <button
                class="button button-danger text-sm"
                @click="removeApiKey(provider.id)"
                :disabled="authStore.isSaving"
              >
                Remove
              </button>
            </div>
          </div>

          <div v-else class="flex flex-col gap-2">
            <input
              :id="`api-key-${provider.id}`"
              :name="`api-key-${provider.id}`"
              v-model="keyInputs[provider.id]"
              type="password"
              :placeholder="provider.placeholder"
              class="px-4 py-2 border border-gray-200 text-sm bg-white text-black w-full focus:outline-none focus:border-accent transition-all"
              :disabled="authStore.isSaving"
            />
            <div class="flex gap-2 mt-1">
              <button
                class="button button-primary"
                @click="saveApiKey(provider.id)"
                :disabled="!keyInputs[provider.id] || authStore.isSaving"
              >
                {{ authStore.isSaving ? 'Saving...' : 'Save Key' }}
              </button>
              <button
                v-if="isConfigured(provider.id) && replacingKey[provider.id]"
                class="button button-secondary"
                @click="cancelReplaceKey(provider.id)"
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
