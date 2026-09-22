import { afterEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { flushPromises, mount } from '@vue/test-utils'
import { usePreferencesStore } from '@/stores/preferences'
import PreferencesTab from '@/components/PreferencesTab.vue'
import type { UserPreferences } from '@/types/preferences'
const initial = (): UserPreferences => ({
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
})
const response = (data: unknown) => ({ ok: true, json: async () => data })
afterEach(() => {
  vi.unstubAllGlobals()
  vi.useRealTimers()
})
describe('Preference persistence', () => {
  it('serializes rapid checkbox saves without overwriting the newest selection', async () => {
    const pinia = createPinia()
    setActivePinia(pinia)
    const pending: Array<{ body: UserPreferences; resolve: (value: unknown) => void }> = []
    vi.stubGlobal(
      'fetch',
      vi.fn((_url: string, options?: RequestInit) => {
        if (!options?.body) return Promise.resolve(response(initial()))
        return new Promise((resolve) =>
          pending.push({ body: JSON.parse(String(options.body)), resolve }),
        )
      }),
    )
    const store = usePreferencesStore()
    await store.fetchPreferences()
    const wrapper = mount(PreferencesTab, { global: { plugins: [pinia] } })
    for (const type of ['Full-time', 'Part-time', 'Contract'])
      await wrapper.get(`input[id="${type}"]`).setValue(true)
    expect(pending).toHaveLength(1)
    pending[0]!.resolve(response(pending[0]!.body))
    await flushPromises()
    expect(store.preferences?.preferred_job_types).toEqual(['Full-time', 'Part-time', 'Contract'])
    expect(pending).toHaveLength(2)
    pending[1]!.resolve(response(pending[1]!.body))
    await flushPromises()
    pending[2]!.resolve(response(pending[2]!.body))
    await flushPromises()
    expect(store.isSaving).toBe(false)
    expect(store.hasUnsavedChanges).toBe(false)
    expect(pending[2]!.body.preferred_job_types).toEqual(['Full-time', 'Part-time', 'Contract'])
    expect(wrapper.text()).toContain('All changes saved')
    wrapper.unmount()
  })
  it('retains failed edits and retries with the visible Save button', async () => {
    const pinia = createPinia()
    setActivePinia(pinia)
    let fail = true
    vi.stubGlobal(
      'fetch',
      vi.fn(async (_url: string, options?: RequestInit) => {
        if (!options?.body) return response(initial())
        return fail
          ? { ok: false, json: async () => ({ error: 'Connection interrupted' }) }
          : response(JSON.parse(String(options.body)))
      }),
    )
    const store = usePreferencesStore()
    await store.fetchPreferences()
    const wrapper = mount(PreferencesTab, { global: { plugins: [pinia] } })
    await wrapper.get('input[id="Full-time"]').setValue(true)
    await flushPromises()
    expect(store.preferences?.preferred_job_types).toEqual(['Full-time'])
    expect(wrapper.text()).toContain('Retry saving')
    fail = false
    await wrapper
      .findAll('button')
      .find((b) => b.text() === 'Save preferences')!
      .trigger('click')
    await flushPromises()
    expect(store.error).toBeNull()
    expect(store.hasUnsavedChanges).toBe(false)
    wrapper.unmount()
  })
  it('keeps commas within one location tag, deduplicates and removes it', async () => {
    const pinia = createPinia()
    setActivePinia(pinia)
    const writes: UserPreferences[] = []
    vi.stubGlobal(
      'fetch',
      vi.fn(async (_url: string, options?: RequestInit) => {
        if (!options?.body) return response(initial())
        const body = JSON.parse(String(options.body))
        writes.push(body)
        return response(body)
      }),
    )
    const store = usePreferencesStore()
    await store.fetchPreferences()
    const wrapper = mount(PreferencesTab, { global: { plugins: [pinia] } })
    const input = wrapper.get('#desired-locations')
    await input.setValue('Nairobi, Kenya')
    await input.trigger('keydown', { key: 'Enter' })
    await flushPromises()
    expect(store.preferences?.desired_locations.map((v) => v.location_name)).toEqual([
      'Nairobi, Kenya',
    ])
    await input.setValue('nairobi, kenya')
    await input.trigger('keydown', { key: 'Enter' })
    await flushPromises()
    expect(store.preferences?.desired_locations).toHaveLength(1)
    await wrapper.get('button[aria-label="Remove Nairobi, Kenya"]').trigger('click')
    await flushPromises()
    expect(writes[writes.length - 1]?.desired_locations).toEqual([])
    wrapper.unmount()
  })
})
