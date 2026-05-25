export interface Education {
  id: string
  user_id: string
  institution_name: string
  degree_type: string | null
  degree_name: string | null
  field_of_study: string | null
  start_date: string | null
  end_date: string | null
  is_current: boolean
  description: string | null
  is_highlighted: boolean
  created_at?: string
  updated_at?: string
}

export interface EducationInput {
  institution_name: string
  degree_type: string | null
  degree_name: string | null
  field_of_study: string | null
  start_date: string | null
  end_date: string | null
  is_current: boolean
  description: string | null
}
