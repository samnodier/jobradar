# Personal hosting options

Reviewed September 22, 2026. The original deployment decision remains the always-on home server (see `.context/PLAN.md` and `PERSONAL_WORKFLOW.md`). Deployment has not happened simply because these instructions exist.

## Recommended zero-hosting-bill setup

Run the built Go application (which serves the Vue assets), PostgreSQL and Redis on an always-on Linux host using `compose.production.yml`. There is no runtime Node development server or local language model requirement. Eight GB RAM is a reasonable starting point for this personal workload, but measure with `docker stats`; image builds can use substantially more memory than serving. Electricity and internet still cost money.

Keep the machine awake, powered and connected. Fetching and email need outbound internet, not a publicly exposed website. Sleep or outages stop collection and delivery; the scheduler catches the latest due slot after recovery, not every missed slot. Keep database backups off the machine and rehearse restoration.

For your own devices, [Tailscale Serve](https://tailscale.com/docs/features/tailscale-serve) provides private HTTPS. Its [Personal plan](https://tailscale.com/pricing) currently allows up to six users free. Set the HTTPS application URL, allowed origin and GitHub callback consistently. Keep PostgreSQL and Redis private. A hosted external workflow cannot automatically reach a private tailnet address; it needs a network-access arrangement or a separately secured public application endpoint.

## Why Render Free is not the current recommendation

[Render's official free-instance limits](https://render.com/docs/free) include sleeping web services after 15 minutes without incoming traffic, PostgreSQL expiry after 30 days, and nonpersistent free Key Value storage. Background workers are not a free service type. The current Go process owns scheduled fetching and email, so a sleeping process cannot reliably perform this work. A free static frontend would not solve that backend requirement. An always-running paid service or a scheduler redesign would be a different deployment decision.

## Supabase can host the database, not the entire application

[Supabase Free](https://supabase.com/pricing) currently includes a 500 MB PostgreSQL database and pauses projects after one week of inactivity. You still need to run the Go application and Redis elsewhere. Monitor database growth and maintain backups. Local PostgreSQL is sufficient for personal email delivery; a cloud database is optional.

For persistent Go connections, use the direct connection where IPv6 works, or the shared **session pooler on port 5432** on an IPv4-only network. The transaction pooler on port 6543 does not support prepared statements and does not preserve session state. Use TLS and the exact connection string from the project dashboard. See [Supabase's connection guide](https://supabase.com/docs/guides/database/connecting-to-postgres).

**Configuration caveat:** `compose.production.yml` explicitly overrides `DB_URL` with its local PostgreSQL container URL for both the app and migration tool. Putting a Supabase URL into `.env` alone does not switch production Compose to Supabase. An explicit external-database configuration/override must update both services and their database dependencies, followed by migration and connectivity verification. Do not put database credentials in the frontend or commit them.

## Settings design references

The original `JOBRADAR_PLAN.md` specifies Linear light mode: whitespace, subtle borders, one accent color and sidebar navigation. The application board also explicitly planned dragging cards between status columns.

[Linear Preferences](https://linear.app/docs/account-preferences), [Linear Notifications](https://linear.app/docs/notifications), and [GitHub notification configuration](https://docs.github.com/en/subscriptions-and-notifications/get-started/configuring-notifications) provide useful references for separating account settings and delivery channels. Our adaptation uses focused sections for matching, sources, delivery and account access, while retaining this project's square corners and shared color tokens.
