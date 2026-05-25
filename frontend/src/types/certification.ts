export interface Certification {
  id: string
  user_id: string
  certification_name: string
  issuing_organization: string
  issue_date: string | null
  expiration_date: string | null
  does_not_expire: boolean
  credential_id: string | null
  credential_url: string | null
  is_in_progress: boolean
  created_at?: string
  updated_at?: string
}

export interface CertificationInput {
  certification_name: string
  issuing_organization: string
  issue_date: string | null
  expiration_date: string | null
  does_not_expire: boolean
  credential_id: string | null
  credential_url: string | null
  is_in_progress: boolean
}
