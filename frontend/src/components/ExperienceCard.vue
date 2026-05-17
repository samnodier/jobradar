<script setup lang="ts">
import type { Experience } from '@/types/experience'
import { Pencil, Trash2 } from '@lucide/vue'
import { computed } from 'vue'

const props = defineProps<{ exp: Experience }>()
const emit = defineEmits<{
  (e: 'edit'): void
  (e: 'delete'): void
}>()

const experience = computed(() => props.exp)

const formatDate = (dateStr: string | null) => {
  if (!dateStr) return ''
  // If it's already YYYY-MM, return as is. If YYYY-MM-DD, take first 7 chars.
  return dateStr.length > 7 ? dateStr.substring(0, 7) : dateStr
}
</script>

<template>
  <div class="experience-card">
    <div class="card-header">
      <div class="header-left">
        <h3 class="role-title">{{ experience.role_title }}</h3>
        <p class="company-line">
          <span class="company-name">{{ experience.company_name }}</span>
          <span v-if="experience.employment_type" class="dot-separator">•</span>
          <span v-if="experience.employment_type" class="employment-type">{{
            experience.employment_type
          }}</span>
        </p>
      </div>
      <div class="header-actions">
        <button class="action-btn edit-btn" @click="emit('edit')" aria-label="Edit experience">
          <Pencil :size="14" />
        </button>
        <button
          class="action-btn delete-btn"
          @click="emit('delete')"
          aria-label="Delete experience"
        >
          <Trash2 :size="14" />
        </button>
      </div>
    </div>

    <div class="card-meta">
      <span class="date-range">
        {{ formatDate(experience.start_date) }} —
        {{ experience.is_current ? 'Present' : formatDate(experience.end_date) }}
      </span>
      <span v-if="experience.exp_location" class="dot-separator">•</span>
      <span v-if="experience.exp_location" class="location">{{ experience.exp_location }}</span>
    </div>

    <p v-if="experience.description" class="description">
      {{ experience.description }}
    </p>

    <ul v-if="experience.achievements?.length" class="achievements-list">
      <li v-for="(achievement, index) in experience.achievements" :key="index">
        {{ achievement }}
      </li>
    </ul>

    <div v-if="experience.skills?.length" class="skills-wrap">
      <span v-for="skill in experience.skills" :key="skill.id" class="skill-tag">
        {{ skill.name }}
      </span>
    </div>
  </div>
</template>

<style scoped>
.experience-card {
  padding: var(--spacing-5);
  background: var(--color-bg-primary);
  border: 1px solid var(--color-border);
  transition: border-color var(--transition-fast);
}

.experience-card:hover {
  border-color: var(--color-accent);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: var(--spacing-2);
}

.role-title {
  font-size: var(--text-base);
  font-weight: var(--font-semibold);
  color: var(--color-text-primary);
  letter-spacing: -0.1px;
}

.company-line {
  font-size: var(--text-sm);
  color: var(--color-text-muted);
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
}

.company-name {
  color: var(--color-text-primary);
  font-weight: var(--font-medium);
}

.header-actions {
  display: flex;
  gap: var(--spacing-2);
}

.action-btn {
  background: none;
  border: 1px solid var(--color-border);
  color: var(--color-text-muted);
  padding: var(--spacing-1);
  cursor: pointer;
  display: grid;
  place-items: center;
  transition: all var(--transition-fast);
}

.edit-btn:hover {
  color: var(--color-accent);
  border-color: var(--color-accent);
}

.delete-btn:hover {
  color: var(--color-error, #b42318);
  border-color: var(--color-error, #b42318);
}

.card-meta {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  font-size: var(--text-xs);
  color: var(--color-text-tertiary);
  margin-bottom: var(--spacing-4);
}

.dot-separator {
  color: var(--color-border);
}

.description {
  font-size: var(--text-sm);
  line-height: 1.6;
  color: var(--color-text-muted);
  margin-bottom: var(--spacing-4);
  white-space: pre-wrap;
}

.achievements-list {
  padding-left: var(--spacing-4);
  margin-bottom: var(--spacing-4);
}

.achievements-list li {
  font-size: var(--text-sm);
  color: var(--color-text-muted);
  margin-bottom: var(--spacing-1);
  line-height: 1.5;
}

.skills-wrap {
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-2);
}

.skill-tag {
  font-size: 11px;
  padding: 2px 8px;
  background: var(--color-bg-secondary);
  border: 1px solid var(--color-border);
  color: var(--color-text-muted);
  font-weight: var(--font-medium);
}
</style>
