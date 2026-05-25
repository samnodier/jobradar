<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { storeToRefs } from 'pinia'
import { usePreferencesStore } from '@/stores/preferences'
import { useToast } from '@/composables/useToast'
import { Plus, SquarePen, X } from '@lucide/vue'
import { useAuthStore } from '@/stores/auth'
import { useRouter } from 'vue-router'
import { useProfileStore } from '@/stores/profile'
import ExperienceCard from '@/components/ExperienceCard.vue'
import ExperienceForm from '@/components/ExperienceForm.vue'
import EditProfileForm from '@/components/EditProfileForm.vue'
import EducationCard from '@/components/EducationCard.vue'
import EducationForm from '@/components/EducationForm.vue'
import ProjectCard from '@/components/ProjectCard.vue'
import ProjectForm from '@/components/ProjectForm.vue'
import CertificationCard from '@/components/CertificationCard.vue'
import CertificationForm from '@/components/CertificationForm.vue'
import LanguageCard from '@/components/LanguageCard.vue'
import LanguageForm from '@/components/LanguageForm.vue'
import AppSidebar from '@/components/AppSidebar.vue'

import type { Education, EducationInput } from '@/types/education'
import type { Project, ProjectInput } from '@/types/project'
import type { Certification, CertificationInput } from '@/types/certification'
import type { Language, LanguageInput } from '@/types/language'
import type { Experience, ExperienceInput } from '@/types/experience'
import type { User } from '@/types/user'
import type { DesiredRolePreference, SkillPreference } from '@/types/preferences'

type ProfileTab = 'overview' | 'experience' | 'preferences' | 'settings'

const router = useRouter()
const authStore = useAuthStore()
const profileStore = useProfileStore()
const { experiences, educations, projects, certifications, languages } = storeToRefs(profileStore)
const toast = useToast()

const activeTab = ref<ProfileTab>('overview')
const saved = ref(false)
const savedJobsCount = ref(0)
const applicationsCount = ref(0)
const isDeleteConfirmVisible = ref(false)
const typedEmail = ref('')
const deleteError = ref('')
const isEditExperienceOpen = ref(false)
const isEditProfileOpen = ref(false)
const selectedExperience = ref<Experience | undefined>(undefined)

const profile = ref<User | undefined>(undefined)
const isEditEducationOpen = ref(false)
const selectedEducation = ref<Education | undefined>(undefined)

const isEditProjectOpen = ref(false)
const selectedProject = ref<Project | undefined>(undefined)

const isEditCertificationOpen = ref(false)
const selectedCertification = ref<Certification | undefined>(undefined)

const isEditLanguageOpen = ref(false)
const selectedLanguage = ref<Language | undefined>(undefined)

const preferencesStore = usePreferencesStore()
const { preferences, isSaving } = storeToRefs(preferencesStore)

const tabs: Array<{ key: ProfileTab; label: string }> = [
  { key: 'overview', label: 'Overview' },
  { key: 'experience', label: 'Resume Details' },
  { key: 'preferences', label: 'Preferences' },
  { key: 'settings', label: 'Settings' },
]

const availabilityLabels: Record<string, string> = {
  immediate: 'Immediately available',
  two_weeks: 'Available in 2 weeks',
  one_month: 'Available in 1 month',
  three_months: 'Available in 3 months',
  not_looking: 'Not actively looking',
}

const industriesText = computed({
  get() {
    return preferences?.value ? preferences.value?.preferred_industries?.join(', ') : ''
  },
  set(val: string) {
    if (preferences?.value) {
      preferences.value.preferred_industries = val
        .split(', ')
        .map((s) => s.trim())
        .filter((s) => s !== '')
    }
  },
})

// bridge simple dropdown and desired_locations array
const preferredLocationSelect = computed({
  get() {
    const locs = preferences?.value ? preferences.value?.desired_locations : []
    if (locs.length === 0) return 'Any'
    const first = locs[0]
    if (first?.location_name.toLowerCase() === 'remote') return 'Remote'
    if (first?.is_remote) return 'Hybrid'
    return 'On-site'
  },
  set(val: string) {
    if (!preferences?.value) return
    if (val === 'Any') {
      preferences.value.desired_locations = []
    } else if (val === 'Remote') {
      preferences.value.desired_locations = [
        { location_name: 'Remote', is_remote: true, priority: 1 },
      ]
    } else if (val === 'On-site') {
      preferences.value.desired_locations = [
        { location_name: 'On-site', is_remote: false, priority: 1 },
      ]
    } else if (val === 'Hybrid') {
      preferences.value.desired_locations = [
        { location_name: 'Hybrid', is_remote: true, priority: 1 },
      ]
    }
  },
})

const removeRole = (roleTitle: string) => {
  if (preferences?.value && preferences.value.desired_roles) {
    preferences.value.desired_roles = preferences.value.desired_roles.filter(
      (r: DesiredRolePreference) => r.role_title !== roleTitle,
    )
  }
}

const addRole = (e: Event) => {
  const target = e.target as HTMLInputElement
  const val = target.value.trim()
  if (val && preferences?.value) {
    if (!preferences.value.desired_roles) {
      preferences.value.desired_roles = []
    }
    const exists = preferences.value.desired_roles.some(
      (r: DesiredRolePreference) => r.role_title.toLowerCase() === val.toLowerCase(),
    )
    if (!exists) {
      preferences.value.desired_roles.push({
        role_title: val,
        priority: null,
      })
    }
    target.value = ''
  }
}

