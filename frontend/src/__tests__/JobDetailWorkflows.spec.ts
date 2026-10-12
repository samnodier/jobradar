import { afterEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createPinia } from 'pinia'
import JobDetail from '@/components/JobDetail.vue'
import { useAuthStore } from '@/stores/auth'
import type { Job } from '@/types/job'
import type { User } from '@/types/user'
const job: Job = {
  id: 'job-1',
  external_id: '1',
  job_source: 'manual',
  title: 'Go Engineer',
  company_name: 'Example',
  source_url: 'https://example.test/job',
  skills: [],
  is_saved: false,
  is_applied: false,
  is_matched: false,
}
const response = (value: unknown) => ({
  ok: true,
  json: async () => value,
  text: async () => String(value),
})
function render() {
  const pinia = createPinia()
  useAuthStore(pinia).$patch({ user: { id: 'owner' } as User })
  return mount(JobDetail, { props: { job }, global: { plugins: [pinia] } })
}
afterEach(() => vi.unstubAllGlobals())
describe('job preparation and feedback', () => {
  it('copies the selected preparation mode without sending an AI request', async () => {
    const fetch = vi.fn((url: string) =>
      Promise.resolve(response(url.includes('ai-context') ? 'Resume facts' : { sentiment: 0 })),
    )
    const copy = vi.fn().mockResolvedValue(undefined)
    vi.stubGlobal('fetch', fetch)
    vi.stubGlobal('navigator', { clipboard: { writeText: copy } })
    const view = render()
    await flushPromises()
    await view.get('select').setValue('resume')
    await view
      .findAll('button')
      .find((button) => button.text() === 'Copy for AI')!
      .trigger('click')
    await flushPromises()
    expect(fetch).toHaveBeenCalledWith('/api/jobs/job-1/ai-context?mode=resume', {
      credentials: 'include',
    })
    expect(copy).toHaveBeenCalledWith('Resume facts')
    expect(fetch.mock.calls.every(([url]) => url.startsWith('/api/jobs/'))).toBe(true)
    view.unmount()
  })
  it('saves and reverses explicit feedback', async () => {
    const fetch = vi.fn().mockResolvedValue(response({ sentiment: 0 }))
    vi.stubGlobal('fetch', fetch)
    const view = render()
    await flushPromises()
    await view
      .findAll('button')
      .find((button) => button.text() === 'Not relevant')!
      .trigger('click')
    await flushPromises()
    expect(fetch).toHaveBeenLastCalledWith(
      '/api/jobs/job-1/feedback',
      expect.objectContaining({ method: 'PUT', body: JSON.stringify({ sentiment: -1 }) }),
    )
    expect(
      view
        .findAll('button')
        .find((button) => button.text() === 'Not relevant')!
        .attributes('aria-pressed'),
    ).toBe('true')
    await view
      .findAll('button')
      .find((button) => button.text() === 'Clear feedback')!
      .trigger('click')
    await flushPromises()
    expect(fetch).toHaveBeenLastCalledWith(
      '/api/jobs/job-1/feedback',
      expect.objectContaining({ method: 'DELETE' }),
    )
    view.unmount()
  })
  it('does not copy an old export after the selected job changed', async () => {
    let finish!: (value: unknown) => void
    const fetch = vi.fn((url: string) =>
      url.includes('ai-context')
        ? new Promise((resolve) => {
            finish = resolve
          })
        : Promise.resolve(response({ sentiment: 0 })),
    )
    const copy = vi.fn()
    vi.stubGlobal('fetch', fetch)
    vi.stubGlobal('navigator', { clipboard: { writeText: copy } })
    const view = render()
    await flushPromises()
    await view
      .findAll('button')
      .find((button) => button.text() === 'Copy for AI')!
      .trigger('click')
    await view.setProps({ job: { ...job, id: 'job-2' } })
    finish(response('Old job context'))
    await flushPromises()
    expect(copy).not.toHaveBeenCalled()
    expect(view.find('textarea[aria-label="AI application context"]').exists()).toBe(false)
    view.unmount()
  })
})
