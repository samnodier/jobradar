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
  is_matched: boolean // true if this job has been through the matcher for this user
  match_score?: number | null // 0–100 final weighted score, null if not yet matched
  title_score?: number | null // raw Jaro-Winkler sub-score, null if not matched
  skill_score?: number | null // raw skill coverage sub-score, null if not matched
  experience_score?: number | null // raw experience overlap sub-score, null if not matched
  matched_skills?: string[] | null
  missing_skills?: string[] | null
  ai_summary?: string | null // null until LLM enrichment runs
  is_enriched?: boolean | null // false = algorithmic only, true = LLM has run
  match_created_at?: string | null
  match_updated_at?: string | null
}
