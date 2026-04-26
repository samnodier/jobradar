<script setup lang="ts">
import { ref } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useRouter } from 'vue-router'

const authStore = useAuthStore()
const router = useRouter()

const username = ref('')
const name = ref('')
const error = ref('')

async function completeOnboarding() {
  if (!username.value || !name.value) {
    error.value = 'Please enter a username and name'
    return
  }

  try {
    const res = await fetch('/api/users/onboarding', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ username: username.value, name: name.value }),
      credentials: 'include',
    })

    if (!res.ok) {
      throw new Error('Failed to complete onboarding')
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
        <label for="username">Username</label>
        <input type="text" id="username" v-model="username" required />
      </div>
      <div>
        <label for="name">Name</label>
        <input type="text" id="name" v-model="name" required />
      </div>
      <button type="submit">Complete</button>
      <p v-if="error" class="error">{{ error }}</p>
    </form>
  </div>
</template>
