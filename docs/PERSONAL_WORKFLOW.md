# Personal job-search workflow

## Configure once

Keep provider secrets in `.env` on the host. Do not paste keys into chats or frontend settings.

- Set Adzuna app ID/key. Profile → Preferences supplies desired roles and locations; Profile → Settings supplies the supported Adzuna market. Adzuna reads these at each 12-hour fetch. RemoteOK and selected Greenhouse boards are collected separately and then matched to the profile.
- Select Resend or GoHighLevel and configure the fixed recipient/sender. The recipient must be the account email used to sign into Jobradar. Keep `DIGEST_ENABLED=false` while previewing.
- In Preferences, enter actual desired locations separated by semicolons, roles, skills, pay expectations and job types. In Settings, choose work mode, missing-data policy, minimum score, timezone, hours and email size. A reasonable starting schedule is 8 and 18 in your own IANA timezone, with ten jobs per email. These are editable values, not claims about your location.
- Enable the master Job Recommendations switch and the individual email/in-app channels you want.

## A job link from a community

1. Add Job → paste the public job-posting URL.
2. Jobradar reads structured JobPosting metadata when available. Otherwise it tries an optional configured AI extractor, or uses page title/text without a key. Review every field; fill missing details before confirming.
3. If the URL is private, inaccessible or blocked, select manual entry and paste the description. A WhatsApp invite/message URL is not necessarily a publicly fetchable job posting.
4. Open the saved private job and choose **Copy for AI**. Review the Markdown, then paste it into your own AI chat. It includes your saved resume sections and instructions to avoid fabricated experience or claims.
5. Review the generated resume/letter yourself, apply using the original posting, and update the application tracker. Exporting context never submits an application or calls a model.

## Jobs found for you

1. Complete your profile and let matching finish. An empty profile cannot receive a meaningful score.
2. Open Profile → Best Matches or Jobs → Best matches. Both use the profile minimum score and eligibility filters; pending rematches are temporarily excluded.
3. New qualifying matches create one in-app notification each. Opening one marks it read and opens that job, including jobs outside the first feed page. The bell counts unread items among the latest 50 currently eligible notifications.
4. Preview email on Dashboard. Enable delivery only after the sender and recipient are correct. Up to two digests send per local day, with at most the selected number of jobs each. Already emailed jobs are never repeated; fewer qualifying jobs means a shorter digest or no email.
5. Raise the threshold to reduce results; disable a channel to stop it. This does not guarantee a fixed number of suitable openings every day.

## What the score means

The deterministic ranking combines title similarity (45%), proficiency-weighted required-skill coverage (30%), and experience text overlap (25%). Missing signals are removed and remaining weights rescaled. The public-feed title gate skips very weak title matches; explicitly imported jobs bypass that gate. A title-only job can score highly because little evidence is available. This is not a calibrated hiring probability.

The professional-fit score above has weight 80. Each configured preference with known evidence—work mode, desired location, employment type, complete compatible-currency salary range, industry/category, company stage—has weight 5 and contributes 100 for a match or zero for a mismatch. The result is divided by active weights. These preferences also remain eligibility gates, so a high score cannot override a known mismatch. Missing facts follow the Include Unknown setting; details explain exclusions. Salary comparisons assume comparable pay periods. No currency conversion is performed. Location matching uses normalized text phrases rather than a geographic database: spell out alternatives, and verify work authorization, residency, hybrid arrangements and salary period yourself. Remote does not imply worldwide.

Industry/category and company-stage preferences now affect numeric ranking and eligibility when explicit metadata exists. Adzuna supplies categories; imports accept reviewed industry/category/stage fields. Preferred industry/category phrases are literal matches, not a semantic taxonomy. Unknown metadata has no numeric weight and follows Include Unknown for eligibility. Adzuna may omit explicit skills; Greenhouse may omit pay/remote metadata. Unknown fields must not be represented as confirmed facts.

## Email costs and sender identity

