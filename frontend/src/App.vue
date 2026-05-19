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
  <div class="h-screen flex flex-col bg-white overflow-hidden">
    <!-- Topbar -->
    <header
      class="flex items-center justify-between flex-row gap-6 px-4 py-1 border-b border-slate-400/20 bg-white/90 backdrop-blur-md shrink-0 flex-nowrap whitespace-nowrap min-h-(--topbar-height) sticky top-0 z-100"
    >
      <!-- Brand -->
      <RouterLink to="/" class="flex items-center gap-4">
        <div
          class="w-7.5 h-7.5 grid place-items-center text-white"
          :style="{ background: 'var(--color-accent)' }"
        >
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
        <p class="m-0 text-[1.1rem] font-medium">JobRadar</p>
      </RouterLink>

      <div style="flex: 1" />

      <!-- Auth actions -->
      <div>
        <div v-if="authStore.user" class="flex items-center gap-4">
          <RouterLink
            :to="'/@' + authStore.user.username"
            class="inline-flex items-center gap-[0.85rem] px-2 py-1 bg-white/80"
          >
            <img
              v-if="authStore.user.avatar_url"
              :src="authStore.user.avatar_url"
              alt="User avatar"
              class="w-8.5 h-8.5 rounded-full"
            />
            <span class="font-semibold text-gray-900 max-[820px]:hidden">
              {{ authStore.user.full_name }}
            </span>
          </RouterLink>
          <button class="button button-secondary" @click="logout">Logout</button>
        </div>
        <div v-else>
          <RouterLink class="button button-primary" to="/login">Login</RouterLink>
        </div>
      </div>
    </header>

    <!-- Content -->
    <main class="flex-1 flex flex-col min-h-0 relative">
      <div
        v-if="authStore.loading"
        class="p-4 px-5 rounded-[18px] bg-indigo-500/6 text-gray-900 border border-indigo-500/12"
      >
        Loading auth state…
      </div>
      <div
        v-else-if="authStore.error"
        class="p-4 px-5 rounded-[18px] bg-red-400/8 text-red-700 border border-red-400/18"
      >
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
  height: 100%;
}
body {
  margin: 0;
  font-family: var(--font-base);
  color: #111827;
}
* {
  box-sizing: border-box;
}
a {
  text-decoration: none;
  color: inherit;
}
</style>
