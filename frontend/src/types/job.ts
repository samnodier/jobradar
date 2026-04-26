export interface Job {
  id: string
  external_id: string
  source: string
  title: string
  company: string
  description?: string | null
  url: string
  salary_min?: number | null
  salary_max?: number | null
  currency?: string | null
  location?: string | null
  is_remote?: boolean | null
  status?: string | null
  employment_type?: string | null
  experience_level?: string | null
  skills: string[]
  posted_at?: string | null
  expires_at?: string | null
  created_at?: string | null
  updated_at?: string | null
  logo_url?: string | null
}
