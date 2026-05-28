export function useJobFormatting() {
  function scoreBadgeClass(score: number) {
    if (score >= 70) return 'bg-green-100 text-green-800'
    if (score >= 40) return 'bg-yellow-100 text-yellow-800'
    return 'bg-gray-100 text-gray-800'
  }

  function timeAgo(dateStr: string | null | undefined): string {
    if (!dateStr) return ''
    const diff = Date.now() - new Date(dateStr).getTime()
    const hours = Math.floor(diff / 3600000)
    if (hours < 1) return 'just now'
    if (hours < 24) return `${hours}h ago`
    const days = Math.floor(hours / 24)
    return `${days}d ago`
  }
  function formatSalary(job: {
    salary_min?: number | null
    salary_max?: number | null
    currency?: string | null
  }) {
    const currency = job.currency || 'USD'
    return `${currency} ${job.salary_min ?? 0} – ${job.salary_max ?? 0}`
  }

  return { scoreBadgeClass, timeAgo, formatSalary }
}
