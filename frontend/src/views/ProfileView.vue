<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import AppSidebar from '@/components/AppSidebar.vue'
import { Plus, SquarePen, ArrowLeft } from '@lucide/vue'
import { useAuthStore } from '@/stores/auth'
import { useRouter } from 'vue-router'
import { useProfileStore } from '@/stores/profile'
import ExperienceCard from '@/components/ExperienceCard.vue'
import ExperienceForm from '@/components/ExperienceForm.vue'
import type { Experience } from '@/types/experience'
import { storeToRefs } from 'pinia'

type ProfileTab = 'overview' | 'experience' | 'preferences' | 'settings'

const router = useRouter()
const authStore = useAuthStore()
const profileStore = useProfileStore()
const { experiences } = storeToRefs(profileStore)

const activeTab = ref<ProfileTab>('overview')
const saved = ref(false)
const savedCount = ref(12)
const isDeleteConfirmVisible = ref(false)
const typedEmail = ref('')
const deleteError = ref('')
const isFormOpen = ref(false)
const selectedExperience = ref<Experience | undefined>(undefined)
const detailOpen = computed(() => isFormOpen.value)
function closeDetail() {
  isFormOpen.value = false
}

const preferences = ref({
  jobTypes: ['Full-time', 'Contract'],
  location: 'Remote',
  salaryMin: 150000,
  salaryMax: 250000,
  experience: '5-10',
  skills: ['React', 'TypeScript', 'Node.js'],
  companyStage: ['Growth', 'Enterprise'],
  industries: 'SaaS, AI/ML, Developer Tools',
  notifyJobs: true,
  visibleToRecruiters: true,
})

const tabs: Array<{ key: ProfileTab; label: string }> = [
  { key: 'overview', label: 'Overview' },
  { key: 'experience', label: 'Experience' },
  { key: 'preferences', label: 'Preferences' },
  { key: 'settings', label: 'Settings' },
]

function openAddForm() {
  selectedExperience.value = undefined
  isFormOpen.value = true
}

function openEditForm(exp: Experience) {
  selectedExperience.value = exp
  isFormOpen.value = true
}

async function handleSaveExperience(payload: Omit<Experience, 'id' | 'user_id' | 'created_at' | 'updated_at'>) {
  if (selectedExperience.value) {
    await profileStore.updateExperience(selectedExperience.value.id, payload as Partial<Experience>)
  } else {
    await profileStore.addExperience(payload)
  }
  isFormOpen.value = false
}

async function handleDeleteExperience(id: string) {
  if (confirm('Are you sure you want to delete this experience?')) {
    await profileStore.deleteExperience(id)
  }
}

const removeSkill = (skill: string) => {
  const idx = preferences.value.skills.indexOf(skill)
  if (idx > -1) {
    preferences.value.skills.splice(idx, 1)
  }
}

const addSkill = (e: Event) => {
  const target = e.target as HTMLInputElement
  if (target.value && !preferences.value.skills.includes(target.value)) {
    preferences.value.skills.push(target.value)
    target.value = ''
  }
}

function savePreferences() {
  saved.value = true
  setTimeout(() => {
    saved.value = false
  }, 2500)
}

// Cancel delete to void multi-statement
function cancelDelete() {
  isDeleteConfirmVisible.value = false
  typedEmail.value = ''
  deleteError.value = ''
}

// Delete account
async function deleteAccount() {
  if (typedEmail.value !== authStore.user?.email) {
    deleteError.value = 'Please enter your email address to confirm account deletion.'
    return
  }

  try {
    // Send a delete request to the backend
    const response = await fetch('/api/users/me', {
      method: 'DELETE',
      credentials: 'include',
    })
    if (!response.ok) {
      deleteError.value = 'Failed to delete account. Please try again.'
      return
    }
    authStore.user = null
    router.push('/login')
  } catch (err) {
    deleteError.value =
      'Something went wrong. Please try again' + (err instanceof Error ? err.message : String(err))
    return
  }
}

