import type { UserPreferences } from '@/types/preferences'
import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

// Serialize writes and never replace a newer local draft with an older response.
export const usePreferencesStore = defineStore('preferences', () => {
  const preferences = ref<UserPreferences | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)
  const pending = ref(0)
  const savedSnapshot = ref('')
  const lastSavedAt = ref<Date | null>(null)
  let revision = 0
  let tail: Promise<void> = Promise.resolve()
  const isSaving = computed(() => pending.value > 0)
  const hasUnsavedChanges = computed(
    () => !!preferences.value && JSON.stringify(preferences.value) !== savedSnapshot.value,
  )

  async function fetchPreferences() {
    loading.value = true
    error.value = null
    const startedAt = revision
    try {
      const response = await fetch('/api/profile/preferences', { credentials: 'include' })
      if (!response.ok) throw new Error('Failed to fetch preferences')
      const data: UserPreferences = await response.json()
      if (revision === startedAt && !isSaving.value && !hasUnsavedChanges.value) {
        preferences.value = data
        savedSnapshot.value = JSON.stringify(data)
      }
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to load preferences'
    } finally {
      loading.value = false
    }
  }

  function updatePreferences(patch: Partial<UserPreferences>): Promise<void> {
    const snapshot = JSON.parse(JSON.stringify(patch)) as Partial<UserPreferences>
    if (preferences.value) Object.assign(preferences.value, snapshot)
    const draftAtSave = JSON.stringify(preferences.value)
    const requestRevision = ++revision
    pending.value++
    error.value = null
    const write = tail
      .then(async () => {
        const response = await fetch('/api/profile/preferences', {
          method: 'PATCH',
          headers: { 'Content-Type': 'application/json' },
          credentials: 'include',
          body: JSON.stringify(snapshot),
        })
        if (!response.ok) {
          const data = await response.json().catch(() => null)
          throw new Error(data?.error ?? 'Failed to update preferences')
        }
        const data: UserPreferences = await response.json()
        savedSnapshot.value = JSON.stringify(data)
        // A user can type while a debounced save is still waiting to be queued.
        if (requestRevision === revision && JSON.stringify(preferences.value) === draftAtSave)
          preferences.value = data
        if (requestRevision === revision) {
          error.value = null
          lastSavedAt.value = new Date()
        }
      })
      .catch((err: unknown) => {
        if (requestRevision === revision)
          error.value = err instanceof Error ? err.message : 'Failed to update preferences'
        throw err
      })
      .finally(() => {
        pending.value--
      })
    // A failure must not poison the queue; the explicit Save button can retry.
    tail = write.catch(() => {})
    return write
  }
  async function toggleNotifyJobs() {
    if (!preferences.value) return
    await updatePreferences({ notify_jobs: !preferences.value.notify_jobs })
  }
  return {
    preferences,
    loading,
    error,
    isSaving,
    hasUnsavedChanges,
    lastSavedAt,
    fetchPreferences,
    updatePreferences,
    toggleNotifyJobs,
  }
})
