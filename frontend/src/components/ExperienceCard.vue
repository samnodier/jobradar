<script setup lang="ts">
import type { Experience } from '@/types/experience'

const props = defineProps<{ exp: Experience }>()
</script>
<template>
  <h2 class="section-title">Skills & Experience</h2>

  <div class="setting-group">
    <label class="setting-label">Years of Experience</label>
    <select v-model="props.exp" class="form-input">
      <option value="0-2">0-2 years</option>
      <option value="2-5">2-5 years</option>
      <option value="5-10">5-10 years</option>
      <option value="10+">10+ years</option>
    </select>
  </div>

  <div class="setting-group">
    <label class="setting-label">Primary Skills</label>
    <div class="skills-tags">
      <span v-for="skill in experiences.skills" :key="skill" class="skill-tag">
        {{ skill }}
        <button @click="removeSkill(skill)" class="tag-close">×</button>
      </span>
    </div>
    <input
      type="text"
      placeholder="Add a skill and press Enter"
      @keydown.enter="addSkill"
      class="form-input"
    />
  </div>
</template>
