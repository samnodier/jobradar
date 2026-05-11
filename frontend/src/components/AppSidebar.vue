<script setup lang="ts">
import { RouterLink } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import {
  AlertCircle,
  BarChart3,
  Bookmark,
  CheckCircle,
  LayoutDashboard,
  User,
  Zap,
} from '@lucide/vue'

const authStore = useAuthStore()
</script>

<template>
  <aside class="sidebar-container">
    <nav class="sidebar-nav">
      <div class="nav-section">
        <h2 class="nav-section-title">Discovery</h2>
        <RouterLink to="/jobs" class="nav-item">
          <Zap class="nav-icon" />
          Browse Roles
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
        <div class="nav-section-title">Workspace</div>
        <RouterLink to="/dashboard" class="nav-item">
          <LayoutDashboard class="nav-icon" />
          Overview
        </RouterLink>
        <RouterLink
          :to="{
            path: '/jobs',
            query: { filter: 'saved' },
          }"
          class="nav-item"
        >
          <Bookmark class="nav-icon" />
          Saved Jobs
        </RouterLink>
        <RouterLink to="/applications" class="nav-item">
          <CheckCircle class="nav-icon" />
          Tracker
        </RouterLink>
      </div>

      <div class="nav-section" v-if="authStore.user">
        <div class="nav-section-title">Personal</div>
        <RouterLink :to="'/@' + authStore.user?.username" class="nav-item">
          <User class="nav-icon" />
          Profile
        </RouterLink>
      </div>

      <div class="nav-section" v-if="authStore.user">
        <h2 class="nav-section-title">System</h2>
        <RouterLink to="/admin" class="nav-item">
          <BarChart3 class="nav-icon" />
          Console
        </RouterLink>

        <RouterLink to="/status" class="nav-item">
          <AlertCircle class="nav-icon" />
          Health
        </RouterLink>
      </div>
    </nav>
  </aside>
</template>

<style scoped>
.sidebar-container {
  width: 220px;
  min-width: 220px;
  height: 100%;
  background: var(--color-bg-primary);
  border-right: 1px solid var(--color-border);
  display: flex;
  flex-direction: column;
  padding: 14px 0;
  overflow-y: auto;
}

.sidebar-nav {
  padding: 0 8px;
}

.nav-item.router-link-active {
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

.nav-item.router-link-active .nav-icon {
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
