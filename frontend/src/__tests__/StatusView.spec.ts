import { afterEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import StatusView from '@/views/StatusView.vue'
const snapshot = {
  database: true,
  redis: true,
  pending_matches: 3,
  sources: [
    {
      name: 'adzuna',
      status: 'success',
      last_success: '2026-09-22T10:00:00Z',
      last_checked: '2026-09-22T12:00:00Z',
      history: [
        { day: '2026-09-22', successes: 3, failures: 1 },
        { day: '2026-09-21', successes: 1, failures: 0 },
      ],
    },
  ],
}
function render() {
  return mount(StatusView, { global: { stubs: { AppSidebar: true } } })
}
afterEach(() => {
  vi.useRealTimers()
  vi.unstubAllGlobals()
})
describe('system health', () => {
  it('shows observed rates and 30 accessible day cells without inventing history', async () => {
    vi.useFakeTimers()
    vi.setSystemTime(new Date('2026-09-22T14:00:00Z'))
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: true, json: async () => snapshot }))
    const view = render()
    await flushPromises()
    expect(view.text()).toContain('80.0% fetch success · 5 recorded attempts')
    expect(view.findAll('.history-cell')).toHaveLength(30)
    expect(view.findAll('.history-unknown')).toHaveLength(29) // 28 days plus legend
    expect(view.get('.history-mixed[role="img"]').attributes('aria-label')).toBe(
      '2026-09-22 UTC: 3 successful, 1 failed fetches',
    )
    expect(view.get('details').text()).toContain('Last attempt:')
    expect(view.text()).toContain('latest fetches successful')
    view.unmount()
  })
  it('reports source failures and unknown collection history separately', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({
        ok: true,
        json: async () => ({
          ...snapshot,
          sources: [{ ...snapshot.sources[0], status: 'failing', history: [] }],
        }),
      }),
    )
    const view = render()
    await flushPromises()
    expect(view.get('[role="status"]').text()).toContain('need attention')
    expect(view.text()).toContain('No recorded attempts in this period')
    expect(view.text()).not.toContain('100.0%')
    view.unmount()
  })
  it('preserves but labels the old snapshot after refresh failure and retries', async () => {
    const fetcher = vi
      .fn()
      .mockResolvedValueOnce({ ok: true, json: async () => snapshot })
      .mockResolvedValueOnce({ ok: false })
      .mockResolvedValue({ ok: true, json: async () => snapshot })
    vi.stubGlobal('fetch', fetcher)
    const view = render()
    await flushPromises()
    await view.get('button').trigger('click')
    await flushPromises()
    expect(view.get('[role="alert"]').text()).toContain('last successful snapshot')
    expect(view.get('[role="status"]').text()).toContain('out of date')
    await view.get('button').trigger('click')
    await flushPromises()
    expect(view.find('[role="alert"]').exists()).toBe(false)
    view.unmount()
  })
  it('polls, avoids overlapping requests, and aborts on unmount', async () => {
    vi.useFakeTimers()
    const fetcher = vi
      .fn()
      .mockResolvedValueOnce({ ok: true, json: async () => ({ ...snapshot, sources: [] }) })
      .mockImplementation(() => new Promise(() => {}))
    vi.stubGlobal('fetch', fetcher)
    const view = render()
    await flushPromises()
    expect(view.get('[role="status"]').text()).toContain('incomplete')
    await vi.advanceTimersByTimeAsync(30000)
    expect(fetcher).toHaveBeenCalledTimes(2)
    await vi.advanceTimersByTimeAsync(90000)
    expect(fetcher).toHaveBeenCalledTimes(2)
    expect(view.get('[role="status"]').text()).toContain('out of date')
    const signal = fetcher.mock.calls[1]![1].signal as AbortSignal
    view.unmount()
    expect(signal.aborted).toBe(true)
    await vi.advanceTimersByTimeAsync(30000)
    expect(fetcher).toHaveBeenCalledTimes(2)
  })
})
