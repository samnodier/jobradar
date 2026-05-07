export const statusLabels: Record<string, string> = {
  applied: 'Applied',
  recruiter_screen: 'Recruiter Screen',
  interview: 'Interview',
  offer: 'Offer',
  accepted: 'Accepted',
  rejected: 'Rejected',
  withdrawn: 'Withdrawn',
  other: 'Other',
}

export const statusColors: Record<string, string> = {
  applied: '#dbeafe',
  recruiter_screen: '#e0e7ff',
  interview: '#fef9c3',
  offer: '#dcfce7',
  accepted: '#bbf7d0',
  rejected: '#fee2e2',
  withdrawn: '#f3f4f6',
  other: '#d3d3d3',
}

export const statusOrder = [
  'applied',
  'recruiter_screen',
  'interview',
  'offer',
  'accepted',
  'rejected',
  'withdrawn',
  'other',
]
