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

The moment you upgrade a signal from presence to magnitude, you inherit a missing-data problem you didn't have before. A name is binary — it's in the list or it isn't, there's no such thing as a "missing" name. But a level can be absent while the thing it describes is present. Flat counting let you ignore missingness entirely; weighting forces you to confront it. That's not proficiency being annoying — that's the unavoidable cost of every richer signal. Renormalize over what you have (which you already do for whole dimensions), and use a neutral prior for missing values within a dimension. Never let "unknown" masquerade as 0 or 1.

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

The core rule (return value = retry signal; ask "could a retry succeed?"), the enrich taxonomy table, and the contrast: the matcher needs no taxonomy because it has no flaky external dependency — its only retryable failure is a DB error, which it already handles.

The voice/person/audience of LLM output is controlled by the system prompt's framing ("the candidate" → analyst report; "you" → direct address), and the model defaults to detached third-person unless you name the audience explicitly.

Never put a Docker bind-mount volume (especially DB data, owned by the container user) inside your language's module/source tree — toolchain commands that walk the tree (go mod tidy, go build ./...) will hit permission-denied. Use a named volume.

`errors.As` matches by the exact dynamic type in the error chain — pointer vs value is part of that type. The genai SDK returns its error by value (`return APIError{...}`), so a target of `var apiErr *genai.APIError` never matches: `errors.As(err, &apiErr)` is looking for a `*genai.APIError` in the chain, but the chain holds a value `genai.APIError`. Fix: declare the target as a value (`var apiErr genai.APIError`) so `&apiErr` points at the type that's actually there. The deeper lesson: my table test PASSED while production was broken, because the test built its input as `&genai.APIError{...}` (a pointer) — the wrong shape. A test is only as good as the fidelity of its inputs to what the real dependency returns; a test that asserts against a fiction gives false confidence. Always construct the error the way the SDK actually constructs it (check the SDK source: `return APIError{}` vs `return &APIError{}`).

Full jitter on exponential backoff: instead of sleeping exactly `base * 2^attempt`, sleep a random duration between 0 and that cap (`rand.Float64() * maxBackoff`). Same average growth, but two failures that happen at the same instant pick different delays instead of retrying in lockstep — which is what prevents the thundering herd from re-slamming a recovering service. Plain exponential backoff (no jitter) keeps synchronized failures synchronized.

A retryable (transient) error still ends up in the DLQ if the retry budget is shorter than the failure window. A Gemini free-tier 429 has a ~60s quota window (it even told us "retry in 14s"), but my 3 retries span only ~15s of jittered backoff — so all three fire inside the same blocked minute, all 429 again, and the job dies in the DLQ despite being "retryable." Correct classification (transient) is necessary but not sufficient: the retry schedule has to be able to outlast the thing it's retrying. Fixes: honor the server's `Retry-After`/`RetryInfo`, cap concurrency to the rate limit, or widen the backoff window.

Swapping a Docker bind mount for a named volume (or running `docker compose down -v`) gives you a brand-new, empty database — the new volume has no schema. You must re-run migrations (`make migrate-up`) before the app has any tables; until then every query fails with `relation "x" does not exist` (SQLSTATE 42P01). A fresh volume is a fresh DB.

Validate at the edge for UX; defend in the core for correctness. A required form field improves the common path, but the scoring algorithm must stay robust to missing data anyway — because the UI is never the only way data enters the system (old rows, joined tables, bulk imports). Never let an algorithm assume an upstream guarantee holds. The neutral prior isn't a workaround for a bad form — it's the algorithm doing its job: producing a sane answer regardless of how complete its inputs are.

A nil slice is safe to append to (append allocates); a nil map panics on write. Maps must be make()'d before any assignment. Same var x T syntax, opposite safety — because append owns the slice header, but a map write has no such indirection.

A committed (or rolled-back) transaction is dead — its qtx query object can't be reused. Reads that happen after Commit go through the pool (cfg.db), not the tx. Mixing them up gives a post-commit 500 even though the write succeeded.2. Every 500 branch must log the underlying err. A 500 with no log is a debugging black hole — you end up guessing the cause (you guessed "conflict") instead of reading it. The generic message is for the client; the real err is for your logs.

- ordinal vs interval data — why no proficiency→number mapping is "objective," and why a uniform global scale isn't "per-user opinion"- richer signal ⇒ inherited missing-data problem — the moment you upgrade presence to magnitude, you inherit missingness; renormalize/neutral-prior, never let unknown = 0 or 1- validate-at-edge / defend-in-core — a required form field doesn't let the algorithm assume the data's there (old rows, joined tables, imports)- nil slice vs nil map — append to a nil slice is safe; writing to a nil map panics- build-green ≠ correct — the un-summed accumulator the test caught, not the compiler- tx is dead after commit — post-write read-back goes through the pool; scope a transaction to the atomic unit only- log every 500 — a generic 500 with no logged err is a debugging black hole (you guessed "conflict")- LSP cache vs compiler truth — trust go build over a stale gopls diagnostic; don't refactor real code to satisfy a ghost- free-tier LLM is a per-day demo allowance — 20/day here, not per-minute; meaningful enrichment needs paid billing; a per-day RetryInfo won't refill in 30s

SSRF (Server-Side Request Forgery): an attacker supplies a URL and tricks the server into making the request. Dangerous because the server sits inside the trust boundary — it can reach hosts the public internet can't: localhost services (Redis, Postgres, the app's own API), private-network IPs, and the cloud metadata endpoint (169.254.169.254) which can leak cloud credentials. Two vectors: (1) a direct internal URL, (2) a public URL that redirects inward. Primary defense: resolve the host and reject non-public IPs (loopback/private/link-local). Secondary: disable or re-validate redirects, cap response size, set timeouts, allow only http/https. SSRF = controlling where the request goes; separate from validating what comes back.

CREATE/DROP INDEX CONCURRENTLY can't run in a transaction; migration tools wrap migrations in one by default — so concurrent index ops need an explicit opt-out, and you only reach for them on live production tables, not routine migrations. A linter can be correct in the abstract and still wrong for your context — and following it can actively break things. A tool emits a general heuristic; you own whether it applies here. The skill isn't obeying or ignoring linters — it's dispositioning each warning against your actual context (transaction boundaries, dev-vs-prod, what locks the surrounding statements already take).

The two failure postures — fail loud for violated invariants you control (hardcoded config, the Must… pattern, panic at init), vs fail closed for untrusted runtime input (a security predicate denies on uncertainty rather than crashing — a user-triggerable panic is a DoS).

A `/96` mask on a v4-mapped network (`::ffff:0:0/96`) silently collapses to `0.0.0.0/0` inside Go's `net.IPNet.Contains`. Contains normalizes a v4-mapped IP to 4 bytes and slices the mask to its low 4 bytes — for a /96 those four bytes are all zeros, so the network matches _every_ IPv4 address. In the SSRF classifier this "defense-in-depth" CIDR blocked the entire IPv4 internet (every public IP came back non-public), while real IPv6 slipped past because `len(ip) != len(network)` makes Contains bail early. The table test was the only reason it surfaced before shipping — otherwise a fetcher that refuses every real URL. Lesson: a redundant belt-and-suspenders check isn't free; it can be actively wrong. And the `To4()` normalization at the top of the classifier already handles v4-mapped smuggling (`::ffff:127.0.0.1` → `127.0.0.1` → caught by `IsLoopback`), so the explicit CIDR was both redundant and broken.

SSRF dial-time check closes the DNS-rebinding (TOCTOU) gap. Checking the IP up-front resolves DNS _twice_ — once to validate, once when `http.Client` actually connects — and the attacker controls what the second resolution returns (safe IP first, internal IP second). Putting the `isPublicIP` check in `net.Dialer.Control` collapses those into one: Control fires after resolution, on the exact address about to be dialed, so check and connect see the same IP. Wire it via `Transport.DialContext = dialer.DialContext`; because the client follows redirects through the _same_ Transport, every redirect hop is re-validated for free — which is why "allow redirects" stays safe here and "disable redirects" buys no extra security.

`io.LimitReader(body, maxBytes+1)` (not `maxBytes`) is how you _detect_ an oversized response instead of silently truncating it. A plain `LimitReader(body, maxBytes)` reads at most maxBytes, so an exactly-maxBytes file and a 50× file both come back the same length — indistinguishable. Reading one byte past the cap lets `len(data) > maxBytes` answer "was there more?" and fail loudly, instead of handing a truncated half-page to the LLM. `Content-Type` gating is a fast-fail nicety (a server can lie), not a security control — the real defenses are the IP check, size cap, and timeout.

Untrusted text into an LLM = prompt injection; structured-output schema + private scope + human-in-the-loop confirm are the layered defenses.

