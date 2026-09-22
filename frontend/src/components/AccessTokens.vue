<script setup lang="ts">
import { onMounted, ref } from 'vue'
interface AccessToken {
  id: string
  name: string
  expires_at: string
  revoked_at: string | null
}
const tokens = ref<AccessToken[]>([])
const name = ref('Personal workflow')
const secret = ref('')
const error = ref('')
const busy = ref(false)
async function load() {
  const response = await fetch('/api/profile/access-tokens', { credentials: 'include' })
  if (!response.ok) throw new Error('Could not load access tokens')
  tokens.value = await response.json()
}
onMounted(async () => {
  try {
    await load()
  } catch (e) {
    error.value = String(e)
  }
})
async function create() {
  busy.value = true
  error.value = ''
  secret.value = ''
  try {
    const response = await fetch('/api/profile/access-tokens', {
      method: 'POST',
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name: name.value, expires_days: 90 }),
    })
    if (!response.ok) throw new Error('Could not create token; enter a name of 1–80 characters')
    secret.value = (await response.json()).token
    await load()
  } catch (e) {
    error.value = String(e)
  } finally {
    busy.value = false
  }
}
async function revoke(id: string) {
  busy.value = true
  error.value = ''
  try {
    const response = await fetch(`/api/profile/access-tokens/${id}`, {
      method: 'DELETE',
      credentials: 'include',
    })
    if (!response.ok) throw new Error('Could not revoke token')
    secret.value = ''
    await load()
  } catch (e) {
    error.value = String(e)
  } finally {
    busy.value = false
  }
}
</script>
<template>
  <section class="settings-card">
    <h3 class="font-semibold">Personal workflow access</h3>
    <p class="text-sm text-gray-600">
      Create a read-only token to retrieve your matching jobs and Copy for AI context, including
      your saved resume. Tokens expire after 90 days. Keep them private; revoke access here at any
      time.
    </p>
    <p v-if="error" role="alert" class="text-red-600">{{ error }}</p>
    <form class="flex flex-wrap items-end gap-2" @submit.prevent="create">
      <label class="flex-1 min-w-40 text-sm"
        >Token name<input v-model="name" required maxlength="80" class="form-input mt-1"
      /></label>
      <button :disabled="busy" class="button button-primary" type="submit">Create token</button>
    </form>
    <div v-if="secret" class="space-y-2">
      <label class="block text-sm"
        >Save this token now; it will not be shown again.
        <textarea
          :value="secret"
          readonly
          spellcheck="false"
          class="form-input font-mono"
          aria-label="New access token"
        />
      </label>
      <button class="button button-secondary" @click="secret = ''">Hide token</button>
    </div>
    <ul class="space-y-2">
      <li v-for="token in tokens" :key="token.id" class="flex justify-between gap-3 text-sm">
        <span
          >{{ token.name }} ·
          {{
            token.revoked_at
              ? 'Revoked'
              : `Expires ${new Date(token.expires_at).toLocaleDateString()}`
          }}</span
        >
        <button
          v-if="!token.revoked_at"
          :disabled="busy"
          class="button button-secondary"
          @click="revoke(token.id)"
        >
          Revoke
        </button>
      </li>
    </ul>
    <details class="text-sm break-words">
      <summary class="cursor-pointer text-accent">Use a token in your workflow</summary>
      <p class="mt-2">
        Your workflow can request
        <code>/api/export/jobs?min_score=70&amp;date=today&amp;limit=100</code> with this token in
        the Authorization: Bearer header. Follow next_cursor to retrieve further pages. Use the
        job's ID at <code>/api/export/jobs/{id}/ai-context</code> for Markdown.
      </p>
    </details>
  </section>
</template>
