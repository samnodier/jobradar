# Personal job export API

Create a token in Profile → Settings → Personal workflow access. The secret is shown once; it expires in 90 days and can be revoked there. It grants read access to public jobs, your own private jobs, your scores, and your saved resume/profile through AI context. It cannot change data or create other tokens. Do not commit the secret. Use HTTPS outside localhost.

A browser session also works on these routes. Account ownership always comes from authentication, never a supplied user ID.

```sh
# Set JOBRADAR_URL and JOBRADAR_TOKEN in your workflow's private configuration.
curl --get "$JOBRADAR_URL/api/export/jobs" \
  --header "Authorization: Bearer $JOBRADAR_TOKEN" \
  --data-urlencode 'min_score=70' \
  --data-urlencode 'date=2026-09-21' \
  --data-urlencode 'timezone=Africa/Nairobi' \
  --data-urlencode 'limit=100'
```

Response: `{ "jobs": [...], "next_cursor": "...", "min_score": 70, "date_field": "added", "start_at": "...", "end_at": "...", "eligible_only": true }`.

Repeat with identical filters and `cursor=<next_cursor>` until next_cursor is empty. There is no total-job cap; limit controls each page. Jobs include descriptions, source links, explicit metadata, professional-fit component scores, final match score and preference eligibility/notes. Missing AI enrichment does not exclude a job. Store job IDs in your workflow to avoid processing repeats. Pagination is not a frozen database snapshot: concurrent rescoring can move a job, so rerun/deduplicate if completeness during active updates matters.

| Parameter | Meaning / default |
| --- | --- |
| min_score | 0–100 inclusive; defaults to profile threshold. Does not change profile settings. |
| date | YYYY-MM-DD or today; omitted means all dates. Pin a literal date during pagination. |
| timezone | IANA zone; defaults to profile timezone. |
| date_field | added (first collected), posted (source date), matched (first scored); defaults to added. Unknown posting dates do not match a filtered posted day. |
| eligible_only | true by default; false includes preference mismatches at/above the requested score. |
| source | Exact source identifier, such as adzuna or greenhouse:board. |
| location | Case-insensitive substring of the saved job location; not geographic inference. |
| remote_only | false by default; true requires explicitly remote. |
| exclude_applied | false by default. |
| include_expired | false by default; only known expiry timestamps can be excluded. |
| limit | 1–100; default 50. |
| cursor | next_cursor from previous page. |

Pending recomputation is always excluded so stale scores are not exported. Asking for 70+ does not create new matches or change the matching algorithm's existing title gate.

For each selected job, retrieve Markdown for manual AI use or your external integration:

```sh
curl "$JOBRADAR_URL/api/export/jobs/$JOB_ID/ai-context" \
  --header "Authorization: Bearer $JOBRADAR_TOKEN"
```

This exports saved resume context along with the job; treat the returned document as personal information. Jobradar makes no AI request for either endpoint. Choosing a ChatGPT/Claude or other external integration is separate from authenticating to Jobradar.

Token administration requires browser-session authentication: GET/POST `/api/profile/access-tokens`, DELETE `/api/profile/access-tokens/{id}`. POST accepts `name` and `expires_days` (1–365, default 90). Lists never return the secret or stored hash.