onMounted(async () => {
  await profileStore.fetchExperiences()
})
</script>
<template>
  <div class="profile-container">
    <AppSidebar />

    <main class="profile-main">
      <div class="content-area profile-page">
        <div class="profile-max-width">
          <!-- Profile Header -->
          <div class="profile-header">
            <div class="profile-info">
              <div class="profile-avatar">
                <div class="avatar-shell">
                  <img
                    v-if="authStore.user?.avatar_url"
                    :src="authStore.user.avatar_url"
                    alt="Profile avatar"
                  />
                  <span v-else class="avatar-fallback">{{
                    authStore.user?.username?.[0]?.toUpperCase() || 'U'
                  }}</span>
                </div>
              </div>

              <div class="profile-copy">
                <h1 class="profile-name">
                  {{ authStore.user?.full_name || authStore.user?.username }}
                </h1>
                <p class="profile-email">
                  {{ authStore.user?.email }}
                  <span v-if="authStore.user?.username"> • {{ authStore.user.username }}</span>
                </p>
                <p class="profile-bio">
                  {{ authStore.user?.user_summary || 'Add a summary to your profile.' }}
                </p>
              </div>
            </div>

            <button class="button button-primary button-edit">
              <SquarePen :size="16" />
              Edit Profile
            </button>
          </div>

          <div class="profile-stats">
            <div class="stat-item">
              <span class="stat-num">{{ savedCount }}</span>
              <span class="stat-label">Saved jobs</span>
            </div>
            <div class="stat-item">
              <span class="stat-num">4</span>
              <span class="stat-label">Applications</span>
            </div>
            <div class="stat-item">
              <span class="stat-num">{{ experiences?.length || 0 }}</span>
              <span class="stat-label">Experiences</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Tabs-->
      <div class="content-area">
        <!-- Profile Sections -->
        <div class="profile-max-width tab-layout">
          <nav class="tabs" aria-label="Profile sections">
            <button
              v-for="tab in tabs"
              :key="tab.key"
              type="button"
              class="tab-btn"
              :class="{ 'tab-btn--active': activeTab === tab.key }"
              @click="activeTab = tab.key"
            >
              {{ tab.label }}
            </button>
          </nav>

          <!-- Overview -->
          <section v-if="activeTab === 'overview'" class="tab-panel">
            <section class="profile-section">
              <h2 class="section-heading">About</h2>
              <div class="info-grid">
                <div class="info-item">
                  <span class="info-label">Years of experience</span>
                  <span class="info-value">{{
                    authStore.user?.years_of_experience ?? 'Not set'
                  }}</span>
                </div>
                <div class="info-item">
                  <span class="info-label">Location</span>
                  <span class="info-value">{{ authStore.user?.user_location || 'Not set' }}</span>
                </div>
                <div class="info-item">
                  <span class="info-label">Headline</span>
                  <span class="info-value">{{ authStore.user?.headline || 'Not set' }}</span>
                </div>
              </div>
            </section>

            <section class="profile-section">
              <h2 class="section-heading">Activity</h2>
              <div class="info-item">
                <span class="info-label">Saved Jobs</span>
                <span class="info-value">{{ savedCount }} jobs</span>
              </div>
            </section>
          </section>

          <!-- Experience -->
          <section v-if="activeTab === 'experience'" class="tab-panel">
            <section class="profile-section">
              <div class="section-header">
                <h2 class="section-heading">Work history</h2>
                <button class="button button-primary button-add" @click="openAddForm">
                  <Plus /> Add Experience
                </button>
              </div>
              <p v-if="!experiences?.length" class="empty-state">No experience added yet.</p>
              <div v-else class="experience-list">
                <ExperienceCard
                  v-for="exp in experiences"
                  :key="exp.id"
                  :exp="exp"
                  @edit="openEditForm(exp)"
                  @delete="handleDeleteExperience(exp.id)"
                />
              </div>
            </section>
          </section>

          <!-- Preferences Section -->
          <section v-if="activeTab === 'preferences'" class="tab-panel">
            <!-- Header -->
            <div class="page-header">
              <h1 class="page-title">Preferences & Settings</h1>
              <p class="page-subtitle">
                Fine-tune your job recommendations and notification preferences
              </p>
            </div>

            <div class="settings-grid">
              <section class="settings-section">
                <h2 class="section-title">Job Preferences</h2>

                <div class="setting-group">
                  <label class="setting-label">Preferred Job Types</label>
                  <div class="checkbox-group">
                    <div class="checkbox-item">
                      <input
                        id="ft"
                        v-model="preferences.jobTypes"
                        type="checkbox"
                        value="Full-time"
                      />
                      <label for="ft">Full-time</label>
                    </div>
                    <div class="checkbox-item">
                      <input
                        id="pt"
                        v-model="preferences.jobTypes"
                        type="checkbox"
                        value="Part-time"
                      />
                      <label for="pt">Part-time</label>
                    </div>
                    <div class="checkbox-item">
                      <input
                        id="contract"
                        v-model="preferences.jobTypes"
                        type="checkbox"
                        value="Contract"
                      />
                      <label for="contract">Contract</label>
                    </div>
                  </div>
                </div>

                <div class="setting-group">
                  <label class="setting-label">Work Location</label>
                  <select v-model="preferences.location" class="form-input">
                    <option>Remote</option>
                    <option>On-site</option>
                    <option>Hybrid</option>
                    <option>Any</option>
                  </select>
                </div>

                <div class="setting-group salary-inputs">
                  <div class="settings-label">
                    <label class="setting-label">Min Salary</label>
                    <input
                      v-model="preferences.salaryMin"
                      type="number"
                      placeholder="Min"
                      class="form-input"
                    />
                  </div>
                  <div class="settings-label">
                    <label class="setting-label">Max Salary</label>
                    <input
                      v-model="preferences.salaryMax"
                      type="number"
                      placeholder="Max"
                      class="form-input"
                    />
                  </div>
                </div>
              </section>

              <!-- Company Preferences -->
              <section class="settings-section">
                <h2 class="section-title">Company Preferences</h2>

                <div class="setting-group">
                  <label class="setting-label">Company Stage</label>
                  <div class="checkbox-group">
                    <div class="checkbox-item">
                      <input
                        id="startup"
                        v-model="preferences.companyStage"
                        type="checkbox"
                        value="Startup"
                      />
                      <label for="startup">Startup</label>
                    </div>
                    <div class="checkbox-item">
                      <input
                        id="growth"
                        v-model="preferences.companyStage"
                        type="checkbox"
                        value="Growth"
                      />
                      <label for="growth">Growth Stage</label>
                    </div>
                    <div class="checkbox-item">
                      <input
                        id="enterprise"
                        v-model="preferences.companyStage"
                        type="checkbox"
                        value="Enterprise"
                      />
                      <label for="enterprise">Enterprise</label>
                    </div>
                  </div>
                </div>

                <div class="setting-group">
                  <label class="setting-label">Industries of Interest</label>
                  <input
                    v-model="preferences.industries"
                    type="text"
                    placeholder="e.g., SaaS, AI/ML, FinTech"
                    class="form-input"
                  />
                </div>

                <div class="setting-group">
                  <label class="setting-label">Skills</label>
                  <div class="skills-tags">
                    <span v-for="skill in preferences.skills" :key="skill" class="skill-tag">
                      {{ skill }}
                      <button type="button" class="tag-close" @click="removeSkill(skill)">x</button>
                    </span>
                  </div>
                  <input
                    type="text"
                    placeholder="Type a skill and press Enter"
                    class="form-input"
                    @keydown.enter.prevent="addSkill($event)"
                  />
                </div>
              </section>
            </div>

            <div class="save-section">
              <button class="button button-primary" @click="savePreferences">Save Changes</button>
              <p v-if="saved" class="save-message">Preferences saved successfully</p>
            </div>
          </section>

          <!-- Settings -->
          <section v-if="activeTab === 'settings'" class="tab-panel">
            <div class="settings-grid">
              <p>Experiences {{ experiences?.length }}</p>

              <section class="settings-section">
                <h2 class="section-title">Notifications</h2>
                <div class="setting-group toggle-group">
                  <label class="setting-label">Job Recommendations</label>
                  <button
                    :class="['toggle-btn', { 'toggle-active': preferences.notifyJobs }]"
                    @click="preferences.notifyJobs = !preferences.notifyJobs"
                  >
                    {{ preferences.notifyJobs ? 'On' : 'Off' }}
                  </button>
                </div>
              </section>

              <section class="settings-section">
                <h2 class="section-title">Privacy & Account</h2>

                <div class="setting-group">
                  <button class="button button-secondary">Download My Data</button>
                </div>

                <div class="setting-group" v-if="!isDeleteConfirmVisible">
                  <button @click="isDeleteConfirmVisible = true" class="button button-secondary">
                    Delete My Account
                  </button>
                </div>

                <div class="setting-group" v-else>
                  <p style="font-size: var(--text-xs); color: var(--color-accent)">
                    This action is permanent and cannot be undone.
                  </p>
                  <label class="setting-label">Type your email to confirm</label>
                  <input
                    v-model="typedEmail"
                    type="email"
                    placeholder="your@email.com"
                    class="form-input"
                  />
                  <p v-if="deleteError" class="delete-error">{{ deleteError }}</p>
                  <div class="delete-actions">
                    <button class="button button-secondary" @click="cancelDelete">Cancel</button>
                    <button
                      class="button button-primary"
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
        <div v-if="detailOpen" class="detail-overlay" @click.self="closeDetail">
          <div class="detail-drawer">
            <div class="detail-back" @click="closeDetail">
              <ArrowLeft />
              Back to jobs
            </div>
            <ExperienceForm
              :open="isFormOpen"
              :experience="selectedExperience"
              @close="isFormOpen = false"
              @save="handleSaveExperience"
            />
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<style scoped>
.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-4);
}

