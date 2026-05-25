export interface Language {
  user_id: string
  user_language: string
  proficiency: string | null // e.g., 'native', 'fluent', professional elementary
  created_at?: string
  updated_at?: string
}

export interface LanguageInput {
  user_language: string
  proficiency: string | null
}
