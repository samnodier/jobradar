/**
 * Ensures a URL string has an absolute protocol prefix.
 * Prevents the browser from treating bare domains (e.g. "github.com/sam")
 * as relative paths (e.g. "localhost:5173/github.com/sam").
 */
export function ensureAbsoluteUrl(url: string): string {
  if (!url) return url
  if (/^https?:\/\//i.test(url)) return url
  return `https://${url}`
}
