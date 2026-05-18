import { defineStore } from 'pinia'
import type { User } from '@/types/user'

interface AuthState {
  user: User | null
  isSaving: boolean
  loading: boolean
  error: string | null
}

export const useAuthStore = defineStore('auth', {
  state: (): AuthState => ({
    user: null,
    isSaving: false,
    loading: false,
    error: null,
  }),

  actions: {
    async fetchMe() {
      this.loading = true
      this.error = null
      try {
        const response = await fetch('/auth/users/me', {
          credentials: 'include',
        })
        if (response.status === 401) {
          this.user = null
          return
        }
        if (!response.ok) throw new Error('Failed to fetch current user')
        const user = await response.json()
        this.user = user
      } catch (e) {
        this.error = e instanceof Error ? e.message : 'Unknown error'
        this.user = null
      } finally {
        this.loading = false
      }
    },

    async editProfile(user: Partial<User>) {
      this.isSaving = true
      this.error = null

      try {
        const response = await fetch(`/api/users/me`, {
          method: 'PATCH',
          headers: {
            'Content-Type': 'application/json',
          },
          credentials: 'include',
          body: JSON.stringify(user),
        })

        if (!response.ok) {
          const data = await response.json().catch(() => null)
          throw new Error(data?.error ?? 'Failed to updated the user profile')
        }

        const updatedUser = await response.json()
        this.user = updatedUser
      } catch (err) {
        this.error = err instanceof Error ? err.message : 'Failed to update user profile'
      } finally {
        this.isSaving = false
      }
    },

    async logout() {
      await fetch('/auth/logout', {
        method: 'POST',
        credentials: 'include',
      })
      this.user = null
    },
  },
})
