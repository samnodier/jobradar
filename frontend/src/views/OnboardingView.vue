<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useRouter, useRoute } from 'vue-router'

const authStore = useAuthStore()
const router = useRouter()
const route = useRoute()

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
    const res = await fetch(`/auth/onboarding?token=${token}`)
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
    const res = await fetch(`/auth/onboarding?token=${token}`, {
      method: 'POST',
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
</script>
<template>
  <div>
    <h1>Complete your profile</h1>
    <form @submit.prevent="completeOnboarding">
      <div>
        <label for="email">Email</label>
        <input type="email" id="email" v-model="email" required />
      </div>
      <div>
        <label for="name">Name</label>
        <input type="text" id="name" v-model="name" required />
      </div>
      <div>
        <label for="username">Username</label>
        <input type="text" id="username" v-model="username" required />
      </div>
      <button type="submit">Complete</button>
      <p v-if="error" class="error">{{ error }}</p>
    </form>
  </div>
</template>
