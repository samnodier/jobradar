# Session handoff — profile workflows

Sam wants small focused commits, direct implementation and maintained learning notes. The old unpushed release commit was replaced by 12 focused commits with an identical final tree. Backup branch backup/personal-release-4308bb7 preserves the old local tip; never push that backup. Implementation branch: feature/profile-workflows, based on dev. Integration branch: dev; inspect git status for the currently checked-out branch. No remote history has been changed.

New work: profile-owned score/work-mode/unknown/search-market/digest settings, eligibility gates, two daily digest slots, profile-driven Adzuna searches, best-match feed and profile tab, durable in-app notifications/deep links, authenticated Markdown Copy for AI, and structured/no-key URL import. See PLAN/SCHEMA/ROUTES/DECISIONS and docs/PERSONAL_WORKFLOW.md. New migrations through 20260920000500 are required.

Sam will fill target roles and locations in the profile. Secrets stay in .env. Email provider choice remains pending; Resend fits two daily personal digests, GHL requires the recipient contact ID as well as token. No actual email was sent. Existing development database has not been migrated automatically; all integration tests used isolated containers.

Sam wants to manually test dev before main promotion/push and deployment. Recommend the always-on home server over sleeping/expiring Render free services for this architecture. Do not remove local models or alter home-server services without identifying the deployment machine. Defer external autonomous-agent integration; the manual Markdown flow is implemented and requires no model API calls.
