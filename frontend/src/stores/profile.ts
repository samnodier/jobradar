import type { Experience } from '@/types/experience'
import { defineStore } from 'pinia'

interface ProfileState {
  experiences: Experience[]
  isSaving: boolean
  loading: boolean
  error: string | null
}

export const useProfileStore = defineStore('profile', {
  state: (): ProfileState => ({
    experiences: [],
    isSaving: false,
    loading: false,
    error: null,
  }),
  actions: {
    async fetchExperiences() {
      this.loading = true
      this.isSaving = false
      this.error = null

      try {
        const response = await fetch('/api/profile/experiences', {
          credentials: 'include',
        })
        if (response.status === 401) {
          this.experiences = []
          return
        }
        if (!response.ok) throw new Error('Failed to fetch experiences')
        this.experiences = await response.json()
      } catch (err) {
        this.error = err instanceof Error ? err.message : 'Failed to load experiences'
      } finally {
        this.loading = false
      }
    },

    async addExperience(exp: Partial<Experience>) {
      this.isSaving = true
      this.error = null

      try {
        const response = await fetch('/api/profile/experiences', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          credentials: 'include',
          body: JSON.stringify(exp),
        })

        if (!response.ok) throw new Error(`Request failed with status ${response.status}`)

        const data = await response.json()
        this.experiences.push(data)
        this.experiences.sort((a, b) => b.start_date.localeCompare(a.start_date))
      } catch (err) {
        this.error = err instanceof Error ? err.message : 'Failed to add experience'
      } finally {
        this.isSaving = false
      }
    },

    async updateExperience(id: string, exp: Partial<Experience>) {
      this.isSaving = true
      this.error = null

      try {
        const response = await fetch(`/api/profile/experiences/${id}`, {
          method: 'PATCH',
          headers: {
            'Content-Type': 'application/json',
          },
          credentials: 'include',
          body: JSON.stringify(exp),
        })

        if (!response.ok) {
          const data = await response.json().catch(() => null)
          throw new Error(data?.error ?? 'Failed to update the experience')
        }

        const data = await response.json()

        const index = this.experiences.findIndex((e) => e.id === id)
        if (index !== -1) this.experiences[index] = data
        this.experiences.sort((a, b) => b.start_date.localeCompare(a.start_date))
      } catch (err) {
        this.error = err instanceof Error ? err.message : 'Failed to update experience'
      } finally {
        this.isSaving = false
      }
    },

    async deleteExperience(id: string) {
      this.isSaving = true
      this.error = null

      try {
        const response = await fetch(`/api/profile/experiences/${id}`, {
          method: 'DELETE',
          credentials: 'include',
        })

        if (!response.ok) {
          const data = await response.json().catch(() => null)
          throw new Error(data?.error ?? 'Failed to delete experience')
        }
        this.experiences = this.experiences.filter((e) => e.id !== id)
      } catch (err) {
        this.error = err instanceof Error ? err.message : 'An error occured'
      } finally {
        this.isSaving = false
      }
    },
  },
})
