# Job sources and expansion plan

Research checked September 22, 2026 against the providers' official documentation. Availability and terms can change. More sources are useful, but do not block deploying the existing matching and email pipeline.

## Implemented sources

RemoteOK, Adzuna and configured Greenhouse employer boards already use the shared ingestion pipeline. Jobicy and MyJobMag Kenya are now optional adapters, disabled by default. Enable them in the server environment and restart the backend:

```dotenv
JOBICY_ENABLED=true
MYJOBMAG_ENABLED=true
```

Both new sources run at `max(SCRAPE_INTERVAL, 6h)`. They require no API keys. Jobicy's published minimum interval is one hour; our six-hour minimum is deliberately more conservative. Each adapter uses a 20-second default HTTP timeout, a 16 MiB response limit and source-ID deduplication. These are backend fetches, independent of browser visits.

| Source | Coverage and access | Operational notes |
| --- | --- | --- |
| MyJobMag Kenya | Official detailed RSS at `https://www.myjobmag.co.ke/jobsxml_by_categories.xml`, verified from the official [feed directory](https://www.myjobmag.co.ke/feeds/). | Retains provider job links and category/location facts. The published [terms](https://www.myjobmag.co.ke/terms) permit redistribution of sent alerts and note that expired postings remain online. No numeric polling quota was verified; use the conservative schedule. |
| Jobicy | Keyless JSON at `https://jobicy.com/api/v2/remote-jobs?count=200`; latest remote listings with geographic restrictions. | [Official API and fair-use documentation](https://jobicy.com/jobs-rss-feed) permits job-discovery apps, newsletters and internal tools. Preserve Jobicy attribution and canonical job URLs; poll at most hourly. Original ATS application URLs are a separate paid feature and are not used here. |
| Adzuna | Existing authenticated search using the profile's configured supported market and desired roles. | Default limits: 25 requests/minute, 250/day, 1,000/week and 2,500/month. Budget across all searches/pages; adding locations or accounts multiplies usage. Personal research is a listed permitted use. See [official API terms](https://developer.adzuna.com/docs/terms_of_service). |

The new adapters preserve `Source` and canonical URLs for attribution in job displays and exports. Do not remove source identity when copying records into another workflow.

## Important data limits

- **Nairobi does not automatically match Kenya.** MyJobMag often supplies only a city. Matching currently uses location phrases, not a geographic database; the adapter does not guess country from the board's audience. Add the city as a desired location when appropriate. A geography-normalization feature would need explicit mappings and tests.
- **Remote does not mean worldwide.** Jobicy marks jobs remote but retains restrictions such as USA or Europe. Original employer requirements remain authoritative.
- MyJobMag entries with a recognized expiry date already in the past are skipped. Missing/unknown expiry is not proof a job remains open. The shared source record has no expiry field, so this adapter does not retroactively close previously persisted jobs. An HTTP 200 on a job page also does not prove availability.
- Jobicy salary amounts are not imported yet: the feed can mix hourly and annual intervals, while the shared record has no interval field. Comparing those directly would produce misleading results.
- Neither source provides a verified required-skills list in this adapter. Missing skills stay missing; category labels are not silently treated as required skills.
- Duplicate IDs within one source are removed. The same employer posting appearing on several different platforms can still appear more than once; cross-source deduplication is separate work.
- RSS decoding honors the feed's declared encoding. Existing upstream text corruption may remain; the adapter does not invent replacement text.

## Next additions, in practical order

| Priority | Source | Why add it / constraints |
| --- | --- | --- |
| 1 | [We Work Remotely public RSS](https://weworkremotely.com/remote-job-rss-feed) | International remote jobs with programming category feeds. Officially allows reuse with attribution/backlinks. Feed: `https://weworkremotely.com/remote-jobs.rss`. Its token-based posting API is a different product. |
| 2 | [Arbeitnow](https://www.arbeitnow.com/blog/job-board-api) | Free keyless Europe/Germany feed and a UK feed announced in 2026; visa-sponsorship filtering available. Endpoints: `https://www.arbeitnow.com/api/job-board-api` and `https://www.arbeitnow.co.uk/api/job-board-api`. [API terms](https://www.arbeitnow.com/terms) require a backlink; no numeric public quota verified. |
| 3 | [Lever postings API](https://github.com/lever/postings-api/blob/master/README.md) | Add selected employers using their board slugs, including global/EU instances. Published postings are public; this is not a cross-employer full-text search service. |
| 4 | [Ashby public postings API](https://developers.ashbyhq.com/docs/public-job-posting-api) | Employer watchlists with structured locations, workplace type and optional compensation. Requires a board name; list only `isListed=true` records and preserve salary intervals. |

These remain research recommendations, not implemented integrations. Configure useful existing sources first, verify email delivery, then add each adapter with fixture tests and its own small commit.

## Kenya boards without verified public API access

[BrighterMonday Kenya](https://www.brightermonday.co.ke/faq) documents its own job alerts, but no documented public read/search API was found during this review. [Fuzu](https://www.fuzu.com/kenya/company/fuzu-ltd) has Kenya listings, but public search API access was not verified. This does not prove neither offers partner access. Use official alerts or manual job-link import until a supported integration is established; reverse-engineered browser requests are not an official API.

[Remotive's public API](https://remotive.com/remote-jobs/api) is another remote source, but its terms require attribution/backlinks, delay public results by 24 hours and restrict displaying listings to collect signups/email addresses. Confirm compatibility with any future public onboarding flow before adding it.

## RemoteOK feed versus local pagination

[RemoteOK's official feed help](https://remoteok.featurebase.app/help/articles/3140840-is-there-an-api-or-rssjson-feed-of-remote-jobs) documents `/api`, RSS and tag filters; it does not document a next-page cursor or offset for historical backfill. The current adapter ingests returned jobs and retains earlier stored records via upserts. Jobradar pagination queries its own accumulated database. Do not invent source page parameters or delete jobs merely because they drop out of the feed; age-based retention is the only automatic deletion.

The public homepage lists the newest public jobs. Public postings older than `JOB_RETENTION_DAYS` (default 90; posting date, else first-seen date) are deleted hourly and are not re-collected; jobs you saved, applied to or imported are kept. To seed production, restore the current database backup; this also preserves profiles, saves, applications and the digest ledger.
