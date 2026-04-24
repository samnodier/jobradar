<script setup lang="ts">
import { useAuthStore } from '@/stores/auth'
import { RouterLink, useRouter } from 'vue-router'

const authStore = useAuthStore()
const router = useRouter()

async function handleLogout() {
  await authStore.logout()
  router.push('/login')
}
</script>

<template>
  <main class="profile-page">
    <div v-if="authStore.loading" class="status-card">
      <p>Loading profile…</p>
    </div>

    <div v-else-if="!authStore.user" class="status-card">
      <p class="status-title">You are not logged in.</p>
      <RouterLink to="/login" class="button button-primary">Go to login</RouterLink>
    </div>

    <section v-else class="profile-card">
      <div class="profile-header">
        <div class="avatar-shell">
          <img
            v-if="authStore.user.avatar_url"
            :src="authStore.user.avatar_url"
            alt="Profile avatar"
          />
          <span v-else class="avatar-fallback">{{ authStore.user.username?.[0]?.toUpperCase() || 'U' }}</span>
        </div>
        <div>
          <p class="eyebrow">Your profile</p>
          <h1>{{ authStore.user.name || authStore.user.username }}</h1>
          <p class="subtitle">@{{ authStore.user.username }}</p>
        </div>
      </div>

      <div class="profile-details">
        <div class="detail-row">
          <span>Email</span>
          <p>{{ authStore.user.email }}</p>
        </div>
        <div class="detail-row">
          <span>Username</span>
          <p>{{ authStore.user.username }}</p>
        </div>
      </div>

      <div class="profile-actions">
        <button class="button button-ghost" @click="handleLogout">Logout</button>
      </div>
    </section>
  </main>
</template>

<style scoped>
.profile-page {
  max-width: 720px;
  margin: 2rem auto;
  padding: 0 1rem;
}

.status-card,
.profile-card {
  background: #ffffff;
  border: 1px solid #e7ebf0;
  border-radius: 22px;
  padding: 28px;
  box-shadow: 0 18px 40px rgba(15, 23, 42, 0.06);
}

.status-card {
  display: grid;
  gap: 1rem;
  text-align: center;
}

.status-title {
  margin: 0;
  font-size: 1.15rem;
  font-weight: 700;
}

.profile-header {
  display: flex;
  align-items: center;
  gap: 1.25rem;
  margin-bottom: 1.75rem;
}

.avatar-shell {
  width: 72px;
  height: 72px;
  border-radius: 18px;
  background: #f5f7fb;
  display: grid;
  place-items: center;
  overflow: hidden;
}

.avatar-shell img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.avatar-fallback {
  color: #0f172a;
  font-weight: 700;
  font-size: 1.25rem;
}

.eyebrow {
  margin: 0 0 0.45rem;
  text-transform: uppercase;
  letter-spacing: 0.14em;
  font-size: 0.75rem;
  color: #667085;
}

h1 {
  margin: 0;
  font-size: 2rem;
  color: #0f172a;
}

.subtitle {
  margin: 0.5rem 0 0;
  color: #475569;
}

.profile-details {
  display: grid;
  gap: 1rem;
  margin-bottom: 1.75rem;
}

.detail-row {
  display: grid;
  gap: 0.35rem;
}

.detail-row span {
  color: #667085;
  font-size: 0.85rem;
}

.detail-row p {
  margin: 0;
  color: #0f172a;
  font-weight: 600;
}

.profile-actions {
  display: flex;
  justify-content: flex-end;
}

.button {
  border: 1px solid transparent;
  border-radius: 999px;
  padding: 0.85rem 1.4rem;
  font-weight: 600;
  cursor: pointer;
}

.button-primary {
  color: #ffffff;
  background: #4f46e5;
}

.button-ghost {
  color: #0f172a;
  background: #f8fafc;
  border-color: #e2e8f0;
}

.button:hover {
  opacity: 0.95;
}
</style>
