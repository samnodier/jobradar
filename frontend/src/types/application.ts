export interface Application {
  id: string
  user_id: string
  job_id: string
  application_status: string
  applied_at: string | null
  last_status_changed_at: string
  follow_up_at: string | null
  notes: string | null
  created_at: string
  updated_at: string
  job_title: string
  company_name: string
  job_location: string | null
  source_url: string
  job_is_remote: boolean | null
  job_logo_url: string | null
}
