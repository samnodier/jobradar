import { afterEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import PasswordSettings from '@/components/PasswordSettings.vue'
import AccessTokens from '@/components/AccessTokens.vue'
const response = (data: unknown) => ({ ok: true, json: async () => data })
afterEach(() => vi.unstubAllGlobals())
describe('Account settings', () => {
  it('validates confirmation and clears password fields after successful change', async () => {
    const settings = { email: 'test@example.com', password_configured: true, github_enabled: true }
    const fetch = vi.fn().mockResolvedValue(response(settings))
    vi.stubGlobal('fetch', fetch)
    const wrapper = mount(PasswordSettings)
    await flushPromises()
    const inputs = wrapper.findAll('input')
    await inputs[0]!.setValue('old-password-value')
    await inputs[1]!.setValue('new-password-value')
    await inputs[2]!.setValue('wrong')
    await wrapper.get('form').trigger('submit')
    await flushPromises()
    expect(wrapper.get('[role="alert"]').text()).toBe('Passwords do not match')
    expect(fetch).toHaveBeenCalledTimes(1)
    await inputs[2]!.setValue('new-password-value')
    await wrapper.get('form').trigger('submit')
    await flushPromises()
    expect(JSON.parse(fetch.mock.calls[1]![1].body)).toEqual({
      current_password: 'old-password-value',
      password: 'new-password-value',
    })
    expect(inputs.every((i) => (i.element as HTMLInputElement).value === '')).toBe(true)
    expect(wrapper.emitted('changed')).toHaveLength(1)
    expect(wrapper.get('button').classes()).toContain('button-primary')
    expect(wrapper.get('a').classes()).toContain('button-secondary')
  })
  it('creates a one-time token then revokes it and hides the secret', async () => {
    const token = {
      id: 'token-1',
      name: 'Personal workflow',
      expires_at: '2027-01-01',
      revoked_at: null,
    }
    const fetch = vi
      .fn()
      .mockResolvedValueOnce(response([]))
      .mockResolvedValueOnce(response({ token: 'one-time-test-secret' }))
      .mockResolvedValueOnce(response([token]))
      .mockResolvedValueOnce(response({}))
      .mockResolvedValueOnce(response([{ ...token, revoked_at: '2026-09-22' }]))
    vi.stubGlobal('fetch', fetch)
    const wrapper = mount(AccessTokens)
    await flushPromises()
    await wrapper.get('form').trigger('submit')
    await flushPromises()
    expect(JSON.parse(fetch.mock.calls[1]![1].body)).toEqual({
      name: 'Personal workflow',
      expires_days: 90,
    })
    expect((wrapper.get('textarea').element as HTMLTextAreaElement).value).toBe(
      'one-time-test-secret',
    )
    await wrapper
      .findAll('button')
      .find((b) => b.text() === 'Revoke')!
      .trigger('click')
    await flushPromises()
    expect(fetch.mock.calls[3]![0]).toBe('/api/profile/access-tokens/token-1')
    expect(fetch.mock.calls[3]![1].method).toBe('DELETE')
    expect(wrapper.find('textarea').exists()).toBe(false)
    expect(wrapper.text()).toContain('Revoked')
  })
  it('reports a failed token creation and permits retry', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValueOnce(response([])).mockResolvedValueOnce({ ok: false }),
    )
    const wrapper = mount(AccessTokens)
    await flushPromises()
    await wrapper.get('form').trigger('submit')
    await flushPromises()
    expect(wrapper.get('[role="alert"]').text()).toContain('Could not create token')
    expect(wrapper.get('button[type="submit"]').attributes('disabled')).toBeUndefined()
  })
})
