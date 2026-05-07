import { useAuthStore } from '@/stores/auth'
import { createRouter, createWebHistory } from 'vue-router'

import Home from '@/views/HomeView.vue'
import Login from '@/views/LoginView.vue'
import Jobs from '@/views/JobsView.vue'
import Onboarding from '@/views/OnboardingView.vue'
import Profile from '@/views/ProfileView.vue'
import Applications from '@/views/ApplicationsView.vue'
import Dashboard from '@/views/DashboardView.vue'
import Admin from '@/views/AdminView.vue'
import Status from '@/views/StatusView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: Home,
    },
    {
      path: '/login',
      name: 'login',
      component: Login,
    },
    {
      path: '/jobs',
      name: 'jobs',
      component: Jobs,
    },
    {
      path: '/dashboard',
      name: 'dashboard',
      component: Dashboard,
      meta: { requiresAuth: true },
    },
    {
      path: '/auth/onboarding',
      name: 'onboarding',
      component: Onboarding,
      beforeEnter: (to) => {
        if (!to.query.token) {
          return '/login'
        }
      },
    },
    {
      path: '/profile',
      name: 'profile',
      component: Profile,
      meta: { requiresAuth: true },
    },
    {
      path: '/@:username',
      name: 'user-profile',
      component: Profile,
      meta: { requiresAuth: true },
    },
    {
      path: '/applications',
      name: 'applications',
      component: Applications,
      meta: { requiresAuth: true },
    },
    {
      path: '/admin',
      name: 'admin',
      component: Admin,
      meta: { requiresAuth: true },
    },
    {
      path: '/status',
      name: 'status',
      component: Status,
      meta: { requiresAuth: true },
    },
  ],
})

router.beforeEach(async (to) => {
  const authStore = useAuthStore()
  if (!authStore.user) {
    await authStore.fetchMe()
  }

  if (to.meta.requiresAuth && !authStore.user) {
    return '/login'
  }
})

export default router
