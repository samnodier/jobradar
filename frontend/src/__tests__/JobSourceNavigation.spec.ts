import { afterEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createMemoryHistory, createRouter } from 'vue-router'
import { createPinia } from 'pinia'
import JobsView from '@/views/JobsView.vue'
import AppSidebar from '@/components/AppSidebar.vue'

afterEach(() => {
  vi.useRealTimers()
  vi.unstubAllGlobals()
})
function router() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [{ path: '/:pathMatch(.*)*', component: { template: '<div />' } }],
  })
}

describe('job source navigation', () => {
  it('links source shortcuts to working feeds and gives Best Matches an active state', async () => {
    const routing = router()
    await routing.push('/jobs?filter=matched')
    const view = mount(AppSidebar, { global: { plugins: [routing, createPinia()] } })
    expect(view.get('a[href="/jobs?source=remoteok"]').text()).toBe('RemoteOK')
    expect(view.get('a[href="/jobs?source=adzuna"]').text()).toBe('Adzuna')
    expect(view.get('a[href="/jobs?filter=matched"]').classes()).toContain('bg-accent-lighter')
    expect(view.find('a[href="/jobs?filter=matched"] svg').exists()).toBe(true)
    expect(view.find('a[href="/admin"]').exists()).toBe(false)
    await routing.push('/jobs?source=adzuna')
    expect(view.get('a[href="/jobs?source=adzuna"]').classes()).toContain('bg-accent-lighter')
    expect(view.get('a[href="/jobs"]').classes()).not.toContain('bg-accent-lighter')
    view.unmount()
  })

  it('sends source to the server and preserves it through filters, search and navigation', async () => {
    vi.useFakeTimers()
    const fetch = vi.fn().mockResolvedValue({ ok: true, json: async () => [] })
    vi.stubGlobal('fetch', fetch)
    const routing = router()
    await routing.push('/jobs?source=adzuna&q=Engineer')
    const view = mount(JobsView, {
      global: {
        plugins: [routing],
        stubs: { AppSidebar: true, JobRow: true, JobDetail: true, ImportJobPanel: true },
      },
    })
    await flushPromises()
    expect(fetch.mock.calls[0]![0]).toContain('source=adzuna')
    expect(fetch.mock.calls[0]![0]).toContain('q=Engineer')
    const saved = view.findAll('button').find((button) => button.text() === 'Saved')!
    await saved.trigger('click')
    await flushPromises()
    expect(routing.currentRoute.value.query).toMatchObject({
      source: 'adzuna',
      filter: 'saved',
      q: 'Engineer',
    })
    await view.get('#search').setValue('Go')
    await flushPromises()
    await vi.advanceTimersByTimeAsync(300)
    expect(routing.currentRoute.value.query).toMatchObject({
      source: 'adzuna',
      filter: 'saved',
      q: 'Go',
    })
    expect(fetch.mock.calls[fetch.mock.calls.length - 1]![0]).toContain('source=adzuna')
    expect(fetch.mock.calls[fetch.mock.calls.length - 1]![0]).toContain('q=Go')
    await view
      .findAll('button')
      .find((button) => button.text() === 'All')!
      .trigger('click')
    await flushPromises()
    expect(routing.currentRoute.value.query.source).toBe('adzuna')
    expect(routing.currentRoute.value.query.q).toBe('Go')
    await routing.push('/jobs?source=remoteok')
    await vi.advanceTimersByTimeAsync(300)
    expect(fetch.mock.calls[fetch.mock.calls.length - 1]![0]).toContain('source=remoteok')
    view.unmount()
  })
})
