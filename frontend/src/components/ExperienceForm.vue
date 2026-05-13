<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { X, Plus, Trash2 } from '@lucide/vue'
import type { Experience, Skill } from '@/types/experience'

//  Props & emits
const props = defineProps<{
  /** Pass an existing experience to edit, or leave undefined to create */
  experience?: Experience
  /** Whether the form is visible */
  open: boolean
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'save', payload: Omit<Experience, 'id' | 'user_id' | 'created_at' | 'updated_at'>): void
}>()

// Employment types
const employmentTypes = [
  'Full-time',
  'Part-time',
  'Contract',
  'Freelance',
  'Internship',
  'Apprenticeship',
  'Volunteer',
]

// Industries
const industries = [
  'Software / Technology',
  'FinTech',
  'HealthTech',
  'EdTech',
  'SaaS',
  'AI / ML',
  'E-commerce',
  'Media & Entertainment',
  'Consulting',
  'Government',
  'Non-profit',
  'Other',
]

// Form state
function blankForm() {
  return {
    company_name: '',
    company_url: '' as string | null,
    role_title: '',
    exp_location: '' as string | null,
    industry: '' as string | null,
    employment_type: '' as string | null,
    description: '' as string | null,
    achievements: [] as string[],
    start_date: '',
    end_date: '' as string | null,
    is_current: false,
    skills: [] as Skill[],
  }
}

const form = ref(blankForm())

// New achievement being typed
const newAchievement = ref('')

// New skill being typed
const newSkillName = ref('')

// Errors
const errors = ref<Record<string, string>>({})

// Populate form when editing
watch(
  () => props.open,
  (isOpen) => {
    if (isOpen) {
      if (props.experience) {
        form.value = {
          company_name: props.experience.company_name,
          company_url: props.experience.company_url ?? '',
          role_title: props.experience.role_title,
          exp_location: props.experience.exp_location ?? '',
          industry: props.experience.industry ?? '',
          employment_type: props.experience.employment_type ?? '',
          description: props.experience.description ?? '',
          achievements: props.experience.achievements ? [...props.experience.achievements] : [],
          start_date: props.experience.start_date,
          end_date: props.experience.end_date ?? '',
          is_current: props.experience.is_current,
          skills: props.experience.skills ? [...props.experience.skills] : [],
        }
      } else {
        form.value = blankForm()
      }
      errors.value = {}
      newAchievement.value = ''
      newSkillName.value = ''
    }
  },
  { immediate: true },
)

//  Computed
const isEditing = computed(() => !!props.experience)

const endDateDisabled = computed(() => form.value.is_current)

//  Achievements
function addAchievement() {
  const text = newAchievement.value.trim()
  if (!text) return
  form.value.achievements.push(text)
  newAchievement.value = ''
}

function removeAchievement(index: number) {
  form.value.achievements.splice(index, 1)
}

//  Skills
function addSkill() {
  const name = newSkillName.value.trim()
  if (!name) return
  // Avoid duplicates (case-insensitive)
  const already = form.value.skills.some((s) => s.name.toLowerCase() === name.toLowerCase())
  if (already) return
  form.value.skills.push({ id: crypto.randomUUID(), name })
  newSkillName.value = ''
}

function removeSkill(id: string) {
  form.value.skills = form.value.skills.filter((s) => s.id !== id)
}

//  Validation
function validate(): boolean {
  errors.value = {}

  if (!form.value.company_name.trim()) {
    errors.value.company_name = 'Company name is required.'
  }
  if (!form.value.role_title.trim()) {
    errors.value.role_title = 'Role title is required.'
  }
  if (!form.value.start_date) {
    errors.value.start_date = 'Start date is required.'
  }
  if (!form.value.is_current && !form.value.end_date) {
    errors.value.end_date = 'End date is required unless this is your current role.'
  }
  if (
    form.value.start_date &&
    form.value.end_date &&
    !form.value.is_current &&
    form.value.end_date < form.value.start_date
  ) {
    errors.value.end_date = 'End date must be after start date.'
  }

  return Object.keys(errors.value).length === 0
}

//  Submit
function handleSubmit() {
  if (!validate()) return

  emit('save', {
    company_name: form.value.company_name.trim(),
    company_url: form.value.company_url?.trim() || null,
    role_title: form.value.role_title.trim(),
    exp_location: form.value.exp_location?.trim() || null,
    industry: form.value.industry || null,
    employment_type: form.value.employment_type || null,
    description: form.value.description?.trim() || null,
    achievements: form.value.achievements.length ? form.value.achievements : null,
    start_date: form.value.start_date,
    end_date: form.value.is_current ? null : form.value.end_date || null,
    is_current: form.value.is_current,
    skills: form.value.skills,
  })
}

