export interface User {
  id: string
  username: string
  full_name: string | null
  email: string
  avatar_url: string | null
  phone: string | null
  user_location: string | null
  website_url: string | null
  linkedin_url: string | null
  github_url: string | null
  headline: string | null
  user_summary: string | null
  availability: string | null
  min_salary: number | null
  max_salary: number | null
  salary_currency: string | null
  years_of_experience: number | null
  preferred_job_types: string[]
  preferred_industries: string[]
  company_stage_preference: string[]
  notify_jobs: boolean
  is_admin: boolean | null
  configured_providers: string[]
}