As checked on 2026-09-20, [Resend's documented free transactional quota](https://resend.com/docs/knowledge-base/account-quotas-and-limits) is 100 emails/day and 3,000/month, including received mail. Two messages a day to one recipient use about 60/month. A digest containing ten jobs is one email, not ten.

For testing, [Resend's resend.dev sender](https://resend.com/docs/knowledge-base/403-error-resend-dev-domain) is restricted to the email associated with the Resend account. A verified sending domain is needed for other recipients. The Jobradar recipient must also match your signed-in account, so align those addresses or use a verified sender. Provider acceptance does not prove inbox delivery.

GoHighLevel remains an alternative when your subaccount already has email delivery configured. This adapter needs a contact ID whose email matches the recipient, plus API permissions to read contacts and send messages. A location/subaccount ID is not a contact ID. Account email charges can still apply; the API token alone does not establish free sending.

## External workflow access

Profile → Settings → Personal workflow access now creates expiring, revocable read-only tokens. Use `GET /api/export/jobs?min_score=70&date=today&limit=100` and follow `next_cursor` for further pages. Request filters do not change profile notification settings. `GET /api/export/jobs/{jobID}/ai-context` returns the job and your saved profile/resume as Markdown. Tokens authenticate these export routes only. Full arguments and examples: [Job export API](JOB_EXPORT_API.md).

Jobradar does not make an AI call for exports. Your external workflow can fetch jobs independently of any AI subscription; how that workflow connects to an AI product is a separate choice. Manual Copy for AI remains available.

## Running on an always-on Linux host

The Go server, built static frontend, PostgreSQL and Redis are a modest personal workload and should fit 8 GB RAM without local inference. Measure actual usage with `docker stats`; a frontend/Go image build can use more memory than normal serving. Existing hardware avoids a hosting subscription, but still uses electricity and your internet connection.

Use `compose.production.yml` and the deployment guide. The app serves frontend assets itself, so no frontend development server is needed. Keep the laptop awake, connected and powered; restart policies restart containers after Docker starts. A sleeping/offline laptop cannot collect jobs or send mail. Resume catches up the latest due slot, not every missed email.

No inbound public port is needed merely to fetch jobs and send email. For access from your own devices, an optional private HTTPS setup is [Tailscale Serve](https://tailscale.com/kb/1312/serve), subject to the [current personal plan](https://tailscale.com/pricing). Set the app base URL, allowed origin and GitHub OAuth callback to that HTTPS address. Keep PostgreSQL and Redis private. A remote AI service cannot automatically reach a private home-server address.

[Render free instances](https://render.com/docs/free) sleep after 15 minutes without inbound traffic, and free PostgreSQL expires after 30 days. That is unsuitable for this release's persistent database and in-process scheduled workers without redesign or paid services. Keeping a website awake artificially is not a dependable scheduler design.

Ollama is not a dependency. On the deployment machine, first stop/disable its service if unused; remove model files only after identifying what you want to keep. This repository work does not uninstall software or delete models on your home server.

## Local review before deployment

Run migrations, start dev, and exercise both scenarios above with your profile. Verify filtering, notification links, Copy for AI, digest preview and real delivery. Then promote the reviewed dev commit to main and push the release. Do not push backup branches or replace the remote history with the old large commit.


## Editor diagnostics

Targeted Go compilation and language-server checks passed. A broad `go list ./...` still encounters the existing PostgreSQL-owned `pgdata` directory; use `make test` for source packages. `.lazy.lua` adds project-local gopls exclusions for the installed LazyVim setup. Restart Neovim from the repository, accept its project configuration if prompted, and reconnect/restart gopls. If a marker remains, inspect its actual diagnostic text rather than assuming every marker is a compile failure. Database ownership/permissions were not changed.


## Local morning worker

The local installation uses `deploy/jobradar-digest.service` copied to `~/.config/systemd/user/`, with `bin/jobradar --digest-worker`. It runs the scheduler independently of the development HTTP process. Build with `go build -o bin/jobradar ./cmd/jobradar`, then restart using `systemctl --user restart jobradar-digest`. Inspect logs with `journalctl --user -u jobradar-digest`.

The sender must match the exact verified domain: `jobs@mail.example.com`. Set the registered account as DIGEST_RECIPIENT. Delivery time and the maximum job count live in Profile → Settings. DIGEST_START_DATE optionally delays the first scheduled delivery until that YYYY-MM-DD in the profile timezone.

`bin/jobradar --digest-preview` writes HTML without sending. `bin/jobradar --digest-test` sends a real preview to the configured personal recipient without reserving jobs or consuming the morning slot. Repeating the same preview on the same UTC date uses the same provider idempotency key. `--digest-once` runs the real delivery cycle, which does reserve jobs.

This worker uses local infrastructure: the computer must stay awake and online, and PostgreSQL must be running. Keep the main Jobradar server running for new source collection and matching. Existing matches can be emailed by the independent worker. App links currently use APP_BASE_URL (localhost for development); original job links work directly from the email.

## Feedback and posting checks

In a job's detail panel, Good match / Not relevant records explicit account-specific feedback. Similar postings receive a bounded adjustment of up to ten ranking points. Not relevant excludes that job from new digest selection; Clear feedback queues restoration of its score. Existing prepared emails remain immutable.

Check posting link records reachable / unavailable / unknown, cached for 24 hours. Reachable is not confirmation applications are still open. A 404 or 410 excludes a posting from new digest preparation; 403, timeouts and other uncertain results do not. Nothing deletes your application history. Scheduled digests also check selected links before preparing mail. The email can contain fewer jobs when links are unavailable.

Copy for AI offers combined preparation or focused evaluation, resume, cover-letter, interview and outreach modes. All use your saved facts and explicit instructions not to invent credentials; you choose where to paste the text. No model API key is needed.

Applications displays due follow-ups for active stages and a table of current outcomes by source and score. After following up, clear or reschedule its date. Current stages/scores are not historical conversion rates.
