export interface SkillPreference {
  skill_name: string
  skill_category: string | null
  proficiency: string | null
  years_experience: number | null
  is_featured: boolean | null
  endorsed_by_ai: boolean | null
  display_order: number | null
}

export interface DesiredLocationPreference {
  location_name: string
  is_remote: boolean | null
  priority: number | null
}

export interface DesiredRolePreference {
  role_title: string
  priority: number | null
}

export interface UserPreferences {
  min_salary: number | null
  max_salary: number | null
  salary_currency: string | null
  preferred_job_types: string[]
  preferred_industries: string[]
  company_stage_preference: string[]
  notify_jobs: boolean
  skills: SkillPreference[]
  desired_locations: DesiredLocationPreference[]
  desired_roles: DesiredRolePreference[]
}
