import type { Experience, ExperienceInput } from '@/types/experience'
import type { Education, EducationInput } from '@/types/education'
import type { Project, ProjectInput } from '@/types/project'
import type { Certification, CertificationInput } from '@/types/certification'
import type { Language, LanguageInput } from '@/types/language'
import { defineStore } from 'pinia'

interface ProfileState {
  experiences: Experience[]
  educations: Education[]
  projects: Project[]
  certifications: Certification[]
  languages: Language[]
  isSaving: boolean
  loading: boolean
  error: string | null
}

export const useProfileStore = defineStore('profile', {
  state: (): ProfileState => ({
    experiences: [],
    educations: [],
    projects: [],
    certifications: [],
    languages: [],
    isSaving: false,
    loading: false,
    error: null,
  }),
  actions: {
    // ─── EXPERIENCES ───
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

    async addExperience(exp: ExperienceInput) {
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

    async updateExperience(id: string, exp: ExperienceInput) {
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
        this.error = err instanceof Error ? err.message : 'An error occurred'
      } finally {
        this.isSaving = false
      }
    },

    // ─── EDUCATIONS ───
    async fetchEducations() {
      this.loading = true
      this.error = null
      try {
        const response = await fetch('/api/profile/educations', {
          credentials: 'include',
        })
        if (response.status === 401) {
          this.educations = []
          return
        }
        if (!response.ok) throw new Error('Failed to fetch educations')
        this.educations = await response.json()
      } catch (err) {
        this.error = err instanceof Error ? err.message : 'Failed to load educations'
      } finally {
        this.loading = false
      }
    },

    async addEducation(edu: EducationInput) {
      this.isSaving = true
      this.error = null
      try {
        const response = await fetch('/api/profile/educations', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          credentials: 'include',
          body: JSON.stringify(edu),
        })
        if (!response.ok) throw new Error(`Request failed with status ${response.status}`)

        const data = await response.json()
        this.educations.push(data)
        this.educations.sort((a, b) => (b.start_date || '').localeCompare(a.start_date || ''))
      } catch (err) {
        this.error = err instanceof Error ? err.message : 'Failed to add education'
      } finally {
        this.isSaving = false
      }
    },

    async updateEducation(id: string, edu: EducationInput) {
      this.isSaving = true
      this.error = null
      try {
        const response = await fetch(`/api/profile/educations/${id}`, {
          method: 'PATCH',
          headers: {
            'Content-Type': 'application/json',
          },
          credentials: 'include',
          body: JSON.stringify(edu),
        })
        if (!response.ok) {
          const data = await response.json().catch(() => null)
          throw new Error(data?.error ?? 'Failed to update the education')
        }
        const data = await response.json()
        const index = this.educations.findIndex((e) => e.id === id)
        if (index !== -1) this.educations[index] = data
        this.educations.sort((a, b) => (b.start_date || '').localeCompare(a.start_date || ''))
      } catch (err) {
        this.error = err instanceof Error ? err.message : 'Failed to update education'
      } finally {
        this.isSaving = false
      }
    },

    async deleteEducation(id: string) {
      this.isSaving = true
      this.error = null
      try {
        const response = await fetch(`/api/profile/educations/${id}`, {
          method: 'DELETE',
          credentials: 'include',
        })
        if (!response.ok) {
          const data = await response.json().catch(() => null)
          throw new Error(data?.error ?? 'Failed to delete education')
        }
        this.educations = this.educations.filter((e) => e.id !== id)
      } catch (err) {
        this.error = err instanceof Error ? err.message : 'An error occurred'
      } finally {
        this.isSaving = false
      }
    },

    // ─── PROJECTS ───
    async fetchProjects() {
      this.loading = true
      this.error = null
      try {
        const response = await fetch('/api/profile/projects', {
          credentials: 'include',
        })
        if (response.status === 401) {
          this.projects = []
          return
        }
        if (!response.ok) throw new Error('Failed to fetch projects')
        this.projects = await response.json()
      } catch (err) {
        this.error = err instanceof Error ? err.message : 'Failed to load projects'
      } finally {
        this.loading = false
      }
    },

    async addProject(project: ProjectInput) {
      this.isSaving = true
      this.error = null
      try {
        const response = await fetch('/api/profile/projects', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          credentials: 'include',
          body: JSON.stringify(project),
        })
        if (!response.ok) throw new Error(`Request failed with status ${response.status}`)

        const data = await response.json()
        this.projects.push(data)
        this.projects.sort((a, b) => (b.start_date || '').localeCompare(a.start_date || ''))
      } catch (err) {
        this.error = err instanceof Error ? err.message : 'Failed to add project'
      } finally {
        this.isSaving = false
      }
    },

    async updateProject(id: string, project: ProjectInput) {
      this.isSaving = true
      this.error = null
      try {
        const response = await fetch(`/api/profile/projects/${id}`, {
          method: 'PATCH',
          headers: {
            'Content-Type': 'application/json',
          },
          credentials: 'include',
          body: JSON.stringify(project),
        })
        if (!response.ok) {
          const data = await response.json().catch(() => null)
          throw new Error(data?.error ?? 'Failed to update the project')
        }
        const data = await response.json()
        const index = this.projects.findIndex((p) => p.id === id)
        if (index !== -1) this.projects[index] = data
        this.projects.sort((a, b) => (b.start_date || '').localeCompare(a.start_date || ''))
      } catch (err) {
        this.error = err instanceof Error ? err.message : 'Failed to update project'
      } finally {
        this.isSaving = false
      }
    },

    async deleteProject(id: string) {
      this.isSaving = true
      this.error = null
      try {
        const response = await fetch(`/api/profile/projects/${id}`, {
          method: 'DELETE',
          credentials: 'include',
        })
        if (!response.ok) {
          const data = await response.json().catch(() => null)
          throw new Error(data?.error ?? 'Failed to delete project')
        }
        this.projects = this.projects.filter((p) => p.id !== id)
      } catch (err) {
        this.error = err instanceof Error ? err.message : 'An error occurred'
      } finally {
        this.isSaving = false
      }
    },

    // ─── CERTIFICATIONS ───
    async fetchCertifications() {
      this.loading = true
      this.error = null
      try {
        const response = await fetch('/api/profile/certifications', {
          credentials: 'include',
        })
        if (response.status === 401) {
          this.certifications = []
          return
        }
        if (!response.ok) throw new Error('Failed to fetch certifications')
        this.certifications = await response.json()
      } catch (err) {
        this.error = err instanceof Error ? err.message : 'Failed to load certifications'
      } finally {
        this.loading = false
      }
    },

    async addCertification(cert: CertificationInput) {
      this.isSaving = true
      this.error = null
      try {
        const response = await fetch('/api/profile/certifications', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          credentials: 'include',
          body: JSON.stringify(cert),
        })
        if (!response.ok) throw new Error(`Request failed with status ${response.status}`)

        const data = await response.json()
        this.certifications.push(data)
        this.certifications.sort((a, b) => (b.issue_date || '').localeCompare(a.issue_date || ''))
      } catch (err) {
        this.error = err instanceof Error ? err.message : 'Failed to add certification'
      } finally {
        this.isSaving = false
      }
    },

    async updateCertification(id: string, cert: CertificationInput) {
      this.isSaving = true
      this.error = null
      try {
        const response = await fetch(`/api/profile/certifications/${id}`, {
          method: 'PATCH',
          headers: {
            'Content-Type': 'application/json',
          },
          credentials: 'include',
          body: JSON.stringify(cert),
        })
        if (!response.ok) {
          const data = await response.json().catch(() => null)
          throw new Error(data?.error ?? 'Failed to update the certification')
        }
        const data = await response.json()
        const index = this.certifications.findIndex((c) => c.id === id)
        if (index !== -1) this.certifications[index] = data
        this.certifications.sort((a, b) => (b.issue_date || '').localeCompare(a.issue_date || ''))
      } catch (err) {
        this.error = err instanceof Error ? err.message : 'Failed to update certification'
      } finally {
        this.isSaving = false
      }
    },

    async deleteCertification(id: string) {
      this.isSaving = true
      this.error = null
      try {
        const response = await fetch(`/api/profile/certifications/${id}`, {
          method: 'DELETE',
          credentials: 'include',
        })
        if (!response.ok) {
          const data = await response.json().catch(() => null)
          throw new Error(data?.error ?? 'Failed to delete certification')
        }
        this.certifications = this.certifications.filter((c) => c.id !== id)
      } catch (err) {
        this.error = err instanceof Error ? err.message : 'An error occurred'
      } finally {
        this.isSaving = false
      }
    },

    // ─── LANGUAGES ───
    async fetchLanguages() {
      this.loading = true
      this.error = null
      try {
        const response = await fetch('/api/profile/languages', {
          credentials: 'include',
        })
        if (response.status === 401) {
          this.languages = []
          return
        }
        if (!response.ok) throw new Error('Failed to fetch languages')
        this.languages = await response.json()
      } catch (err) {
        this.error = err instanceof Error ? err.message : 'Failed to load languages'
      } finally {
        this.loading = false
      }
    },

    async addLanguage(lang: LanguageInput) {
      this.isSaving = true
      this.error = null
      try {
        const response = await fetch('/api/profile/languages', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          credentials: 'include',
          body: JSON.stringify(lang),
        })
        if (!response.ok) throw new Error(`Request failed with status ${response.status}`)

        const data = await response.json()
        this.languages.push(data)
      } catch (err) {
        this.error = err instanceof Error ? err.message : 'Failed to add language'
      } finally {
        this.isSaving = false
      }
    },

    async updateLanguage(userLanguage: string, lang: LanguageInput) {
      this.isSaving = true
      this.error = null
      try {
        // We hit the URL parameter endpoint, but Go relies on JSON body for the updated name
        const response = await fetch(`/api/profile/languages/${userLanguage}`, {
          method: 'PATCH',
          headers: {
            'Content-Type': 'application/json',
          },
          credentials: 'include',
          body: JSON.stringify(lang),
        })
        if (!response.ok) {
          const data = await response.json().catch(() => null)
          throw new Error(data?.error ?? 'Failed to update the  language')
        }
        const data = await response.json()
        const index = this.languages.findIndex((l) => l.user_language === userLanguage)
        if (index !== -1) this.languages[index] = data
      } catch (err) {
        this.error = err instanceof Error ? err.message : 'Failed to update language'
      } finally {
        this.isSaving = false
      }
    },

    async deleteLanguage(userLanguage: string) {
      this.isSaving = true
      this.error = null
      try {
        const response = await fetch(`/api/profile/languages/${userLanguage}`, {
          method: 'DELETE',
          credentials: 'include',
        })
        if (!response.ok) {
          const data = await response.json().catch(() => null)
          throw new Error(data?.error ?? 'Failed to delete language')
        }
        this.languages = this.languages.filter((l) => l.user_language !== userLanguage)
      } catch (err) {
        this.error = err instanceof Error ? err.message : 'An error occurred'
      } finally {
        this.isSaving = false
      }
    },
  },
})
