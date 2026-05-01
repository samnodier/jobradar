<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useRouter, useRoute } from 'vue-router'

const authStore = useAuthStore()
const router = useRouter()
const route = useRoute()

const currentStep = ref(0)
const totalSteps = 2

const form = ref({
  username: ref(''),
  name: ref(''),
  email: ref(''),
})

const username = ref('')
const name = ref('')
const email = ref('')
const error = ref('')
const rawToken = route.query.token
const token = Array.isArray(rawToken) ? rawToken[0] : rawToken

onMounted(async () => {
  if (!token) {
    error.value = 'Missing onboarding token. Please try logging in again.'
    return
  }

  try {
    const res = await fetch(`/api/auth/onboarding?token=${token}`)
    if (!res.ok) {
      const errText = await res.text()
      throw new Error(`Failed to load profile data: ${errText}`)
    }
    const data = await res.json()
    name.value = data.name || ''
    email.value = data.email || ''
    username.value = data.suggested_username || ''
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'An unknown error occurred.'
  }
})

async function completeOnboarding() {
  if (!username.value || !name.value || !email.value) {
    error.value = 'Please fill out all fields.'
    return
  }
  if (!token) {
    error.value = 'Onboarding token is missing. Cannot complete signup.'
    return
  }

  try {
    const res = await fetch(`/api/auth/onboarding?token=${token}`, {
      method: 'POST',
      credentials: 'include',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        username: username.value,
        name: name.value,
        email: email.value,
      }),
    })

    if (!res.ok) {
      const errData = await res.json().catch(() => ({}))
      throw new Error(errData.error || 'Failed to complete onboarding')
    }

    await authStore.fetchMe()
    router.push('/')
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  }
}

// Progress percent bar
const progressPercent = computed(() => {
  return ((currentStep.value + 1) / totalSteps) * 100
})
</script>

<template>
  <div class="onboarding-page">
    <div class="onboarding-container">
      <!-- Progress -->
      <div class="progress-bar">
        <div class="progress-fill" :style="{ width: progressPercent + '%' }"></div>
      </div>
      <div class="progress-text">Step {{ currentStep + 1 }} of {{ totalSteps }}</div>

      <!-- Steps -->
      <div v-show="currentStep === 0" class="step">
        <h2 class="step-title">Welcome! Let's Get Started</h2>
        <p class="step-subtitle">Complete your profile information to get started</p>
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
          <button type="submit">Complete</button>
          <p v-if="error" class="error">{{ error }}</p>
        </form>
      </div>

      <div v-show="currentStep === 3" class="step">
        <h2 class="step-title">You're All Set!</h2>
        <p class="step-subtitle">Your account has been created. You can now start using the app.</p>
        <div class="completion-message">
          <h3>Profile Created Successfully</h3>
          <p>Your profile is ready. Let's find your next opportunity!</p>
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
  background: linear-gradient(135deg, var(--color-accent-lighter) 0%, var(--color-bg-secondary) 100%);
  padding: var(--spacing-8);
}

.onboarding-container {
  background: var(--color-bg-primary);
  border-radius: var(--radius-lg);
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.08);
  width: 100%;
  max-width: 500px;
  padding: var(--spacing-8);
}

.progress-bar {
  height: 4px;
  background: var(--color-bg-secondary);
  border-radius: 2px;
  overflow: hidden;
  margin-bottom: var(--spacing-3);
}

.progress-fill {
  height: 100%;
  background: var(--color-accent);
  transition: width var(--transition-base);
}

.progress-text {
  font-size: var(--text-xs);
  color: var(--color-text-tertiary);
  text-align: right;
  margin-bottom: var(--spacing-8);
}

.step {
  margin-bottom: var(--spacing-8);
}

.step-title {
  font-size: 24px;
  font-weight: var(--font-weight-bold);
  color: var(--color-text-primary);
  margin-bottom: var(--spacing-2);
}

.step-subtitle {
  font-size: var(--text-base);
  color: var(--color-text-secondary);
  margin-bottom: var(--spacing-6);
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-2);
  margin-bottom: var(--spacing-4);
}

.form-label {
  font-size: var(--text-sm);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
}

.form-input {
  padding: var(--spacing-3) var(--spacing-4);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  font-size: var(--text-base);
  font-family: var(--font-family);
  background: var(--color-bg-secondary);
  color: var(--color-text-primary);
  transition: all var(--transition-fast);
}

.form-input:focus {
  outline: none;
  border-color: var(--color-accent);
  background: var(--color-bg-primary);
  box-shadow: 0 0 0 3px var(--color-accent-lighter);
}

.completion-message p {
  font-size: var(--text-base);
  color: var(--color-text-secondary);
}

.step-navigation {
  display: flex;
  gap: var(--spacing-3);
  margin-top: var(--spacing-8);
}

.step-navigation .btn {
  flex: 1;
}

</style>