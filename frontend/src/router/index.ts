import { useAuthStore } from '@/stores/auth'
import { createRouter, createWebHistory } from 'vue-router'

import Home from '@/views/HomeView.vue'
import Login from '@/views/LoginView.vue'
import Jobs from '@/views/JobsView.vue'
import Onboarding from '@/views/OnboardingView.vue'
import Profile from '@/views/ProfileView.vue'

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
      path: '/auth/onboarding',
      name: 'onboarding',
      component: Onboarding,
      // beforeEnter: (to) => {
      //   if (!to.query.token) {
      //     return '/login'
      //   }
      // },
    },
    {
      path: '/profile',
      name: 'profile',
      component: Profile,
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
