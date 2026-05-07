<template>
  <div class="mock-view">
    <AppSidebar />
    <main class="mock-main">
      <header class="page-header">
        <h1 class="page-title">System Status</h1>
        <p class="page-subtitle">Real-time health monitoring for JobRadar services.</p>
      </header>
      <div class="status-grid">
        <div class="status-card" v-for="service in services" :key="service.name">
          <div class="status-indicator" :class="service.status"></div>
          <div class="status-info">
            <h3 class="service-name">{{ service.name }}</h3>
            <p class="service-status">{{ service.statusText }}</p>
          </div>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import AppSidebar from '@/components/AppSidebar.vue'

const services = [
  { name: 'API Server', status: 'up', statusText: 'Healthy' },
  { name: 'Database', status: 'up', statusText: 'Healthy' },
  { name: 'Redis Cache', status: 'up', statusText: 'Healthy' },
  { name: 'Scrapers', status: 'warning', statusText: 'Degraded' },
]
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

.status-grid {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-3);
}

.status-card {
  display: flex;
  align-items: center;
  gap: var(--spacing-4);
  background: var(--color-bg-primary);
  padding: var(--spacing-4);
  border: 1px solid var(--color-border);
}

.status-indicator {
  width: 12px;
  height: 12px;
}

.status-indicator.up {
  background: #22c55e;
}
.status-indicator.warning {
  background: #eab308;
}
.status-indicator.down {
  background: #ef4444;
}

.service-name {
  font-size: var(--text-sm);
  font-weight: var(--font-semibold);
}

.service-status {
  font-size: var(--text-xs);
  color: var(--color-text-muted);
}
</style>
