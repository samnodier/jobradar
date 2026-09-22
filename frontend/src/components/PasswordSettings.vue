<script setup lang="ts">
import { onMounted, ref } from 'vue'
const emit = defineEmits<{ changed: [] }>()
interface LoginSettings {
  email: string
  password_configured: boolean
  github_enabled: boolean
}
const settings = ref<LoginSettings | null>(null)
const current = ref('')
const password = ref('')
const confirmation = ref('')
const error = ref('')
const message = ref('')
const busy = ref(false)
async function load() {
  const response = await fetch('/auth/password', { credentials: 'include' })
  if (!response.ok) throw new Error('Could not load sign-in settings')
  settings.value = await response.json()
}
onMounted(async () => {
  try {
    await load()
  } catch (e) {
    error.value = String(e)
  }
})
async function save() {
  error.value = ''
  message.value = ''
  if (password.value !== confirmation.value) {
    error.value = 'Passwords do not match'
    return
  }
  busy.value = true
  try {
    const response = await fetch('/auth/password', {
      method: 'PUT',
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ current_password: current.value, password: password.value }),
    })
    if (!response.ok) {
      const data = await response.json().catch(() => null)
      throw new Error(data?.error || 'Could not save password')
    }
    await load()
    message.value =
      'Password saved. You can now sign in with email. Other sessions and workflow tokens have been revoked.'
    emit('changed')
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    current.value = ''
    password.value = ''
    confirmation.value = ''
    busy.value = false
  }
}
</script>
<template>
  <section class="settings-card">
    <h3 class="font-semibold">Sign-in methods</h3>
    <p v-if="settings" class="text-sm">Email: {{ settings.email }}</p>
    <p v-if="error" role="alert" class="text-sm text-red-600">{{ error }}</p>
    <p v-if="message" role="status" class="text-sm text-green-700">{{ message }}</p>
    <form v-if="settings" class="space-y-3" @submit.prevent="save">
      <p v-if="!settings.password_configured" class="text-sm text-gray-600">
        Add a password within ten minutes of signing in with GitHub. Your profile and saved jobs
        stay on the same account.
      </p>
      <label v-if="settings.password_configured" class="block text-sm"
        >Current password<input
          v-model="current"
          type="password"
          autocomplete="current-password"
          required
          maxlength="1024"
          class="form-input mt-1"
      /></label>
      <label class="block text-sm"
        >New password<input
          v-model="password"
          type="password"
          autocomplete="new-password"
          required
          minlength="15"
          maxlength="1024"
          class="form-input mt-1"
      /></label>
      <label class="block text-sm"
        >Confirm new password<input
          v-model="confirmation"
          type="password"
          autocomplete="new-password"
          required
          maxlength="1024"
          class="form-input mt-1"
      /></label>
      <p class="text-xs text-gray-600">
        At least 15 characters. Saving signs out other sessions and revokes existing workflow
        tokens.
      </p>
      <button :disabled="busy" type="submit" class="button button-primary">
        {{ busy ? 'Saving…' : settings.password_configured ? 'Change password' : 'Add password' }}
      </button>
    </form>
    <div v-if="settings?.github_enabled" class="text-sm space-y-2">
      <p>
        Created your account locally? Connect GitHub using the same verified primary email. Sign in
        again first if your session is more than ten minutes old.
      </p>
      <a href="/auth/github/login?link=1" class="button button-secondary">Connect GitHub</a>
    </div>
  </section>
</template>
