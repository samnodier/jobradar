# Personal hosting options

Reviewed September 22, 2026. The home server was the earlier low-cost recommendation; Sam has now requested a broader comparison before choosing a deployment target. Deployment has not happened simply because these instructions exist.

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


## Current shortlist for this application

The Go process serves the built Vue app and runs collection/matching/email. PostgreSQL stores all durable account/job/delivery state; Redis supports sessions and queues. A static-site deployment alone is insufficient. Prices below were checked September 22, 2026, exclude tax/backups, and are not total-cost guarantees.

| Option | Cost basis | Why choose it | Tradeoff |
| --- | --- | --- | --- |
| Existing home server | No hosting subscription; power/internet remain | Lowest recurring cost, reuse production Compose | Home power/network/sleep directly affect delivery; own backups and updates |
| Linux VPS with Docker Compose | [DigitalOcean Basic](https://www.digitalocean.com/pricing/droplets): 2 GiB $12/month, 4 GiB $24/month | Predictable single-server deployment; independence from laptop; useful ops learning | You maintain OS, database and backups; 2 GiB is a starting estimate, build remotely and monitor memory |
| [Hetzner Cloud](https://www.hetzner.com/cloud/cost-optimized/) | Check live region/currency quote and stock; page showed entry tiers unavailable during review | Compare with DigitalOcean for a lower-cost VPS | Regional availability, IP/backups and tax affect final price; same self-management duties |
| [Railway Hobby](https://railway.com/pricing) | $5 minimum including $5 usage; app, database, Redis and storage usage can exceed it | Less server administration, integrated service deployment | Variable total bill; configure persistent volumes/backups and keep the worker-bearing app awake |
| [Render paid services](https://render.com/docs/service-types) | App + Postgres + Key Value priced separately | Managed application/database service options and deployment workflow | More separate service costs; free sleeping web/database plans do not fit the current scheduler |

My recommendation for predictable overnight operation is one small paid VPS if Sam is comfortable maintaining Linux, or Railway if reducing server administration is the priority. Home hosting remains valid when minimizing the bill is more important than power/network independence. Supabase is an optional database, not a complete hosting solution for this stack. Vercel/Netlify frontend hosting would still leave the always-on Go process and stores to deploy.

## Preserve the existing database during deployment

Deploying code does not require discarding data. Take a PostgreSQL custom-format dump of the current development database, restore it into a fresh production database, run outstanding migrations, then verify job/profile/application and email ledger counts. Preserve ENCRYPTION_KEY so existing encrypted provider keys remain readable. Database volumes persist across application container rebuilds. Do not delete volumes or copy a live pgdata directory.

Use one production environment for collection and scheduled mail. Keep production sending disabled during restore/verification. Disable the local digest service before enabling production email to avoid separate copied databases sending independently. The advisory lock only coordinates processes sharing the same database.

Then configure HTTPS on an application subdomain (for example jobs.example.com), APP_BASE_URL/allowed origins/GitHub callback, backups, provider secrets and sender identity. Verify public browsing, private account isolation, collection and one email before enabling the regular schedule. The application domain is independent of the verified Resend sender subdomain. No paid resources have been provisioned yet.