.button-sm {
  padding: var(--spacing-1) var(--spacing-3);
  font-size: var(--text-xs);
}

.experience-list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-4);
}

.empty-state {
  text-align: center;
  padding: var(--spacing-8);
  color: var(--color-text-muted);
  border: 1px dashed var(--color-border);
}
.profile-container {
  display: grid;
  grid-template-columns: 220px minmax(0, 1fr);
  height: 100%;
  overflow: hidden;
}

.profile-main {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  background: var(--color-bg-primary);
  overflow-y: auto;
  min-height: 0;
}

.content-area {
  margin: 0 auto;
  padding: var(--spacing-8);
  width: 100%;
}

.avatar-shell {
  width: 80px;
  height: 80px;
  border-radius: 50%;
  background: var(--color-bg-secondary);
  border: 1px solid var(--color-border);
  display: grid;
  place-items: center;
  overflow: hidden;
  box-shadow:
    0 0 0 4px var(--color-bg-primary),
    0 0 0 5px var(--color-border);
}

.avatar-shell img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.avatar-fallback {
  color: var(--color-text-primary);
  font-weight: var(--font-bold);
  font-size: var(--text-2xl);
}

.profile-max-width {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-8);
  border-bottom: 1px solid var(--color-border);
  padding-bottom: var(--spacing-6);
}

