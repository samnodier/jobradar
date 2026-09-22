import type { User } from '@/types/user'
import type { Experience } from '@/types/experience'
import type { Education } from '@/types/education'
import type { Project } from '@/types/project'
import type { Certification } from '@/types/certification'
import type { Language } from '@/types/language'

export interface ResumeData {
  user: User
  experiences: Experience[]
  educations: Education[]
  projects: Project[]
  certifications: Certification[]
  languages: Language[]
}
const escape = (value: unknown) =>
  String(value ?? '').replace(
    /[&<>"']/g,
    (c) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' })[c]!,
  )
const text = (value: unknown) => (value ? `<p>${escape(value)}</p>` : '')
const date = (value: string | null) => (value ? value.slice(0, 7) : '')
const period = (start: string | null, end: string | null, current: boolean) =>
  [date(start), current ? 'Present' : date(end)].filter(Boolean).join(' – ')
const section = (name: string, entries: string[]) =>
  entries.length
    ? `<section><h2>${name}</h2>${entries.map((entry) => `<article>${entry}</article>`).join('')}</section>`
    : ''

export function resumeHtml(data: ResumeData): string {
  const { user, experiences, educations, projects, certifications, languages } = data
  const name = user.full_name || user.username
  return `<!doctype html><html lang="en"><head><meta charset="utf-8"><title>${escape(name)} — Resume</title>
  <style>
  @page { size: auto; margin: 18mm; }
  body { max-width: 760px; margin: 32px auto; padding: 0 20px; font: 11pt/1.5 system-ui, sans-serif; color: #172033; }
  h1 { font-size: 25pt; margin: 0; } h2 { font-size: 13pt; border-bottom: 1px solid #cbd5e1; padding-bottom: 4px; margin: 22px 0 10px; break-after: avoid; }
  h3 { font-size: 11pt; margin: 0; break-after: avoid; } p { margin: 4px 0; white-space: pre-wrap; overflow-wrap: anywhere; }
  article { margin: 0 0 14px; } ul { margin: 6px 0; padding-left: 20px; } li { white-space: pre-wrap; }
  .help { background: #f1f5f9; padding: 12px; margin-bottom: 24px; }
  @media print { body { max-width: none; margin: 0; padding: 0; } .help { display: none; } }
  </style></head><body>
  <p class="help">Use your browser’s Print command and select Save as PDF. You can close this tab when finished.</p>
  <header><h1>${escape(name)}</h1>${text(user.headline)}${text([user.email, user.phone, user.user_location].filter(Boolean).join(' · '))}${text([user.website_url, user.linkedin_url, user.github_url].filter(Boolean).join(' · '))}</header>
  ${user.user_summary ? section('Summary', [text(user.user_summary)]) : ''}
  ${section(
    'Experience',
    experiences.map(
      (e) =>
        `<h3>${escape(e.role_title)} · ${escape(e.company_name)}</h3>${text([period(e.start_date, e.end_date, e.is_current), e.exp_location].filter(Boolean).join(' · '))}${text(e.description)}${e.achievements?.length ? `<ul>${e.achievements.map((a) => `<li>${escape(a)}</li>`).join('')}</ul>` : ''}${text(e.skills?.map((s) => s.skill_name).join(' · '))}`,
    ),
  )}
  ${section(
    'Education',
    educations.map(
      (e) =>
        `<h3>${escape(e.institution_name)}</h3>${text([e.degree_name || e.degree_type, e.field_of_study].filter(Boolean).join(' · '))}${text(period(e.start_date, e.end_date, e.is_current))}${text(e.description)}`,
    ),
  )}
  ${section(
    'Projects',
    projects.map(
      (p) =>
        `<h3>${escape(p.title)}</h3>${text(p.role_title)}${text(period(p.start_date, p.end_date, p.is_current))}${text(p.description)}${text(p.impact)}${text([p.project_url, p.repository_url].filter(Boolean).join(' · '))}`,
    ),
  )}
  ${section(
    'Certifications',
    certifications.map(
      (c) =>
        `<h3>${escape(c.certification_name)}</h3>${text(c.issuing_organization)}${text([date(c.issue_date), c.is_in_progress ? 'In progress' : null, c.does_not_expire ? 'No expiration' : c.expiration_date ? `Expires ${date(c.expiration_date)}` : null].filter(Boolean).join(' · '))}${text(c.credential_id)}${text(c.credential_url)}`,
    ),
  )}
  ${section(
    'Languages',
    languages.map((l) => text([l.user_language, l.proficiency].filter(Boolean).join(' · '))),
  )}
  </body></html>`
}

export async function printProfile(): Promise<void> {
  // Open synchronously from the click so popup blockers can recognize user intent.
  const preview = window.open('', '_blank')
  if (!preview) throw new Error('Allow popups to open the PDF preview.')
  preview.opener = null
  preview.document.body.textContent = 'Preparing your résumé…'
  try {
    const paths = [
      '/api/users/me',
      '/api/profile/experiences',
      '/api/profile/educations',
      '/api/profile/projects',
      '/api/profile/certifications',
      '/api/profile/languages',
    ]
    const values = await Promise.all(
      paths.map(async (path) => {
        const response = await fetch(path, { credentials: 'include' })
        if (!response.ok) throw new Error('Could not load the complete profile. Please try again.')
        return response.json()
      }),
    )
    if (preview.closed) return
    const [user, experiences, educations, projects, certifications, languages] = values
    preview.document.open()
    preview.document.write(
      resumeHtml({
        user,
        experiences: experiences ?? [],
        educations: educations ?? [],
        projects: projects ?? [],
        certifications: certifications ?? [],
        languages: languages ?? [],
      }),
    )
    preview.document.close()
    await preview.document.fonts?.ready
    preview.focus()
    preview.print()
  } catch (error) {
    preview.close()
    throw error
  }
}
