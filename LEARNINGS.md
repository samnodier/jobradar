<!-- markdownlint-disable MD013 -->

# Learnings

Explain the difference between these three branches. Specifically, why would a senior engineer be upset if you committed a broken API fetcher directly to main?

Explain why we use feature branches instead of everyone pushing to main. Mention "stability" and "isolation."

The difference between git branch -m (rename) and git checkout -b (create and switch).

Why databases need Volumes in Docker. 5. What a Go module path is and why it should match your repository URL.

The difference between Host Ports and Container Ports in Docker.

Never commit sensitive credentials to version control. Use .env files and add them to .gitignore.

What a "service" represents in Docker Compose.

The difference between a Docker Image (the blueprint) and a Container (the running instance).

What happens to Docker data if you don't use a volume. 11. The difference between a "Host" port and a "Container" port.

Explain the difference between an Image and a Container in your own words (the Class vs. Instance analogy is great).

Explain why your Go code uses localhost as the host even though the DB is in Docker (Port Mapping).

What the -d flag does in docker compose up.

Explain the purpose of the internal/ directory in Go and how it helps with encapsulation.

Explain why sql.Open doesn't mean you are actually connected yet.

What is a "Driver" in the context of Go's database/sql?

Explain what go mod tidy does and why it's good practice.

Why we use a capital letter for Client and NewClient (Exporting).

