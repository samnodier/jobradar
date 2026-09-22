import { afterEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import DownloadProfile from '../components/DownloadProfile.vue'
import { fetchProfileDownload } from '../utils/profileDownload'

afterEach(() => {
  vi.restoreAllMocks()
  vi.unstubAllGlobals()
  vi.useRealTimers()
})
function mockProfile(failed = '') {
  return vi.fn(async (path: string) => ({
    ok: path !== failed,
    status: path === failed ? 401 : 200,
    json: async () =>
      path === '/api/users/me'
        ? {
            full_name: 'Sam',
            email: 'sam@example.test',
            password_hash: 'secret',
            access_token: 'secret',
          }
        : path.endsWith('/preferences')
          ? {
              desired_locations: [{ location_name: 'Nairobi, Kenya', secret: 'secret' }],
              skills: [{ skill_name: 'Go', secret: 'secret' }],
            }
          : path.endsWith('/matching')
            ? { min_score: 70, encrypted_key: 'secret' }
            : path.endsWith('/experiences')
              ? [
                  {
                    company_name: 'Example',
                    skills: [{ skill_name: 'Go', secret: 'secret' }],
                    user_id: 'private-id',
                  },
                ]
              : [],
  }))
}
describe('profile download', () => {
  it('selects allowed fields recursively and authenticates every request', async () => {
    const fetcher = mockProfile()
    vi.stubGlobal('fetch', fetcher)
    const data = await fetchProfileDownload()
    expect(data.profile).toEqual({ full_name: 'Sam', email: 'sam@example.test' })
    expect(JSON.stringify(data)).not.toContain('secret')
    expect(JSON.stringify(data)).not.toContain('private-id')
    expect(data.preferences.desired_locations).toEqual([{ location_name: 'Nairobi, Kenya' }])
    expect(fetcher).toHaveBeenCalledTimes(8)
    expect(fetcher).toHaveBeenCalledWith('/api/users/me', {
      credentials: 'include',
      cache: 'no-store',
    })
  })
  it('downloads JSON and releases its object URL', async () => {
    vi.useFakeTimers()
    vi.stubGlobal('fetch', mockProfile())
    const create = vi.fn(() => 'blob:profile')
    const revoke = vi.fn()
    vi.stubGlobal('URL', { createObjectURL: create, revokeObjectURL: revoke })
    const click = vi.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(() => {})
    const wrapper = mount(DownloadProfile)
    await wrapper.get('button').trigger('click')
    await flushPromises()
    expect(click).toHaveBeenCalledOnce()
    expect(create).toHaveBeenCalledWith(expect.any(Blob))
    expect(wrapper.get('[role="status"]').text()).toContain('prepared')
    await vi.advanceTimersByTimeAsync(1000)
    expect(revoke).toHaveBeenCalledWith('blob:profile')
    wrapper.unmount()
  })
  it('does not download partial results and allows retry after session failure', async () => {
    vi.stubGlobal('fetch', mockProfile('/api/profile/experiences'))
    const click = vi.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(() => {})
    const wrapper = mount(DownloadProfile)
    await wrapper.get('button').trigger('click')
    await flushPromises()
    expect(wrapper.get('[role="alert"]').text()).toContain('sign in again')
    expect(click).not.toHaveBeenCalled()
    expect(wrapper.get('button').attributes('disabled')).toBeUndefined()
    wrapper.unmount()
  })
})