function handleClose() {
  emit('close')
}
</script>

<template>
  <!-- Overlay -->
  <div
    v-if="open"
    class="overlay"
    @click.self="handleClose"
    role="dialog"
    aria-modal="true"
    :aria-label="isEditing ? 'Edit experience' : 'Add experience'"
  >
    <div v-if="open" class="panel">
      <!-- Header -->
      <div class="panel-header">
        <h2 class="panel-title">{{ isEditing ? 'Edit experience' : 'Add experience' }}</h2>
        <button class="close-btn" type="button" @click="handleClose" aria-label="Close">
          <X :size="18" />
        </button>
      </div>

      <!-- Form -->
      <form class="panel-body" @submit.prevent="handleSubmit" novalidate>
        <!--  Row: Company + URL  -->
        <div class="form-row">
          <div class="field" :class="{ 'field--error': errors.company_name }">
            <label class="field-label" for="company_name">
              Company name <span class="required">*</span>
            </label>
            <input
              id="company_name"
              v-model="form.company_name"
              type="text"
              class="form-input"
              placeholder="e.g. Andela"
              autocomplete="organization"
            />
            <p v-if="errors.company_name" class="field-error">{{ errors.company_name }}</p>
          </div>

          <div class="field">
            <label class="field-label" for="company_url">Company website</label>
            <input
              id="company_url"
              v-model="form.company_url"
              type="url"
              class="form-input"
              placeholder="https://company.com"
            />
          </div>
        </div>

        <!--  Row: Role + Employment type  -->
        <div class="form-row">
          <div class="field" :class="{ 'field--error': errors.role_title }">
            <label class="field-label" for="role_title">
              Role title <span class="required">*</span>
            </label>
            <input
              id="role_title"
              v-model="form.role_title"
              type="text"
              class="form-input"
              placeholder="e.g. Senior Engineer"
            />
            <p v-if="errors.role_title" class="field-error">{{ errors.role_title }}</p>
          </div>

          <div class="field">
            <label class="field-label" for="employment_type">Employment type</label>
            <select id="employment_type" v-model="form.employment_type" class="form-input">
              <option value="">Select type</option>
              <option v-for="type in employmentTypes" :key="type" :value="type">
                {{ type }}
              </option>
            </select>
          </div>
        </div>

        <!--  Row: Location + Industry  -->
        <div class="form-row">
          <div class="field">
            <label class="field-label" for="exp_location">Location</label>
            <input
              id="exp_location"
              v-model="form.exp_location"
              type="text"
              class="form-input"
              placeholder="e.g. Kigali, Rwanda or Remote"
            />
          </div>

          <div class="field">
            <label class="field-label" for="industry">Industry</label>
            <select id="industry" v-model="form.industry" class="form-input">
              <option value="">Select industry</option>
              <option v-for="ind in industries" :key="ind" :value="ind">{{ ind }}</option>
            </select>
          </div>
        </div>

        <!--  Row: Start + End dates  -->
        <div class="form-row">
          <div class="field" :class="{ 'field--error': errors.start_date }">
            <label class="field-label" for="start_date">
              Start date <span class="required">*</span>
            </label>
            <input id="start_date" v-model="form.start_date" type="month" class="form-input" />
            <p v-if="errors.start_date" class="field-error">{{ errors.start_date }}</p>
          </div>

          <div class="field" :class="{ 'field--error': errors.end_date }">
            <label class="field-label" for="end_date">End date</label>
            <input
              id="end_date"
              v-model="form.end_date"
              type="month"
              class="form-input"
              :disabled="endDateDisabled"
            />
            <p v-if="errors.end_date" class="field-error">{{ errors.end_date }}</p>
          </div>
        </div>

        <!--  Current role checkbox  -->
        <div class="field checkbox-field">
          <input id="is_current" v-model="form.is_current" type="checkbox" class="checkbox" />
          <label for="is_current" class="checkbox-label"> I currently work here </label>
        </div>

        <!--  Description  -->
        <div class="field">
          <label class="field-label" for="description">Description</label>
          <textarea
            id="description"
            v-model="form.description"
            class="form-input form-textarea"
            placeholder="Describe your role, responsibilities, and impact..."
            rows="4"
          />
        </div>

        <!--  Achievements  -->
        <div class="field">
          <label class="field-label">Key achievements</label>
          <p class="field-hint">
            Add bullet points that highlight impact — think metrics, outcomes, and deliverables.
          </p>

          <ul v-if="form.achievements.length" class="achievement-list" role="list">
            <li v-for="(item, idx) in form.achievements" :key="idx" class="achievement-item">
              <span class="achievement-bullet">•</span>
              <span class="achievement-text">{{ item }}</span>
              <button
                type="button"
                class="achievement-remove"
                @click="removeAchievement(idx)"
                :aria-label="`Remove achievement: ${item}`"
              >
                <Trash2 :size="14" />
              </button>
            </li>
          </ul>

          <div class="achievement-input-row">
            <input
              v-model="newAchievement"
              type="text"
              class="form-input"
              placeholder="e.g. Reduced API latency by 40% by optimising query plans"
              @keydown.enter.prevent="addAchievement"
            />
            <button
              type="button"
              class="add-btn"
              @click="addAchievement"
              :disabled="!newAchievement.trim()"
            >
              <Plus :size="16" />
              Add
            </button>
          </div>
        </div>

        <!--  Skills  -->
        <div class="field">
          <label class="field-label">Skills used</label>
          <p class="field-hint">Tag the technologies and skills relevant to this role.</p>

          <div v-if="form.skills.length" class="skill-tags">
            <span v-for="skill in form.skills" :key="skill.id" class="skill-tag">
              {{ skill.name }}
              <button
                type="button"
                class="tag-remove"
                @click="removeSkill(skill.id)"
                :aria-label="`Remove skill ${skill.name}`"
              >
                ×
              </button>
            </span>
          </div>

          <div class="skill-input-row">
            <input
              v-model="newSkillName"
              type="text"
              class="form-input"
              placeholder="e.g. Go, PostgreSQL, Docker"
              @keydown.enter.prevent="addSkill"
            />
            <button
              type="button"
              class="add-btn"
              @click="addSkill"
              :disabled="!newSkillName.trim()"
            >
              <Plus :size="16" />
              Add
            </button>
          </div>
        </div>

        <!--  Footer actions  -->
        <div class="panel-footer">
          <button type="button" class="button button-secondary" @click="handleClose">Cancel</button>
          <button type="submit" class="button button-primary">
            {{ isEditing ? 'Save changes' : 'Add experience' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<style scoped>
/*  Overlay  */
.overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.45);
  display: flex;
  align-items: flex-start;
  justify-content: flex-end;
  z-index: 100;
}

