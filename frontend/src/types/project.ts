export interface Project {
  id: string
  user_id: string
  title: string
  role_title: string | null
  description: string | null
  impact: string | null
  project_url: string | null
  repository_url: string | null
  start_date: string | null
  end_date: string | null
  is_current: boolean
  is_featured: boolean
  created_at?: string
  updated_at?: string
}

export interface ProjectInput {
  title: string
  role_title: string | null
  description: string | null
  impact: string | null
  project_url: string | null
  repository_url: string | null
  start_date: string | null
  end_date: string | null
  is_current: boolean
  is_featured: boolean
}