const removeSkill = (skillName: string) => {
  if (preferences?.value && preferences.value.skills) {
    preferences.value.skills = preferences.value.skills.filter(
      (s: SkillPreference) => s.skill_name !== skillName,
    )
  }
}

const addSkill = (e: Event) => {
  const target = e.target as HTMLInputElement
  const val = target.value.trim()
  if (val && preferences?.value) {
    // if skills list doesn't exist yet, initialize it
    if (!preferences.value.skills) {
      preferences.value.skills = []
    }
    const exists = preferences.value.skills.some(
      (s: SkillPreference) => s.skill_name.toLowerCase() === val.toLowerCase(),
    )
    if (!exists) {
      preferences.value.skills.push({
        skill_name: val,
        skill_category: null,
        proficiency: null,
        years_experience: null,
        is_featured: null,
        endorsed_by_ai: null,
        display_order: null,
      })
    }
    target.value = ''
  }
}

function openProfileEditForm(user: User) {
  profile.value = user
  isEditProfileOpen.value = true
}

// Experience handlers
function openExperienceAddForm() {
  selectedExperience.value = undefined
  isEditExperienceOpen.value = true
}

function openExperienceEditForm(exp: Experience) {
  selectedExperience.value = exp
  isEditExperienceOpen.value = true
}

async function handleSaveExperience(
  payload: Omit<Experience, 'id' | 'user_id' | 'created_at' | 'updated_at'>,
) {
  const formattedPayload: ExperienceInput = {
    ...payload,
    skills: payload.skills.map((s) => ({ name: s.skill_name })),
  }
  if (selectedExperience.value) {
    await profileStore.updateExperience(selectedExperience.value.id, formattedPayload)
  } else {
    await profileStore.addExperience(formattedPayload)
  }

  if (profileStore.error) {
    toast.error(profileStore.error)
  } else {
    toast.success(
      selectedExperience.value
        ? 'Experience updated successfully'
        : 'Experience added successfully',
    )
    isEditExperienceOpen.value = false
  }
}

async function handleDeleteExperience(id: string) {
  if (confirm('Are you sure you want to delete this experience?')) {
    await profileStore.deleteExperience(id)
  }
}

async function handleSaveProfile(payload: Partial<User>) {
  await authStore.editProfile(payload)

  if (authStore.error) {
    toast.error(authStore.error)
  } else {
    toast.success('User updated successfully')
    isEditProfileOpen.value = false
  }
}

async function handleSavePreferences() {
  if (!preferences?.value) return
  try {
    await preferencesStore.updatePreferences(preferences.value)
    saved.value = true
    setTimeout(() => (saved.value = false), 2500)
  } catch (err) {
    toast.error(err instanceof Error ? err.message : 'Failed to save preferences.')
  }
}

function cancelDelete() {
  isDeleteConfirmVisible.value = false
  typedEmail.value = ''
  deleteError.value = ''
}

async function deleteAccount() {
  if (typedEmail.value !== authStore.user?.email) {
    deleteError.value = 'Please enter your email address to confirm account deletion.'
    return
  }
  try {
    const response = await fetch('/api/users/me', { method: 'DELETE', credentials: 'include' })
    if (!response.ok) {
      deleteError.value = 'Failed to delete account. Please try again.'
      return
    }
    authStore.user = null
    router.push('/login')
  } catch (err) {
    deleteError.value =
      'Something went wrong. Please try again' + (err instanceof Error ? err.message : String(err))
  }
}

// Education handlers
function openEducationAddForm() {
  selectedEducation.value = undefined
  isEditEducationOpen.value = true
}

function openEducationEditForm(edu: Education) {
  selectedEducation.value = edu
  isEditEducationOpen.value = true
}

async function handleSaveEducation(payload: EducationInput) {
  if (selectedEducation.value) {
    await profileStore.updateEducation(selectedEducation.value.id, payload)
  } else {
    await profileStore.addEducation(payload)
  }
  if (profileStore.error) {
    toast.error(profileStore.error)
  } else {
    toast.success(
      selectedEducation.value ? 'Education updated successfully' : 'Education added successfully',
    )
    isEditEducationOpen.value = false
  }
}

async function handleDeleteEducation(id: string) {
  if (confirm('Are you sure you want to delete this education?')) {
    await profileStore.deleteEducation(id)
    if (profileStore.error) {
      toast.error(profileStore.error)
    } else {
      toast.success('Education deleted successfully')
    }
  }
}

// Project handlers
function openProjectAddForm() {
  selectedProject.value = undefined
  isEditProjectOpen.value = true
}

function openProjectEditForm(project: Project) {
  selectedProject.value = project
  isEditProjectOpen.value = true
}

async function handleSaveProject(payload: ProjectInput) {
  if (selectedProject.value) {
    await profileStore.updateProject(selectedProject.value.id, payload)
  } else {
    await profileStore.addProject(payload)
  }
  if (profileStore.error) {
    toast.error(profileStore.error)
  } else {
    toast.success(
      selectedProject.value ? 'Project updated successfully' : 'Project added successfully',
    )
    isEditProjectOpen.value = false
  }
}

