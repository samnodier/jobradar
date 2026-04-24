<script setup lang="ts">
import { onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { RouterLink, RouterView, useRouter } from 'vue-router'

const authStore = useAuthStore()
const router = useRouter()

onMounted(async () => {
  await authStore.fetchMe()
})

</script>

<template>
  <main>
    <nav>
      <RouterLink to="/">Home</RouterLink>
      <RouterLink to="/jobs">Jobs</RouterLink>
      <RouterLink to="/login">Login</RouterLink>
      <RouterLink to="/profile">Profile</RouterLink>
      <span v-if="authStore.user">{{ authStore.user.username }}</span>
    </nav>

    <div v-if="authStore.loading">Loading...</div>
    <div v-else-if="authStore.error">Error: {{ authStore.error }}</div>
    <RouterView v-else />
  </main>
</template>

<style scoped>
nav {
  display: flex;
  gap: 1rem;
  margin-bottom: 1.5rem;
}
</style>
