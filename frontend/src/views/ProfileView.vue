<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import AppSidebar from '@/components/AppSidebar.vue'
import { Plus, SquarePen, X } from '@lucide/vue'
import { useAuthStore } from '@/stores/auth'
import { useRouter } from 'vue-router'
import { useProfileStore } from '@/stores/profile'
import ExperienceCard from '@/components/ExperienceCard.vue'
import ExperienceForm from '@/components/ExperienceForm.vue'
import EditProfileForm from '@/components/EditProfileForm.vue'
import type { Experience } from '@/types/experience'
import type { User } from '@/types/user'
import { storeToRefs } from 'pinia'
import { usePreferencesStore } from '@/stores/preferences'
import type { SkillPreference } from '@/types/preferences'
import { useToast } from '@/composables/useToast'

type ProfileTab = 'overview' | 'experience' | 'preferences' | 'settings'

const router = useRouter()
const authStore = useAuthStore()
const profileStore = useProfileStore()
const { experiences } = storeToRefs(profileStore)
const toast = useToast()

const activeTab = ref<ProfileTab>('overview')
const saved = ref(false)
const savedCount = ref(12)
const isDeleteConfirmVisible = ref(false)
const typedEmail = ref('')
const deleteError = ref('')
const isEditExperienceOpen = ref(false)
const isEditProfileOpen = ref(false)
const selectedExperience = ref<Experience | undefined>(undefined)
const profile = ref<User | undefined>(undefined)

const preferencesStore = usePreferencesStore
const { preferences, loading, error, isSaving } = storeToRefs(preferencesStore)

const tabs: Array<{ key: ProfileTab; label: string }> = [
  { key: 'overview', label: 'Overview' },
  { key: 'experience', label: 'Experience' },
  { key: 'preferences', label: 'Preferences' },
  { key: 'settings', label: 'Settings' },
]

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
    if (first.location_name.toLowerCase() === 'remote') return 'Remote'
    if (first.is_remote) return 'Hybrid'
    return 'On-Site'
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
        { location_name: 'Hybrid', is_remote: false, priority: 1 },
      ]
    }
  },
})

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
  if (val && preferences?.value && preferences.value.skills) {
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

function openExperienceAddForm() {
  selectedExperience.value = undefined
  isEditExperienceOpen.value = true
}

function openExperienceEditForm(exp: Experience) {
  selectedExperience.value = exp
  isEditExperienceOpen.value = true
}

function openProfileEditForm(user: User) {
  profile.value = user
  isEditProfileOpen.value = true
}

async function handleSaveExperience(
  payload: Omit<Experience, 'id' | 'user_id' | 'created_at' | 'updated_at'>,
) {
  if (selectedExperience.value) {
    await profileStore.updateExperience(selectedExperience.value.id, payload as Partial<Experience>)
  } else {
    await profileStore.addExperience(payload)
  }
  isEditExperienceOpen.value = false
}

async function handleDeleteExperience(id: string) {
  if (confirm('Are you sure you want to delete this experience?')) {
    await profileStore.deleteExperience(id)
  }
}

async function savePreferences() {
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

onMounted(async () => {
  await profileStore.fetchExperiences()
  await preferencesStore.fetchPreferences()
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
                class="button button-primary"
                @click="authStore.user && openProfileEditForm(authStore.user)"
              >
                <SquarePen :size="16" />
                Edit Profile
              </button>
              <button class="button button-secondary">Export To PDF</button>
            </div>
          </div>

          <!-- Stats -->
          <div class="grid grid-cols-3 gap-4 mb-2">
            <div
              v-for="stat in [
                { num: savedCount, label: 'Saved jobs' },
                { num: 4, label: 'Applications' },
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
          <section v-if="activeTab === 'overview'" class="flex flex-col">
            <div class="p-6 border border-gray-200 mb-4 last:mb-0">
              <h2 class="text-lg font-semibold text-gray-900 mb-4 tracking-tight">About</h2>
              <div class="flex flex-col gap-4">
                <div
                  v-for="item in [
                    {
                      label: 'Years of experience',
                      value: authStore.user?.years_of_experience ?? 'Not set',
                    },
                    { label: 'Location', value: authStore.user?.user_location || 'Not set' },
                    { label: 'Headline', value: authStore.user?.headline || 'Not set' },
                  ]"
                  :key="item.label"
                  class="flex flex-col gap-1"
                >
                  <span class="text-xs font-semibold uppercase text-gray-400 tracking-wide">{{
                    item.label
                  }}</span>
                  <span class="text-sm text-gray-900">{{ item.value }}</span>
                </div>
              </div>
            </div>

            <div class="p-6 border border-gray-200">
              <h2 class="text-lg font-semibold text-gray-900 mb-4 tracking-tight">Activity</h2>
              <div class="flex flex-col gap-1">
                <span class="text-xs font-semibold uppercase text-gray-400 tracking-wide"
                  >Saved Jobs</span
                >
                <span class="text-sm text-gray-900">{{ savedCount }} jobs</span>
              </div>
            </div>
          </section>

          <!-- Experience -->
          <section v-if="activeTab === 'experience'" class="flex flex-col">
            <div class="p-6 border border-gray-200">
              <div class="flex justify-between items-center mb-4">
                <h2 class="text-lg font-semibold text-gray-900 tracking-tight">Work history</h2>
                <button class="button button-primary" @click="openExperienceAddForm">
                  <Plus /> Add Experience
                </button>
              </div>
              <p
                v-if="!experiences?.length"
                class="text-center py-8 text-gray-400 border border-dashed border-gray-200"
              >
                No experience added yet.
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
          </section>

          <!-- Preferences -->
          <section v-if="activeTab === 'preferences'" class="flex flex-col gap-6">
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
                        v-model="preferences.company_stage_preferences"
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
              <button class="button button-primary" @click="savePreferences">Save Changes</button>
              <p v-if="saved" class="mt-3 text-sm font-medium text-green-600">
                Preferences saved successfully
              </p>
            </div>
          </section>

          <!-- Settings -->
          <section v-if="activeTab === 'settings'" class="flex flex-col gap-6">
            <div
              class="grid gap-6"
              style="grid-template-columns: repeat(auto-fit, minmax(400px, 1fr))"
            >
              <p class="text-sm text-gray-500">Experiences {{ experiences?.length }}</p>

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
                    {{ preferences.notifyJobs ? 'On' : 'Off' }}
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
                    class="px-3 py-2 border border-gray-200 text-sm bg-white text-gray-900 w-full focus:outline-none focus:border-accent focus:shadow-[0_0_0_1px_var(--color-accent)] transition-all"
                  />
                  <p v-if="deleteError" class="text-xs text-red-600">{{ deleteError }}</p>
                  <div class="flex gap-2 mt-1">
                    <button class="button button-secondary" @click="cancelDelete">Cancel</button>
                    <button
                      class="button button-primary disabled:opacity-50"
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
        </div>
      </div>
    </main>

    <Teleport to="body">
      <Transition name="detail-slide">
        <ExperienceForm
          :open="isEditExperienceOpen"
          :experience="selectedExperience"
          @close="isEditExperienceOpen = false"
          @save="handleSaveExperience"
        />
      </Transition>
    </Teleport>
    <Teleport to="body">
      <Transition name="detail-slide">
        <EditProfileForm :open="isEditProfileOpen" @close="isEditProfileOpen = false" />
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
