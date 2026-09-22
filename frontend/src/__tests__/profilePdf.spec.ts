import { afterEach, describe, expect, it, vi } from 'vitest'
import { printProfile, resumeHtml, type ResumeData } from '@/utils/profilePdf'
const data: ResumeData = {
  user: {
    full_name: 'Sam <script>alert(1)</script>',
    username: 'sam',
    email: 'sam@example.test',
    min_salary: 123456,
    configured_providers: ['private-provider'],
  } as ResumeData['user'],
  experiences: [],
  educations: [],
  projects: [],
  certifications: [],
  languages: [],
}
afterEach(() => {
  vi.restoreAllMocks()
  vi.unstubAllGlobals()
})
describe('profile PDF', () => {
  it('escapes profile text and excludes settings and empty sections', () => {
    const html = resumeHtml(data)
    expect(html).toContain('&lt;script&gt;')
    expect(html).not.toContain('<script>')
    expect(html).toContain('sam@example.test')
    expect(html).not.toContain('123456')
    expect(html).not.toContain('private-provider')
    expect(html).not.toContain('<h2>Experience</h2>')
  })
  it('loads all saved sections before printing and handles null lists', async () => {
    const preview = {
      opener: window,
      closed: false,
      document: { body: { textContent: '' }, open: vi.fn(), write: vi.fn(), close: vi.fn() },
      focus: vi.fn(),
      print: vi.fn(),
      close: vi.fn(),
    }
    vi.spyOn(window, 'open').mockReturnValue(preview as unknown as Window)
    const fetcher = vi
      .fn()
      .mockResolvedValueOnce({ ok: true, json: async () => data.user })
      .mockResolvedValue({ ok: true, json: async () => null })
    vi.stubGlobal('fetch', fetcher)
    await printProfile()
    expect(fetcher).toHaveBeenCalledTimes(6)
    expect(preview.document.write).toHaveBeenCalledWith(expect.stringContaining('sam@example.test'))
    expect(preview.print).toHaveBeenCalledOnce()
    expect(preview.opener).toBeNull()
  })
  it('does not print a partial resume after a failed request', async () => {
    const preview = { document: { body: { textContent: '' } }, print: vi.fn(), close: vi.fn() }
    vi.spyOn(window, 'open').mockReturnValue(preview as unknown as Window)
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: false }))
    await expect(printProfile()).rejects.toThrow('complete profile')
    expect(preview.print).not.toHaveBeenCalled()
    expect(preview.close).toHaveBeenCalledOnce()
  })
  it('explains a blocked popup', async () => {
    vi.spyOn(window, 'open').mockReturnValue(null)
    await expect(printProfile()).rejects.toThrow('Allow popups')
  })
})
