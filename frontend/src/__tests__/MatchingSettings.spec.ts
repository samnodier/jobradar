import { afterEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import MatchingSettings from '@/components/MatchingSettings.vue'
const initial = () => ({
  min_score: 60,
  work_mode: 'any',
  include_unknown: true,
  search_country: 'us',
  email_enabled: false,
  notifications_enabled: true,
  digest_timezone: 'Africa/Nairobi',
  digest_hours: [8, 18],
  digest_max_jobs: 10,
})
const response = (data: unknown) => ({ ok: true, json: async () => data })
const render = () =>
  mount(MatchingSettings, { global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } } })
afterEach(() => vi.unstubAllGlobals())
describe('Matching settings', () => {
  it('preserves edits made during a save for the next explicit save', async () => {
    let resolve!: (v: unknown) => void
    const fetch = vi
      .fn()
      .mockResolvedValueOnce(response(initial()))
      .mockImplementationOnce(
        () =>
          new Promise((r) => {
            resolve = r
          }),
      )
      .mockImplementationOnce((_url, options) =>
        Promise.resolve(response(JSON.parse(options.body))),
      )
    vi.stubGlobal('fetch', fetch)
    const wrapper = render()
    await flushPromises()
    expect(wrapper.findAll('h3').map((h) => h.text())).toEqual([
      'Match criteria',
      'Job sources',
      'Delivery channels',
      'Email schedule',
    ])
    const score = wrapper.get('input[type="number"]')
    await score.setValue(70)
    await wrapper.get('form').trigger('submit')
    await score.setValue(85)
    resolve(response({ ...initial(), min_score: 70 }))
    await flushPromises()
    expect((score.element as HTMLInputElement).value).toBe('85')
    expect(wrapper.text()).toContain('Unsaved changes')
    await wrapper.get('form').trigger('submit')
    await flushPromises()
    expect(JSON.parse(fetch.mock.calls[2]![1].body).min_score).toBe(85)
    expect(wrapper.text()).toContain('All changes saved')
  })
  it('rejects duplicate, out-of-range and excessive hours before a request', async () => {
    const fetch = vi.fn().mockResolvedValue(response(initial()))
    vi.stubGlobal('fetch', fetch)
    const wrapper = render()
    await flushPromises()
    for (const hours of ['8, 8', '24', '8, 12, 18', 'noon']) {
      await wrapper.get('input[placeholder="8, 18"]').setValue(hours)
      await wrapper.get('form').trigger('submit')
      await flushPromises()
      expect(wrapper.get('[role="alert"]').text()).toContain('one or two whole hours')
    }
    expect(fetch).toHaveBeenCalledTimes(1)
  })
  it('keeps failed changes available to retry', async () => {
    const fetch = vi
      .fn()
      .mockResolvedValueOnce(response(initial()))
      .mockResolvedValueOnce({ ok: false, json: async () => ({ error: 'Save failed' }) })
      .mockImplementationOnce((_url, options) =>
        Promise.resolve(response(JSON.parse(options.body))),
      )
    vi.stubGlobal('fetch', fetch)
    const wrapper = render()
    await flushPromises()
    await wrapper.get('input[type="number"]').setValue(90)
    await wrapper.get('form').trigger('submit')
    await flushPromises()
    expect(wrapper.get('[role="alert"]').text()).toBe('Save failed')
    expect((wrapper.get('input[type="number"]').element as HTMLInputElement).value).toBe('90')
    await wrapper.get('form').trigger('submit')
    await flushPromises()
    expect(wrapper.text()).toContain('All changes saved')
  })
  it('disables save after a load failure', async () => {
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new Error('Offline')))
    const wrapper = render()
    await flushPromises()
    expect(wrapper.get('[role="alert"]').text()).toContain('Offline')
    expect(wrapper.get('button').attributes('disabled')).toBeDefined()
  })
})
