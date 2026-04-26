<template>
  <aside class="sidebar">
    <div class="logo">
      <div class="logo-mark">
        <svg viewBox="0 0 13 13" fill="none" xmlns="http://www.w3.org/2000/svg">
          <circle cx="6.5" cy="6.5" r="5" stroke="white" stroke-width="1.5"/>
          <circle cx="6.5" cy="6.5" r="2" fill="white"/>
        </svg>
      </div>
      JobRadar
    </div>

    <nav class="nav">
      <RouterLink to="/jobs" class="nav-item" active-class="active">
        <IconGrid class="nav-icon" />
        All jobs
      </RouterLink>
      <RouterLink to="/jobs?filter=saved" class="nav-item" active-class="active">
        <IconStar class="nav-icon" />
        Saved
      </RouterLink>
      <RouterLink to="/jobs?filter=applied" class="nav-item" active-class="active">
        <IconCheck class="nav-icon" />
        Applied
      </RouterLink>
    </nav>

    <div class="nav-section-label">Sources</div>
    <nav class="nav">
      <div class="nav-item nav-disabled">
        <span class="source-dot remoteok" />
        RemoteOK
      </div>
      <div class="nav-item nav-disabled">
        <span class="source-dot adzuna" />
        Adzuna
      </div>
    </nav>

    <div class="sidebar-footer">
      <RouterLink to="/profile" class="nav-item" active-class="active">
        <IconUser class="nav-icon" />
        Profile
      </RouterLink>
    </div>
  </aside>
</template>

<script setup lang="ts">
import { RouterLink } from 'vue-router'
import IconGrid from '@/components/icons/IconGrid.vue'
import IconStar from '@/components/icons/IconStar.vue'
import IconCheck from '@/components/icons/IconCheck.vue'
import IconUser from '@/components/icons/IconUser.vue'
</script>

<style scoped>
.sidebar {
  width: 220px;
  min-width: 220px;
  background: #fff;
  border-right: 1px solid #e8e8e4;
  display: flex;
  flex-direction: column;
  padding: 14px 0;
}

.logo {
  padding: 0 14px 18px;
  font-size: 14.5px;
  font-weight: 500;
  color: #1a1a1a;
  display: flex;
  align-items: center;
  gap: 8px;
}

.logo-mark {
  width: 22px;
  height: 22px;
  background: #5e6ad2;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.logo-mark svg {
  width: 13px;
  height: 13px;
}

.nav {
  padding: 0 8px;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 10px;
  border-radius: 10px;
  cursor: pointer;
  font-size: 13.5px;
  color: #666;
  text-decoration: none;
  transition: background 0.1s, color 0.1s;
}

.nav-item:hover {
  background: #f5f5f3;
  color: #1a1a1a;
}

.nav-item.active {
  background: #f5f5f3;
  color: #1a1a1a;
  font-weight: 500;
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

.nav-section-label {
  font-size: 11px;
  color: #bbb;
  padding: 14px 16px 4px;
  letter-spacing: 0.04em;
  text-transform: uppercase;
}

.source-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  flex-shrink: 0;
}

.source-dot.remoteok { background: #5e6ad2; }
.source-dot.adzuna { background: #26c6a6; }

.nav-disabled {
  opacity: 0.7;
}

.sidebar-footer {
  margin-top: auto;
  padding: 0 8px;
  border-top: 1px solid #f0f0ec;
  padding-top: 8px;
}
</style>
