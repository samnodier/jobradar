import { afterEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import HomeView from '@/views/HomeView.vue'
import { useAuthStore } from '@/stores/auth'
vi.mock('vue-router', () => ({
  useRouter: () => ({ push: vi.fn() }),
  RouterLink: { props: ['to'], template: '<a><slot /></a>' },
}))
const ok = (data: unknown) => ({ ok: true, json: async () => data })
function render(signedIn = false) {
  const pinia = createPinia()
  setActivePinia(pinia)
  if (signedIn) useAuthStore().$patch({ user: { id: 'owner' } as never })
  return mount(HomeView, { global: { plugins: [pinia] } })
}
afterEach(() => vi.unstubAllGlobals())
describe('public homepage', () => {
  it('shows real public jobs and statistics without synthetic recommendations', async () => {
    const fetcher = vi
      .fn()
      .mockImplementation((url: string) =>
        Promise.resolve(
          url.includes('stats')
            ? ok({
                total_jobs: 27,
                new_jobs_today: 2,
                sources: ['remoteok'],
                latest_scrape_at: null,
              })
            : ok([{ id: 'job', title: 'Actual public role', company_name: 'Real employer' }]),
        ),
      )
    vi.stubGlobal('fetch', fetcher)
    const view = render()
    await flushPromises()
    expect(view.text()).toContain('Actual public role')
    expect(view.text()).toContain('27')
    expect(view.text()).not.toContain('12,483')
    expect(view.text()).not.toContain('92%')
    expect(view.text()).not.toContain('Recommended for you')
    expect(fetcher).toHaveBeenCalledWith(expect.stringContaining('order=newest'), {
      credentials: 'omit',
    })
    view.unmount()
  })
  it('does not fabricate totals or jobs during an outage', async () => {
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new Error('Offline')))
    const view = render()
    await flushPromises()
    expect(view.text()).toContain('Could not load recent jobs')
    expect(view.text()).not.toContain('12,483')
    expect(view.text()).not.toContain('241')
    expect(view.text()).toContain('Not recorded')
    view.unmount()
  })
  it('uses the real score-ranked recommendation response for the owner', async () => {
    vi.stubGlobal(
      'fetch',
      vi
        .fn()
        .mockImplementation((url: string) =>
          Promise.resolve(
            url.includes('filter=matched')
              ? ok([
                  { id: 'match', title: 'Scored role', company_name: 'Employer', match_score: 67 },
                ])
              : url.includes('stats')
                ? ok({ total_jobs: 0, new_jobs_today: 0, sources: [], latest_scrape_at: null })
                : ok([]),
          ),
        ),
    )
    const view = render(true)
    await flushPromises()
    expect(view.text()).toContain('67% match')
    expect(view.text()).not.toContain('92%')
    expect(view.text()).toContain('No recent public jobs yet')
    view.unmount()
  })
})
