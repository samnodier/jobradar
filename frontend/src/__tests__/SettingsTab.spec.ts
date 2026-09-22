import { afterEach, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { createMemoryHistory, createRouter } from 'vue-router'
import SettingsTab from '@/components/SettingsTab.vue'
import { usePreferencesStore } from '@/stores/preferences'
afterEach(() => vi.unstubAllGlobals())
it('keeps master notification failures visible and retries from the explicit Save button', async () => {
  const pinia = createPinia()
  setActivePinia(pinia)
  const store = usePreferencesStore()
  const initial = {
    min_salary: null,
    max_salary: null,
    salary_currency: 'USD',
    preferred_job_types: [],
    preferred_industries: [],
    company_stage_preference: [],
    notify_jobs: true,
    skills: [],
    desired_locations: [],
    desired_roles: [],
  }
  const fetch = vi
    .fn()
    .mockResolvedValueOnce({ ok: true, json: async () => initial })
    .mockResolvedValueOnce({ ok: false, json: async () => ({ error: 'Offline' }) })
    .mockImplementationOnce((_url, options) =>
      Promise.resolve({ ok: true, json: async () => JSON.parse(options.body) }),
    )
  vi.stubGlobal('fetch', fetch)
  await store.fetchPreferences()
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [{ path: '/', component: { template: '<div/>' } }],
  })
  await router.push('/')
  const wrapper = mount(SettingsTab, {
    props: { user: null },
    global: {
      plugins: [pinia, router],
      stubs: {
        MatchingSettings: true,
        PasswordSettings: true,
        AccessTokens: true,
        DownloadProfile: true,
      },
    },
  })
  await wrapper.get('[role="switch"]').trigger('click')
  await flushPromises()
  expect(wrapper.get('[role="alert"]').text()).toBe('Offline')
  expect(store.preferences?.notify_jobs).toBe(false)
  await wrapper
    .findAll('button')
    .find((b) => b.text() === 'Save preferences')!
    .trigger('click')
  await flushPromises()
  expect(wrapper.find('[role="alert"]').exists()).toBe(false)
  expect(JSON.parse(fetch.mock.calls[2]![1].body).notify_jobs).toBe(false)
  expect(wrapper.text()).toContain('Preferences saved automatically')
  wrapper.unmount()
})
