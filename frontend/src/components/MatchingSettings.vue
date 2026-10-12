<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { timezones } from '@/constants/timezones'
import { useToast } from '@/composables/useToast'
interface Settings {
  min_score: number
  work_mode: string
  include_unknown: boolean
  search_country: string
  search_location?: string
  email_enabled: boolean
  notifications_enabled: boolean
  digest_timezone: string
  digest_hours: number[]
  digest_max_jobs: number
}
const settings = ref<Settings | null>(null)
const hours = ref('')
const timezoneOptions = computed(() =>
  [...new Set([...timezones, settings.value?.digest_timezone ?? 'UTC'])].sort(),
)
const saving = ref(false)
const error = ref('')
const toast = useToast()
const saved = ref('')
const draft = computed(() => JSON.stringify({ settings: settings.value, hours: hours.value }))
const dirty = computed(() => !!settings.value && draft.value !== saved.value)
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
    saved.value = draft.value
  } catch (e) {
    error.value = String(e)
  }
})
async function save() {
  if (!settings.value || saving.value) return
  saving.value = true
  error.value = ''
  try {
    const values = hours.value.split(',').map((v) => v.trim())
    if (
      values.length < 1 ||
      values.length > 2 ||
      values.some((v) => !/^\d{1,2}$/.test(v) || Number(v) > 23) ||
      new Set(values.map(Number)).size !== values.length
    )
      throw new Error('Use one or two whole hours, for example 8, 18')
    const submitted = draft.value
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
    const result: Settings = await r.json()
    // A successful response acknowledges the submitted draft, not edits made while waiting.
    saved.value = submitted
    if (draft.value === submitted) {
      settings.value = result
      hours.value = result.digest_hours.join(', ')
      saved.value = draft.value
    }
    toast.success('Matching settings saved. Scores will refresh in the background.')
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Could not save'
  } finally {
    saving.value = false
  }
}
</script>
<template>
  <form class="contents" @submit.prevent="save">
    <section class="settings-card settings-save-bar">
      <div>
        <h2 class="text-lg font-semibold">Matching and delivery</h2>
        <p class="text-sm text-gray-600">
          Edit these four cards, then save. Roles, skills, locations and salary are in Preferences.
        </p>
        <p role="status" class="text-sm mt-2">
          {{
            saving
              ? 'Saving…'
              : error
                ? 'Could not save or load settings'
                : !settings
                  ? 'Loading…'
                  : dirty
                    ? 'Unsaved changes'
                    : 'All changes saved'
          }}
        </p>
        <p v-if="error" role="alert" class="text-sm text-red-700">{{ error }}</p>
      </div>
      <button class="button button-primary shrink-0" :disabled="saving || !settings || !dirty">
        Save matching settings
      </button>
    </section>
    <template v-if="settings">
      <section class="settings-card">
        <h3 class="font-semibold">Match criteria</h3>
        <label class="block text-sm"
          >Minimum match score (0–100)<input
            v-model.number="settings.min_score"
            type="number"
            min="0"
            max="100"
            required
            class="form-input mt-1"
        /></label>
        <label class="block text-sm"
          >Work mode<select v-model="settings.work_mode" class="form-input mt-1">
            <option value="any">Any</option>
            <option value="remote">Remote</option>
            <option value="onsite">On-site</option>
          </select></label
        >
        <label class="flex items-start gap-2 text-sm"
          ><input
            v-model="settings.include_unknown"
            type="checkbox"
            class="mt-1 accent-accent"
          />Include jobs with missing work-mode, salary or location details</label
        >
        <p class="text-xs text-gray-600">
          Remote does not mean worldwide. Desired locations also apply to remote jobs. Review
          eligibility in the original posting.
        </p>
        <details class="text-sm">
          <summary class="cursor-pointer text-accent">How the score works</summary>
          <p class="mt-2 text-gray-600">
            Title similarity: 45%; required-skill coverage weighted by proficiency: 30%; experience
            text overlap: 25%. Missing signals are excluded and remaining weights rescaled.
            Professional fit has weight 80; each known, configured preference has weight 5:
            location, work mode, employment type, salary, industry/category and company stage.
            Matching preferences add credit; mismatches lower the score and exclude recommendations.
            Unknown metadata follows your missing-information setting. This is literal text
            matching, not a hiring probability.
          </p>
        </details>
      </section>
      <section class="settings-card">
        <h3 class="font-semibold">Job sources</h3>
        <label class="block text-sm"
          >Adzuna search market<select v-model="settings.search_country" class="form-input mt-1">
            <option v-for="c in countries" :key="c" :value="c">{{ c.toUpperCase() }}</option>
          </select></label
        >
        <p class="text-sm text-gray-600">
          Select the market to search, not your residence. Searches use up to five desired roles and
          an optional location within that market. Your desired locations still control matching.
        </p>
        <label class="block text-sm"
          >Location within the selected market
          <input
            v-model="settings.search_location"
            maxlength="100"
            class="form-input mt-1"
            placeholder="Blank: whole market; ZA: Johannesburg"
          />
        </label>
        <p class="text-xs text-gray-600">
          Kenya is not supported by this Adzuna integration. You can import Kenyan links or enable
          the MyJobMag feed on your server. Hybrid work is not reliably identified by every source.
        </p>
        <RouterLink to="/status" class="button button-secondary">View source health</RouterLink>
      </section>
      <section class="settings-card">
        <h3 class="font-semibold">Delivery channels</h3>
        <label class="flex items-start gap-2 text-sm"
          ><input
            v-model="settings.notifications_enabled"
            type="checkbox"
            class="mt-1 accent-accent"
          />In-app match notifications</label
        >
        <label class="flex items-start gap-2 text-sm"
          ><input
            v-model="settings.email_enabled"
            type="checkbox"
            class="mt-1 accent-accent"
          />Email digests</label
        >
        <p class="text-sm text-gray-600">
          Both channels use your match criteria. The Job Recommendations master switch must also be
          enabled.
        </p>
        <p class="text-xs text-gray-600">
          Email requires a configured sender and recipient on your server. Turning this on does not
          configure the provider.
        </p>
      </section>
      <section class="settings-card">
        <h3 class="font-semibold">Email schedule</h3>
        <label class="block text-sm"
          >Timezone<select
            v-model="settings.digest_timezone"
            required
            aria-label="Digest timezone"
            class="form-input mt-1"
          >
            <option v-for="zone in timezoneOptions" :key="zone" :value="zone">
              {{ zone.replace(/_/g, ' ') }}
            </option>
          </select>
        </label>
        <p class="text-xs text-gray-600">
          Delivery hours use this timezone, including daylight-saving changes. For morning and
          evening, enter two hours such as 7, 18.
        </p>
        <label class="block text-sm"
          >Delivery hours (24-hour clock)<input
            v-model="hours"
            required
            placeholder="8, 18"
            class="form-input mt-1"
        /></label>
        <label class="block text-sm"
          >Maximum jobs per email<input
            v-model.number="settings.digest_max_jobs"
            type="number"
            min="1"
            max="25"
            required
            class="form-input mt-1"
        /></label>
        <p class="text-xs text-gray-600">
          One or two digests per day. Empty digests are skipped. Each job is emailed once, even if
          rescored later.
        </p>
      </section>
    </template>
  </form>
</template>
