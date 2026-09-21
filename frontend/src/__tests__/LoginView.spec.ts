import { afterEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createPinia } from 'pinia'
import { createMemoryHistory, createRouter } from 'vue-router'
import LoginView from '../views/LoginView.vue'

afterEach(() => {
  vi.unstubAllGlobals()
  window.history.replaceState(null, '', '/')
})
async function render() {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/login', component: LoginView },
      { path: '/jobs', component: { template: '<p>Jobs</p>' } },
      { path: '/terms', component: { template: '<p>Terms</p>' } },
      { path: '/privacy', component: { template: '<p>Privacy</p>' } },
    ],
  })
  await router.push('/login')
  await router.isReady()
  const wrapper = mount(LoginView, { global: { plugins: [createPinia(), router] } })
  await flushPromises()
  return { wrapper, router }
}
describe('Email sign-in', () => {
  it('signs in by email and preserves the optional GitHub choice', async () => {
    const fetcher = vi.fn(async (path: string) => ({
      ok: true,
      json: async () =>
        path === '/auth/methods'
          ? { github: true }
          : path === '/auth/users/me'
            ? { id: 'owner' }
            : {},
    }))
    vi.stubGlobal('fetch', fetcher)
    const { wrapper, router } = await render()
    expect(wrapper.find('a[href="/auth/github/login"]').exists()).toBe(true)
    await wrapper.get('input[type="email"]').setValue('owner@example.test')
    await wrapper.get('input[type="password"]').setValue('a long personal password')
    await wrapper.get('form').trigger('submit')
    await flushPromises()
    const call = fetcher.mock.calls.find(([path]) => path === '/auth/password/login')
    expect(call).toBeDefined()
    expect(router.currentRoute.value.path).toBe('/jobs')
    wrapper.unmount()
  })
  it('uses a setup fragment and prevents mismatched password submission', async () => {
    window.history.replaceState(null, '', '/login#setup=private-one-time-token')
    const fetcher = vi.fn(async () => ({ ok: true, json: async () => ({ github: false }) }))
    vi.stubGlobal('fetch', fetcher)
    const { wrapper } = await render()
    expect(wrapper.text()).toContain('Choose your password')
    expect(wrapper.find('input[type="email"]').exists()).toBe(false)
    expect(wrapper.find('a[href="/auth/github/login"]').exists()).toBe(false)
    const fields = wrapper.findAll('input[type="password"]')
    await fields[0]!.setValue('a long personal password')
    await fields[1]!.setValue('a different personal password')
    await wrapper.get('form').trigger('submit')
    await flushPromises()
    expect(wrapper.get('[role="alert"]').text()).toBe('Passwords do not match')
    expect(fetcher).toHaveBeenCalledTimes(1)
    wrapper.unmount()
  })
})
