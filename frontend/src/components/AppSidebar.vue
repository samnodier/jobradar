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
  <aside
    class="w-220px min-w-220px h-full bg-white border-r border-ui-border flex flex-col py-3.5 overflow-y-auto"
  >
    <nav class="px-2 space-y-4">
      <!-- Discovery Section -->
      <div class="space-y-1">
        <h2 class="text-[0.7rem] font-medium text-gray-400 px-4 mb-2 tracking-widest uppercase">
          Discovery
        </h2>
        <RouterLink
          to="/jobs"
          class="flex items-center gap-3 px-4 py-2 text-sm text-gray-500 hover:bg-gray-100/50 hover:text-gray-900 rounded-lg transition-all group"
          active-class="bg-accent-lighter text-accent font-semibold"
        >
          <Zap
            class="w-4 h-4 opacity-60 group-hover:opacity-100 shrink-0"
            :class="{ 'opacity-100': $route.path === '/jobs' }"
          />
          Browse Roles
        </RouterLink>

        <div class="flex items-center gap-3 px-4 py-2 text-sm text-gray-400 opacity-70">
          <span class="w-7px h-7px rounded-full bg-[#5e6ad2] shrink-0" />
          RemoteOK
        </div>
        <div class="flex items-center gap-3 px-4 py-2 text-sm text-gray-400 opacity-70">
          <span class="w-7px h-7px rounded-full bg-[#26c6a6] shrink-0" />
          Adzuna
        </div>
      </div>

      <!-- Workspace Section -->
      <div class="pt-4 border-t border-ui-border-soft space-y-1">
        <div class="text-[0.7rem] font-medium text-gray-400 px-4 mb-2 tracking-widest uppercase">
          Workspace
        </div>
        <RouterLink
          to="/dashboard"
          class="flex items-center gap-3 px-4 py-2 text-sm text-gray-500 hover:bg-gray-100/50 hover:text-gray-900 rounded-lg transition-all group"
          active-class="bg-accent-lighter text-accent font-semibold"
        >
          <LayoutDashboard
            class="w-4 h-4 opacity-60 group-hover:opacity-100 shrink-0"
            :class="{ 'opacity-100': $route.path === '/dashboard' }"
          />
          Overview
        </RouterLink>
        <RouterLink
          :to="{
            path: '/jobs',
            query: { filter: 'saved' },
          }"
          class="flex items-center gap-3 px-4 py-2 text-sm text-gray-500 hover:bg-gray-100/50 hover:text-gray-900 rounded-lg transition-all group"
          active-class="bg-accent-lighter text-accent font-semibold"
        >
          <Bookmark
            class="w-4 h-4 opacity-60 group-hover:opacity-100 shrink-0"
            :class="{ 'opacity-100': $route.query.filter === 'saved' }"
          />
          Saved Jobs
        </RouterLink>
        <RouterLink
          to="/applications"
          class="flex items-center gap-3 px-4 py-2 text-sm text-gray-500 hover:bg-gray-100/50 hover:text-gray-900 rounded-lg transition-all group"
          active-class="bg-accent-lighter text-accent font-semibold"
        >
          <CheckCircle
            class="w-4 h-4 opacity-60 group-hover:opacity-100 shrink-0"
            :class="{ 'opacity-100': $route.path === '/applications' }"
          />
          Tracker
        </RouterLink>
      </div>

      <!-- Personal Section -->
      <div v-if="authStore.user" class="pt-4 border-t border-ui-border-soft space-y-1">
        <div class="text-[0.7rem] font-medium text-gray-400 px-4 mb-2 tracking-widest uppercase">
          Personal
        </div>
        <RouterLink
          :to="'/@' + authStore.user?.username"
          class="flex items-center gap-3 px-4 py-2 text-sm text-gray-500 hover:bg-gray-100/50 hover:text-gray-900 rounded-lg transition-all group"
          active-class="bg-accent-lighter text-accent font-semibold"
        >
          <User
            class="w-4 h-4 opacity-60 group-hover:opacity-100 shrink-0"
            :class="{ 'opacity-100': $route.path.startsWith('/@') }"
          />
          Profile
        </RouterLink>
      </div>

      <!-- System Section -->
      <div v-if="authStore.user" class="pt-4 border-t border-ui-border-soft space-y-1">
        <h2 class="text-[0.7rem] font-medium text-gray-400 px-4 mb-2 tracking-widest uppercase">
          System
        </h2>
        <RouterLink
          to="/admin"
          class="flex items-center gap-3 px-4 py-2 text-sm text-gray-500 hover:bg-gray-100/50 hover:text-gray-900 rounded-lg transition-all group"
          active-class="bg-accent-lighter text-accent font-semibold"
        >
          <BarChart3
            class="w-4 h-4 opacity-60 group-hover:opacity-100 shrink-0"
            :class="{ 'opacity-100': $route.path === '/admin' }"
          />
          Console
        </RouterLink>

        <RouterLink
          to="/status"
          class="flex items-center gap-3 px-4 py-2 text-sm text-gray-500 hover:bg-gray-100/50 hover:text-gray-900 rounded-lg transition-all group"
          active-class="bg-accent-lighter text-accent font-semibold"
        >
          <AlertCircle
            class="w-4 h-4 opacity-60 group-hover:opacity-100 shrink-0"
            :class="{ 'opacity-100': $route.path === '/status' }"
          />
          Health
        </RouterLink>
      </div>
    </nav>
  </aside>
</template>
