import { afterEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import ApplicationsView from '@/views/ApplicationsView.vue'

const application = {
  id: 'app-1',
  application_status: 'applied',
  job_title: 'Backend Engineer',
  company_name: 'Example',
  notes: 'Existing note',
  updated_at: '2026-09-22T10:00:00Z',
}
const response = (data: unknown, ok = true) => ({ ok, json: async () => structuredClone(data) })
function render() {
  return mount(ApplicationsView, {
    global: {
      stubs: {
        AppSidebar: true,
        ApplicationDetail: true,
        RouterLink: { template: '<a><slot /></a>' },
      },
    },
  })
}
afterEach(() => {
  vi.unstubAllGlobals()
})

describe('application board', () => {
  it('persists an accessible status change and keeps the card details', async () => {
    const fetch = vi
      .fn()
      .mockResolvedValueOnce(response([application]))
      .mockResolvedValueOnce(
        response({ application_status: 'interview', updated_at: application.updated_at }),
      )
    vi.stubGlobal('fetch', fetch)
    const view = render()
    await flushPromises()
    await view.get('select').setValue('interview')
    await flushPromises()
    expect(fetch).toHaveBeenLastCalledWith(
      '/api/applications/app-1',
      expect.objectContaining({
        method: 'PATCH',
        credentials: 'include',
        body: JSON.stringify({ application_status: 'interview' }),
      }),
    )
    expect(view.get('[data-status="interview"]').text()).toContain('Backend Engineer')
    expect(view.get('[data-status="applied"]').find('[data-application-id]').exists()).toBe(false)
    expect(view.get('[role="status"]').text()).toContain('moved to Interview')
    view.unmount()
  })

  it('moves a dragged card and only allows one pending move per application', async () => {
    let finish!: (value: unknown) => void
    const fetch = vi
      .fn()
      .mockResolvedValueOnce(response([application]))
      .mockImplementationOnce(
        () =>
          new Promise((resolve) => {
            finish = resolve
          }),
      )
    vi.stubGlobal('fetch', fetch)
    const view = render()
    await flushPromises()
    await view
      .get('[data-application-id="app-1"]')
      .trigger('dragstart', { dataTransfer: { setData: vi.fn() } })
    await view.get('[data-status="offer"]').trigger('drop')
    expect(view.get('[data-status="offer"] select').attributes('disabled')).toBeDefined()
    expect(view.get('[data-status="offer"] [data-application-id]').attributes('draggable')).toBe(
      'false',
    )
    await view.get('select').trigger('change')
    expect(fetch).toHaveBeenCalledTimes(2)
    finish(response({ ...application, application_status: 'offer' }))
    await flushPromises()
    expect(view.get('select').attributes('disabled')).toBeUndefined()
    view.unmount()
  })

  it('restores the original column after a failed save while keeping the board available', async () => {
    vi.stubGlobal(
      'fetch',
      vi
        .fn()
        .mockResolvedValueOnce(response([application]))
        .mockResolvedValueOnce(response({}, false)),
    )
    const view = render()
    await flushPromises()
    await view.get('select').setValue('rejected')
    await flushPromises()
    expect(view.get('[data-status="applied"]').text()).toContain('Backend Engineer')
    expect(view.get('[role="alert"]').text()).toContain('Could not move')
    expect(view.get('select').attributes('disabled')).toBeUndefined()
    view.unmount()
  })

  it('ignores external drags and same-status changes', async () => {
    const fetch = vi.fn().mockResolvedValue(response([application]))
    vi.stubGlobal('fetch', fetch)
    const view = render()
    await flushPromises()
    await view
      .get('[data-status="offer"]')
      .trigger('drop', { dataTransfer: { getData: () => 'app-1' } })
    await view.get('select').setValue('applied')
    expect(fetch).toHaveBeenCalledTimes(1)
    expect(view.find('option[value="other"]').exists()).toBe(false)
    view.unmount()
  })

  it('handles HTTP failures without trying to render the error body as applications', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(response({ error: 'unauthorized' }, false)))
    const view = render()
    await flushPromises()
    expect(view.text()).toContain('Failed to fetch applications')
    expect(view.find('[data-status]').exists()).toBe(false)
    view.unmount()
  })
})

it('shows due follow-ups only for active applications and groups current outcomes', async () => {
  vi.stubGlobal(
    'fetch',
    vi.fn().mockResolvedValue(
      response([
        {
          ...application,
          follow_up_at: '2000-01-01T00:00:00Z',
          job_source: 'adzuna',
          match_score: 72,
        },
        {
          ...application,
          id: 'app-2',
          application_status: 'accepted',
          follow_up_at: '2000-01-01T00:00:00Z',
          job_source: 'adzuna',
          match_score: 72,
        },
        { ...application, id: 'app-3', follow_up_at: '2099-01-01T00:00:00Z' },
      ]),
    ),
  )
  const view = render()
  await flushPromises()
  expect(view.get('[aria-label="Due follow-ups"]').text()).toContain('Follow-ups due · 1')
  const rows = view.findAll('tbody tr')
  expect(
    rows
      .find((row) => row.text().includes('adzuna'))
      ?.findAll('td')
      .map((cell) => cell.text()),
  ).toEqual(['adzuna', '60–79', '2', '0', '1', '0'])
  view.unmount()
})
