import { defineStore } from 'pinia'
import type { User } from '@/types/user'

interface AuthState {
  user: User | null
  loading: boolean
  error: string | null
}

export const useAuthStore = defineStore('auth', {
  state: (): AuthState => ({
    user: null,
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

    async logout() {
      await fetch('/auth/logout', {
        method: 'POST',
        credentials: 'include',
      })
      this.user = null
    },
  },
})
