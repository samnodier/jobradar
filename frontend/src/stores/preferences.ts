import type { UserPreferences } from '@/types/preferences'
import { defineStore } from 'pinia'

interface PreferencesState {
  preferences: UserPreferences | null
  isSaving: boolean
  loading: boolean
  error: string | null
}

export const usePreferencesStore = defineStore('preferences', {
  state: (): PreferencesState => ({
    preferences: null,
    isSaving: false,
    loading: false,
    error: null,
  }),
  actions: {
    async fetchPreferences() {
      this.loading = true
      this.error = null
      try {
        const response = await fetch('/api/profile/preferences', {
          credentials: 'include',
        })
        if (!response.ok) throw new Error('Failed to fetch preferences')
        this.preferences = await response.json()
      } catch (err) {
        this.error = err instanceof Error ? err.message : 'Failed to load preferences'
      } finally {
        this.loading = false
      }
    },

    async updatePreferences(preferences: Partial<UserPreferences>) {
      this.isSaving = true
      this.error = null

      try {
        const response = await fetch('/api/profile/preferences', {
          method: 'PATCH',
          headers: { 'Content-Type': 'application/json' },
          credentials: 'include',
          body: JSON.stringify(preferences),
        })
        if (!response.ok) {
          const data = await response.json().catch(() => null)
          throw new Error(data?.error ?? 'Failed to update preferences')
        }
        this.preferences = await response.json()
      } catch (err) {
        this.error = err instanceof Error ? err.message : 'Failed to update preferences'
        throw err
      } finally {
        this.isSaving = false
      }
    },

    async toggleNotifyJobs() {
      if (!this.preferences) return
      const newVal = !this.preferences.notify_jobs
      await this.updatePreferences({ notify_jobs: newVal })
    },
  },
})
