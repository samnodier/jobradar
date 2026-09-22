import { afterEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import StatusView from '@/views/StatusView.vue'

const health = {
  database: true,
  redis: false,
  pending_matches: 3,
  sources: [{ name: 'remoteok', status: 'ok', last_success: null }],
}
const response = (ok = true) => ({ ok, json: async () => health })
function render() {
  return mount(StatusView, { global: { stubs: { AppSidebar: true } } })
}
afterEach(() => {
  vi.useRealTimers()
  vi.unstubAllGlobals()
})

describe('system health', () => {
  it('renders dependency state and stored source history, with manual refresh and last checked time', async () => {
    const fetch = vi.fn().mockResolvedValue(response())
    vi.stubGlobal('fetch', fetch)
    const view = render()
    await view.vm.$nextTick()
    expect(view.get('button').attributes('disabled')).toBeDefined()
    await flushPromises()
    expect(view.text()).toContain('Connected')
    expect(view.text()).toContain('Unavailable')
    expect(view.text()).toContain('Pending matches')
    expect(view.text()).toContain('remoteok')
    expect(view.text()).toContain('not a live uptime check')
    expect(view.get('[role="status"]').text()).toContain('Last checked:')
    await view.get('button').trigger('click')
    await flushPromises()
    expect(fetch).toHaveBeenCalledTimes(2)
    view.unmount()
  })

  it('polls every 30 seconds without overlapping requests and aborts on unmount', async () => {
    vi.useFakeTimers()
    let finish!: (value: unknown) => void
    const fetch = vi
      .fn()
      .mockImplementationOnce(
        () =>
          new Promise((resolve) => {
            finish = resolve
          }),
      )
      .mockResolvedValue(response())
    vi.stubGlobal('fetch', fetch)
    const view = render()
    const initialSignal = fetch.mock.calls[0]![1].signal as AbortSignal
    await vi.advanceTimersByTimeAsync(60000)
    expect(fetch).toHaveBeenCalledTimes(1)
    finish(response())
    await flushPromises()
    await vi.advanceTimersByTimeAsync(30000)
    expect(fetch).toHaveBeenCalledTimes(2)
    const latestSignal = fetch.mock.calls[1]![1].signal as AbortSignal
    view.unmount()
    expect(initialSignal.aborted).toBe(false)
    expect(latestSignal.aborted).toBe(true)
    await vi.advanceTimersByTimeAsync(60000)
    expect(fetch).toHaveBeenCalledTimes(2)
  })

  it('keeps the last successful result clearly marked stale when a refresh fails and allows retry', async () => {
    vi.stubGlobal(
      'fetch',
      vi
        .fn()
        .mockResolvedValueOnce(response())
        .mockResolvedValueOnce(response(false))
        .mockResolvedValueOnce(response()),
    )
    const view = render()
    await flushPromises()
    const checked = view.get('[role="status"]').text()
    await view.get('button').trigger('click')
    await flushPromises()
    expect(view.get('[role="alert"]').text()).toContain('Showing the last successful check')
    expect(view.text()).toContain('remoteok')
    expect(view.get('[role="status"]').text()).toBe(checked)
    expect(view.get('button').attributes('disabled')).toBeUndefined()
    await view.get('button').trigger('click')
    await flushPromises()
    expect(view.find('[role="alert"]').exists()).toBe(false)
    view.unmount()
  })
})