Retry belongs where no human is blocked. A background worker can retry itself (return error → backoff → re-run, as the enrich worker does over Asynq) because nobody is waiting on the result — it can burn 90 seconds across attempts and no one notices. A synchronous request handler is the opposite: a user is staring at a spinner, so every backoff step is latency _they_ feel. The dominant pattern there is fail fast and hand the retry to the user — return a clear error + status, and the frontend shows a "Retry?" button. That is client-driven retry: the loop condition is a human clicking, which is why it can never infinite-loop. So a two-layer system retries in its workers and fails fast in its handlers — the same operation, two different retry postures, decided entirely by "is someone blocked?"

The two preconditions for ANY retry — idempotency and retryability. (1) Idempotency: retrying must not double-create, double-charge, or double-send. A GET/fetch is naturally idempotent; a "create" is not, unless guarded by a uniqueness key — which is exactly why CreateImportedJob upserts on (user*id, source_url): the dedup key makes the write idempotent by design, so a retry physically cannot make duplicates. (2) Retryability: only transient failures are worth repeating (429, 503, network timeout). A permanent failure (400, 401 bad key, 404) will fail identically on attempt 2 — retrying it just fails 3× slower and hides the real problem. This is why classifyGeminiError + ErrPermanent exist: the classification is the thing that makes retry \_safe*, not the retry loop itself. And a retry budget must outlast the failure window it is retrying — 3 retries spanning 15s cannot ride out a 60s rate-limit quota; correct classification is necessary but not sufficient.