async function handleDeleteProject(id: string) {
  if (confirm('Are you sure you want to delete this project?')) {
    await profileStore.deleteProject(id)
    if (profileStore.error) {
      toast.error(profileStore.error)
    } else {
      toast.success('Project deleted successfully')
    }
  }
}

// Certification handlers
function openCertificationAddForm() {
  selectedCertification.value = undefined
  isEditCertificationOpen.value = true
}

function openCertificationEditForm(certification: Certification) {
  selectedCertification.value = certification
  isEditCertificationOpen.value = true
}

async function handleSaveCertification(payload: CertificationInput) {
  if (selectedCertification.value) {
    await profileStore.updateCertification(selectedCertification.value.id, payload)
  } else {
    await profileStore.addCertification(payload)
  }
  if (profileStore.error) {
    toast.error(profileStore.error)
  } else {
    toast.success(
      selectedCertification.value
        ? 'Certification updated successfully'
        : 'Certification added successfully',
    )
    isEditCertificationOpen.value = false
  }
}

async function handleDeleteCertification(id: string) {
  if (confirm('Are you sure you want to delete this certification?')) {
    await profileStore.deleteCertification(id)
    if (profileStore.error) {
      toast.error(profileStore.error)
    } else {
      toast.success('Certification deleted successfully')
    }
  }
}

// Language handlers
function openLanguageAddForm() {
  selectedLanguage.value = undefined
  isEditLanguageOpen.value = true
}

function openLanguageEditForm(language: Language) {
  selectedLanguage.value = language
  isEditLanguageOpen.value = true
}

async function handleSaveLanguage(payload: LanguageInput) {
  if (selectedLanguage.value) {
    await profileStore.updateLanguage(selectedLanguage.value.user_language, payload)
  } else {
    await profileStore.addLanguage(payload)
  }
  if (profileStore.error) {
    toast.error(profileStore.error)
  } else {
    toast.success(
      selectedLanguage.value ? 'Language updated successfully' : 'Language added successfully',
    )
    isEditLanguageOpen.value = false
  }
}

async function handleDeleteLanguage(id: string) {
  if (confirm('Are you sure you want to delete this language?')) {
    await profileStore.deleteLanguage(id)
    if (profileStore.error) {
      toast.error(profileStore.error)
    } else {
      toast.success('Language deleted successfully')
    }
  }
}

async function fetchStats() {
  try {
    const savedRes = await fetch('/api/saved_jobs', {
      credentials: 'include',
    })
    if (savedRes.ok) {
      const savedData = await savedRes.json()
      savedJobsCount.value = savedData.length
    }
    const appsRes = await fetch('/api/applications', {
      credentials: 'include',
    })
    if (appsRes.ok) {
      const appsData = await appsRes.json()
      applicationsCount.value = appsData.length
    }
  } catch (err) {
    toast.error('Failed to fetch stats' + err)
  }
}

onMounted(async () => {
  await profileStore.fetchExperiences()
  await profileStore.fetchEducations()
  await profileStore.fetchProjects()
  await profileStore.fetchCertifications()
  await profileStore.fetchLanguages()
  await preferencesStore.fetchPreferences()
  await fetchStats()
})
</script>

