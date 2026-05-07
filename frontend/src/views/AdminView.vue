<template>
  <div class="mock-view">
    <AppSidebar />
    <main class="mock-main">
      <header class="page-header">
        <h1 class="page-title">Admin Console</h1>
        <p class="page-subtitle">Administrative tools and user management.</p>
      </header>
      <div class="mock-content">
        <div class="placeholder-card" v-for="i in 4" :key="i"></div>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import AppSidebar from '@/components/AppSidebar.vue'
</script>

<style scoped>
.mock-view {
  display: flex;
  height: 100%;
  background: var(--color-bg-secondary);
}

.mock-main {
  flex: 1;
  padding: var(--spacing-8);
  display: flex;
  flex-direction: column;
  gap: var(--spacing-6);
  overflow-y: auto;
}

.page-header {
  border-bottom: 1px solid var(--color-border);
  padding-bottom: var(--spacing-4);
}

.page-title {
  font-size: var(--text-2xl);
  font-weight: var(--font-bold);
}

.mock-content {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: var(--spacing-4);
}

.placeholder-card {
  height: 150px;
  background: var(--color-bg-primary);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
}
</style>
