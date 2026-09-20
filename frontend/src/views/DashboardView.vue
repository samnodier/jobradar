<script setup lang="ts">
import { ref, onMounted } from 'vue'
import AppSidebar from '@/components/AppSidebar.vue'
interface DigestStatus {
  enabled: boolean
  hours: number[]
  timezone: string
  min_score: number
  history: {
    id: string
    digest_date: string
    status: string
    attempts: number
    sent_at: string | null
    last_error: string | null
  }[]
}
const digest = ref<DigestStatus | null>(null)
const error = ref('')
onMounted(async () => {
  try {
    const response = await fetch('/api/digest', { credentials: 'include' })
    if (!response.ok) throw new Error('Could not load digest status')
    digest.value = await response.json()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Could not load dashboard'
  }
})
</script>
<template>
  <div class="flex h-full">
    <AppSidebar />
    <main class="flex-1 p-8 overflow-y-auto space-y-6">
      <h1 class="text-2xl font-bold">Your job search</h1>
      <p>
        <RouterLink to="/jobs" class="button button-primary">Browse jobs</RouterLink>
        <RouterLink to="/profile" class="button button-secondary"
          >Update matching preferences</RouterLink
        >
      </p>
      <p v-if="error" role="alert">{{ error }}</p>
      <section v-if="digest" class="border p-5 space-y-3">
        <h2 class="text-lg font-semibold">Daily email digest</h2>
        <p v-if="digest.enabled">
          Scheduled after {{ digest.hours.map((h) => String(h).padStart(2, '0') + ':00').join(' and ') }} {{ digest.timezone }} for
          new matches scoring at least {{ digest.min_score }}%. Your notification preference
          controls delivery.
        </p>
        <p v-else>
          Email delivery is not enabled for this account. You can still browse and track jobs.
        </p>
        <a
          v-if="digest.enabled"
          href="/api/digest/preview"
          target="_blank"
          rel="noopener"
          class="button button-secondary"
          >Preview next digest</a
        >
        <p v-if="!digest.history.length">No digest has been prepared yet.</p>
        <ul v-else class="space-y-3">
          <li v-for="item in digest.history" :key="item.id" class="border-t pt-3">
            <strong>{{ item.digest_date }} — {{ item.status }}</strong>
            <p v-if="item.last_error">{{ item.last_error }}</p>
            <p v-if="item.status === 'sent'">
              Accepted by the email provider. Check your inbox and spam folder.
            </p>
            <p v-if="item.status === 'uncertain'">
              Delivery needs review in your email provider before retrying.
            </p>
          </li>
        </ul>
      </section>
    </main>
  </div>
</template>