<template>
  <div class="grid h-full overflow-hidden" style="grid-template-columns: 220px minmax(0, 1fr)">
    <AppSidebar />

    <main class="flex flex-col gap-6 bg-white overflow-y-auto min-h-0">
      <!-- Profile Header area -->
      <div class="w-full mx-auto px-8 py-8">
        <div class="flex flex-col gap-8 border-b border-gray-200 pb-6">
          <!-- Header -->
          <div class="flex items-start justify-between gap-6">
            <div class="flex items-start gap-4 flex-1">
              <!-- Avatar -->
              <div class="shrink-0">
                <div
                  class="w-20 h-20 rounded-full bg-black/4 border border-gray-200 grid place-items-center overflow-hidden"
                  style="
                    box-shadow:
                      0 0 0 4px white,
                      0 0 0 5px #e5e7eb;
                  "
                >
                  <img
                    v-if="authStore.user?.avatar_url"
                    :src="authStore.user.avatar_url"
                    alt="Profile avatar"
                    class="w-full h-full object-cover"
                  />
                  <span v-else class="text-gray-900 font-bold text-2xl">
                    {{ authStore.user?.username?.[0]?.toUpperCase() || 'U' }}
                  </span>
                </div>
              </div>

              <!-- Copy -->
              <div>
                <h1 class="text-2xl font-bold text-gray-900 mb-1 tracking-tight">
                  {{ authStore.user?.full_name || authStore.user?.username }}
                </h1>
                <p class="text-sm text-gray-500 mb-3">
                  {{ authStore.user?.email }}
                  <span v-if="authStore.user?.username"> • {{ authStore.user.username }}</span>
                </p>
                <p class="text-base text-gray-500 leading-relaxed">
                  {{ authStore.user?.user_summary || 'Add a summary to your profile.' }}
                </p>
              </div>
            </div>

            <!-- Actions -->
            <div class="flex flex-row items-center justify-center gap-3">
              <button
                class="inline-flex items-center bg-accent text-accent-foreground px-4 py-2 gap-2 transition-colors duration-150 hover:bg-accent/90 active:bg-accent/80 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-accent focus-visible:ring-offset-2 disabled:opacity-50 disabled:pointer-events-none"
                @click="authStore.user && openProfileEditForm(authStore.user)"
              >
                <SquarePen :size="16" />
                Edit Profile
              </button>
              <button class="border border-gray-200 bg-gray-100 px-4 py-2">Export To PDF</button>
            </div>
          </div>

          <!-- Stats -->
          <div class="grid grid-cols-3 gap-4 mb-2">
            <div
              v-for="stat in [
                { num: savedJobsCount, label: 'Saved jobs' },
                { num: applicationsCount, label: 'Applications' },
                { num: experiences?.length || 0, label: 'Experiences' },
              ]"
              :key="stat.label"
              class="bg-black/4 border border-gray-200 p-4 text-center flex flex-col"
            >
              <span class="text-3xl font-bold text-gray-900">{{ stat.num }}</span>
              <span class="text-sm font-normal text-gray-500 mt-3">{{ stat.label }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Tabs + content -->
      <div class="w-full mx-auto px-8 pb-8">
        <div class="flex flex-col gap-8 border-b border-gray-200 pb-6">
          <!-- Tab nav -->
          <nav class="flex gap-0 border-b border-gray-200" aria-label="Profile sections">
            <button
              v-for="tab in tabs"
              :key="tab.key"
              type="button"
              class="px-4 py-2 text-sm font-medium border-b-2 -mb-px transition-all cursor-pointer"
              :class="
                activeTab === tab.key
                  ? 'border-b-2 text-accent'
                  : 'border-transparent text-gray-500 hover:text-gray-900'
              "
              :style="activeTab === tab.key ? { borderBottomColor: 'var(--color-accent)' } : {}"
              @click="activeTab = tab.key"
            >
              {{ tab.label }}
            </button>
          </nav>

          <!-- Overview -->
          <section
            v-if="activeTab === 'overview'"
            class="grid gap-6"
            style="grid-template-columns: repeat(auto-fit, minmax(320px, 1fr))"
          >
            <!-- Professional Info Card -->
            <div class="p-6 border border-gray-200 flex flex-col gap-4 bg-white">
              <h2 class="text-sm font-semibold uppercase text-gray-400 tracking-wider m-0">
                Professional Summary
              </h2>

              <div class="flex flex-col gap-1">
                <span class="text-xs text-gray-400 font-semibold uppercase">Headline</span>
                <span class="text-sm font-medium text-gray-900">{{
                  authStore.user?.headline || 'Not set'
                }}</span>
              </div>

              <div class="flex flex-col gap-1">
                <span class="text-xs text-gray-400 font-semibold uppercase"
                  >Availability Status</span
                >
                <div>
                  <span
                    class="inline-flex items-center gap-1.5 px-2.5 py-0.5 text-xs font-semibold border"
                    :class="
                      authStore.user?.availability === 'not_looking'
                        ? 'bg-gray-50 border-gray-200 text-gray-600'
                        : 'bg-green-50 border-green-200 text-green-700'
                    "
                  >
                    <span
                      class="w-1.5 h-1.5 rounded-full"
                      :class="
                        authStore.user?.availability === 'not_looking'
                          ? 'bg-gray-400'
                          : 'bg-green-500'
                      "
                    ></span>
                    {{ availabilityLabels[authStore.user?.availability || 'immediate'] }}
                  </span>
                </div>
              </div>
            </div>

            <!-- Contact & Social Links Card -->
            <div class="p-6 border border-gray-200 flex flex-col gap-4 bg-white">
              <h2 class="text-sm font-semibold uppercase text-gray-400 tracking-wider m-0">
                Contact & Socials
              </h2>

              <div class="flex flex-col gap-3">
                <div v-if="authStore.user?.email" class="flex justify-between items-center text-sm">
                  <span class="text-gray-400 font-medium">Email</span>
                  <a
                    :href="`mailto:${authStore.user.email}`"
                    class="text-accent no-underline hover:underline"
                    >{{ authStore.user.email }}</a
                  >
                </div>

                <div class="flex justify-between items-center text-sm">
                  <span class="text-gray-400 font-medium">Phone</span>
                  <span class="text-gray-900 font-medium">{{
                    authStore.user?.phone || 'Not set'
                  }}</span>
                </div>

                <div class="flex justify-between items-center text-sm">
                  <span class="text-gray-400 font-medium">GitHub</span>
                  <a
                    v-if="authStore.user?.github_url"
                    :href="authStore.user.github_url"
                    target="_blank"
                    rel="noreferrer"
                    class="text-accent no-underline hover:underline"
                    >GitHub Profile</a
                  >
                  <span v-else class="text-gray-400 italic">Not linked</span>
                </div>

                <div class="flex justify-between items-center text-sm">
                  <span class="text-gray-400 font-medium">LinkedIn</span>
                  <a
                    v-if="authStore.user?.linkedin_url"
                    :href="authStore.user.linkedin_url"
                    target="_blank"
                    rel="noreferrer"
                    class="text-accent no-underline hover:underline"
                    >LinkedIn Profile</a
                  >
                  <span v-else class="text-gray-400 italic">Not linked</span>
                </div>

                <div class="flex justify-between items-center text-sm">
                  <span class="text-gray-400 font-medium">Website</span>
                  <a
                    v-if="authStore.user?.website_url"
                    :href="authStore.user.website_url"
                    target="_blank"
                    rel="noreferrer"
                    class="text-accent no-underline hover:underline"
                    >Personal Website</a
                  >
                  <span v-else class="text-gray-400 italic">Not set</span>
                </div>
              </div>
            </div>

            <!-- Target Job Preferences Card -->
            <div class="p-6 border border-gray-200 flex flex-col gap-4 bg-white">
              <h2 class="text-sm font-semibold uppercase text-gray-400 tracking-wider m-0">
                Target Job preferences
              </h2>

              <div class="flex justify-between items-center text-sm">
                <span class="text-gray-400 font-medium">Years of Experience</span>
                <span class="text-gray-900 font-semibold">{{
                  authStore.user?.years_of_experience ?? 'Not set'
                }}</span>
              </div>

              <div class="flex justify-between items-center text-sm">
                <span class="text-gray-400 font-medium">Target Salary</span>
                <span
                  v-if="authStore.user?.min_salary || authStore.user?.max_salary"
                  class="text-gray-900 font-semibold"
                >
                  {{ authStore.user.salary_currency || 'USD' }}
                  {{ authStore.user.min_salary ?? 0 }} — {{ authStore.user.max_salary ?? 0 }}
                </span>
                <span v-else class="text-gray-400 italic">Not set</span>
              </div>

              <div class="flex flex-col gap-2">
                <span class="text-xs text-gray-400 font-semibold uppercase"
                  >Preferred Job Types</span
                >
                <div
                  v-if="authStore.user?.preferred_job_types?.length"
                  class="flex flex-wrap gap-1.5"
                >
                  <span
                    v-for="type in authStore.user.preferred_job_types"
                    :key="type"
                    class="text-[11px] px-2 py-0.5 bg-black/4 border border-gray-200 text-gray-600 font-medium"
                    >{{ type }}</span
                  >
                </div>
                <span v-else class="text-sm text-gray-400 italic">Any job type</span>
              </div>
            </div>
          </section>

          <!-- Experience -->
          <section v-if="activeTab === 'experience'" class="flex flex-col">
            <!-- Work History Section -->

            <div class="p-6 border border-gray-200">
              <div class="flex justify-between items-center mb- 4">
                <h2 class="text-lg font-semibold text-gray-900 tracking-tight">
                  Work history ({{ experiences?.length || 0 }})
                </h2>
                <button
                  class="bg-accent text-white px-4 py-2 inline- flex gap-2 text-sm font-semibold cursor-pointer transition- transform hover:-translate-y-px"
                  @click="openExperienceAddForm"
                >
                  <Plus /> Add Experience
                </button>
              </div>
              <p
                v-if="!experiences?.length"
                class="text-center py-8 text-gray-400 border border-dashed border-gray-200 text-sm"
              >
                No work history added yet.
              </p>
              <div v-else class="flex flex-col gap-4">
                <ExperienceCard
                  v-for="exp in experiences"
                  :key="exp.id"
                  :exp="exp"
                  @edit="openExperienceEditForm(exp)"
                  @delete="handleDeleteExperience(exp.id)"
                />
              </div>
            </div>

            <!-- Education Section -->
            <div class="p-6 border border-gray-200">
              <div class="flex justify-between items-center mb- 4">
                <h2 class="text-lg font-semibold text-gray-900 tracking-tight">
                  Education ({{ educations?.length || 0 }})
                </h2>
                <button
                  class="bg-accent text-white px-4 py-2 inline- flex gap-2 text-sm font-semibold cursor-pointer transition- transform hover:-translate-y-px"
                  @click="openEducationAddForm"
                >
                  <Plus /> Add Education
                </button>
              </div>
              <p
                v-if="!educations?.length"
                class="text-center py-8 text-gray-400 border border-dashed border-gray-200 text-sm"
              >
                No education records added yet.
              </p>
              <div v-else class="flex flex-col gap-4">
                <EducationCard
                  v-for="edu in educations"
                  :key="edu.id"
                  :edu="edu"
                  @edit="openEducationEditForm(edu)"
                  @delete="handleDeleteEducation(edu.id)"
                />
              </div>
            </div>

            <!-- Projects Section -->
            <div class="p-6 border border-gray-200">
              <div class="flex justify-between items-center mb- 4">
                <h2 class="text-lg font-semibold text-gray-900 tracking-tight">
                  Projects ({{ projects?.length || 0 }})
                </h2>
                <button
                  class="bg-accent text-white px-4 py-2 inline- flex gap-2 text-sm font-semibold cursor-pointer transition- transform hover:-translate-y-px"
                  @click="openProjectAddForm"
                >
                  <Plus /> Add Project
                </button>
              </div>
              <p
                v-if="!projects?.length"
                class="text-center py-8 text-gray-400 border border-dashed border-gray-200 text-sm"
              >
                No projects added yet.
              </p>
              <div v-else class="flex flex-col gap-4">
                <ProjectCard
                  v-for="project in projects"
                  :key="project.id"
                  :project="project"
                  @edit="openProjectEditForm(project)"
                  @delete="handleDeleteProject(project.id)"
                />
              </div>
            </div>

            <!-- Certifications and Languages Grid -->
            <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
              <!-- Certifications -->
              <div class="p-6 border border-gray-200 flex flex- col h-full">
                <div class="flex justify-between items-center mb- 4">
                  <h2 class="text-lg font-semibold text-gray-900 tracking-tight">
                    Certifications ({{ certifications?.length || 0 }})
                  </h2>
                  <button
                    class="bg-accent text-white px-3 py-1.5 inline-flex gap-1.5 text-xs font-semibold cursor-pointer transition-transform hover:-translate-y-px"
                    @click="openCertificationAddForm"
                  >
                    <Plus :size="14" /> Add Cert
                  </button>
                </div>
                <p
                  v-if="!certifications?.length"
                  class="text-center py-8 text-gray-400 border border-dashed border-gray-200 text-sm flex-1 grid place-items- center"
                >
                  No certifications added yet.
                </p>
                <div v-else class="flex flex-col gap-3 flex-1 overflow-y-auto">
                  <CertificationCard
                    v-for="cert in certifications"
                    :key="cert.id"
                    :cert="cert"
                    @edit="openCertificationEditForm(cert)"
                    @delete="handleDeleteCertification(cert.id)"
                  />
                </div>
              </div>

              <!-- Spoken Languages -->
              <div class="p-6 border border-gray-200 flex flex- col h-full">
                <div class="flex justify-between items-center mb- 4">
                  <h2 class="text-lg font-semibold text-gray-900 tracking-tight">
                    Languages ({{ languages?.length || 0 }})
                  </h2>
                  <button
                    class="bg-accent text-white px-3 py-1.5 inline-flex gap-1.5 text-xs font-semibold cursor-pointer transition-transform hover:-translate-y-px"
                    @click="openLanguageAddForm"
                  >
                    <Plus :size="14" /> Add Language
                  </button>
                </div>
                <p
                  v-if="!languages?.length"
                  class="text-center py-8 text-gray-400 border border-dashed border-gray-200 text-sm flex-1 grid place-items- center"
                >
                  No languages added yet.
                </p>
                <div v-else class="grid grid-cols-1 gap-3 flex-1 overflow-y-auto">
                  <LanguageCard
                    v-for="lang in languages"
                    :key="lang.user_language"
                    :lang="lang"
                    @edit="openLanguageEditForm(lang)"
                    @delete="handleDeleteLanguage(lang.user_language)"
                  />
                </div>
              </div>
            </div>
          </section>

          <!-- Preferences -->
          <section v-if="activeTab === 'preferences' && preferences" class="flex flex-col gap-6">
            <div class="pb-6 border-b border-gray-200">
              <h1 class="text-2xl font-bold text-gray-900 mb-2">Preferences & Settings</h1>
              <p class="text-base text-gray-500">
                Fine-tune your job recommendations and notification preferences
              </p>
            </div>

            <div
              class="grid gap-6"
              style="grid-template-columns: repeat(auto-fit, minmax(400px, 1fr))"
            >
              <!-- Job Preferences -->
              <section class="p-6 bg-black/4 border border-gray-200">
                <h2 class="text-lg font-semibold text-gray-900 mb-4">Job Preferences</h2>

                <div class="flex flex-col gap-2 mb-4">
                  <label class="text-sm font-semibold text-gray-900">Preferred Job Types</label>
                  <div class="flex flex-col gap-2">
                    <div
                      v-for="type in ['Full-time', 'Part-time', 'Contract']"
                      :key="type"
                      class="flex items-center gap-2"
                    >
                      <input
                        :id="type"
                        v-model="preferences.preferred_job_types"
                        type="checkbox"
                        :value="type"
                        class="w-4 h-4 cursor-pointer"
                        :style="{ accentColor: 'var(--color-accent)' }"
                      />
                      <label :for="type" class="cursor-pointer text-sm text-gray-900">{{
                        type
                      }}</label>
                    </div>
                  </div>
                </div>

                <div class="flex flex-col gap-2 mb-4">
                  <label class="text-sm font-semibold text-gray-900"> Desired Roles</label>
                  <div class="flex flex-wrap gap-2 mb-2">
                    <span
                      v-for="role in preferences.desired_roles"
                      :key="role.role_title"
                      class="flex items-center gap-2 px-2 py-1 bg-white border text-xs font-medium"
                    >
                      {{ role.role_title }}
                      <button
                        type="button"
                        class="bg-transparent border-none cursor-pointer text-sm p-0 leading-none"
                        @click="removeRole(role.role_title)"
                      >
                        <X :size="12" />
                      </button>
                    </span>
                  </div>
                  <input
                    type="text"
                    placeholder="Type a role (e.g. Frontend Engineer) and press Enter"
                    class="px-3 py-2 border border-gray-200 text-sm bg-white text-gray-900 w-full focus:outline-none focus:border-accent focus:ring-1 focus:ring-accent transition-all"
                    @keydown.enter.prevent="addRole($event)"
                  />
                </div>

                <div class="flex flex-col gap-2 mb-4">
                  <label class="text-sm font-semibold text-gray-900">Work Location</label>
                  <select
                    v-model="preferredLocationSelect"
                    class="px-3 py-2 border border-gray-200 text-sm bg-white text-gray-900 w-full focus:outline-none focus:border-accent focus:shadow-[0_0_0_1px_var(--color-accent)] transition-all"
                  >
                    <option>Remote</option>
                    <option>On-site</option>
                    <option>Hybrid</option>
                    <option>Any</option>
                  </select>
                </div>

                <div class="flex gap-4">
                  <div class="flex flex-col gap-2 flex-1">
                    <label class="text-sm font-semibold text-gray-900">Min Salary</label>
                    <input
                      v-model="preferences.min_salary"
                      type="number"
                      placeholder="Min"
                      class="px-3 py-2 border border-gray-200 text-sm bg-white text-gray-900 w-full focus:outline-none focus:border-accent focus:shadow-[0_0_0_1px_var(--color-accent)] transition-all"
                    />
                  </div>
                  <div class="flex flex-col gap-2 flex-1">
                    <label class="text-sm font-semibold text-gray-900">Max Salary</label>
                    <input
                      v-model="preferences.max_salary"
                      type="number"
                      placeholder="Max"
                      class="px-3 py-2 border border-gray-200 text-sm bg-white text-gray-900 w-full focus:outline-none focus:border-accent focus:shadow-[0_0_0_1px_var(--color-accent)] transition-all"
                    />
                  </div>
                  <div class="flex flex-col gap-2 w-28">
                    <label class="text-sm font-semibold text-gray-900">Currency</label>

                    <select
                      v-model="preferences.salary_currency"
                      class="px-3 py-2 border border-gray-200 text-sm bg-white text-gray-900 w-full focus:outline-none focus:border-accent focus:ring-1 focus:ring-accent transitional-all"
                    >
                      <option value="USD">USD ($)</option>
                      <option value="EUR">EUR (€)</option>
                      <option value="GBP">GBP (£)</option>
                      <option value="CAD">CAD ($)</option>
                    </select>
                  </div>
                </div>
              </section>

              <!-- Company Preferences -->
              <section class="p-6 bg-black/4 border border-gray-200">
                <h2 class="text-lg font-semibold text-gray-900 mb-4">Company Preferences</h2>

                <div class="flex flex-col gap-2 mb-4">
                  <label class="text-sm font-semibold text-gray-900">Company Stage</label>
                  <div class="flex flex-col gap-2">
                    <div
                      v-for="stage in [
                        { id: 'startup', value: 'Startup', label: 'Startup' },
                        { id: 'growth', value: 'Growth', label: 'Growth Stage' },
                        { id: 'enterprise', value: 'Enterprise', label: 'Enterprise' },
                      ]"
                      :key="stage.id"
                      class="flex items-center gap-2"
                    >
                      <input
                        :id="stage.id"
                        v-model="preferences.company_stage_preference"
                        type="checkbox"
                        :value="stage.value"
                        class="w-4 h-4 cursor-pointer"
                        :style="{ accentColor: 'var(--color-accent)' }"
                      />
                      <label :for="stage.id" class="cursor-pointer text-sm text-gray-900">{{
                        stage.label
                      }}</label>
                    </div>
                  </div>
                </div>

                <div class="flex flex-col gap-2 mb-4">
                  <label class="text-sm font-semibold text-gray-900">Industries of Interest</label>
                  <input
                    v-model="industriesText"
                    type="text"
                    placeholder="e.g., SaaS, AI/ML, FinTech"
                    class="px-3 py-2 border border-gray-200 text-sm bg-white text-gray-900 w-full focus:outline-none focus:border-accent focus:shadow-[0_0_0_1px_var(--color-accent)] transition-all"
                  />
                </div>

                <div class="flex flex-col gap-2">
                  <label class="text-sm font-semibold text-gray-900">Skills</label>
                  <div class="flex flex-wrap gap-2 mb-2">
                    <span
                      v-for="skill in preferences.skills"
                      :key="skill.skill_name"
                      class="flex items-center gap-1 px-2 py-1 bg-white border text-xs font-medium"
                      :style="{ borderColor: 'var(--color-accent)', color: 'var(--color-accent)' }"
                    >
                      {{ skill.skill_name }}
                      <button
                        type="button"
                        class="bg-transparent border-none cursor-pointer text-sm p-0 leading-none"
                        :style="{ color: 'var(--color-accent)' }"
                        @click="removeSkill(skill.skill_name)"
                      >
                        <X />
                      </button>
                    </span>
                  </div>
                  <input
                    type="text"
                    placeholder="Type a skill and press Enter"
                    class="px-3 py-2 border border-gray-200 text-sm bg-white text-gray-900 w-full focus:outline-none focus:border-accent focus:shadow-[0_0_0_1px_var(--color-accent)] transition-all"
                    @keydown.enter.prevent="addSkill($event)"
                  />
                </div>
              </section>
            </div>

            <div class="text-center p-6 bg-black/4 border border-gray-200">
              <button
                class="bg-accent text-accent-foreground px-4 py-2 disabled:opacity-50 disabled:cursor-not-allowed"
                :disabled="isSaving"
                @click="handleSavePreferences"
              >
                {{ isSaving ? 'Saving Changes...' : 'Save Changes' }}
              </button>
              <p v-if="saved && !isSaving" class="mt-3 text-sm font-medium text-green-600">
                Preferences saved successfully
              </p>
            </div>
          </section>

          <!-- Settings -->
          <section v-if="activeTab === 'settings' && preferences" class="flex flex-col gap-6">
            <div
              class="grid gap-6"
              style="grid-template-columns: repeat(auto-fit, minmax(400px, 1fr))"
            >
              <!-- Notifications -->
              <section class="p-6 bg-black/4 border border-gray-200">
                <h2 class="text-lg font-semibold text-gray-900 mb-4">Notifications</h2>
                <div class="flex flex-row justify-between items-center gap-2 mb-4">
                  <label class="text-sm font-semibold text-gray-900">Job Recommendations</label>
                  <button
                    class="px-2 py-1 border text-xs font-semibold cursor-pointer transition-all min-w-12.5"
                    :class="
                      preferences.notify_jobs
                        ? 'text-white border-transparent'
                        : 'bg-white border-gray-200 text-gray-500 hover:border-accent'
                    "
                    :style="
                      preferences.notify_jobs
                        ? { background: 'var(--color-accent)', borderColor: 'var(--color-accent)' }
                        : {}
                    "
                    @click="preferences.notify_jobs = !preferences.notify_jobs"
                  >
                    {{ preferences.notify_jobs ? 'On' : 'Off' }}
                  </button>
                </div>
              </section>

              <!-- Privacy & Account -->
              <section class="p-6 bg-black/4 border border-gray-200">
                <h2 class="text-lg font-semibold text-gray-900 mb-4">Privacy & Account</h2>

                <div class="flex flex-col gap-2 mb-4">
                  <button class="button button-secondary">Download My Data</button>
                </div>

                <div v-if="!isDeleteConfirmVisible" class="flex flex-col gap-2">
                  <button class="button button-secondary" @click="isDeleteConfirmVisible = true">
                    Delete My Account
                  </button>
                </div>

                <div v-else class="flex flex-col gap-2">
                  <p class="text-xs" :style="{ color: 'var(--color-accent)' }">
                    This action is permanent and cannot be undone.
                  </p>
                  <label class="text-sm font-semibold text-gray-900"
                    >Type your email to confirm</label
                  >
                  <input
                    v-model="typedEmail"
                    type="email"
                    placeholder="your@email.com"
                    class="px-4 py-2 border border-gray-200 text-sm bg-white text-gray-900 w-full focus:outline-none focus:border-accent focus:shadow-[0_0_0_1px_var(--color-accent)] transition-all"
                  />
                  <p v-if="deleteError" class="text-xs text-red-600">{{ deleteError }}</p>
                  <div class="flex gap-2 mt-1">
                    <button
                      class="px-4 py-2 bg-gray-400 border-gray-200 button button-secondary"
                      @click="cancelDelete"
                    >
                      Cancel
                    </button>
                    <button
                      class="bg-accent text-white px-3 py-2 disabled:opacity-50"
                      :disabled="typedEmail !== authStore.user?.email"
                      @click="deleteAccount"
                    >
                      Confirm Delete
                    </button>
                  </div>
                </div>
              </section>
            </div>
          </section>
          <div v-else-if="activeTab === 'settings' || activeTab === 'preferences'">
            Loading data...
          </div>
        </div>
      </div>
    </main>

    <!-- Experience Form Modal -->
    <Teleport to="body">
      <Transition name="detail-slide">
        <ExperienceForm
          :open="isEditExperienceOpen"
          :experience="selectedExperience"
          :is-saving="profileStore.isSaving"
          @close="isEditExperienceOpen = false"
          @save="handleSaveExperience"
        />
      </Transition>
    </Teleport>

    <!-- Profile Form Modal -->
    <Teleport to="body">
      <Transition name="detail-slide">
        <EditProfileForm
          v-if="authStore.user"
          :open="isEditProfileOpen"
          :is-saving="authStore.isSaving"
          :user="authStore.user"
          @save="handleSaveProfile"
          @close="isEditProfileOpen = false"
        />
      </Transition>
    </Teleport>

    <!-- Education Form Modal -->
    <Teleport to="body">
      <Transition name="detail-slide">
        <EducationForm
          :open="isEditEducationOpen"
          :education="selectedEducation"
          :is-saving="profileStore.isSaving"
          @close="isEditEducationOpen = false"
          @save="handleSaveEducation"
        />
      </Transition>
    </Teleport>

    <!-- Project Form Modal -->
    <Teleport to="body">
      <Transition name="detail-slide">
        <ProjectForm
          :open="isEditProjectOpen"
          :project="selectedProject"
          :is-saving="profileStore.isSaving"
          @close="isEditProjectOpen = false"
          @save="handleSaveProject"
        />
      </Transition>
    </Teleport>

    <!-- Certification Form Modal -->
    <Teleport to="body">
      <Transition name="detail-slide">
        <CertificationForm
          :open="isEditCertificationOpen"
          :certification="selectedCertification"
          :is-saving="profileStore.isSaving"
          @close="isEditCertificationOpen = false"
          @save="handleSaveCertification"
        />
      </Transition>
    </Teleport>

    <!-- Language Form Modal -->
    <Teleport to="body">
      <Transition name="detail-slide">
        <LanguageForm
          :open="isEditLanguageOpen"
          :lang="selectedLanguage"
          :is-saving="profileStore.isSaving"
          @close="isEditLanguageOpen = false"
          @save="handleSaveLanguage"
        />
      </Transition>
    </Teleport>
  </div>
</template>

<style scoped>
/* Only kept for grid breakpoints — Tailwind can't conditionally swap auto-fit minmax */
@media (max-width: 768px) {
  div[style*='grid-template-columns'] {
    grid-template-columns: 1fr !important;
  }

  .toggle-group {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>
