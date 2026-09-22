import { afterEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import ApplicationDetail from '@/components/ApplicationDetail.vue'
import type { Application } from '@/types/application'

vi.mock('@/composables/useToast', () => ({ useToast: () => ({ error: vi.fn() }) }))
const application = (id: string) =>
  ({
    id,
    job_title: id,
    company_name: 'Example',
    application_status: 'applied',
    notes: '',
    updated_at: '2026-09-22T10:00:00Z',
  }) as Application
function render() {
  return mount(ApplicationDetail, {
    props: { app: application('first') },
    global: { stubs: { VueDatePicker: true } },
  })
}
afterEach(() => {
  vi.useRealTimers()
  vi.unstubAllGlobals()
})

describe('application detail notes', () => {
  it('saves pending notes against the original application when selection changes', async () => {
    vi.useFakeTimers()
    const fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ ...application('first'), notes: 'Interview notes' }),
    })
    vi.stubGlobal('fetch', fetch)
    const view = render()
    await view.get('textarea').setValue('Interview notes')
    await view.setProps({ app: application('second') })
    await flushPromises()
    expect(fetch).toHaveBeenCalledTimes(1)
    expect(fetch).toHaveBeenCalledWith(
      '/api/applications/first',
      expect.objectContaining({ body: expect.stringContaining('Interview notes') }),
    )
    expect((view.get('textarea').element as HTMLTextAreaElement).value).toBe('')
    await vi.advanceTimersByTimeAsync(2000)
    expect(fetch).toHaveBeenCalledTimes(1)
    view.unmount()
  })

  it('flushes notes when the detail panel closes and only emits changed fields', async () => {
    vi.useFakeTimers()
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({
        ok: true,
        json: async () => ({ ...application('first'), notes: 'Saved on close' }),
      }),
    )
    const view = render()
    await view.get('textarea').setValue('Saved on close')
    view.unmount()
    await flushPromises()
    expect(fetch).toHaveBeenCalledWith(
      '/api/applications/first',
      expect.objectContaining({ body: expect.stringContaining('Saved on close') }),
    )
  })

  it('serializes notes writes so an older request cannot overwrite newer text', async () => {
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
      .mockResolvedValue({
        ok: true,
        json: async () => ({ ...application('first'), notes: 'Newest' }),
      })
    vi.stubGlobal('fetch', fetch)
    const view = render()
    await view.get('textarea').setValue('Earlier')
    await vi.advanceTimersByTimeAsync(2000)
    await view.get('textarea').setValue('Newest')
    await vi.advanceTimersByTimeAsync(2000)
    expect(fetch).toHaveBeenCalledTimes(1)
    finish({ ok: true, json: async () => ({ ...application('first'), notes: 'Earlier' }) })
    await flushPromises()
    expect(fetch).toHaveBeenCalledTimes(2)
    expect(fetch.mock.calls[1]![1].body).toContain('Newest')
    const updates = view.emitted('updated')!
    expect(updates[0]![0]).not.toHaveProperty('application_status')
    expect(updates[1]![0]).toMatchObject({ notes: 'Newest' })
    view.unmount()
  })
})
