<script setup lang="ts">
import { routeLocationKey, RouterLink, useRoute } from 'vue-router'
import IconGrid from '@/components/icons/IconGrid.vue'
import { useAuthStore } from '@/stores/auth'
import { AlertCircle, BarChart3, Bookmark, CheckCircle, User, Zap } from 'lucide-vue-next'
import { computed } from 'vue'

const authStore = useAuthStore()
const route = useRoute()

const isSavedActive = computed(() => {
  return route.path === '/jobs' && route.query.filter === 'saved'
})

const isAppliedActive = computed(() => {
  return route.path === '/jobs' && route.query.filter === 'applied'
})
</script>

<template>
  <aside class="sidebar-container">

    <nav class="sidebar-nav">
      <div class="nav-section">
        <h2 class="nav-section-title">Discover</h2>
        <RouterLink to="/jobs" class="nav-item" active-class="active">
          <Zap class="nav-icon" />
          All jobs
        </RouterLink>

        <div class="nav-item nav-disabled">
          <span class="source-dot remoteok" />
          RemoteOK
        </div>
        <div class="nav-item nav-disabled">
          <span class="source-dot adzuna" />
          Adzuna
        </div>
      </div>

      <div class="nav-section">
        <div class="nav-section-title">Account</div>
        <RouterLink to="/dashboard" class="nav-item">
          <IconGrid class="nav-icon" />
          Dashboard
        </RouterLink>
        <RouterLink to="/profile" class="nav-item">
          <User class="nav-icon" />
          Profile
        </RouterLink>
        <RouterLink :to="{
          path: '/jobs',
          query: { filter: 'saved' }
        }" class="nav-item" :class="{ active: isSavedActive }">
          <Bookmark class="nav-icon" />
          Saved
        </RouterLink>
        <RouterLink :to="{
          path: '/jobs',
          query: { filter: 'applied' }
        }" class="nav-item" :class="{ active: isAppliedActive }">
          <CheckCircle class="nav-icon" />
          Applied
        </RouterLink>
      </div>

      <div class="nav-section" v-if="authStore.user">
        <h2 class="nav-section-title">Admin</h2>
        <div>
          <RouterLink to="/admin" class="nav-item">
            <BarChart3 class="nav-icon" />
            Dashboard
          </RouterLink>

          <RouterLink to="/status" class="nav-item">
            <AlertCircle class="nav-icon" />
            System status
          </RouterLink>
        </div>
      </div>
    </nav>
  </aside>
</template>


<style scoped>
.sidebar-container {
  width: 220px;
  min-width: 220px;
  background: var(--color-surface);
  border-right: 1px solid var(--color-border);
  display: flex;
  flex-direction: column;
  padding: 14px 0;
}

.sidebar-nav {
  padding: 0 8px;
}

.nav-item.active {
  background-color: var(--color-accent-lighter);
  color: var(--color-accent);
  font-weight: var(--font-weight-semibold);
  border-left-color: var(--color-accent);
}

.nav-icon {
  width: 15px;
  height: 15px;
  opacity: 0.65;
  flex-shrink: 0;
}

.nav-item.active .nav-icon {
  opacity: 1;
}

.nav-section {
  padding: var(--spacing-3);
  border-top: 1px solid var(--color-border);
  margin: var(--spacing-3) 0;
}

.nav-section-title {
  font-size: var(--text-xs);
  font-weight: var(--font-weight-medium);
  color: var(--muted);
  padding: 0 var(--spacing-4);
  letter-spacing: 0.04em;
  text-transform: uppercase;
}

.source-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  flex-shrink: 0;
}

.source-dot.remoteok {
  background: #5e6ad2;
}

.source-dot.adzuna {
  background: #26c6a6;
}

.nav-disabled {
  opacity: 0.7;
}
</style>
