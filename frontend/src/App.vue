<script setup lang="ts">
import { onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { RouterLink, RouterView, useRouter } from 'vue-router'
import AppToast from './components/AppToast.vue'

const authStore = useAuthStore()
const router = useRouter()

onMounted(async () => {
  await authStore.fetchMe()
})

async function logout() {
  await authStore.logout()
  router.push('/login')
}
</script>

<template>
  <div class="app-shell">
    <header class="topbar">
      <RouterLink to="/" class="brand-shell">
        <div class="brand-mark">
          <svg
            xmlns="http://www.w3.org/2000/svg"
            height="24px"
            viewBox="0 -960 960 960"
            width="24px"
            fill="#e3e3e3"
          >
            <path
              d="M324-111.5Q251-143 197-197t-85.5-127Q80-397 80-480t31.5-156Q143-709 197-763t127-85.5Q397-880 480-880t156 31.5Q709-817 763-763t85.5 127Q880-563 880-480t-31.5 156Q817-251 763-197t-127 85.5Q563-80 480-80t-156-31.5ZM480-160q56 0 105.5-17.5T676-227l-57-57q-29 21-64.5 32.5T480-240q-100 0-170-70t-70-170q0-100 70-170t170-70q100 0 170 70t70 170q0 39-12 75t-33 65l57 57q32-41 50-91t18-106q0-134-93-227t-227-93q-134 0-227 93t-93 227q0 134 93 227t227 93Zm0-160q22 0 42.5-5.5T561-342l-61-61q-5 2-10 2.5t-10 .5q-33 0-56.5-23.5T400-480q0-33 23.5-56.5T480-560q33 0 56.5 23.5T560-480q0 6-.5 11.5T557-458l60 60q11-18 17-38.5t6-43.5q0-66-47-113t-113-47q-66 0-113 47t-47 113q0 66 47 113t113 47Z"
            />
          </svg>
        </div>
        <div>
          <p class="brand-title">JobRadar</p>
        </div>
      </RouterLink>

      <nav class="main-nav">
        <RouterLink to="/">Home</RouterLink>
        <RouterLink to="/jobs">Jobs</RouterLink>
      </nav>

      <div>
        <div class="action-group" v-if="authStore.user">
          <RouterLink to="/profile" class="user-chip">
            <img
              v-if="authStore.user.avatar_url"
              :src="authStore.user.avatar_url"
              alt="User avatar"
              class="avatar"
            />
            <span class="user-name">{{ authStore.user.name }}</span>
          </RouterLink>
          <button class="button button-secondary" @click="logout">Logout</button>
        </div>
        <div v-else>
          <RouterLink class="button button-primary" to="/login">Login</RouterLink>
        </div>
      </div>
    </header>

    <main class="content-shell">
      <div v-if="authStore.loading" class="status-banner">Loading auth state…</div>
      <div v-else-if="authStore.error" class="status-banner status-error">
        Error: {{ authStore.error }}
      </div>
      <RouterView v-else />
    </main>
  </div>
  <AppToast />
</template>

<style>
html,
body,
#app {
  min-height: 100%;
}

body {
  margin: 0;
  font-family: var(--font-base);
  background: var(--color-surface);
  color: var(--color-text);
}

* {
  box-sizing: border-box;
}

a {
  text-decoration: none;
  color: inherit;
}

.app-shell {
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: var(--color-surface);
  overflow: hidden;
}

.content-shell {
  flex: 1;
  overflow-y: auto;
}

.topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1.5rem;
  padding: var(--spacing-1) var(--spacing-4);
  border-bottom: 1px solid rgba(148, 163, 184, 0.18);
  background: rgba(255, 255, 255, 0.9);
  backdrop-filter: blur(16px);
  flex-shrink: 0;

  position: sticky;
  top: 0;
  z-index: 100;
  background-color: var(--color-surface);
}

.brand-shell {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.brand-mark {
  width: 30px;
  height: 30px;
  border-radius: 100%;
  background: var(--color-accent);
  display: grid;
  place-items: center;
  color: white;
}

.brand-title {
  margin: 0;
  font-size: 1.1rem;
}

.main-nav {
  display: flex;
  gap: 1.5rem;
  flex-wrap: wrap;
}

.main-nav a {
  color: var(--muted);
  padding: 0.6rem 0;
  transition: color 180ms ease;
}

.main-nav a.router-link-active,
.main-nav a:hover {
  color: var(--text);
}

.action-group {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.user-chip {
  display: inline-flex;
  align-items: center;
  gap: 0.85rem;
  padding: var(--spacing-1) var(--spacing-2);
  background: rgba(255, 255, 255, 0.8);
}

.avatar {
  width: 34px;
  height: 34px;
  border-radius: 50%;
  background: var(--accent-soft);
  color: var(--accent);
  display: grid;
  place-items: center;
  font-weight: 700;
}

.user-name {
  font-weight: 600;
  color: var(--text);
}

.status-banner {
  padding: 1rem 1.25rem;
  border-radius: 18px;
  background: rgba(79, 70, 229, 0.06);
  color: var(--text);
  border: 1px solid rgba(79, 70, 229, 0.12);
}

.status-error {
  background: rgba(244, 63, 94, 0.08);
  color: #b91c1c;
  border-color: rgba(244, 63, 94, 0.18);
}

@media (max-width: 820px) {
  .topbar {
    flex-direction: column;
    align-items: stretch;
  }

  .main-nav {
    justify-content: space-between;
    width: 100%;
    gap: 0.75rem;
  }

  .action-group {
    justify-content: space-between;
    width: 100%;
  }
}
</style>