.profile-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--spacing-6);
}

.profile-info {
  display: flex;
  align-items: flex-start;
  gap: var(--spacing-4);
  flex: 1;
}

.profile-avatar {
  flex-shrink: 0;
}

.profile-name {
  font-size: var(--text-2xl);
  font-weight: var(--font-bold);
  color: var(--color-text-primary);
  margin-bottom: var(--spacing-1);
  letter-spacing: -0.3px;
}

.profile-email {
  font-size: var(--text-sm);
  color: var(--color-text-muted);
  margin-bottom: var(--spacing-3);
}

.profile-bio {
  font-size: var(--text-base);
  color: var(--color-text-muted);
  line-height: 1.5;
}

.profile-stats {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: var(--spacing-4);
  margin-bottom: var(--spacing-2);
}

.stat-item {
  background: var(--color-bg-secondary);
  padding: var(--spacing-4);
  text-align: center;
  display: flex;
  flex-direction: column;
  border: 1px solid var(--color-border);
}

.stat-num {
  font-size: var(--text-3xl);
  font-weight: var(--font-bold);
}

.stat-label {
  font-size: var(--text-sm);
  font-weight: var(--font-normal);
  margin-top: var(--spacing-3);
}

.profile-section {
  padding: var(--spacing-6);
  border: 1px solid var(--color-border);
  margin-bottom: var(--spacing-4);
}

.profile-section:last-child {
  margin-bottom: unset;
}