/*  Slide-in panel (right drawer style)  */
.panel {
  width: 100%;
  max-width: 560px;
  height: 100dvh;
  background: var(--color-bg-primary);
  border-left: 1px solid var(--color-border);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--spacing-5) var(--spacing-6);
  border-bottom: 1px solid var(--color-border);
  flex-shrink: 0;
}

.panel-title {
  font-size: var(--text-lg);
  font-weight: var(--font-semibold);
  color: var(--color-text-primary);
}

.close-btn {
  display: grid;
  place-items: center;
  width: 32px;
  height: 32px;
  background: none;
  border: 1px solid var(--color-border);
  color: var(--color-text-muted);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.close-btn:hover {
  color: var(--color-text-primary);
  border-color: var(--color-text-muted);
}

/*  Scrollable body  */
.panel-body {
  flex: 1;
  overflow-y: auto;
  padding: var(--spacing-6);
  display: flex;
  flex-direction: column;
  gap: var(--spacing-5);
}

/*  Form rows (two columns)  */
.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--spacing-4);
}

/*  Fields  */
.field {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-2);
}

.field-label {
  font-size: var(--text-sm);
  font-weight: var(--font-semibold);
  color: var(--color-text-primary);
}

.required {
  color: var(--color-error, #b42318);
  margin-left: 2px;
}

.field-hint {
  font-size: var(--text-xs);
  color: var(--color-text-muted);
  margin-top: calc(-1 * var(--spacing-1));
}

.field-error {
  font-size: var(--text-xs);
  color: var(--color-error, #b42318);
}

.field--error .form-input {
  border-color: var(--color-error, #b42318);
}

/*  Inputs  */
.form-input {
  padding: var(--spacing-2) var(--spacing-3);
  border: 1px solid var(--color-border);
  font-size: var(--text-sm);
  background: var(--color-bg-primary);
  color: var(--color-text-primary);
  transition:
    border-color var(--transition-fast),
    box-shadow var(--transition-fast);
  width: 100%;
  font-family: inherit;
}

.form-input:focus {
  outline: none;
  border-color: var(--color-accent);
  box-shadow: 0 0 0 1px var(--color-accent);
}

.form-input:disabled {
  opacity: 0.45;
  cursor: not-allowed;
  background: var(--color-bg-secondary);
}

.form-textarea {
  resize: vertical;
  min-height: 96px;
  line-height: 1.6;
}

/*  Checkbox  */
.checkbox-field {
  flex-direction: row;
  align-items: center;
  gap: var(--spacing-2);
}

.checkbox {
  width: 16px;
  height: 16px;
  accent-color: var(--color-accent);
  cursor: pointer;
  flex-shrink: 0;
}

.checkbox-label {
  font-size: var(--text-sm);
  color: var(--color-text-primary);
  cursor: pointer;
}

/*  Achievements list  */
.achievement-list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-2);
  list-style: none;
  padding: 0;
  margin-bottom: var(--spacing-2);
}

.achievement-item {
  display: flex;
  align-items: flex-start;
  gap: var(--spacing-2);
  padding: var(--spacing-2) var(--spacing-3);
  background: var(--color-bg-secondary);
  border: 1px solid var(--color-border);
}

.achievement-bullet {
  color: var(--color-accent);
  font-weight: var(--font-bold);
  flex-shrink: 0;
  line-height: 1.5;
}

.achievement-text {
  flex: 1;
  font-size: var(--text-sm);
  color: var(--color-text-primary);
  line-height: 1.5;
}

.achievement-remove {
  background: none;
  border: none;
  color: var(--color-text-muted);
  cursor: pointer;
  display: grid;
  place-items: center;
  padding: var(--spacing-1);
  flex-shrink: 0;
  transition: color var(--transition-fast);
}

.achievement-remove:hover {
  color: var(--color-error, #b42318);
}

.achievement-input-row {
  display: flex;
  gap: var(--spacing-2);
}

/*  Skills  */
.skill-tags {
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-2);
  margin-bottom: var(--spacing-2);
}

.skill-tag {
  display: inline-flex;
  align-items: center;
  gap: var(--spacing-1);
  padding: var(--spacing-1) var(--spacing-2);
  background: var(--color-bg-primary);
  border: 1px solid var(--color-accent);
  font-size: var(--text-xs);
  font-weight: var(--font-medium);
  color: var(--color-accent);
}

.tag-remove {
  background: none;
  border: none;
  color: var(--color-accent);
  cursor: pointer;
  font-size: var(--text-sm);
  line-height: 1;
  padding: 0;
}

.skill-input-row {
  display: flex;
  gap: var(--spacing-2);
}

/*  Add button  */
.add-btn {
  display: inline-flex;
  align-items: center;
  gap: var(--spacing-1);
  padding: var(--spacing-2) var(--spacing-3);
  background: var(--color-bg-secondary);
  border: 1px solid var(--color-border);
  font-size: var(--text-sm);
  font-weight: var(--font-semibold);
  color: var(--color-text-primary);
  cursor: pointer;
  white-space: nowrap;
  transition: all var(--transition-fast);
  flex-shrink: 0;
}

.add-btn:hover:not(:disabled) {
  border-color: var(--color-accent);
  color: var(--color-accent);
}

.add-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

/*  Footer  */
.panel-footer {
  display: flex;
  justify-content: flex-end;
  gap: var(--spacing-3);
  padding-top: var(--spacing-4);
  border-top: 1px solid var(--color-border);
  margin-top: auto;
}

.button {
  display: inline-flex;
  align-items: center;
  gap: var(--spacing-2);
  padding: var(--spacing-2) var(--spacing-5);
  border: 1px solid transparent;
  font-size: var(--text-sm);
  font-weight: var(--font-semibold);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.button-primary {
  background: var(--color-accent);
  color: white;
  border-color: var(--color-accent);
}

.button-secondary {
  background: var(--color-bg-primary);
  color: var(--color-text-primary);
  border-color: var(--color-border);
}

/*  Transitions  */
.overlay-enter-active,
.overlay-leave-active {
  transition: opacity 180ms ease;
}

.overlay-enter-from,
.overlay-leave-to {
  opacity: 0;
}

.panel-enter-active,
.panel-leave-active {
  transition: transform 220ms cubic-bezier(0.16, 1, 0.3, 1);
}

.panel-enter-from,
.panel-leave-to {
  transform: translateX(100%);
}

/*  Responsive  */
@media (max-width: 640px) {
  .overlay {
    align-items: flex-end;
    justify-content: stretch;
  }

  .panel {
    max-width: 100%;
    height: 92dvh;
    border-left: none;
    border-top: 1px solid var(--color-border);
  }

  .form-row {
    grid-template-columns: 1fr;
  }
}
</style>