Explain why sql.Open needs a driver to be imported with a \_.Why we used sslmode=disable in our connection string (Hint: it's for local development).

The difference between %v and %w.

Why we return nil instead of an empty struct when an error occurs.

The command to run a Go program located in a sub-folder.

What is a UNIQUE constraint and how does it help with Idempotency?

Why we use TIMESTAMP WITH TIME ZONE instead of just a local timestamp.

The difference between a Primary Key (for the DB) and an External ID (from the API).

What is a Makefile and why is it used in Go projects to manage tasks like migrations and builds?

Explain the difference between json.Unmarshal (using a byte slice) and json.NewDecoder (using a stream).

Why do we need to check res.StatusCode before trying to decode the body? (What happens if the API sends a 404?)

What is an HTML Entity (like &amp;) and why do we need to decode it?

What does ON CONFLICT DO NOTHING do in a Postgres INSERT statement?

Why is it a bad idea to let your fetcher package talk directly to your database package? (Hint: Think about "Dependency Cycles").

What is Type Safety in the context of a database?

Why does SQLC need access to your Goose migration files to work?

What is a Router and why do we use one instead of just if/else statements?

What is Middleware in the context of a web server?

Why do we use json.NewEncoder(w).Encode() instead of fmt.Fprintf(w, ...)?

Why do we use a respondWithJSON helper instead of encoding JSON inside every handler?

Why is r.Context() passed to the database query instead of context.Background()? (Hint: Think about what happens if the user closes their browser tab while the query is running).

Goroutine: A way to run a function in the background without blocking the rest of the program.

Ticker: A tool to trigger events at regular time intervals.

Blocking: When a line of code (like <-ticker.C) makes the program wait before moving to the next line.

What is CORS and why does it exist? (Security against malicious websites).

Why do we use chi.URLParam instead of reading from the query string for IDs?
What is the difference between LIKE and ILIKE in Postgres?

The difference between Authentication (Who are you?) and Authorization (What are you allowed to do?).

The concept of Delegated Identity (using Google/GitHub to handle the "Who are you?" part).

JWT Rule: Never put sensitive data (like passwords or secrets) inside a JWT payload because it is readable by anyone who has the token. It is only used to verify integrity, not to hide data.

`access_token` and `refresh_token` should be encrypted at rest in production. That's the right call to make when you have the core system working.

Git Stash: How to move uncommitted work between branches.

Orphaned Rows: Why Foreign Keys with ON DELETE CASCADE are necessary in normalized schemas.

The Blacklist Tradeoff: Balancing the speed of JWTs with the security of instant revocation.

`CITEXT` is a Postgres extension that makes text comparisons case-insensitive. Important for emails because `Sam@gmail.com` and `sam@gmail.com` should be treated as the same user.

the `state` parameter prevents CSRF attacks (Cross Site Request Forgery), not XSS. Here's the idea: your app generates a random string, sends it to GitHub, and when GitHub redirects back you verify the same string came back. This prevents a malicious site from tricking your server into thinking it initiated the login.

Request the minimum OAuth scopes needed; for GitHub profile plus email access, read:user and user:email are usually more appropriate than the broader user scope.”

Redis is good for temporary onboarding/session state because it supports TTL and shared access across instances, but permanent account truth belongs in Postgres.

Authentication systems often contain multiple independent state artifacts: OAuth state, onboarding session, email verification token, and application session; confusing them leads to bad design.

For sensitive changes like updating an email address, the new email should be verified before it replaces the currently trusted account email.

The OAuth callback URL is just the route in my app that receives the provider redirect after authorization; it is not special by itself, but it must exactly match what is registered with the provider.

golang.org/x/oauth2 handles OAuth client flow in Go: AuthCodeURL builds the provider redirect URL and Exchange swaps the callback code for a token.

OAuth state can be stored in a cookie for simple implementations, but production systems typically combine client-side storage (cookie) with server-side storage (e.g., Redis) to ensure one-time use, prevent replay attacks, and support distributed systems.

`Secure: true` on cookies means they only travel over HTTPS; always make this configurable or environment-aware so local development still works."

Namespace Redis keys consistently by feature prefix so related keys are easy to identify, scan, and manage together — e.g. `auth:pending_signup:` not just `pending:`

Use strconv.FormatInt(id, 10) to convert int64 to string when passing numeric IDs as text to a database query.

The difference between UNIQUE constraint and a UNIQUE index, and when partial unique indexes are useful. (Hint: unique only when a condition is true)

What .PHONY does in a makefile and why you need it for commands that don't produce files. tells Make: these targets arenot files, they are just commands. Always run them.

Vue component lifecycle. onMounted runs after the component is inserted into the DOM. It's where you put data fetching that needs to happen on page load.

what the SPA fallback problem is, why it happens, and how serving index.html as a fallback fixes it.

Never edit a migration that has already been run. Goose tracks which migrations have executed. If you change a file that already ran, your schema and your migration files are out of sync. Other developers (or your future self on a new machine) will have a broken database. Always move forward with a new migration file.

what .PHONY does in a Makefile and why you need it for commands that don't produce files.

the difference between INNER JOIN and LEFT JOIN, and when you use LEFT JOIN for optional relationships.

always scope destructive operations (DELETE, UPDATE) to both the resource ID and the authenticated user's ID. Never trust that the client owns a resource just because they know its ID.

REST conventions. URLs describe resources, HTTP methods describe actions. Never put actions in URLs.

sharing connections vs duplicating them. In Go you pass pointers to connection clients, so multiple structs can use the same underlying connection safely.

Moving shared types to internal/models is a good refactor when you notice multiple packages need the same type and neither should own it. It prevents import cycles and keeps types neutral. Do it when the pain is real, not preemptively.

Always use an unexported custom type for context keys. type contextKey string is the idiomatic pattern. It prevents key collisions across packages without any runtime cost

flag: credentials: 'include' is required for any fetch() call that needs to send cookies cross-origin or in non-trivial browser contexts. Without it, cookie-based auth silently fails with no error — just a 401.

In Go, the zero value for a pointer is nil. This is helpful with SQLC because we can omit nullable fields in a struct literal, and Go will default them to nil, which SQLC then inserts as a NULL in the database.

When do we use URL params vs. Request Bodies? URL params (/api/jobs/{id}) are great for identifying a resource you want to GET or DELETE. Request Bodies are better for POST requests when you are sending data to create a new record.

Why do we JOIN the jobs table when fetching saved\*jobs? (Because the join table only has IDs, and we need the human-readable data like Title and Company).

In Vue SFCs, PascalCase is preferred for components to distinguish them from native HTML elements. `<RouterLink>` and `<router-link>` are functionally identical.

API Design: When using POST, the backend expects a JSON body. We define a Go struct with json tags to tell json.NewDecoder how to map the keys.

State Management: After an API call modifies data on the server, we must also update the local frontend state (e.g., the Pinia store or a ref array) to keep the UI in sync without a page refresh.

SQL Joins for User State: Using LEFT JOIN allows us to combine global data (Jobs) with user-specific data (Saved/Applied) in a single query. IS NOT NULL is a handy way to turn a join result into a boolean is_saved flag.

The "Immediate" Watcher: Why immediate: true is necessary when a component needs to react to props as soon as it's born.

Hoisting in `<script setup>`: Why constants like formatters must be defined at the top before they are used in watchers or lifecycle hooks.

The "Stale State" Problem: Why the parent needs to update its list when a child updates a single item (and how the Spread Operator helps keep joined data intact).

Debouncing: Why we use clearTimeout and setTimeout to protect the server from excessive API calls.The pattern is: When the user types, cancel the previous pending timer Start a new timer for 800ms If they don't type again before it fires, save Not ref — correct. ref() creates a reactive container that Vue watches for changes to update the DOM. A timer ID is never read in the template, so reactivity is wasted overhead. Plain let is enough. Not const — correct. The value reassigns on every keystroke (each setTimeout returns a new ID). Not inside the function — this is the one you partially got. The real reason: every time a function is called, its local variables are created from scratch. If debounceTimer lives inside, each call starts at null. clearTimeout(null) is a no-op. You need the ID from the previous call to cancel it — which means the variable must outlive the function invocation. Module scope is the only place that persists between calls.

SQL Data Loss: Why RETURNING \* in an update doesn't return joined columns.

Computed vs. Watch: Use computed when you need to transform data for the template (like filtering or grouping). Use watch when you need to perform "side effects" (like making an API call or saving to localStorage) in response to a change.

Fail Fast (Validation): Validate data at the "Edge" (the API handler) before sending it to the "Core" (the Database). This saves database resources and provides better error messages to the user.

Database vs. API Constraints: Keep database constraints strict (NOT NULL) for data integrity, and use SQL casting or pointers in the application layer to handle optional/partial updates.

In Single Page Applications (SPAs), setting `height: 100%` on `html`, `body`, and the root `#app` is essential. Using `flex: 1` combined with `min-height: 0` on child containers allows them to correctly fill the viewport and support internal scrolling without breaking the layout shell.

For Kanban boards, the main page should usually be `overflow: hidden`. Scroll responsibility should be delegated to the individual `column-body` elements to keep headers and navigation sticky.

During early development, grouping related tables (e.g., all Resume sections like Experience/Education) into a single migration file keeps the schema manageable. Once in production, migrations should remain strictly atomic (one file per change).

Always preserve timestamps in filenames (e.g., `202604..._name.sql`). Migration tools (like Goose) rely on these for execution order; removing them breaks automation.

Avoid overlapping fields. Instead of having an `open_to_remote` boolean on the `users` table AND a `user_desired_locations` table, it is cleaner to let the location table handle "Remote" as a specific entry.

Separating "What I did" (Description) from "What I accomplished" (Achievements) in a schema is vital for AI. LLMs can tailor resumes much more effectively when they can pull quantified wins (Achievements) separately from general job duties.

Linking skills to specific \_experiences\* and _projects_ (not just the user) allows the system to calculate "Total years of experience" per tech stack, providing much higher signal for job matching.

Always scope destructive operations (DELETE, UPDATE) to both the resource ID and the authenticated user's ID. Never trust that the client owns a resource just because they know its ID.

The N+1 problem occurs when you fetch a list of items and then execute a separate query for each item to fetch related data. Use JOINs or JSON_AGG to fetch related data in a single query.

A Database Transaction ensures that a series of operations either all succeed or all fail. If we create an Experience but failing to link its Skills, the transaction 'rolls back' so we don't end up with partial data.

Always preserve the 'User ID' scope in every query. Even if you have the record ID, checking against the User ID prevents one user from deleting or editing another user's data.

Decoupled (What we are doing): Skills and Experiences have their own endpoints. Creating an experience doesn't require skills in the same request. This makes the code simpler and the UI more interactive.

Atomic: One giant request creates the experience AND all its skills in one database transaction. Great for data integrity (like a bank transfer), but more complex to code and harder to build a snappy "tagging" UI for.

SQLC COALESCE Pattern: While COALESCE(sqlc.narg('field'), field) is great for partial updates (PATCH), it makes it impossible to explicitly set a field to NULL. Using a CASE statement tied to a boolean (like is_current) is a clean way to handle mandatory nullification. The deciding question isn't "is this column nullable?" — it's "can the caller legitimately have no value to write?" PATCH from a client: yes → COALESCE. System write after a successful operation: no → write the value directly, so a missing value surfaces as a bug instead of silently preserving stale data.

Date Normalization: When bridging HTML5 `<input type="month">` and SQL DATE types, normalization (appending -01) should happen at the boundary (the conversion helper) to maintain type safety without cluttering the business logic.

You should note the difference between PUT (replace the whole resource) and PATCH (update specific fields) and why we use PATCH for profile updates.

TypeScript's strict null checks protect you from de-referencing null pointers. You can resolve these by using short-circuiting guards (&&) or by re-evaluating if the data needs to be passed as a prop if the child already has access to a global store (Pinia).

Backend validation is mandatory even if the frontend validates. Never trust the client. Also, COALESCE in PATCH requests makes it impossible to 'clear' a field using NULL values—I need to decide on a convention (like empty strings) for clearing data.

In PATCH requests, validation is conditional. If a field is sent (pointer is not nil), it must meet the requirements. strings.TrimSpace is essential to prevent users from bypassing 'required' fields with whitespace.

The 'Handoff' pattern: Frontend validates for UX, Backend validates for Security. Stores manage the Network state so Components can stay focused on the UI. A clean architecture means removing data-passing (props) if a global store already provides that truth.

PostgreSQL requires explicit type casting (::type) for nullable parameters (sqlc.narg) inside complex SQL structures like CASE and COALESCE because it cannot always infer the type before execution."

Since they have defaults like '{}', the backend will likely always send an array (even if empty) rather than a null value. In TypeScript, it's often cleaner to type these as string[] to avoid having to do null-checks every time you want to use .map() or .length in your Vue templates.

1. **Vue Router Query-Aware Active Highlights**: Vue Router's `active-class` matches on the URL path (`/jobs`). Differentiating between `/jobs` and `/jobs?filter=saved` requires query-aware highlighting. We solve this by removing `active-class` and using manual `:class` bindings that inspect `$route.query` directly.
2. **Relative URLs in HTML Anchors**: If a user inputs an external URL without a protocol (e.g. `github.com`), the browser treats it as relative (e.g. `localhost:5173/github.com`). We must normalize inputs in the app layer by checking for `http://` or `https://` prefixes and prepending `https://` if absent.
3. **SQL Schema Changes & Query SELECT lists**: When adding new fields (e.g. `last_status_changed_at`), we must update all corresponding queries that SELECT fields for that table and run `sqlc generate` to ensure Go database clients match the database state.
4. **Auto-Saving Settings Toggles**: In-app setting toggles (like email notifications) should automatically commit to the server on change rather than relying on a separate "Save Changes" button in a different tab, providing a much cleaner and responsive UX.
5. **Blocking Pop (`BRPOP`) vs. Polling (`RPOP`)**: Standard `RPOP` requires a loop to constantly poll Redis for new jobs. To avoid pinning the CPU at 100%, we would need a sleep interval, introducing processing latency. `BRPOP` blocks the connection at the database level, putting the worker goroutine to sleep (0% CPU) until a job is pushed (`LPUSH`), ensuring instant processing and optimal CPU efficiency.
6. **Graceful Worker Shutdown via Go Contexts & WaitGroup**: To stop background workers without interrupting active jobs, we listen for cancellation signals using `ctx.Done()`. Go goroutines cannot be forcibly killed, so a worker naturally finishes its current handler execution. By using `sync.WaitGroup` (`wg.Add`/`wg.Done`), the main thread can block using `wg.Wait()` until all workers complete their active tasks and exit.

7. **Set Representation with `map[string]struct{}`**: In Go, a `bool` occupies 1 byte of memory, whereas an empty struct `struct{}` occupies **0 bytes**. For high-performance sets, mapping keys to `struct{}` saves memory. Since `struct{}` has no value, we must check membership explicitly using the comma-ok syntax: `if _, ok := myMap[key]; ok`.
8. **Jaro-Winkler Similarity & Prefix Boosting**: Jaro similarity measures character matches and transpositions within a dynamic sliding window. Jaro-Winkler extends Jaro by adding a prefix boost: if two strings share a matching prefix (up to 4 characters), it increases the similarity score, which is highly useful for matching job titles.
9. **Token Sorting for Swapped Title Matches**: Standard string similarity metrics fail when word orders are swapped (e.g. "Go Software Engineer" vs "Software Engineer in Go"). Token Sorting normalizes, tokenizes, filters stop words, sorts alphabetically, and joins tokens. This makes the string sequences identical, resulting in a perfect match.
10. **Aho-Corasick O(N) Multi-Pattern Matcher**: Searching for $K$ skills in a text of length $N$ by checking substrings or running multiple patterns is slow. Aho-Corasick builds a Trie with failure links. When a match fails, the state machine transition jump targets the node representing the longest valid suffix without backtracking in the text, ensuring a single $O(N)$ pass.
11. **BFS for Building Failure Links**: The failure link of a node at depth $D$ always points to a suffix node at depth $< D$. Therefore, we must compute failure links for all shallower nodes first. BFS processes the tree level-by-level, ensuring that all target links are already computed before we process deeper nodes.

When dealing with queues, a job should never exist only in server memory unless it is actively being executed. If you use a fire-and-forget goroutine with time.Sleep, that job is held hostage in RAM. If the server is restarted, crashes, or is deployed, that RAM is wiped, and the job is gone forever.

Why You Cannot Blanket-Return 404s HTTP status codes are the primary way a backend communicates with the outside world (frontends, mobile apps, other microservices). 404 (Not Found): This tells the frontend, "The ID you gave me does not exist. Do not try this exact request again, because it will never work." The frontend might then redirect the user to a "Page Not Found" screen. 500 (Internal Server Error): This tells the frontend, "You did everything right, but my database is down or something broke on my end. Try again in a few minutes." The frontend might then show a "Server is experiencing issues, please try again" toast notification. The Danger: If your database crashes and you return 404s, your frontend will think the jobs were deleted and might show the user a "Job not found" page. Worse, your server monitoring tools (like Datadog or Sentry) will just see a bunch of 404s—which are normal—and will not trigger any alarms that your database is actually offline.

Your reviewer is completely right. What you currently have is an Absolute Timeout (the session dies exactly 7 days after creation, no matter what). What modern applications use is an Idle Timeout or Sliding Session (the session dies only if the user is inactive for 7 days). Fixing this requires updating two things, not just one. You need to extend the TTL in Redis, and you need to extend the expiration date of the cookie in the user's browser. If you only update Redis, the browser will automatically delete the cookie after 7 days anyway, logging them out.

BRPOP fixes both. It sleeps the goroutine at the Redis level — 0% CPU — and wakes it instantly the moment a job is pushed. The 2-second timeout isn't a grace period — it's the max time the worker sleeps before waking up to check ctx.Done(). If the timeout were infinite (BRPOP queue 0), the goroutine would sleep forever and you'd never get a clean shutdown. Write this in LEARNINGS.md — specifically the contrast: RPOP = poll loop + sleep = CPU waste + latency. BRPOP = sleep at DB level = 0% CPU + instant wakeup. The timeout exists for shutdown awareness.

Write this in LEARNINGS.md: When filtering on a LEFT JOIN's outer table, the filter must go in the ON clause. Putting it in WHERE silently converts the LEFT JOIN to an INNER JOIN by discarding rows where the outer table had no match.

Write this in LEARNINGS.md. Specifically: why COALESCE in PATCH updates, why NOT in system writes, and the stale-data trap that would happen if you misapplied it.

json.Unmarshal works on a []byte already in memory, json.NewDecoder wraps an io.Reader for streaming — HTTP bodies, files, anything where you don't want to load the whole thing into memory first. For a queue payload that's already a []byte field in a struct, json.Unmarshal is the right call.

Background workers have no HTTP context. They can't use middleware like TryAuth because there's no request. When they need user data, they query the database directly. This is the fundamental difference between a request handler and a background worker.

In Go, a struct's zero value has all fields set to their type's zero. For pgx nullable types like pgtype.Float8, the zero value means NULL because Valid defaults to false. This is why pgtype.Float8{} works as NULL without any helper function.

In Go, len(s) returns bytes, not characters. For Unicode-safe string operations, convert to []rune first and use len(r). A job title like "Développeur" has 11 characters but 12 bytes — len(s) lies to you.

A sanitization function should never return something worse than its input.

Now, the WHY of the encoding direction — think through it step by step: Byte 0xE0 is a raw byte in the original UTF-8 text. A broken pipeline reads those bytes using Windows-1252 and asks: "what Unicode character is byte 0xE0?" The answer is U+00E0 — the character à. So 0xE0 becomes à in the mangled string. Now you have à (U+00E0) and you need to get back to byte 0xE0. You ask the Windows-1252 encoder: "what byte represents U+00E0?" Answer: 0xE0. You're back to the original byte. The round-trip works because mojibake is its own inverse — the same table that broke the text can un-break it, just run in the other direction. Decoder maps bytes → characters. Encoder maps characters → bytes. Same table, opposite direction.

In production you can't reset — adding a uniqueness constraint to a column that already has duplicates requires a data-cleanup migration first. That's a real interview answer.

Skill score denominator flaw — matched/userSkills penalizes broad skill sets; matched/jobSkills is the correct signal

Production uniqueness constraint caveat — can't add a UNIQUE constraint to a column with existing duplicates; requires a data-cleanup migration first (repoint FKs, delete dupes, then add constraint)

"Background workers have no HTTP context." The enqueue is conceptually background work that you're kicking off from a handler. You don't want it killed because the user closed their tab. context.Background() (or better, a context derived from your app's root context that dies on server shutdown) is the right call for the enqueue loop — even though the DB query and transaction above it correctly used r.Context().

A handler can run two kinds of work: work that belongs to the request (use r.Context() — cancellable when the user leaves) and work that's merely triggered by the request but outlives it (use context.Background() or the app root context — must survive the user disconnecting). The transaction is the former; the fire-and-forget re-match enqueue is the latter. Passing r.Context() to background work is a silent bug: it half-completes when the client disconnects.

Storing the app root context on apiConfig (rootCtx) lets handlers kick off background work with a lifetime tied to the server, not the request. Using context.Background() as that root is fine — it just never cancels, so in-flight work isn't interrupted on shutdown. Getting true shutdown-cancellation would mean storing the cancellable context derived from the root, but that's a chicken-and-egg with struct construction order and not worth it for a fast enqueue loop. Knowing the gap is the lesson; context.Background() on rootCtx is the right call here.

Re-matching scale, part 1 (coalescing): Re-match-everything-on-every-change is O(jobs × saves). Fine at portfolio scale, a fire at production scale — auto-save makes it trivial to fire 10 saves in 10 seconds, and with 10,000 jobs that's 100,000 enqueues. First defense is to not enqueue on every save: mark the user "dirty" and enqueue one re-match after a quiet period (server-side debounce). Ten toggles become one job, not ten thousand.

Re-matching scale, part 2 (move the fan-out into the worker, and pipeline): Two more techniques beyond coalescing. (a) Instead of enqueueing N per-job match jobs from the handler, enqueue ONE batch job — RematchUser{user_id} — and let the worker loop over all jobs internally. The queue holds 1 item, not 200; the fan-out happens where you control concurrency, not where it explodes. (b) If you do push many items, use a Redis pipeline/MULTI so they go in one network round-trip instead of N. The current loop does N separate Enqueue calls = N round-trips. Order of reach: coalesce first (cheapest, biggest win), then batch-job fan-out, then pipelining only if a single operation still pushes a lot.

Background activity shouldn't depend on the browser context" is the what. The why it matters for the logs you're about to watch: if you'd left it as r.Context() and the loop is pushing 200 jobs when the response finishes, the context dies mid-loop and the remaining Enqueue calls fail silently — you'd see match jobs for some jobs in the logs, not all, and no error explaining why. With rootCtx you'll see all of them process regardless. That partial-completion-with-no-error is exactly the silent bug.

A missing signal is "unknown," not 0 or 1. Renormalize over the weights of the dimensions you actually have, instead of substituting a constant. The accumulator's denominator must collect the weights of active dimensions, not the running weighted sum — and the first dimension hides that bug because its weight equals its contribution. Also: 0-because-no-data and 0-because-no-overlap are different, and the code has to tell them apart.

You cannot enforce uniqueness on values encrypted with a random nonce/IV — identical plaintext yields different ciphertext by design.

Never send a secret to the client just to render UI state. The backend stores the encrypted Gemini key, but `GetUserByID` must not select the ciphertext — and there is no plaintext-returning endpoint at all (decryption happens only server-side, in the worker, right before the API call). To let the UI show "key configured" vs "no key," compute a derived boolean in SQL — `(encrypted_gemini_api_key IS NOT NULL)::boolean AS has_gemini_key` — and put that on the wire. The rule: the boolean crosses the wire, never the secret.

Optimistic UI update vs refetch: after a successful mutation, update the local store state directly (`this.user.has_gemini_key = true`) instead of re-fetching the whole user. The UI flips instantly with no extra round-trip. Only refetch when the server computes something you can't predict client-side.

Smart (container) vs dumb (presentational) Vue components. A dumb component takes everything via props and emits events — it knows nothing about stores/fetching, so it's reusable anywhere; you pay for that in parent boilerplate. A smart component reaches into stores and orchestrates its own mutations. The deciding question is reuse: a route-level feature section that lives in exactly one place (e.g. a Settings tab) should be smart — pushing all its data and handlers through the parent just bloats the parent for zero reuse benefit. A genuinely reusable widget should be dumb. Don't mix styles inside one component (one mutation through a store, another emitted up to the parent) — that inconsistency is the smell to avoid.

Don't check the same condition in two places. If a parent gates a child with `v-if="cond"`, the child can't also own a `v-else` loading branch for `!cond` — that branch is dead code, because the child only mounts when `cond` is already true. Pick one owner of the condition. Corollary: passing a prop a component is already gated on is dead weight, and passing a prop the child doesn't declare silently becomes a fallthrough HTML attribute on the child's root element.

Vue Router `router.push({ name })` needs the registered route _name_, not a path. If routes are named `'login'`/`'home'`, then `{ name: '/login' }` matches nothing — navigation silently fails with only a console warning. TypeScript won't catch it because names are plain strings. Use the exact registered name, or push the path as a string (`router.push('/login')`).

Pick one error-propagation convention per store and stick to it. Either actions rethrow (and components wrap calls in try/catch) or actions swallow and set `this.error` (and components read the flag after awaiting). Mixing both in the same store — one action rethrows, another swallows — forces future-you to guess which pattern each action uses.

A queue job's payload is a contract between enqueuer and handler. Give each job type its own payload struct — even when they're identical today — so each contract evolves for its own reasons (low coupling). The cost is duplication now; the justification is expected divergence, which you must be able to name. Duplicating without that reason is premature abstraction.

Don't enqueue work that's guaranteed to fail — gate it at the producer (same instinct as validating at the API edge before hitting the DB). A pre-enqueue check can go stale by the time the job runs, so the handler must still validate defensively. The pre-check is an optimization; the handler's check is the guarantee.

Producer and consumer of a queue job can live in different files. The enqueue decision belongs with whoever triggers it (the match handler); the job's processing logic belongs in its own file per job type. Keeps each file single-concern and makes "add a job type" mean "add a file," not "grow a file."

Retry granularity — a handler that returns an error retries the whole handler, not the failed line. Before return err, ask "what re-runs?" If a cheap sub-step fails but retrying re-does expensive work that already succeeded (and is idempotent), return nil and let a higher-level mechanism recover it. Match the retry to the cost of the work it repeats.

A registered handler is a runtime wiring step, not a compile-time one — forgetting RegisterHandler compiles green but the job type silently has no consumer. "It builds" never proves the plumbing is connected; only running it and seeing the log does.

When the same resource is served by two endpoints — one returning the raw DB struct, one returning a hand-mapped DTO — a new column added to the query appears automatically on the raw endpoint but is silently dropped by the DTO until you add the field by hand. Two response shapes for one resource = a drift hazard. Either converge on one shape, or treat "add field to every DTO" as part of the migration checklist.

In filter-then-enrich, the cheap filter must own the score — because the score is what gates the expensive step. If the LLM produced the score, you'd have to run the LLM on everything to decide whether to run the LLM, and the filter collapses. The expensive stage can only ever see pre-filtered inputs.

A vendor SDK couples you to that vendor — it does not make swapping vendors easier; it makes it harder. Provider-portability comes from your own interface (Enricher) that the rest of the app depends on, with the SDK hidden behind one implementation. Open/closed: a new provider = a new implementation of your interface, not edits to the caller. The SDK reduces per-vendor boilerplate; your interface buys the portability.

`var _ Iface = (*T)(nil)` is a compile-time assertion that `*T` implements `Iface`. Go interfaces are satisfied implicitly, so without it a mismatch only surfaces at the call site; with it, the error appears on the type itself, immediately, and documents the intent. Typed nil = a zero-cost value of the right type; the whole line vanishes at compile time.
