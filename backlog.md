# Jobradar backlog

Current personal-release implementation and verification: [.context/PLAN.md](.context/PLAN.md).

Before deployment: finish live account configuration, confirm searches and recipient, preview/send a real digest, HTTPS/OAuth setup, backup/restore verification.

After personal v1:
- Improve relevance using reviewed examples; explicit geographic/work-authorization and salary eligibility rules.
- Track source disappearance correctly (full board feed versus paginated search).
- Add Ashby/Lever adapters as needed; cross-source deduplication.
- Cursor pagination if concurrent ingestion makes offset pagination insufficient.
- Version AI enrichment so stale provider responses cannot overwrite a newer profile summary.
- Add reconciliation UI for failed/uncertain email only if manual provider checks become frequent.
- Resume import and tailored resume PDF, then optional Discord and public multi-user features.

Avoid rebuilding completed CRUD or AI plumbing from the older roadmap. Learning questions are maintained directly in LEARNINGS.md.
