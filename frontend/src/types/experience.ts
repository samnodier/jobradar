export interface Skill {
  id: string
  skill_name: string
  category: string | null
  created_at?: string
  updated_at?: string
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

export interface ExperienceSkillInput {
  name: string
}

// Represents the exact shape expected by the Go POST/PATCH API
export interface ExperienceInput {
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
  skills: ExperienceSkillInput[]
}
