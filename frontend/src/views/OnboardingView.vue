<script setup lang="ts">
import { reactive, ref, onMounted, computed } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useRouter, useRoute } from 'vue-router'

const authStore = useAuthStore()
const router = useRouter()
const route = useRoute()

const currentStep = ref(0)
const totalSteps = 2
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
    form.username.trim().length > 0 &&
    form.name.trim().length > 0 &&
    form.email.trim().length > 0
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
    const res = await fetch(`/api/auth/onboarding?token=${token}`, {
      credentials: 'include'
    })

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
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        username: form.username.trim(),
        name: form.name.trim(),
        email: form.email.trim(),
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
  <div class="onboarding-page">
    <div class="onboarding-container">
      <!-- Progress -->
      <div class="header">
        <h1 class="step-title">Welcome! Let's Get Started</h1>
        <p class="step-subtitle">Complete your profile information to finish setting up your account.</p>
      </div>

      <!-- Steps -->
      <div v-if="currentStep === 0" class="step">
        <form @submit.prevent="completeOnboarding">
          <div class="form-group">
            <label class="form-label" for="email">Email</label>
            <input class="form-input" type="email" id="email" v-model="form.email" required
              placeholder="john@example.com" />
          </div>
          <div class="form-group">
            <label class="form-label" for="name">Name</label>
            <input class="form-input" type="text" id="name" v-model="form.name" required placeholder="John Doe" />
          </div>
          <div class="form-group">
            <label class="form-label" for="username">Username</label>
            <input class="form-input" type="text" id="username" v-model="form.username" required
              placeholder="john_doe" />
          </div>
          <button class="button button-primary" type="submit" :disabled="loading || !isFormValid">{{ loading ?
            'Saving...' : 'Complete Onboarding' }}</button>
          <p v-if="error" class="error">{{ error }}</p>
        </form>
      </div>

      <div v-else-if="currentStep === 1" class="step">
        <h2 class="step-title">You're All Set!</h2>
        <p class="step-subtitle">Your account has been created successfully.</p>
        <div class="completion-message">
          <h3>Profile Created</h3>
          <p>Your profile is ready. Redirecting your now...</p>
        </div>"
      </div>
    </div>
  </div>
</template>

<style scoped>
.onboarding-page {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  padding: var(--spacing-8);
  background: linear-gradient(135deg,
      var(--color-accent-lighter) 0%,
      var(--color-bg-secondary) 100%);
}

.onboarding-container {
  width: 100%;
  max-width: 500px;
  padding: var(--spacing-8);
  background: var(--color-bg-primary);
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.08);
}

.progress-bar {
  height: 4px;
  margin-bottom: var(--spacing-3);
  overflow: hidden;
  background: var(--color-bg-secondary);
}

.progress-fill {
  height: 100%;
  background: var(--color-accent);
  transition: width var(--transition-base);
}

.progress-text {
  margin-bottom: var(--spacing-8);
  color: var(--color-text-tertiary);
  font-size: var(--text-xs);
  text-align: right;
}

.header {
  margin-bottom: var(--spacing-6)
}

.step-title {
  margin-bottom: var(--spacing-2);
  color: var(--color-text-primary);
  font-size: 24px;
  font-weight: var(--font-weight-bold);
}

.step-subtitle {
  margin-bottom: var(--spacing-6);
  color: var(--color-text-secondary);
  font-size: var(--text-base);
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-2);
  margin-bottom: var(--spacing-4);
}

.form-label {
  color: var(--color-text-primary);
  font-size: var(--text-sm);
  font-weight: var(--font-weight-semibold);
}

.form-input {
  padding: var(--spacing-3) var(--spacing-4);
  color: var(--color-text-primary);
  font-size: var(--text-sm);
  font-family: var(--font-family);
  background: var(--color-bg-secondary);
  border: 1px solid var(--color-border);
  transition: all var(--transition-fast);
}

.form-input:focus {
  outline: none;
  background: var(--color-bg-primary);
  border-color: var(--color-accent);
  box-shadow: 0 0 0 3px var(--color-accent-lighter);
}

.submit-button {
  width: 100%;
  padding: var(--spacing-3) var(--spacing-4);
  color: white;
  font-size: var(--text-base);
  font-weight: var(--font-weight-semibold);
  background: var(--color-accent);
  border: none;
  cursor: pointer;
  transition: opacity var(--transition-fast);
}

.submit-button:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.error {
  margin-top: var(--spacing-4);
  color: var(--color-danger, #dc2626);
  font-size: var(--text-sm);
}

.completion-message {
  padding: var(--spacing-4);
  background: var(--color-bg-secondary);
}

.completion-message h3 {
  margin-bottom: var(--spacing-2);
  color: var(--color-text-primary);
}

.completion-message p {
  color: var(--color-text-secondary);
  font-size: var(--text-base);
}
</style>