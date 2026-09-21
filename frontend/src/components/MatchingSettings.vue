<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useToast } from '@/composables/useToast'
interface Settings {
  min_score: number
  work_mode: string
  include_unknown: boolean
  search_country: string
  email_enabled: boolean
  notifications_enabled: boolean
  digest_timezone: string
  digest_hours: number[]
  digest_max_jobs: number
}
const settings = ref<Settings | null>(null)
const hours = ref('')
const saving = ref(false)
const error = ref('')
const toast = useToast()
const countries = [
  'at',
  'au',
  'be',
  'br',
  'ca',
  'ch',
  'de',
  'es',
  'fr',
  'gb',
  'in',
  'it',
  'mx',
  'nl',
  'nz',
  'pl',
  'sg',
  'us',
  'za',
]
onMounted(async () => {
  try {
    const r = await fetch('/api/profile/matching', { credentials: 'include' })
    if (!r.ok) throw new Error('Could not load matching settings')
    settings.value = await r.json()
    hours.value = settings.value!.digest_hours.join(', ')
  } catch (e) {
    error.value = String(e)
  }
})
async function save() {
  if (!settings.value) return
  saving.value = true
  error.value = ''
  try {
    const values = hours.value.split(',').map((v) => v.trim())
    if (values.some((v) => !/^\d{1,2}$/.test(v)))
      throw new Error('Use one or two whole hours, for example 8, 18')
    const r = await fetch('/api/profile/matching', {
      method: 'PUT',
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ ...settings.value, digest_hours: values.map(Number) }),
    })
    if (!r.ok) {
      const b = await r.json()
      throw new Error(b.error || 'Could not save settings')
    }
    settings.value = await r.json()
    toast.success('Matching settings saved. Scores will refresh in the background.')
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Could not save'
  } finally {
    saving.value = false
  }
}
</script>
<template>
  <section class="p-6 border border-gray-200 bg-white space-y-4">
    <h2 class="text-lg font-semibold">Matching and delivery</h2>
    <p class="text-sm text-gray-600">
      Your roles, skills, locations and salary are in Preferences. These settings control which
      results reach Best Matches, notifications and email.
    </p>
    <p v-if="error" role="alert" class="text-red-700">{{ error }}</p>
    <form v-if="settings" class="grid gap-4" @submit.prevent="save">
      <label
        >Minimum match score (0–100)<input
          v-model.number="settings.min_score"
          type="number"
          min="0"
          max="100"
          required
          class="border p-2 block w-full"
      /></label>
      <label
        >Work mode<select v-model="settings.work_mode" class="border p-2 block w-full">
          <option value="any">Any</option>
          <option value="remote">Remote</option>
          <option value="onsite">On-site</option>
        </select></label
      >
      <label class="flex gap-2"
        ><input v-model="settings.include_unknown" type="checkbox" />Include jobs with missing
        work-mode, salary or location details</label
      >
      <p class="text-xs text-gray-600">
        Remote does not mean worldwide. Desired locations also apply to remote jobs. Hybrid cannot
        be reliably identified by every source. Review eligibility in the original posting.
      </p>
      <label
        >Adzuna search market<select
          v-model="settings.search_country"
          class="border p-2 block w-full"
        >
          <option v-for="c in countries" :key="c" :value="c">{{ c.toUpperCase() }}</option>
        </select></label
      >
      <p class="text-xs text-gray-600">
        This selects Adzuna's supported market, not your residence. Kenya is not currently in this
        list; Kenyan links can still be imported manually. Searches use up to five desired roles and
        the first desired location.
      </p>
      <label class="flex gap-2"
        ><input v-model="settings.notifications_enabled" type="checkbox" />In-app match
        notifications</label
      >
      <label class="flex gap-2"
        ><input v-model="settings.email_enabled" type="checkbox" />Email digests (requires the
        configured sender)</label
      >
      <label
        >Timezone<input
          v-model="settings.digest_timezone"
          required
          placeholder="Africa/Nairobi"
          class="border p-2 block w-full"
      /></label>
      <label
        >Delivery hours (24-hour clock)<input
          v-model="hours"
          required
          placeholder="8, 18"
          class="border p-2 block w-full"
      /></label>
      <label
        >Maximum jobs per email<input
          v-model.number="settings.digest_max_jobs"
          type="number"
          min="1"
          max="25"
          required
          class="border p-2 block w-full"
      /></label>
      <p class="text-xs text-gray-600">
        One or two digests per day. Empty digests are skipped. A job is emailed once, even if it is
        rescored later. The master Job Recommendations switch must also be enabled.
      </p>
      <button class="button button-primary" :disabled="saving">
        {{ saving ? 'Saving…' : 'Save matching settings' }}
      </button>
    </form>
    <details class="text-sm">
      <summary>How the score works</summary>
      <p class="mt-2">
        Title similarity: 45%; required-skill coverage weighted by your proficiency: 30%; experience
        text overlap: 25%. Missing signals are left out and the remaining weights are rescaled. This
        is deterministic text matching, not an AI judgement or hiring probability. The professional-fit score has weight 80; each known, configured preference has weight 5:
        location, work mode, employment type, salary, industry/category and company stage. Matching
        preferences add credit; mismatches lower the score and exclude the job from recommendations.
        Unknown metadata adds no score weight and follows your missing-information setting.
        Industry/category matching uses literal phrases, not inferred company information.
      </p>
    </details>
  </section>
</template>
