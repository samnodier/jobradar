export interface Skill {
  id: string
  name: string
}

export interface Experience {
  id: string
  user_id: string
  company_name: string
  company_url: string | null
  role_title: string
  exp_location: string | null
  industry: string | null
  employment_type: string | null
  description: string | null
  achievements: string[] | null
  start_date: string
  end_date: string | null
  is_current: boolean
  skills: Skill[]
  created_at?: string
  updated_at?: string
}
