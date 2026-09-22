// Export only explicitly supported profile fields, even if API responses grow later.
type Shape = { [key: string]: true | Shape }
function fields(names: string): Shape {
  return Object.fromEntries(names.split(' ').map((name) => [name, true]))
}
const schemas: Record<string, { path: string; shape: Shape; list?: boolean }> = {
  profile: {
    path: '/api/users/me',
    shape: fields(
      'username full_name email avatar_url phone user_location website_url linkedin_url github_url headline user_summary availability years_of_experience',
    ),
  },
  preferences: {
    path: '/api/profile/preferences',
    shape: {
      ...fields(
        'min_salary max_salary salary_currency preferred_job_types preferred_industries company_stage_preference notify_jobs',
      ),
      skills: fields(
        'skill_name skill_category proficiency years_experience is_featured display_order',
      ),
      desired_locations: fields('location_name is_remote priority'),
      desired_roles: fields('role_title priority'),
    },
  },
  matching: {
    path: '/api/profile/matching',
    shape: fields(
      'min_score work_mode include_unknown search_country email_enabled notifications_enabled digest_timezone digest_hours digest_max_jobs',
    ),
  },
  experiences: {
    path: '/api/profile/experiences',
    list: true,
    shape: {
      ...fields(
        'company_name company_url role_title exp_location industry employment_type description achievements start_date end_date is_current',
      ),
      skills: fields('skill_name category'),
    },
  },
  educations: {
    path: '/api/profile/educations',
    list: true,
    shape: fields(
      'institution_name degree_type degree_name field_of_study start_date end_date is_current description is_highlighted',
    ),
  },
  projects: {
    path: '/api/profile/projects',
    list: true,
    shape: fields(
      'title role_title description impact project_url repository_url start_date end_date is_current is_featured',
    ),
  },
  certifications: {
    path: '/api/profile/certifications',
    list: true,
    shape: fields(
      'certification_name issuing_organization issue_date expiration_date does_not_expire credential_id credential_url is_in_progress',
    ),
  },
  languages: {
    path: '/api/profile/languages',
    list: true,
    shape: fields('user_language proficiency'),
  },
}
const scalar = (value: unknown) =>
  value === null || ['string', 'number', 'boolean'].includes(typeof value)
function select(value: unknown, shape: Shape): Record<string, unknown> {
  if (!value || typeof value !== 'object' || Array.isArray(value))
    throw new Error('Unexpected profile response. Please retry.')
  const input = value as Record<string, unknown>
  const output: Record<string, unknown> = {}
  for (const [key, rule] of Object.entries(shape)) {
    const item = input[key]
    if (item === undefined) continue
    if (rule === true) {
      if (scalar(item)) output[key] = item
      else if (Array.isArray(item) && item.every(scalar)) output[key] = item
    } else if (Array.isArray(item)) output[key] = item.map((row) => select(row, rule))
    else if (item === null) output[key] = []
  }
  return output
}
export async function fetchProfileDownload() {
  const sections = await Promise.all(
    Object.entries(schemas).map(async ([name, { path, shape, list }]) => {
      const response = await fetch(path, { credentials: 'include', cache: 'no-store' })
      if (!response.ok)
        throw new Error(
          response.status === 401
            ? 'Please sign in again before downloading your profile.'
            : 'Could not download your complete profile. Please retry.',
        )
      const data: unknown = await response.json()
      if (list && data !== null && !Array.isArray(data))
        throw new Error('Unexpected profile response. Please retry.')
      return [
        name,
        list ? ((data ?? []) as unknown[]).map((row) => select(row, shape)) : select(data, shape),
      ]
    }),
  )
  return {
    format_version: 1,
    exported_at: new Date().toISOString(),
    ...Object.fromEntries(sections),
  }
}
