<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { computed } from 'vue'
import { X } from '@lucide/vue'
import { useToast } from '@/composables/useToast'
import type { DesiredRolePreference, SkillPreference } from '@/types/preferences'
import { usePreferencesStore } from '@/stores/preferences'

const preferencesStore = usePreferencesStore()
const { preferences } = storeToRefs(preferencesStore)
const toast = useToast()
const props = defineProps<{ activeTab: string }>()

let debounceTimer: ReturnType<typeof setTimeout> | null = null

const removeIndustry = (name: string) => {
  if (preferences?.value && preferences.value.preferred_industries) {
    preferences.value.preferred_industries = preferences.value.preferred_industries.filter(
      (i: string) => i !== name,
    )
  }
  handleSavePreferences()
}

const addIndustry = (e: Event) => {
  const target = e.target as HTMLInputElement
  const val = target.value.trim()
  if (val && preferences?.value) {
    if (!preferences.value.preferred_industries) {
      preferences.value.preferred_industries = []
    }
    const exists = preferences.value.preferred_industries.some(
      (i: string) => i.toLowerCase() === val.toLowerCase(),
    )
    if (!exists) {
      preferences.value.preferred_industries.push(val)
    }
    target.value = ''
    handleSavePreferences()
  }
}

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
  handleSavePreferences()
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
    handleSavePreferences()
  }
}

const removeSkill = (skillName: string) => {
  if (preferences?.value && preferences.value.skills) {
    preferences.value.skills = preferences.value.skills.filter(
      (s: SkillPreference) => s.skill_name !== skillName,
    )
  }
  handleSavePreferences()
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
    handleSavePreferences()
  }
}

function handleSalarySave() {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    handleSavePreferences()
  }, 800)
}

async function handleSavePreferences() {
  if (!preferences?.value) return
  try {
    await preferencesStore.updatePreferences(preferences.value)
    toast.success('Preferences saved.')
  } catch (err) {
    toast.error(err instanceof Error ? err.message : 'Failed to save preferences.')
  }
}
</script>

<template>
  <section v-if="activeTab === props.activeTab && preferences" class="flex flex-col gap-6">
    <div class="grid gap-6" style="grid-template-columns: repeat(auto-fit, minmax(400px, 1fr))">
      <!-- Job Preferences -->
      <section class="p-6 bg-white border border-gray-200">
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
                @change="handleSavePreferences"
                type="checkbox"
                :value="type"
                class="w-4 h-4 cursor-pointer"
                :style="{ accentColor: 'var(--color-accent)' }"
              />
              <label :for="type" class="cursor-pointer text-sm text-gray-900">{{ type }}</label>
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
            @change="handleSavePreferences"
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
              @input="handleSalarySave"
              type="number"
              placeholder="Min"
              class="px-3 py-2 border border-gray-200 text-sm bg-white text-gray-900 w-full focus:outline-none focus:border-accent focus:shadow-[0_0_0_1px_var(--color-accent)] transition-all"
            />
          </div>
          <div class="flex flex-col gap-2 flex-1">
            <label class="text-sm font-semibold text-gray-900">Max Salary</label>
            <input
              v-model="preferences.max_salary"
              @input="handleSalarySave"
              type="number"
              placeholder="Max"
              class="px-3 py-2 border border-gray-200 text-sm bg-white text-gray-900 w-full focus:outline-none focus:border-accent focus:shadow-[0_0_0_1px_var(--color-accent)] transition-all"
            />
          </div>
          <div class="flex flex-col gap-2 w-28">
            <label class="text-sm font-semibold text-gray-900">Currency</label>

            <select
              v-model="preferences.salary_currency"
              @change="handleSavePreferences"
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
      <section class="p-6 bg-white border border-gray-200">
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
                @change="handleSavePreferences"
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
          <div class="flex flex-wrap gap-2 mb-2">
            <span
              v-for="ind in preferences.preferred_industries"
              :key="ind"
              class="flex items-center gap-2 px-2 py-1 bg-white border text-xs font-medium"
            >
              {{ ind }}
              <button
                type="button"
                class="bg-transparent border-none cursor-pointer text-sm p-0 leading-none"
                @click="removeIndustry(ind)"
              >
                <X :size="12" />
              </button>
            </span>
          </div>
          <input
            type="text"
            placeholder="Type an industry (e.g. SaaS) and press Enter"
            class="px-3 py-2 border border-gray-200 text-sm bg-white text-gray-900 w-full focus:outline-none focus:border-accent focus:shadow-[0_0_0_1px_var(--color-accent)] transition-all"
            @keydown.enter.prevent="addIndustry($event)"
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
  </section>
</template>
