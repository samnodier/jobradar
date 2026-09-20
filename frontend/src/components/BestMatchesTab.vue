<script setup lang="ts">
import { onMounted, ref } from 'vue'
import type { Job } from '@/types/job'
const jobs = ref<Job[]>([])
const loading = ref(true)
const error = ref('')
onMounted(async () => {
  try {
    const r = await fetch('/api/jobs?filter=matched&limit=20', { credentials: 'include' })
    if (!r.ok) throw new Error('Could not load best matches')
    jobs.value = await r.json()
  } catch (e) {
    error.value = String(e)
  } finally {
    loading.value = false
  }
})
</script>
<template>
  <section class="p-6 space-y-4">
    <h2 class="text-lg font-semibold">Your best matches</h2>
    <p class="text-sm text-gray-600">
      Highest-scoring jobs that meet your current preferences and minimum score. Scores refresh
      after profile edits. Enrichment is optional.
    </p>
    <p v-if="loading">Loading…</p>
    <p v-else-if="error" role="alert">{{ error }}</p>
    <p v-else-if="!jobs.length">
      No eligible matches yet. Complete your roles and skills, then check Settings and source
      status.
    </p>
    <RouterLink
      v-for="job in jobs"
      :key="job.id"
      :to="{ path: '/jobs', query: { job: job.id } }"
      class="block border p-4 hover:bg-gray-50"
      ><strong>{{ job.title }}</strong
      ><span class="float-right">{{ job.match_score }}%</span>
      <p class="text-sm">
        {{ job.company_name }} · {{ job.job_location || 'Location unspecified' }}
      </p></RouterLink
    >
    <RouterLink to="/jobs?filter=matched" class="button button-secondary"
      >Browse all matching jobs</RouterLink
    >
  </section>
</template>
