import { describe, it, expect, vi, afterEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia } from 'pinia'
import { createRouter, createMemoryHistory } from 'vue-router'
import App from '../App.vue'
import { useAuthStore } from '../stores/auth'

afterEach(() => vi.unstubAllGlobals())
describe('App', () => {
  it('renders the public shell when no session exists', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: false, status: 401 }))
    const router = createRouter({
      history: createMemoryHistory(),
      routes: [
        { path: '/', component: { template: '<p>Public jobs</p>' } },
        { path: '/login', component: { template: '<p>Login</p>' } },
      ],
    })
    await router.push('/')
    await router.isReady()
    const wrapper = mount(App, { global: { plugins: [createPinia(), router] } })
    await flushPromises()
    expect(wrapper.text()).toContain('JobRadar')
    expect(wrapper.text()).toContain('Public jobs')
    expect(wrapper.find('a[href="/login"]').exists()).toBe(true)
    wrapper.unmount()
  })
  it('keeps the current page mounted while the session is re-checked', async () => {
    let finish: (value: unknown) => void = () => {}
    const fetcher = vi
      .fn()
      .mockResolvedValueOnce({ ok: false, status: 401 })
      .mockImplementation(() => new Promise((resolve) => (finish = resolve)))
    vi.stubGlobal('fetch', fetcher)
    const router = createRouter({
      history: createMemoryHistory(),
      routes: [
        { path: '/', component: { template: '<p>Public jobs</p>' } },
        { path: '/login', component: { template: '<p>Login</p>' } },
      ],
    })
    await router.push('/')
    await router.isReady()
    const pinia = createPinia()
    const wrapper = mount(App, { global: { plugins: [pinia, router] } })
    await flushPromises()
    const pending = useAuthStore(pinia).fetchMe()
    await flushPromises()
    expect(wrapper.text()).toContain('Public jobs')
    expect(wrapper.text()).not.toContain('Loading auth state')
    finish({ ok: false, status: 401 })
    await pending
    wrapper.unmount()
  })
})
