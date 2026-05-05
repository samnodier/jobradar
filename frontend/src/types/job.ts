export interface Job {
  id: string
  external_id: string
  job_source: string
  title: string
  company_name: string
  description?: string | null
  source_url: string
  salary_min?: number | null
  salary_max?: number | null
  currency?: string | null
  job_location?: string | null
  is_remote?: boolean | null
  job_status?: string | null
  employment_type?: string | null
  experience_level?: string | null
  skills: string[]
  posted_at?: string | null
  expires_at?: string | null
  created_at?: string | null
  updated_at?: string | null
  logo_url?: string | null
  is_saved: boolean
  is_applied: boolean
}
