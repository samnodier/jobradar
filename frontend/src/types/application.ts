export interface Application {
  id: string
  user_id: string
  job_id: string
  status: 'saved' | 'applied' | 'interview' | 'offer' | 'rejected'
  applied_at: string | null
  last_status_changed_at: string
  follow_up_at: string | null
  notes: string | null
  created_at: string
  updated_at: string
  job_title: string
  job_company: string
  job_location: string | null
  job_url: string
  job_is_remote: boolean
  job_logo_url: string | null
}
