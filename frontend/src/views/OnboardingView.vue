<script setup lang="ts">
import { reactive, ref, onMounted, computed } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useRouter, useRoute } from 'vue-router'

const authStore = useAuthStore()
const router = useRouter()
const route = useRoute()

const currentStep = ref(0)
const loading = ref(false)
const error = ref('')

const rawToken = route.query.token
const token = Array.isArray(rawToken) ? rawToken[0] : rawToken

const form = reactive({
  username: '',
  name: '',
  email: '',
})

const isFormValid = computed(() => {
  return (
    form.username.trim().length > 0 && form.name.trim().length > 0 && form.email.trim().length > 0
  )
})

onMounted(async () => {
  if (!token) {
    error.value = 'Missing onboarding token. Please try logging in again.'
    return
  }
  loading.value = true
  error.value = ''
  try {
    const res = await fetch(`/api/auth/onboarding?token=${token}`, { credentials: 'include' })
    if (!res.ok) {
      const errText = await res.text()
      throw new Error(`Failed to load profile data: ${errText}`)
    }
    const data = await res.json()
    form.name = data.name || ''
    form.email = data.email || ''
    form.username = data.suggested_username || ''
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'An unknown error occurred.'
  } finally {
    loading.value = false
  }
})

async function completeOnboarding() {
  error.value = ''
  if (!isFormValid.value) {
    error.value = 'Please fill out all fields.'
    return
  }
  if (!token) {
    error.value = 'Onboarding token is missing. Cannot complete signup.'
    return
  }
  loading.value = true
  try {
    const res = await fetch(`/api/auth/onboarding?token=${token}`, {
      method: 'POST',
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        username: form.username.trim(),
        name: form.name.trim(),
      }),
    })
    if (!res.ok) {
      const errData = await res.json().catch(() => ({}))
      throw new Error(errData.error || 'Failed to complete onboarding')
    }
    await authStore.fetchMe()
    router.push('/')
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Something went wrong.'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div
    class="flex items-center justify-center flex-1 p-8"
    style="
      background: linear-gradient(
        135deg,
        var(--color-accent-lighter) 0%,
        var(--color-bg-secondary) 100%
      );
    "
  >
    <div class="w-full max-w-125 p-8 bg-white shadow-[0_10px_40px_rgba(0,0,0,0.08)]">
      <!-- Header -->
      <div class="mb-6">
        <h1 class="text-2xl font-bold text-gray-900 mb-2">Welcome! Let's Get Started</h1>
        <p class="text-sm text-gray-500 mb-6">
          Complete your profile information to finish setting up your account.
        </p>
      </div>

      <!-- Step 0: Form -->
      <div v-if="currentStep === 0">
        <form @submit.prevent="completeOnboarding">
          <div class="flex flex-col gap-2 mb-4">
            <label class="text-sm font-semibold text-gray-900" for="email">Email</label>
            <input
              id="email"
              v-model="form.email"
              type="email"
              required
              disabled="true"
              placeholder="john@example.com"
              class="px-4 py-2 text-sm text-gray-900 bg-black/4 border border-gray-200 transition-all cursor-not-allowed focus:outline-none focus:bg-white focus:border-accent focus:shadow-[0_0_0_3px_var(--color-accent-lighter)]"
            />
          </div>

          <div class="flex flex-col gap-2 mb-4">
            <label class="text-sm font-semibold text-gray-900" for="name">Name</label>
            <input
              id="name"
              v-model="form.name"
              type="text"
              required
              placeholder="John Doe"
              class="px-4 py-2 text-sm text-gray-900 border border-gray-200 transition-all focus:outline-none focus:bg-white focus:border-accent focus:shadow-[0_0_0_3px_var(--color-accent-lighter)]"
            />
          </div>

          <div class="flex flex-col gap-2 mb-4">
            <label class="text-sm font-semibold text-gray-900" for="username">Username</label>
            <input
              id="username"
              v-model="form.username"
              type="text"
              required
              placeholder="john_doe"
              class="px-4 py-2 text-sm text-gray-900 border border-gray-200 transition-all focus:outline-none focus:bg-white focus:border-accent focus:shadow-[0_0_0_3px_var(--color-accent-lighter)]"
            />
          </div>

          <button type="submit" class="bg-accent px-3 py-2" :disabled="loading || !isFormValid">
            {{ loading ? 'Saving...' : 'Complete Onboarding' }}
          </button>

          <p v-if="error" class="mt-4 text-sm text-red-600">{{ error }}</p>
        </form>
      </div>

      <!-- Step 1: Done -->
      <div v-else-if="currentStep === 1">
        <h2 class="text-2xl font-bold text-gray-900 mb-2">You're All Set!</h2>
        <p class="text-base text-gray-500 mb-6">Your account has been created successfully.</p>
        <div class="p-4 bg-black/4">
          <h3 class="font-semibold text-gray-900 mb-2">Profile Created</h3>
          <p class="text-base text-gray-500">Your profile is ready. Redirecting you now...</p>
        </div>
      </div>
    </div>
  </div>
</template>
