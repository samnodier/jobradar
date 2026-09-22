import { afterEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createMemoryHistory, createRouter } from 'vue-router'
import JobsView from '@/views/JobsView.vue'
import JobPagination from '@/components/JobPagination.vue'

const response = (hasMore = false, jobs: unknown[] = []) => ({
  ok: true,
  headers: new Headers({ 'X-Has-More': String(hasMore) }),
  json: async () => jobs,
})
async function render(path = '/jobs') {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [{ path: '/jobs', component: { template: '<div />' } }],
  })
  await router.push(path)
  const view = mount(JobsView, {
    global: {
      plugins: [router],
      stubs: {
        AppSidebar: true,
        JobRow: { props: ['job'], template: '<div>{{ job.title }}</div>' },
        JobDetail: true,
        ImportJobPanel: true,
      },
    },
  })
  await flushPromises()
  return { view, router }
}
afterEach(() => {
  vi.useRealTimers()
  vi.unstubAllGlobals()
})

describe('job pagination', () => {
  it('sorts by score through the API and resets pagination without clearing filters', async () => {
    const fetcher = vi.fn().mockResolvedValue(response(true))
    vi.stubGlobal('fetch', fetcher)
    const { view, router } = await render('/jobs?source=remoteok&filter=saved&page=3')
    expect(view.get('[aria-label="Sort jobs"]').findAll('option')).toHaveLength(4)
    await view.get('[aria-label="Sort jobs"]').setValue('score_asc')
    await flushPromises()
    expect(router.currentRoute.value.query).toMatchObject({
      source: 'remoteok',
      filter: 'saved',
      order: 'score_asc',
    })
    expect(router.currentRoute.value.query.page).toBeUndefined()
    expect(fetcher.mock.lastCall![0]).toContain('order=score_asc')
    view.unmount()
  })

  it('exposes page boundaries and disables controls while loading', async () => {
    const view = mount(JobPagination, {
      props: { page: 1, pageSize: 50, hasNext: false, loading: false },
    })
    expect(
      view.findAll('button').every((button) => button.attributes('disabled') !== undefined),
    ).toBe(true)
    await view.setProps({ page: 2, hasNext: true })
    await view.findAll('button')[0]!.trigger('click')
    await view.findAll('button')[1]!.trigger('click')
    expect(view.emitted('page')).toEqual([[1], [3]])
    await view.get('[aria-label="Jobs per page"]').setValue('500')
    expect(view.emitted('pageSize')).toEqual([[500]])
    await view.setProps({ loading: true })
    expect(view.get('select').attributes('disabled')).toBeDefined()
    expect(
      view.findAll('button').every((button) => button.attributes('disabled') !== undefined),
    ).toBe(true)
    view.unmount()
  })

  it('uses only bottom controls and requests exact offsets without invented totals', async () => {
    const fetch = vi
      .fn()
      .mockResolvedValueOnce(response(true))
      .mockResolvedValueOnce(response(false))
    vi.stubGlobal('fetch', fetch)
    const { view, router } = await render('/jobs?source=remoteok&size=100')
    expect(view.findAll('nav[aria-label^="Job pages"]')).toHaveLength(1)
    expect(fetch.mock.calls[0]![0]).toContain('limit=100&offset=0')
    await view.get('nav[aria-label="Job pages"]').findAll('button')[1]!.trigger('click')
    await flushPromises()
    expect(router.currentRoute.value.query).toMatchObject({
      source: 'remoteok',
      size: '100',
      page: '2',
    })
    expect(fetch.mock.calls[1]![0]).toContain('limit=100&offset=100')
    expect(view.findAll('nav').every((nav) => nav.text().includes('Page 2'))).toBe(true)
    expect(view.get('nav').findAll('button')[1]!.attributes('disabled')).toBeDefined()
    expect(view.text()).not.toContain(' of ')
    view.unmount()
  })

  it('resets page after filters, search and page size, while preserving source', async () => {
    vi.useFakeTimers()
    const fetch = vi.fn().mockResolvedValue(response(true))
    vi.stubGlobal('fetch', fetch)
    const { view, router } = await render('/jobs?source=adzuna&page=4&size=200')
    await view
      .findAll('button')
      .find((button) => button.text() === 'Saved')!
      .trigger('click')
    await flushPromises()
    expect(router.currentRoute.value.query).toMatchObject({
      source: 'adzuna',
      filter: 'saved',
      size: '200',
    })
    expect(router.currentRoute.value.query.page).toBeUndefined()
    await router.push('/jobs?source=adzuna&page=4&size=200')
    await flushPromises()
    await view.get('#search').setValue('Go')
    await vi.advanceTimersByTimeAsync(300)
    await flushPromises()
    expect(router.currentRoute.value.query).toMatchObject({ source: 'adzuna', q: 'Go' })
    expect(router.currentRoute.value.query.page).toBeUndefined()
    await router.push('/jobs?source=adzuna&page=4&size=200')
    await flushPromises()
    await view.get('[aria-label="Jobs per page"]').setValue('500')
    await flushPromises()
    expect(router.currentRoute.value.query).toMatchObject({ source: 'adzuna', size: '500' })
    expect(router.currentRoute.value.query.page).toBeUndefined()
    view.unmount()
  })

  it('ignores an older response even when the transport does not honor abort', async () => {
    let finish!: (value: unknown) => void
    const fetch = vi
      .fn()
      .mockImplementationOnce(
        () =>
          new Promise((resolve) => {
            finish = resolve
          }),
      )
      .mockResolvedValue(response(false, [{ id: 'new', title: 'Current job', created_at: null }]))
    vi.stubGlobal('fetch', fetch)
    const { view, router } = await render()
    await router.push('/jobs?page=2')
    await flushPromises()
    finish(response(true, [{ id: 'old', title: 'Stale job', created_at: null }]))
    await flushPromises()
    expect(view.text()).toContain('Current job')
    expect(view.text()).not.toContain('Stale job')
    expect(view.get('nav').findAll('button')[1]!.attributes('disabled')).toBeDefined()
    view.unmount()
  })

  it('does not filter or reorder a server page locally and shows upgrade guidance for an old backend', async () => {
    vi.stubGlobal(
      'fetch',
      vi
        .fn()
        .mockResolvedValueOnce(
          response(false, [
            {
              id: 'one',
              title: 'Server-selected result',
              job_source: 'RemoteOK',
              is_remote: false,
              created_at: null,
            },
          ]),
        )
        .mockResolvedValueOnce({ ok: true, json: async () => [] }),
    )
    const { view } = await render('/jobs?source=remoteok&filter=remote&q=Engineer')
    expect(view.text()).toContain('Server-selected result')
    await view
      .findAll('button')
      .find((button) => button.text() === 'Refresh')!
      .trigger('click')
    await flushPromises()
    expect(view.text()).toContain('Restart the Go server with make run')
    expect(view.text()).not.toContain('Server-selected result')
    view.unmount()
  })

  it('handles failed page loads without showing the prior page as current', async () => {
    vi.stubGlobal(
      'fetch',
      vi
        .fn()
        .mockResolvedValueOnce(
          response(true, [{ id: 'one', title: 'First page result', created_at: null }]),
        )
        .mockResolvedValueOnce({ ok: false }),
    )
    const { view } = await render()
    await view.get('nav').findAll('button')[1]!.trigger('click')
    await flushPromises()
    expect(view.text()).toContain('Failed to load jobs')
    expect(view.text()).not.toContain('First page result')
    expect(view.get('nav').findAll('button')[0]!.attributes('disabled')).toBeUndefined()
    view.unmount()
  })
})