The whose-fault-is-it rule for HTTP status codes inside a handler. Before every error response, ask "whose fault is this failure?" The database is down, decryption broke, an internal invariant failed → _yours_ → 5xx (and log the real err — a 5xx with no log is a debugging black hole). The user pasted a URL that 404s, set no API key, sent a malformed body → _theirs_ → 4xx (with a message that names the fix). The trap is mapping your own failures to 4xx (the user chases a problem they can't fix) or their input to 5xx (you trip ops alarms and page yourself for user typos). Within 4xx, pick the honest one: 400 = malformed request; 401 = not authenticated; 403 = authenticated but not allowed; 404 = resource doesn't exist; 422 = request well-formed but unprocessable given current state (e.g. no Gemini key on file — the JSON was fine, the account state wasn't).

Defense-in-depth at the handler edge for untrusted input. Two cheap fail-closed guards belong on any endpoint taking user-supplied data: (1) http.MaxBytesReader(w, r.Body, limit) before decoding JSON, so a client can't stream a 500MB body into memory — the request-layer twin of the fetcher's response size cap; (2) validate the _parts_ of a parsed URL, not just that it parsed — url.Parse("http://") succeeds with an empty Host, and url.Parse rarely errors at all, so gate on scheme ∈ {http, https} AND Host != "". Cheap up-front checks fail fast before you spend a network round-trip or an LLM call on input that was never going to work.

Defensive error handling should map to failures that can actually occur. Branches for impossible states aren't free — they encode a wrong belief about your own code (here: "a duplicate import is a conflict," when you specifically built it to be an upsert). The next reader trusts that belief.

A background worker is a "reader" of shared queries too. When GetJobByID gained a public-OR-owned filter, the four HTTP readers were updated but the match worker's UserID: uuid.Nil placeholder wasn't — it kept working for public jobs (their created_by_user_id IS NULL satisfies the OR regardless of the id) and silently failed only for user-private rows. "Compiles + existing data still works" hid it. A nil/placeholder argument that a query's new WHERE clause now depends on is a silent drift hazard.

An LLM extracting from untrusted text will fabricate any field the schema asks for but the page doesn't supply — and a "guessed conventional URL" (company.com/logo.png) is the signature of a hallucinated value: it didn't find a logo, it constructed the canonical path. A general "never fabricate, return null" rule loses to a named schema field with no field-specific guardrail. Two fixes, the second stronger: (a) add an explicit per-field rule, or (b) remove the field from the ResponseSchema entirely — structured output can only emit declared fields, so deletion is a hard structural guarantee, not a behavioral hope. Best of all for a value that's derivable (a logo from the source_url domain): never ask the model — compute it deterministically. Asking a text model for a URL is asking it to guess. Corollary: a structural fix needs no Gemini call to verify — if the field can't be in the schema, it can't be in the output.

A computed is a pure derivation (read-only); a wrong value comes from a write, never from the computed that reads it — debug the assignment, not the derivation.

A field-level error is something the user can fix on that specific input; a toast is the outcome of the whole operation. Validation failures are per-field; network/server failures are per-form. Also think about when you clear errors.title — on resubmit (reset the object) and/or when the user edits that field. A stale "Required" sitting under a now-filled box is its own bug.

(1) an early broad if (status === X) { ...; return } makes any later status === X branch unreachable; overlapping conditions are first-match-wins. (2) Branch on status code, display the server's message; don't hardcode client copy for a condition the server already describes, and don't add a discriminator field until the client needs to branch on it (different text ≠ different behavior).

---

## Personal release review guide — September 2026

This section preserves the repository review findings and the concepts behind the personal release. Older notes above remain a historical learning log; use the current code and `.context/` for implementation status. Knowledge checks are collected here for a walkthrough after the working release, rather than interrupting implementation.

For each topic below, explain it without reading the answer, trace it in the code, and describe a failure scenario. Knowing the vocabulary is not enough: identify the transaction, request, or process boundary where the behavior changes.

### 1. What this application actually does

The reliable path is: external job API → normalized job → PostgreSQL → durable pending match → Redis → algorithmic matcher → stored per-user match → daily digest → email provider. AI summaries and URL extraction are optional. They do not determine whether scheduled job collection or notifications can work.

The original implementation already included a relational database, OAuth, sessions, profiles, matching, import, and tracking. Local Ollama inference was an experiment removed in July. The September review did not prove the historical cause of an API hang: provider errors, two busy workers, and stale frontend state are different failure modes requiring different evidence.

Code: [main](cmd/jobradar/main.go), [source ingestion](cmd/jobradar/scraper_remoteok.go), [matcher](cmd/jobradar/worker_match.go), [digest service](internal/digest/service.go).

Questions:
- Trace one new job from an API response to an email in my inbox. Which steps write data?
- Why can this application work with no AI provider key?
- What evidence distinguishes a frozen HTTP request, a slow queue, and a stale browser view?
- Which components need to stay running when my laptop sleeps? What catches up after restart?
- Which features are necessary for a useful personal release, and which are future product scope?

### 2. Atomicity across PostgreSQL and Redis: the transactional outbox

Saving a job and sending a Redis message are two separate effects. A successful database insert followed by a failed enqueue leaves stored data without processing. Re-fetching the same job does not necessarily repair this, because deduplication may skip it.

The release records `pending_matches` in the SAME PostgreSQL transaction as job/profile changes using database triggers. A dispatcher enqueues these durable records and marks their delivery time only after enqueue succeeds. If Redis is unavailable, the pending record remains. If the process dies after enqueue but before marking, it may deliver twice. Duplicate delivery is safer than silently losing the work, provided processing is idempotent.

Questions from the repository review:
- If a job is saved in PostgreSQL but Redis is unavailable immediately afterward, how should Jobradar discover later that matching still needs to happen?
- Why is wrapping only the PostgreSQL write in a transaction insufficient to make a Redis enqueue atomic?
- Why does the dispatcher mark delivery AFTER enqueue rather than before?
- What does a database trigger buy us? What are its costs in visibility, write overhead, and testing?
- What happens if an experience edit transaction rolls back after its trigger schedules matching?

Answer outline: the same commit persists both the business change and pending work. A periodic dispatcher can recover without knowing which process originally wrote the job. Trigger writes roll back with the original transaction.

Code: [migration and trigger functions](sql/schema/20260920000100_reliable_personal_release.sql), [pending queries](sql/queries/pending_matches.sql), [dispatcher](cmd/jobradar/worker_pending.go).

### 3. Queue delivery, leases, revisions, and acknowledgements

`LPUSH`/`BRPOP` implements FIFO delivery but `BRPOP` removes a task before processing finishes. It is not a crash-safe acknowledgement protocol. The durable database record remains until successful matching; an enqueue timestamp older than five minutes becomes eligible for redelivery.

A revision changes when a job or profile changes. Completion deletes only the revision the worker processed. This prevents an older worker from acknowledging newer work. A transaction-scoped advisory lock serializes deliveries for the same user/job pair. Work is recomputed from current database state, not from a stale profile copied into a message.

Questions:
- What is the difference between at-most-once, at-least-once, and exactly-once delivery?
- Why does a dead-letter queue not recover a task lost during a process crash?
- Why are graceful shutdown and crash recovery separate concerns?
- Why can the same task run twice even after adding a durable outbox?
- How does a revision prevent an old completion from deleting new work?
- How do an advisory lock, a unique constraint, and a delivery lease solve different problems?
- What happens to a transaction-scoped advisory lock when its connection dies?
- Why must a long-term queue failure be visible rather than endlessly hidden in retries?

Code: [Redis queue](internal/queue/queue.go), [worker retries](internal/queue/worker.go), [durable worker](cmd/jobradar/worker_pending.go), [reliability integration tests](cmd/jobradar/reliability_test.go).

### 4. Authentication is not authorization

Authentication identifies a user. Authorization decides whether that user may access a particular resource. A valid UUID is not proof of ownership; UUID unpredictability is not an access policy.

The review found that direct job reads applied public-or-owned filtering, but save/application inserts did not. The repaired SQL inserts from a visible `jobs` selection. If the resource is inaccessible, the query returns no row. Saved/application list queries also enforce visibility. A personal deployment may additionally restrict GitHub sign-in to `PERSONAL_EMAIL`.

Questions:
- Why does checking authentication middleware not protect another user's private imported job?
- Why is `INSERT ... SELECT ... WHERE visible` stronger than trusting a client-provided owner ID?
- When should inaccessible resources return 404 rather than disclose their existence?
- Why must access rules apply to indirect reads such as saved-job lists?
- Why do we scope UPDATE and DELETE by both resource ID and user ID?
- What is the difference between CORS, CSRF protection, cookie flags, and resource authorization?
- How would Laravel policies or a TypeScript service enforce the same rule?

Code: [saved-job SQL](sql/queries/saved_jobs.sql), [application SQL](sql/queries/applications.sql), [auth middleware](internal/auth/middleware.go), [OAuth callback](internal/auth/github.go), [ownership tests](cmd/jobradar/reliability_test.go).

### 5. An interface makes providers replaceable

RemoteOK, Adzuna, and Greenhouse implement `sources.Source`: a name and a context-aware fetch operation returning normalized jobs. Ingestion owns sanitization, persistence, and health reporting. Adapters own provider URLs, authentication, pagination, and response shape. A provider should not make unrelated application code depend on its vendor-specific JSON.

Adzuna uses application credentials in query parameters, so transport errors must not log the full URL. Pagination consumes quota. Five queries with three pages twice per day are up to 30 scheduled requests/day, before retries and restarts. Remote eligibility is not inferred merely because a search contains the word “remote.” Company-board APIs require a watchlist of board slugs rather than a global search.

Questions:
- What belongs in a provider adapter, and what belongs in ingestion?
- Why do optional fields need null/unknown rather than fabricated zero/false values?
- Why distinguish a provider's external ID from our database UUID?
- Why can `(source, external_id)` deduplicate within a source but not guarantee cross-source deduplication?
- What do API rate limits count: jobs returned, HTTP requests, or pages?
- How do retries and restarts affect a monthly quota budget?
- Why must credentials in query strings never appear in logs?
- How would I add a fourth provider without copying database and queue code?

Code: [source interface](internal/sources/sources.go), [Adzuna](internal/sources/adzuna.go), [Greenhouse](internal/sources/greenhouse.go), [configuration](cmd/jobradar/config_sources.go), [source tests](internal/sources/sources_test.go).

Official references: [Adzuna API](https://developer.adzuna.com/overview), [Adzuna limits/attribution](https://developer.adzuna.com/docs/terms_of_service), [Greenhouse job-board API](https://docs.greenhouse.io/job-board.html). Check current terms when changing product usage.

### 6. Idempotent ingestion and job freshness

Repeated fetches update source-owned job fields rather than creating duplicate rows. `last_seen_at` records observation; `posted_at` records source publication when available; `created_at` records first insertion here. These timestamps answer different questions. `updated_at` advances on relevant content changes, avoiding a full rematch on every identical fetch.

The initial schema already contained publication and expiry fields, but old ingestion did not populate publication dates. Digests exclude postings older than 30 days using publication date when known, otherwise insertion date, and respect explicit expiry. This is not proof that a job is still open: source absence/closure and cross-source repost identity remain distinct problems.

Questions:
- Why can “fetched today” differ from “posted today”?
- What does an upsert do on insert versus conflict?
- Why should an unchanged fetch not invalidate every existing match?
- What are partial unique indexes, and why do public and private jobs need different uniqueness rules?
- What should happen if a job disappears from a complete company-board feed versus one page of a search API?
- Why is a migration a versioned transformation rather than a reason to drop a database?

Code: [job queries](sql/queries/jobs.sql), [private/public indexes](sql/schema/20260611025256_jobs_add_created_by_user.sql), [new migration](sql/schema/20260920000100_reliable_personal_release.sql).

### 7. Matching is a heuristic, not a hiring probability

Title similarity, required-skill coverage, and experience overlap produce a ranking. Missing signals renormalize the weights. A score is not calibrated as a probability of receiving an interview. User-imported jobs express intent, so they bypass the title rejection gate.

The review found that prose mentions could inflate skill coverage and that an owned skill could be reported missing when absent from the description. Required-skill coverage now compares source requirements to the user's skill map; missing skills compare against the full known user skill set. Description mentions remain a separate display signal. Adzuna/Greenhouse may not provide a structured skill list, so the matcher cannot pretend it knows every requirement. Location, work authorization, seniority, and salary need human review; changing the scoring model requires representative examples, not just code that produces higher numbers.

Questions:
- Why is unknown information different from a zero score?
- Why compare required skills against the profile rather than only description mentions?
- Why does an exact title match not prove eligibility?
- How can adding a new job source change score distributions without changing the algorithm?
- What examples would I use to evaluate relevance and false positives?
- Why should AI enrichment not overwrite the deterministic score?

Code: [matcher](internal/matcher/matcher.go), [worker policy](cmd/jobradar/worker_match.go), [regression tests](internal/matcher/matcher_test.go).

### 8. Optional AI, bounded work, backoff, and bulkheads

Slow AI requests used to share two workers with scraping and matching. A separate enrichment queue with one worker isolates this latency. Automatic enrichment is disabled by default through `AI_ENRICHMENT_ENABLED`; manual URL extraction can still use a configured key. Provider calls and overall operations have deadlines, prompt text is bounded, and retries carry provider cooldowns where available.

Permanent errors (invalid credentials or unsupported requests) differ from transient errors (429, 503, timeout). When several providers fail, a permanent error from the last provider must not erase an earlier retryable failure. A failed database write must return an error rather than tell the queue that work succeeded.

Questions:
- What is a bulkhead, and why does adding workers alone not solve an overloaded provider?
- What is the difference between concurrency, request rate, and daily quota?
- Why use exponential backoff with jitter? What is a thundering herd?
- Why should `Retry-After` override a shorter generic retry delay?
- Why do we need both per-provider and whole-operation deadlines?
- What does cancelling a Go context actually do, and what requires explicit cooperation?
- Why is valid JSON not sufficient validation of an LLM result?
- How do input limits, secret encryption, and credential redaction address different risks?

Code: [enrichment worker](cmd/jobradar/worker_enrich.go), [Groq errors](internal/llm/groq.go), [worker backoff](internal/queue/worker.go), [import handler](cmd/jobradar/handler_import.go), [encryption](internal/crypto/crypto.go).

### 9. Email outbox and the ambiguous-success problem

A digest is rendered and persisted with its selected jobs before sending. The persisted body is immutable for retries. A unique `(user_id, digest_date)` limits daily preparation, and the delivery ledger reserves a user/job once. Only `DIGEST_RECIPIENT` is eligible; the user must exist and have notifications enabled.

Sending an email and committing “sent” to PostgreSQL cannot be one ordinary database transaction. If the provider accepts a message but the connection or database write fails, we do not know whether the inbox received it. Resend supports a stable idempotency key with a 24-hour retention window. This implementation stops automated retries before that window expires. GoHighLevel does not provide a documented equivalent in the chosen endpoint, so ambiguous sends are marked `uncertain` for manual reconciliation rather than silently sending another copy. “Sent” means accepted by the provider, not proven inbox delivery.

Questions:
- Why is “check ledger, send email, write ledger” not exactly-once delivery?
- What happens if the process crashes immediately before or after the provider accepts a message?
- Why must an idempotency key and request body remain stable across retries?
- Why does a UNIQUE constraint in PostgreSQL not deduplicate an external email by itself?
- Why stop retrying when a provider's deduplication window expires?
- How should a human reconcile `uncertain` delivery? Why not just delete the ledger?
- How do an opt-out and a fixed personal recipient prevent unintended recipients?
- Why does changing provider/sender while mail is pending require review?
- Why are timezone-aware calendar days preferable to “sleep 24 hours” for a morning digest?

Code: [digest service](internal/digest/service.go), [digest queries](sql/queries/digests.sql), [mailer interface/adapters](internal/mailer/mailer.go), [digest integration tests](internal/digest/service_test.go).

References: [Resend idempotency](https://resend.com/docs/dashboard/emails/idempotency-keys), [HighLevel send endpoint](https://marketplace.gohighlevel.com/docs/ghl/conversations/send-a-new-message/). Self-hosting SMTP is possible, but DNS authentication, reputation, abuse prevention, bounce handling, and deliverability become our operational responsibility.

### 10. Frontend state and backend state are separate

Import confirmation returns before asynchronous matching finishes. Updating a Vue array with the confirmation response does not subscribe it to later database writes. The import view briefly polls the imported job, with bounded attempts and cleanup on unmount. Normal feed requests use an AbortController so an older response cannot overwrite a newer search.

The feed uses server-side limits, offsets, search, and filters. Loading every description into the browser is wasteful as the database grows. Empty collections are serialized as `[]`, not `null`. Stable IDs prevent duplicate rows when pagination or imports overlap.

Questions:
- Why can a background job finish while the UI still shows “not matched”?
- When would I use polling, server-sent events, or WebSockets?
- How can two search requests arrive out of order, and how does cancellation help?
- Why clean up intervals when a component unmounts?
- What is the tradeoff between offset and cursor pagination under concurrent inserts?
- Why is a TypeScript interface not runtime validation of an HTTP response?
- Why must untrusted HTML be sanitized even when it came from an API?
- What should optimistic UI updates do when the backend rejects a mutation?

Code: [JobsView](frontend/src/views/JobsView.vue), [job handlers](cmd/jobradar/handler_jobs.go), [JobDetail sanitization](frontend/src/components/JobDetail.vue), [dashboard](frontend/src/views/DashboardView.vue).

### 11. Observability, reproducible setup, and deployment

A hardcoded green status indicator is not a health check. The status page reads PostgreSQL/Redis connectivity, durable backlog size, and persisted source health. The public health endpoint is a basic liveness signal. Source health is the result of the last fetch, not a promise that the provider is currently reachable.

A clean setup must declare every required environment variable and dependency. `ENCRYPTION_KEY` must remain stable across restarts to decrypt stored provider keys. The production image builds the Vue app and Go binary separately, then runs as a non-root user. PostgreSQL and Redis use persistent volumes. The Vue history router requires an index.html fallback, while unknown API routes must still return 404.

Questions:
- What is the difference between liveness, readiness, and a dependency health check?
- Which log fields help locate a failure without leaking secrets?
- Why should deployment not depend on a developer's root node_modules directory?
- What belongs in the image versus an environment variable versus a persistent volume?
- Why do HTTPS, HttpOnly, Secure, and SameSite each matter for session cookies?
- What survives a container restart? What disappears if its volume is deleted?
- How do I restore a PostgreSQL backup and verify it actually works?
- Why should migration execution happen before switching traffic to the new application?
- Why must unknown `/api/...` paths not return the SPA HTML page with status 200?
- Why do we test against dedicated ports and a database whose name ends in `_test`?

Code: [production Dockerfile](Dockerfile), [production Compose](compose.production.yml), [migrations image](Dockerfile.migrations), [SPA/status handlers](cmd/jobradar/handler_system.go), [test database guard](internal/testutil/database.go), [environment example](.env.example).

### 12. Testing behavior rather than implementation details

The review found a frontend starter test that did not provide Pinia and expected old template text. The repaired test creates the required app context and checks the public shell. Integration tests exercise ownership, transactional pending work, stale acknowledgements, duplicate delivery, profile-change rematching, opt-out, provider failures, and stable email retry identity. External-provider unit tests use local HTTP servers; they do not send real email.

Questions:
- Which bugs can a unit test catch, and which require a real PostgreSQL constraint/transaction?
- Why does “the build passes” not prove OAuth or actual email delivery works?
- What makes a mock useful, and what important behavior can it hide?
- Why should retries and timeouts be tested as normal behavior, not rare exceptions?
- How would I simulate a crash between two side effects?
- Why should test setup never flush my development Redis database?
- What does Go's race detector detect, and what database races does it not detect?

Code: [backend integration tests](cmd/jobradar/reliability_test.go), [digest tests](internal/digest/service_test.go), [mailer tests](internal/mailer/mailer_test.go), [frontend test](frontend/src/__tests__/App.spec.ts).

### 13. The same fundamentals in Go, Laravel, and TypeScript

| Concept | This repository | PHP/Laravel equivalent | TypeScript/Next.js equivalent |
| --- | --- | --- | --- |
| HTTP routing | Chi handlers | Routes/controllers | Route handlers/API routes |
| Validation | Go request decoding and checks | Form requests/validation rules | Runtime schema validation |
| Authentication | OAuth + Redis session cookie | OAuth integration + session guard | Auth library + server-side sessions |
| Authorization | User-scoped SQL | Policies + scoped queries | Authorization in service/repository queries |
| Dependency inversion | Go interfaces | PHP interfaces + service container | Interfaces/functions + dependency injection |
| Transactions | pgx + SQLC | DB::transaction | Database client transactions |
| Migrations | Goose SQL files | Laravel migrations | SQL migrations or migration tooling |
| Durable asynchronous work | PostgreSQL pending records + Redis | Outbox + queue jobs | Outbox + dedicated worker |
| Background lifecycle | Goroutines/context | Queue worker processes | Long-lived worker processes; not a request callback |
| Retry/idempotency | Revisions, constraints, email keys | Same guarantees through jobs and constraints | Same guarantees through workers and constraints |
| Reactive UI | Vue refs/computed/Pinia | Blade/Livewire or separate SPA | React state/query cache |
| Safe rendering | Vue escaping, DOMPurify, html/template | Blade escaping/sanitization | React escaping/sanitization |
| Deployment | Binary + static assets + services | PHP runtime + assets + services | Node runtime/serverless + assets + services |

Questions:
- Which guarantees come from PostgreSQL and HTTP rather than Go?
- Which Go techniques would change in a PHP worker while keeping the same delivery guarantees?
- Why is starting background work inside a serverless request unreliable after the response?
- Why does changing an ORM not remove the need to understand indexes, joins, transactions, and query plans?
- Explain the entire application using only language-independent terms, then point to each implementation.

### 14. Review workflow and release discipline

Use a feature branch based on `dev`, stage/commit there, verify, then merge into `dev`; promote to `main` only as a release step. Existing job-import commits were brought into the personal-release feature branch without rewriting them. Documentation must describe current behavior rather than preserve contradictory “done” and “not started” claims.

Questions:
- Why isolate unfinished work on a feature branch?
- What is the difference between a fast-forward merge and rewriting history?
- Why can local origin/* references be stale until fetch?
- What evidence should a release checklist contain besides a green build?
- How do I explain a design tradeoff, reproduce a failure, and locate its fix during an interview?

Walkthrough checklist for later: trace the request; identify the invariant; locate the code; explain the test; describe one alternative; explain why the chosen tradeoff fits a small personal application.


### 15. Connection pools, transaction boundaries, and bounded resources

A connection pool is a limited shared resource. A worker that opens a transaction holds one connection until commit or rollback. If it then runs queries through the pool instead of through the transaction, it needs another connection. Several workers can exhaust a small pool this way: each holds a connection and waits for another. Increasing the pool can hide the design problem while consuming more memory.

The durable matching worker binds SQLC queries to its transaction with `WithTx(tx)`. Reads, score writes and revision acknowledgement use that connection, and score persistence plus acknowledgement commit together. Startup database checks have a five-second deadline and close a failed pool. AI prompts bound each section, including candidate experience, so a long profile cannot create an unexpectedly large provider request. These limits complement request deadlines; neither replaces the other.

Questions for our later walkthrough:
- Where does this worker acquire a connection, and when does it release it?
- Why must queries use the transaction object to participate in that transaction?
- What happens if writing the score succeeds but acknowledging the pending revision fails?
- How could holding one resource while waiting for another deadlock a system?
- Why is increasing a connection pool not always the right fix for slow requests?
- What is the difference between a timeout, a concurrency limit, a queue limit, and an input-size limit?
- How would these same resource limits appear in Laravel workers or a TypeScript service?

Code: [transactional matching](cmd/jobradar/worker_pending.go), [database startup](internal/database/database.go), [prompt bounds](internal/llm/llm.go). Evidence: [release verification](docs/RELEASE_VERIFICATION.md).


### Profile matching settings and small commits

Matching preferences belong to the user profile. Provider secrets remain server configuration. A dedicated settings resource uses authenticated GET/PUT, validates score 0–100, supported search country, timezone and one or two distinct delivery hours. PUT replaces the complete settings resource; the server ignores any supplied user ID and derives ownership from the session. PostgreSQL checks protect invariants even outside HTTP. Relevant preference changes schedule rematching transactionally.

Small commits isolate a feature or fix for review and debugging. The unpushed release commit was regrouped without changing its final tree; a local backup reference preserves recovery. No remote history was rewritten. Future changes should be committed as each coherent change passes its checks.

Questions: Why do preferences belong in the database but API credentials stay in the environment? How do PUT and PATCH differ? Why validate both in HTTP and PostgreSQL? How does the session prevent editing another account? How does a smaller commit help git bisect? What changes when rewriting published rather than local history?

Code: `handler_match_settings.go`, `sql/queries/match_settings.sql`, `20260920000200_match_preferences.sql`.


### Scores versus preference gates

Title similarity contributes 45%, required-skill coverage 30%, and experience text overlap 25%, renormalized over available signals. A score is a ranking heuristic, not a probability of getting hired. Geography, work mode, employment type and compatible salary ranges are separate eligibility gates. The user chooses whether missing metadata is allowed; notes explain missing facts or exclusions. Location matching uses normalized phrases, not geocoding, and remote does not imply worldwide eligibility. Industry/company-stage preferences are recorded but not enforced because sources lack reliable fields. Salary comparisons assume the same pay period and currency; no exchange-rate conversion is performed.

Best matches use the user's threshold and exclude scores awaiting recomputation. Imported jobs remain accessible even when ineligible, allowing deliberate applications outside preferences.

Questions: Why separate ranking from filtering? Why is missing information different from false? Can a 90 score mean a 90% hiring chance? Why must a remote US-only job not automatically qualify someone in Kenya? Why exclude stale scores during rematching? What metadata is missing from each source?

Code: `internal/matcher/preferences.go`, `worker_match.go`, `sql/queries/jobs.sql`.


### Two digest slots and civil time

Email delivery now uses the profile threshold, timezone, one or two hours, and maximum jobs per message. The environment remains a deployment-wide safety switch and fixes the recipient/provider. A unique `(user, local date, scheduled hour)` prevents duplicate slot sends, while the user/job ledger prevents the morning jobs repeating in the evening. After downtime only the latest due slot is prepared, avoiding a burst of old emails. A repeated daylight-saving hour uses the same slot identity; a skipped hour is caught up when later running. Empty slots do not send filler email.

Questions: Why use an IANA timezone instead of a fixed UTC offset? Why do we need both per-slot and per-job deduplication? What happens when the laptop sleeps through the morning? Why can rolling back a uniqueness migration be unsafe once new data exists?

Code: `internal/digest/service.go`, `20260920000400_digest_slots.sql`.


### Preferences UI and honest controls

Work mode and geography are different dimensions. The previous location selector stored labels such as Remote as if they were places. Locations now accept actual alternatives, and matching/delivery settings have their own validated form. Unsupported dimensions are explained rather than silently presented as active filters. The server remains authoritative even when HTML input constraints exist.

Questions: Why can a frontend numeric input not replace server validation? Why distinguish source search market from candidate location? Which preferences act immediately, and which wait for rematching or the next source fetch?

Code: `MatchingSettings.vue`, `PreferencesTab.vue`, `SettingsTab.vue`.


### Profile-driven source searches

Adzuna credentials are configured once on the server. Each scheduled fetch reads the personal account's current desired roles (up to five), first desired location and selected supported search market. Editing profile preferences does not require restarting the server; new searches run on the next 12-hour poll. Sources that expose entire feeds/boards are collected normally and then filtered per profile. A source cannot promise jobs in an unsupported country. With multiple local accounts, PERSONAL_EMAIL selects which account drives the capped Adzuna search budget.

Questions: How do upstream search criteria differ from downstream matching? Why cap role and page fan-out? Why not fetch every desired location for every role without considering quota? How would this design change for a multi-user product?

Code: `config_sources.go`, `internal/sources/adzuna.go`.


### In-app notifications without external delivery

A qualifying match creates one database notification per user/job in the matching transaction. Repeated processing or rescoring cannot flood the user with duplicates. The list joins current scores/preferences and excludes pending rematches, so raising the threshold immediately hides lower-quality results. Opening a notification marks only that user's record read and deep-links to the job. The bell shows unread items among the latest 50 eligible notifications, not a lifetime total. Polling while the app is open uses the existing server, with no separate notification vendor.

Questions: Why use a unique constraint instead of checking first? How does transaction rollback prevent a notification for a score that was never committed? Why recheck current eligibility when listing old notifications? Why scope mark-read by both notification ID and user ID? When would WebSockets be preferable to polling?

Code: `sql/queries/notifications.sql`, `handler_notifications.go`, `worker_match.go`.


### Best-match views, notification polling and deep links

The job feed has a Best Matches filter ordered by score on the server before pagination. The profile tab shows the first 20 with a link to the full feed. The notification bell polls every 30 seconds while visible and cancels work on unmount. Notifications open `/jobs?job=...`; the detail view fetches that specific authorized job even when it is not in the currently loaded page. Client-side sorting only one page would not produce the globally best matches.

Questions: Why must ranking happen before LIMIT? Why should a deep link not assume the job is in the first page? Why cancel fetches and timers when components unmount? How do authenticated server queries protect private links?

Code: `JobsView.vue`, `BestMatchesTab.vue`, `NotificationBell.vue`.


### Copy for AI: portable context without an API bill

The authenticated Markdown export combines the selected authorized job with saved resume sections and deterministic score details. It uses a read-only repeatable-read transaction for a consistent snapshot, explicitly selects safe profile fields, strips job HTML, excludes API/OAuth secrets and primary contact fields, and disables caching. User-authored resume text may itself contain contact information: review the export before sharing. A clipboard fallback exposes selectable text on browsers without clipboard permission or a secure origin. Nothing calls a model or submits applications.

The prompt asks for truthful tailoring and treats posting content as data. This reduces instruction confusion but cannot guarantee an external model obeys it; inspect all generated claims. Markdown is convenient for manual chat use; JSON sections retain structured facts. Ordinary chat subscriptions do not grant generic API credits. A future agent must use the provider's supported authentication and plan, rather than assuming a chat subscription is an unlimited automation API.

Questions: Why not serialize the entire user object? What is a read-only consistent snapshot? Why is clipboard access sometimes unavailable on a local-network HTTP address? How does prompt injection differ from HTML injection? Why must every rewritten experience bullet remain grounded in actual work?

Code: `handler_ai_context.go`, `JobDetail.vue`.


### Job import without a model

Import first checks a fetched page for Schema.org JobPosting JSON-LD, including graph/array forms. If available, it reads title, employer, description and location without an AI request. Otherwise it can use an optional configured extractor, then fall back to page title and visible text. It never guesses missing company, pay or remote status. The review form remains mandatory. Private WhatsApp messages, login-only pages, blocked scrapers and JavaScript-only content cannot reliably be fetched as public HTML; manual entry remains available.

Questions: Why prefer structured metadata over guessing from prose? Why is a review form still necessary? Why doesn't an HTTP fetch render a JavaScript application? Why must fetching links defend against SSRF? How do fallbacks preserve a useful workflow during provider outages?

Code: `internal/extraction/page.go`, `handler_import.go`. Schema reference: https://schema.org/JobPosting .


The earlier environment-driven digest hour/threshold settings were removed once the profile became authoritative. Keeping two conflicting configuration paths would make troubleshooting misleading. Deployment configuration controls whether email is permitted at all; profile settings control relevance, frequency and opt-out.

Question: How do you choose one authoritative source when migrating configuration from environment variables to per-user settings?


A production build caught a Vue template expression that type checking accepted after formatting: multiple statements inside a click attribute were parsed differently. Moving the behavior into a named script function fixes the template and improves readability. Run the actual production compiler as well as the type checker; they verify different parts of the application.

Question: Why can type checking pass while the production template build fails? Code: `NotificationBell.vue`.


### Personal hosting, email economics and future agents

A digest containing ten jobs is one email to one recipient. Two daily digests use roughly sixty monthly messages, within Resend's documented free transactional quota at verification time. Testing sender domains can restrict the recipient; authentication email, sender authorization and provider acceptance are separate concerns. A GoHighLevel location ID selects a subaccount, while a contact ID identifies the recipient.

The personal scheduler runs inside the Go process, so free hosting that sleeps also stops background work. Render's free PostgreSQL lifetime is another constraint. Existing home hardware can run the application stack without a local model, but power, connectivity, storage/backups and wakefulness remain operational responsibilities. Private HTTPS access can avoid exposing the database or app publicly. A hosted external agent cannot automatically access a private LAN endpoint.

Questions: Why does an idle web-service sleep policy break cron-like work inside the process? Why is free compute different from persistent free storage? How would you move scheduling into an external job runner? Why is localhost inside a container different from localhost on the host? Why is a subaccount ID not an email recipient? What should a future read-only agent token allow, and how would it be revoked? Why is a chat plan not automatically an API allowance? What must be backed up besides PostgreSQL data?

Guide and current primary sources: `docs/PERSONAL_WORKFLOW.md`. These are operational tradeoffs, not a promise that external services will remain free forever.


### Editor workspace diagnostics versus compilation

Targeted Go tests and gopls checks passed, but `go list ./...` reproduced a workspace traversal failure on PostgreSQL-owned `pgdata`. Git ignore rules do not control Go or language-server directory traversal. The installed editor is LazyVim/Neovim; its project-local `.lazy.lua` adds gopls directory filters for database storage and generated frontend dependencies, preserving the user's existing analysis settings. Reload/restart the editor to load the project spec. `make test` deliberately checks actual source roots (`./internal/... ./cmd/...`) and does not scan runtime storage. No database permissions or files were altered.

Questions: Why does .gitignore not exclude files from every tool? How can a workspace-loading error differ from a Go compilation error? Why should PostgreSQL's restrictive storage permissions remain intact? Why is runtime data outside a source tree preferable in a new deployment?

Code/config: `.lazy.lua`, `Makefile`; source: https://go.dev/gopls/settings#directoryFilters .

## Preference-aware numeric ranking

The core title/skills/experience score has weight 80. Each configured preference with known evidence (location, work mode, employment type, comparable complete salary range, industry/category, company stage) has weight 5 and contributes 100 for a match or 0 for a mismatch. Divide by the sum of active weights. Unknown evidence is omitted from numeric ranking; the existing include-unknown setting still controls eligibility. A mismatch also excludes a job from recommendations, even if its skills score is high. Scores are heuristics, not probabilities. Industry/category matching uses normalized literal phrases, not a semantic taxonomy or company-industry inference.

Implementation: `internal/matcher/preferences.go`, `internal/matcher/preference_score_test.go`.
Review questions: Why keep eligibility separate from rank? Why does missing metadata differ from a known mismatch? Can two jobs with different available evidence have comparable-looking scores? How would we calibrate weights with real user feedback? How would this pure scoring function translate to Laravel or TypeScript?

### Carrying preference evidence through the application

`20260921000100_preference_ranking.sql` adds nullable job industry/category/company-stage metadata and rematches existing users. Industry and company-stage profile changes now schedule durable rematching. Adzuna category and contract-time metadata flow through the shared ingestion adapter; structured import reads industry, occupational category and employment type. The import review form lets the owner supply explicit missing facts. Unknown facts remain unknown. Preferred industries also accepts source-category phrases (for example IT Jobs); this is literal matching, not equivalence between every provider taxonomy.

The worker applies `PreferenceScore` before persistence, notifications, email selection and AI gating. Existing required-skill scoring still needs explicit job skills; description-only mentions are display evidence, not invented requirements.

Code: `cmd/jobradar/worker_match.go`, `internal/sources/adzuna.go`, `internal/extraction/page.go`, `frontend/src/components/ImportJobPanel.vue`.
Questions: Why can adding a column without updating ingestion leave a feature ineffective? Why must changing profile preferences invalidate cached scores? Why shouldn't an occupational category automatically establish the employer's industry? Where is source evidence reviewed before persistence?

## Read-only access for personal workflows

A Jobradar account can issue a random, expiring, revocable bearer token for explicitly registered export routes. Browser login remains required to create/list/revoke tokens. Only a SHA-256 hash is stored; the original secret is shown once. These tokens can read public/owned jobs and the owner's AI-context resume export, never mutate jobs or change settings. Revocation and expiry are checked on every request. The personal-instance email restriction still applies. Do not put a token in a URL or commit it; send it in the Authorization header over HTTPS.

Code: `internal/auth/read_token.go`, `cmd/jobradar/handler_read_tokens.go`, `sql/queries/read_tokens.sql`.
Questions: How do authentication, authorization and scope differ? Why is a fast hash suitable for a random 256-bit token but not a human password? Why must a token never pick another user's ID? How does revocation differ from a self-contained JWT? What private information is included in AI context?

## Ollama versus optional cloud AI

Jobradar does not read OLLAMA_HOST or GROQ_API_KEY from .env. They are obsolete for this app. Removing an environment variable does not stop an independently running Ollama service or delete model files. Groq/Gemini integration remains optional and uses encrypted per-user keys saved in Profile → Settings. AI_ENRICHMENT_ENABLED=false disables automatic enrichment, but URL import can still use a configured cloud key as a fallback. Remove saved provider keys for the no-model path. Matching, digest emails and Copy for AI do not need a model. Jobradar's own read API is independent of any ChatGPT/Claude subscription; importing those results into an external AI workflow is a separate integration.

Questions: Where does computation happen for local versus hosted inference? Why is disabling a worker different from disabling all call sites? Which settings are read from environment variables versus user-owned database records?

## Filtering and paginating the job export

GET `/api/export/jobs?min_score=70&date=2026-09-21&timezone=Africa/Nairobi&limit=100` returns your scored jobs as JSON. The caller overrides the threshold for that request only. Defaults use profile threshold/timezone, eligible-only=true, all dates, and exclude explicitly expired jobs. `date_field=added|posted|matched` means first imported, source posting timestamp (unknown dates excluded when filtering), or first scored timestamp. `date=today` is supported, but pin a literal date while paginating across midnight. Local-day boundaries use calendar addition so daylight-saving days need not be 24 hours.

Repeat the same filters with `cursor=next_cursor` until next_cursor is empty. Score/UUID keyset pagination avoids offset shifts from insertions; this is not a frozen snapshot, so jobs rescored during a long export can move. A workflow should deduplicate IDs and rerun its date window. Pending scores are excluded. `eligible_only=false` explicitly permits preference mismatches while retaining score/date/ownership guards. Other arguments: source, location (substring), exclude_applied, remote_only, include_expired. Limit is 1–100 per page, not a total-job cap. Each job includes full saved description and score details. GET `/api/export/jobs/{id}/ai-context` returns Markdown with the owner's saved resume context for tailoring.

Code: `cmd/jobradar/handler_export.go`, `sql/queries/export_jobs.sql`, `cmd/jobradar/handler_export_test.go`.
Questions: Why must filters run before LIMIT? What does a date actually refer to? Why does a 23-hour local day break naive arithmetic? What can change between pages without snapshot isolation? Why is request-scoped filtering different from modifying notification preferences?

### Managing workflow tokens in Settings

Profile → Settings → Personal workflow access creates a 90-day token, shows it once in a selectable field, lists expiry and permits immediate revocation. The interface states that AI context includes saved resume information. The token is held in component memory, not localStorage. Code: `frontend/src/components/AccessTokens.vue`.
Question: Why does showing a credential once reduce exposure, and what should a user do after losing it?

## Proposed local email/password login — not implemented

The current schema has no password hash. A future migration should store local credentials separately from exported user/profile fields. Email/password can be the default login with optional GitHub OAuth, sharing the existing Redis session system. Startup currently requires GitHub configuration; that validation must become conditional. For a personal installation, create an owner account through a controlled setup path and disable public registration. PERSONAL_EMAIL is an account allowlist, not an authentication method; password login does not replace it. Changing login method must preserve the same user ID and must not silently link an unverified email to an existing account.

Use salted Argon2id hashes and login throttling; define password change, recovery and session invalidation before treating this as finished. Recovery for a private local owner can use an explicit server-side reset command rather than requiring an email service. Public sign-up/recovery would require a separate verified-email flow. This is a follow-up design, not a claim that password login exists.
References: https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html and https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html
Questions: Why never store plaintext/reversibly encrypted passwords? How is an allowlist different from login? Why must account linking prove control of both identities? How do password changes revoke existing sessions? How would the same design map to Laravel authentication?

## Email/password credentials — storage foundation

Passwords live in a separate `user_passwords` table, never in profile responses or Copy for AI. Argon2id uses a random 16-byte salt, 19 MiB, two passes, one lane and a 32-byte output. The encoded hash includes its algorithm parameters. Passwords require 15 characters, allow spaces/Unicode and are bounded at 1024 bytes. Verification uses constant-time comparison; a two-operation bound limits simultaneous hashing memory. Unknown accounts can run a dummy verification instead of exposing a cheap early exit.

`users.auth_version` enables invalidation of old sessions after password changes. One-time setup-token hashes expire in 15 minutes; they are separate from passwords and are issued by a server-side owner command. No password is placed in an environment variable or command-line argument.

Code: `internal/password/password.go`, `sql/schema/20260921000300_password_login.sql`, `sql/queries/passwords.sql`.
Review: Why is password hashing deliberately slow while API-token hashing is fast? What does the salt protect against? Why bound both password input size and concurrent hashes? Why must a public user serializer never include password_hash?
Reference: https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html

## Email login and owner recovery

POST `/auth/password/login` verifies the password and creates the same Redis/HttpOnly-cookie session used by GitHub. Login failures use a generic response; account and client-address attempt counters are atomic in Redis (10/account and 50/direct-IP per 15 minutes). Forwarded-IP headers are not trusted. JSON and allowed-origin checks protect credential-changing requests from browser CSRF. Missing Redis fails closed. GET `/auth/methods` lets the UI hide GitHub when it is unconfigured; GitHub credentials are no longer needed for a local-only installation.

An existing GitHub account may add a password after signing in within the last ten minutes. Once a password exists, changes require the current password. User-row locking prevents competing credential updates. Each change increments auth_version, revokes workflow tokens and invalidates all outstanding setup links in the same transaction. Every browser session is checked against auth_version; deleted accounts and older sessions cannot authenticate. A fresh session is issued after a successful change.

For a local installation or lost password, the server owner runs `go run ./cmd/jobradar --password-setup you@example.com`. It creates the local account if absent, preserves existing accounts, and prints a private 15-minute link. The secret is in the URL fragment, so it is not sent in ordinary HTTP requests or access logs. The browser exchanges it once for a chosen password. There is no public registration or email-recovery dependency. Treat server/database access as administrative authority; never expose this command through an unauthenticated endpoint.

Code: `internal/auth/password.go`, `internal/auth/password_setup.go`, `internal/auth/sessions.go`, `cmd/jobradar/main.go`. Integration tests: `internal/auth/password_test.go`.
Review: Why doesn't a browser session automatically authorize changing an existing password? How do version numbers revoke sessions without scanning Redis? Why must recovery-token consumption and password update commit together? How is operator provisioning different from public email verification? What changes when serving many users behind one reverse proxy?

## Keeping GitHub and password login on the same account

Existing GitHub users add a password to their existing user ID. A locally provisioned account connects GitHub explicitly from Settings after recent sign-in. OAuth state records the intended user and is consumed atomically with Redis GETDEL. The callback checks the same live recent session and GitHub's verified primary email; a public profile email is not trusted as verification. An unrelated GitHub identity is never silently attached because its email matches. Returning GitHub users retain the same profile, scores and saved jobs.

Code: `internal/auth/github.go`, `internal/auth/github_test.go`. Tests use a mock OAuth provider to verify explicit linking, verified-email enforcement, returning sessions and single-use state. Real GitHub consent remains a manual acceptance test.
Questions: Why is automatic account linking dangerous? What prevents login CSRF? Why must checking and consuming OAuth state be atomic? Why does authentication identity differ from a mutable display email?

## Email-login forms and credential lifecycle

The login page defaults to email/password and asks `/auth/methods` whether GitHub is available. A setup fragment opens a password-confirmation form; its secret is removed from the address bar and retained only in component memory. Password fields use browser password-manager autocomplete hints, clear after submission, and do not enter Pinia/localStorage. Profile Settings lets a freshly signed-in GitHub user add a password or a password user change it. Saving remounts the token list because existing workflow credentials were revoked.

Code: `frontend/src/views/LoginView.vue`, `frontend/src/components/PasswordSettings.vue`, `frontend/src/__tests__/LoginView.spec.ts`, `frontend/vite.config.ts`.
Review: Why must frontend validation be repeated on the server? Why do new auth routes need Vite proxy entries? Why is a password confirmation field a UX check, not an authentication factor? How do we keep two login methods attached to one user ID?

### Local rollout and acceptance evidence

After isolated migration rollback/reapply and automated checks, backed up the existing development database and applied all eight outstanding migrations. Then started PostgreSQL/Redis and the Go/Vite servers. A successful HTTP health check alone does not prove a login works: also checked auth-method discovery, Vite proxy routing, OAuth redirect destination and invalid-password rejection. Automated tests exercise successful email sessions and mock GitHub callbacks; Sam must perform real GitHub consent and choose their own password. Live RemoteOK/Adzuna polling succeeded; source save counts are not unique-new-job or matching-job counts. Email sending remains disabled until separately tested.

Review: How do unit tests, integration tests, smoke checks and user acceptance differ? Why back up before migrating an existing database even after testing a clean schema? Why use localhost consistently when OAuth cookies and redirects are configured for that hostname?

## Review fix: autosave must preserve newer edits

The original preference store sent overlapping full snapshots and replaced the editable state with every server reply. A slow response could remove a checkbox clicked after that request began. Saves now run sequentially with copied request payloads; response reconciliation checks the revision and the current draft. An in-flight response cannot overwrite a newer edit waiting for its debounce. Failed saves keep the draft and do not poison the queue. Persistent Saving/Unsaved/Retry/All changes saved text and a visible Save preferences button supplement autosave.

Desired locations are individual removable tags. Enter/Add commits a tag, so `Nairobi, Kenya` stays a single place; commas are not delimiters. Empty salary inputs become null rather than an invalid empty string.

Code: `frontend/src/stores/preferences.ts`, `frontend/src/components/PreferencesTab.vue`. Regression tests in `frontend/src/__tests__/preferences.spec.ts` reproduce rapid checkbox changes, delayed responses, retry after failure and location tagging.
Questions: Why is debouncing alone insufficient for out-of-order network responses? Why copy a payload before waiting? What is the difference between an editable draft and last acknowledged state? Why must failed saves leave a visible retry path?

## Review fix: PATCH omission is different from null

A nullable Go pointer alone treated a missing salary field and explicit JSON null identically, so clearing salary never persisted. Salary patch fields now retain raw JSON presence: omitted preserves the old value; null clears it; a number sets it. Dedicated SQL updates only preference columns instead of copying unrelated profile fields. Locking the user row before reading/merging serializes simultaneous partial updates. Validation failures roll back the entire transaction, including any list replacement.

Code: `cmd/jobradar/handler_preferences.go`, `sql/queries/users.sql`, `cmd/jobradar/preferences_test.go`.
Questions: Why is PATCH not the same as replacing a resource? Why does SQL COALESCE prevent clearing a nullable field? Why lock before reading rather than after? What would happen if list deletion committed before input validation?

## Application-board drag and drop and delayed edits

The board now moves cards optimistically and PATCHes their status. A failed request restores the previous column and displays an error. Only the pending card is disabled; other cards remain usable. A native status select supplies the same action for keyboard and touch users. External drag IDs are ignored. Other is a display fallback, not a valid persisted database state. Ordering within a column is not implemented.

The earlier notes debounce read the currently selected application when its timer fired, so text could be saved to a different job after switching selection. It now captures application ID and text before waiting, flushes on selection change/close, and serializes writes. PATCH responses merge only changed fields so an older notes response cannot overwrite a newer status.

Code: `frontend/src/views/ApplicationsView.vue`, `frontend/src/components/ApplicationDetail.vue`. Eight regression tests cover moving, rollback, overlapping actions, load failure, note identity and timing in `ApplicationsView.spec.ts` and `ApplicationDetail.spec.ts`.
Questions: Why doesn't debounce prevent concurrent requests? Why capture identity before awaiting? Why roll back optimistic state? Why must drag actions have a non-drag alternative? Why can merging an entire PATCH response lose unrelated edits?

## Profile download: an explicit export contract

The previous Download My Data button had no action. A dedicated component now fetches the authenticated profile/resume/preferences endpoints and downloads JSON. Recursive allowlists exclude credentials and internal account IDs even if future API responses grow. One failed endpoint prevents a misleading partial download. This is a profile copy, not an account backup: saved jobs and applications are excluded and described in the UI.

Code: `frontend/src/components/DownloadProfile.vue`, `frontend/src/utils/profileDownload.ts`; three regression tests in `frontend/src/__tests__/DownloadProfile.spec.ts` cover field selection, authentication, failure and object-URL cleanup.
Questions: Why is an allowlist safer than deleting known secret fields? Why is a browser download not a database backup? Can eight separate GETs guarantee a transactionally consistent snapshot? (No.) Why revoke a generated object URL?

## Settings layout, explicit saves and design tokens

The original design reference is Linear light mode (`JOBRADAR_PLAN.md`); account-preferences and notification screens separate related concerns. The settings grid now has compact cards for match criteria, sources, delivery channels and email schedule, followed by account/integration controls. Shared `settings-card`, `form-input` and `button` classes use the existing accent and square borders instead of one-off rounded purple styles. The grid shrinks to the viewport and aligns cards at the top. Matching settings still use one explicit save because they are one API resource; Preferences retains autosave plus its visible Save button. Master recommendations remains separately autosaved.

A save captures the submitted draft, then only reconciles a response if the user has not edited again. Persistent dirty/saving/error state tells the user whether another save is needed. Duplicate hours, more than two slots and hours outside 0–23 are rejected; the server remains authoritative. The formerly inert download control now renders the scoped profile exporter.

Code: `frontend/src/components/MatchingSettings.vue`, `SettingsTab.vue`, `PasswordSettings.vue`, `AccessTokens.vue`, `frontend/src/styles/globals.css`. Regression tests in `MatchingSettings.spec.ts` and `AccountSettings.spec.ts` cover delayed saves, retries, hours, password confirmation/secret clearing and token create/revoke. The preferences test uses indexed array access to match the configured TypeScript library target.
Questions: Why share design tokens instead of hardcoded colors? Why separate delivery channels from matching eligibility? What is the difference between a dirty draft and a request in flight? Why does client validation not replace server validation? Why does a successful build not prove visual layout or touch interaction?
References: https://linear.app/docs/account-preferences and https://linear.app/docs/notifications (research September 22, 2026).

## Source filters must run before pagination

RemoteOK and Adzuna were decorative sidebar labels. They now set a source query argument, which is normalized and bound as a SQL parameter before LIMIT/OFFSET. Search, source and saved/matched filters compose; changing one retains the others in the URL. Filtering only the current page in the browser would hide matching records on later pages. Best Matches has active navigation styling. The empty Console page is no longer advertised; its old URL redirects to the real Health page.

Code: `cmd/jobradar/handler_jobs.go`, `sql/queries/jobs.sql`, `frontend/src/views/JobsView.vue`, `AppSidebar.vue`, `router/index.ts`. Database tests in `cmd/jobradar/job_source_filter_test.go` verify combined filters/pagination; `frontend/src/__tests__/JobSourceNavigation.spec.ts` tests links and route preservation.
Questions: Why is filtering after LIMIT incorrect? Why parameterize source strings? What belongs in route query state rather than only component memory? Why must UI links have a functioning data path behind them?

## Health reporting: observations have timestamps

Health now shows database, Redis, pending matches and source fetch history in shared square cards. Refresh and 30-second polling share a request guard; leaving the page cancels the request and timer. Failed refreshes keep previous data with a stale-data warning. PostgreSQL stores source health; Redis handles sessions and queue delivery. A successful recorded fetch is not proof that a third-party API is reachable right now, and refreshing this screen does not trigger scraping.

Code: `frontend/src/views/StatusView.vue`, `cmd/jobradar/handler_system.go`; three tests in `frontend/src/__tests__/StatusView.spec.ts` cover polling, overlap prevention, cancellation and failure/retry.
Questions: Why distinguish liveness from historical observations? Why label stale results? How do intervals leak after component unmount? Why does Redis connectivity not prove source collection is healthy?

## Adding official job feeds without AI

Jobicy and MyJobMag Kenya adapt public JSON/RSS into the existing shared job record; persistence and matching remain in shared ingestion. They are opt-in environment flags and poll no more than every six hours. HTTP timeouts, body limits, context cancellation, malformed-feed rejection and duplicate-ID filtering bound external failures. XML must contain a channel and no malformed trailing content; a syntactically valid but wrong document must not look like a successful empty fetch. RSS declared encodings need decoding rather than assuming UTF-8.

Neither adapter invents required skills or country eligibility. Remote listings keep their geographic restrictions. Jobicy salaries are omitted until intervals can be normalized; an hourly amount cannot safely be compared with an annual preference. MyJobMag skips recognized expired entries, but does not retroactively expire previously saved jobs. Location matching remains literal: add Nairobi as an alternative when a source says Nairobi rather than Kenya. Cross-source duplicates remain possible.

Code: `internal/sources/jobicy.go`, `myjobmag.go`, `public_feeds_test.go`, `cmd/jobradar/config_sources.go`, `config_public_feeds_test.go`. Source tests pass with the race detector; configuration tests verify opt-in and conservative schedules. Research, official links, terms, coverage and next adapters are in `docs/JOB_SOURCES.md`.
Questions: Why use a normalized adapter boundary? Why distinguish unknown from false? Why preserve canonical links/attribution? Why doesn't an absent record in a partial feed prove expiry? Why do rate limits depend on queries × pages × users × polling frequency? Why is city-to-country inference a separate feature?

### Master recommendation switch retry

The Settings master switch now also exposes persistent save/error state and a visible Save preferences action. A failed optimistic toggle leaves the desired selection available for retry rather than silently implying persistence. The same serialized store backs this and the Preferences tab. Regression: `frontend/src/__tests__/SettingsTab.spec.ts` reproduces a failed toggle then successful explicit retry.
Question: Why should optimistic UI communicate failure until it is resolved, even after a toast disappears?

## Personal hosting and real email verification

An always-running scheduler needs an awake process. Static frontend hosting or a free web service that sleeps cannot run this process's collection/email timers reliably. Supabase can replace PostgreSQL but does not host Go or Redis; production Compose currently overrides DB_URL, so changing only .env is insufficient. The home server remains the zero-hosting-bill recommendation, with power/internet, HTTPS/private access and backups to plan. See `docs/HOSTING_OPTIONS.md` for official references and alternatives; no deployment has occurred.

The authorized Resend test returned HTTP 403 because samnodier.com is not verified for the configured key. Cloudflare DNS entries must be confirmed by Resend, and the domain/key must belong to the expected account. No email was accepted. Do not confuse an HTTP provider acceptance with confirmed inbox delivery. Scheduled sending remains disabled until configuration and delivery verification. The test used a stable idempotency key to prevent duplicate retries.

Questions: Why does hosting the database not host the application? What happens to timers when a server sleeps? Why does changing an environment file not necessarily override Compose's environment block? How do sender-domain verification, API authentication and inbox delivery differ? How would you back up and restore this deployment? Why can a private tailnet URL work for your laptop but not an arbitrary hosted workflow?

Manual acceptance steps are in `docs/REVIEW_CHECKLIST.md`. Automated coverage grew around concrete regressions; it is not complete coverage of every legacy function. A clean build and component tests do not replace visual review or a real provider test.

## Real pages and a stale running backend

The RemoteOK shortcut looked empty because the running Go process was still the older version: a read-only request with source=remoteok returned 50 Adzuna rows, while PostgreSQL contained 791 RemoteOK records. The old browser-side source filter hid those incorrectly fetched rows. Rebuilding/editing source does not replace a running `go run` process. Restart it explicitly; the agent leaves Sam's servers and logs under Sam's control.

The jobs API now accepts up to 500 records and queries limit+1 to determine whether another page exists. It returns the requested array plus X-Has-More, avoiding an empty extra page when the final page is exactly full. This is offset pagination, not a frozen snapshot: new jobs can shift boundaries between requests. The existing deterministic SQL ordering remains important.

Code: `cmd/jobradar/handler_jobs.go`, `cmd/jobradar/job_source_filter_test.go`. Integration tests verify 50/100/200/500 limits, final/empty pages, source filtering and invalid limits. Questions: Why does editing code not update a running binary? Why fetch one extra row? Why isn't a loaded-page count a total count? How can concurrent inserts shift offset pagination?

### Pagination controls and one source of truth

Both ends of the job list now show Previous/Next, current page and 50/100/200/500 records per page. The URL owns pagination and filter state; changing search/filter/order/page size resets page one, while browser history can restore a page. The server determines membership/order rather than filtering an already paginated array again in the browser. Imported jobs open in their own detail panel without being inserted into an unrelated source page. An AbortController plus request identity prevents delayed responses from replacing a newer page; a failed page does not display old rows under a new page number.

The browser requires X-Has-More and gives an explicit backend-restart message if the old process lacks that contract. CORS exposes that header for deployments using separate frontend/backend origins. Code: `frontend/src/components/JobPagination.vue`, `frontend/src/views/JobsView.vue`, `cmd/jobradar/main.go`; eight focused tests across `JobPagination.spec.ts` and `JobSourceNavigation.spec.ts`. Questions: Why keep filters in the URL? Why reset page when changing criteria? Why do custom response headers need CORS exposure? Why combine cancellation with response identity checks?

## Durable source history instead of invented uptime

The previous service_health row retained only the latest attempt. A new source_health_daily table stores success/failure counters per source and UTC date. A PostgreSQL trigger records each scraper health upsert in the same transaction; unrelated metadata updates do not add attempts. Every observation also prunes days older than the 30-day window. Existing history is not backfilled from guesses. This measures completed collection attempts (including persistence failures), not continuous availability of a third-party API.

Code: `sql/schema/20260922000100_source_daily_health.sql`, `sql/queries/health.sql`, `cmd/jobradar/handler_system.go`, generated SQLC code. The status endpoint returns daily counts and last attempt, never raw provider errors. `cmd/jobradar/source_history_test.go` verifies counters, non-observation updates, retention and safe API output. Migration up/down/up and full backend race/integration suite plus go vet passed on isolated services.
Questions: Why can a single latest-status row not generate historical bars? Why aggregate in UTC? Why do database triggers make related updates atomic? What is the denominator of fetch success rate, and why is it not uptime? What data is lost by daily aggregation? (Individual attempt timing/details.)

### Status-page presentation

Health now follows the supplied status-page reference: summary banner, compact component rows, daily colored bars and expandable fetch details. Green means successful attempts, amber mixed outcomes, red failures and gray no observations. Each day has a keyboard-focusable descriptive label and tooltip; text and counts supplement color. The percentage is explicitly fetch success, not uptime. Refresh failures retain the old snapshot with a stale warning; an incomplete source state does not claim everything is operational. Four component tests cover history/rates, missing data, failures and polling cancellation.

Code: `frontend/src/views/StatusView.vue`, `frontend/src/__tests__/StatusView.spec.ts`. Question: How can a polished status dashboard mislead people if missing measurements are rendered as green? Why expose sample counts next to a percentage?