.section-heading {
  font-size: var(--text-lg);
  font-weight: var(--font-semibold);
  color: var(--color-text-primary);
  margin-bottom: var(--spacing-4);
  letter-spacing: -0.2px;
}

.info-grid {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-4);
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-1);
}

.info-label {
  font-size: var(--text-xs);
  font-weight: var(--font-semibold);
  text-transform: uppercase;
  color: var(--color-text-tertiary);
  letter-spacing: 0.5px;
}

.info-value {
  font-size: var(--text-sm);
  color: var(--color-text-primary);
  font-weight: var(--font-normal);
}

.page-header {
  padding-bottom: var(--spacing-6);
  border-bottom: 1px solid var(--color-border);
}

.page-title {
  font-size: 24px;
  font-weight: var(--font-bold);
  color: var(--color-text-primary);
  margin-bottom: var(--spacing-2);
}

.page-subtitle {
  font-size: var(--text-base);
  color: var(--color-text-muted);
}

.settings-tab {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
  gap: var(--spacing-6);
  margin-bottom: var(--spacing-8);
}

.settings-section {
  padding: var(--spacing-6);
  background: var(--color-bg-secondary);
  border: 1px solid var(--color-border);
}

.section-title {
  font-size: var(--text-lg);
  font-weight: var(--font-semibold);
  color: var(--color-text-primary);
  margin-bottom: var(--spacing-4);
}

.setting-group {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-2);
  margin-bottom: var(--spacing-4);
}

.setting-group:last-child {
  margin-bottom: 0;
}

.setting-label {
  font-size: var(--text-sm);
  font-weight: var(--font-semibold);
  color: var(--color-text-primary);
}

.form-input {
  padding: var(--spacing-2) var(--spacing-3);
  border: 1px solid var(--color-border);
  font-size: var(--text-sm);
  background: var(--color-bg-primary);
  color: var(--color-text-primary);
  transition: all var(--transition-fast);
  width: 100%;
}

.form-input:focus {
  outline: none;
  border-color: var(--color-accent);
  box-shadow: 0 0 0 1px var(--color-accent);
}

.checkbox-group {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-2);
}

.checkbox-item {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
}

.checkbox-item input[type='checkbox'] {
  width: 16px;
  height: 16px;
  cursor: pointer;
  accent-color: var(--color-accent);
}

.checkbox-item label {
  cursor: pointer;
  font-size: var(--text-sm);
  color: var(--color-text-primary);
}

.skills-tags {
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-2);
  margin-bottom: var(--spacing-2);
}

.skill-tag {
  padding: var(--spacing-1) var(--spacing-2);
  background: var(--color-bg-primary);
  border: 1px solid var(--color-accent);
  font-size: var(--text-xs);
  color: var(--color-accent);
  font-weight: var(--font-medium);
  display: flex;
  align-items: center;
  gap: var(--spacing-1);
}

.tag-close {
  background: none;
  border: none;
  color: var(--color-accent);
  cursor: pointer;
  font-size: var(--text-sm);
  padding: 0;
  line-height: 1;
}

.toggle-group {
  flex-direction: row;
  justify-content: space-between;
  align-items: center;
}

.toggle-btn {
  padding: var(--spacing-1) var(--spacing-2);
  background: var(--color-bg-primary);
  border: 1px solid var(--color-border);
  cursor: pointer;
  font-size: var(--text-xs);
  font-weight: var(--font-semibold);
  color: var(--color-text-muted);
  transition: all var(--transition-fast);
  min-width: 50px;
}

.toggle-btn:hover {
  border-color: var(--color-accent);
}

.toggle-active {
  background: var(--color-accent);
  border-color: var(--color-accent);
  color: white;
}

.save-section {
  text-align: center;
  padding: var(--spacing-6);
  background: var(--color-bg-secondary);
  border: 1px solid var(--color-border);
}

.save-message {
  margin-top: var(--spacing-3);
  font-size: var(--text-sm);
  color: var(--color-success);
  font-weight: var(--font-medium);
}

@media (max-width: 768px) {
  .settings-tab {
    grid-template-columns: 1fr;
  }

  .toggle-group {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--spacing-2);
  }
}
</style>
