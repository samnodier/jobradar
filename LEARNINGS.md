<!-- markdownlint-disable MD013 -->

# Learnings

1. Explain the difference between these three branches. Specifically, why would a senior engineer be upset if you committed a broken API fetcher directly to main?
2. Explain why we use feature branches instead of everyone pushing to main. Mention "stability" and "isolation."
3. The difference between git branch -m (rename) and git checkout -b (create and switch).
4. Why databases need Volumes in Docker.
5. What a Go module path is and why it should match your repository URL.
6. The difference between Host Ports and Container Ports in Docker.
7. Never commit sensitive credentials to version control. Use .env files and add them to .gitignore.
8. What a "service" represents in Docker Compose.
9. The difference between a Docker Image (the blueprint) and a Container (the running instance).
10. What happens to Docker data if you don't use a volume.
11. The difference between a "Host" port and a "Container" port.

12. Explain the difference between an Image and a Container in your own words (the Class vs. Instance analogy is great).
13. Explain why your Go code uses localhost as the host even though the DB is in Docker (Port Mapping).
14. What the -d flag does in docker compose up.
15. Explain the purpose of the internal/ directory in Go and how it helps with encapsulation.
16. Explain why sql.Open doesn't mean you are actually connected yet.
17. What is a "Driver" in the context of Go's database/sql?
18. Explain what go mod tidy does and why it's good practice.
19. Why we use a capital letter for Client and NewClient (Exporting).
20. Explain why sql.Open needs a driver to be imported with a \_.Why we used sslmode=disable in our connection string (Hint: it's for local development).
21. The difference between %v and %w.
22. Why we return nil instead of an empty struct when an error occurs.
23. The command to run a Go program located in a sub-folder.

24. What is a UNIQUE constraint and how does it help with Idempotency?
25. Why we use TIMESTAMP WITH TIME ZONE instead of just a local timestamp.
26. The difference between a Primary Key (for the DB) and an External ID (from the API).
27. What is a Makefile and why is it used in Go projects to manage tasks like migrations and builds?

28. Explain the difference between json.Unmarshal (using a byte slice) and json.NewDecoder (using a stream).
29. Why do we need to check res.StatusCode before trying to decode the body? (What happens if the API sends a 404?)
30. What is an HTML Entity (like &amp;) and why do we need to decode it?
31. What does ON CONFLICT DO NOTHING do in a Postgres INSERT statement?
32. Why is it a bad idea to let your fetcher package talk directly to your database package? (Hint: Think about "Dependency Cycles").
33. What is Type Safety in the context of a database?
34. Why does SQLC need access to your Goose migration files to work?

35. What is a Router and why do we use one instead of just if/else statements?
36. What is Middleware in the context of a web server?
37. Why do we use json.NewEncoder(w).Encode() instead of fmt.Fprintf(w, ...)?
38. Why do we use a respondWithJSON helper instead of encoding JSON inside every handler?
39. Why is r.Context() passed to the database query instead of context.Background()? (Hint: Think about what happens if the user closes their browser tab while the query is running).
40. Goroutine: A way to run a function in the background without blocking the rest of the program.
41. Ticker: A tool to trigger events at regular time intervals.
42. Blocking: When a line of code (like <-ticker.C) makes the program wait before moving to the next line.
43. What is CORS and why does it exist? (Security against malicious websites).
44. Why do we use chi.URLParam instead of reading from the query string for IDs?
45. What is the difference between LIKE and ILIKE in Postgres?

46. The difference between Authentication (Who are you?) and Authorization (What are you allowed to do?).
47. The concept of Delegated Identity (using Google/GitHub to handle the "Who are you?" part).
48. JWT Rule: Never put sensitive data (like passwords or secrets) inside a JWT payload because it is readable by anyone who has the token. It is only used to verify integrity, not to hide data.
49. `access_token` and `refresh_token` should be encrypted at rest in production. That's the right call to make when you have the core system working.
50. Git Stash: How to move uncommitted work between branches.
51. Orphaned Rows: Why Foreign Keys with ON DELETE CASCADE are necessary in normalized schemas.
52. The Blacklist Tradeoff: Balancing the speed of JWTs with the security of instant revocation.
53. `CITEXT` is a Postgres extension that makes text comparisons case-insensitive. Important for emails because `Sam@gmail.com` and `sam@gmail.com` should be treated as the same user.
54. the `state` parameter prevents CSRF attacks (Cross Site Request Forgery), not XSS. Here's the idea: your app generates a random string, sends it to GitHub, and when GitHub redirects back you verify the same string came back. This prevents a malicious site from tricking your server into thinking it initiated the login.
55. Request the minimum OAuth scopes needed; for GitHub profile plus email access, read:user and user:email are usually more appropriate than the broader user scope.”
56. Redis is good for temporary onboarding/session state because it supports TTL and shared access across instances, but permanent account truth belongs in Postgres.
57. Authentication systems often contain multiple independent state artifacts: OAuth state, onboarding session, email verification token, and application session; confusing them leads to bad design.
58. For sensitive changes like updating an email address, the new email should be verified before it replaces the currently trusted account email.
59. The OAuth callback URL is just the route in my app that receives the provider redirect after authorization; it is not special by itself, but it must exactly match what is registered with the provider.
60. golang.org/x/oauth2 handles OAuth client flow in Go: AuthCodeURL builds the provider redirect URL and Exchange swaps the callback code for a token.
61. OAuth state can be stored in a cookie for simple implementations, but production systems typically combine client-side storage (cookie) with server-side storage (e.g., Redis) to ensure one-time use, prevent replay attacks, and support distributed systems.
62. `Secure: true` on cookies means they only travel over HTTPS; always make this configurable or environment-aware so local development still works."
63. Namespace Redis keys consistently by feature prefix so related keys are easy to identify, scan, and manage together — e.g. `auth:pending_signup:` not just `pending:`
64. Use strconv.FormatInt(id, 10) to convert int64 to string when passing numeric IDs as text to a database query.
65. The difference between UNIQUE constraint and a UNIQUE index, and when partial unique indexes are useful. (Hint: unique only when a condition is true)
66. What .PHONY does in a makefile and why you need it for commands that don't produce files. tells Make: these targets arenot files, they are just commands. Always run them.
67. Vue component lifecycle. onMounted runs after the component is inserted into the DOM. It's where you put data fetching that needs to happen on page load.
68. what the SPA fallback problem is, why it happens, and how serving index.html as a fallback fixes it.
69. Never edit a migration that has already been run. Goose tracks which migrations have executed. If you change a file that already ran, your schema and your migration files are out of sync. Other developers (or your future self on a new machine) will have a broken database. Always move forward with a new migration file.
70. what .PHONY does in a Makefile and why you need it for commands that don't produce files.
71. the difference between INNER JOIN and LEFT JOIN, and when you use LEFT JOIN for optional relationships.
72. always scope destructive operations (DELETE, UPDATE) to both the resource ID and the authenticated user's ID. Never trust that the client owns a resource just because they know its ID.
73. REST conventions. URLs describe resources, HTTP methods describe actions. Never put actions in URLs.
74. sharing connections vs duplicating them. In Go you pass pointers to connection clients, so multiple structs can use the same underlying connection safely.
75. Moving shared types to internal/models is a good refactor when you notice multiple packages need the same type and neither should own it. It prevents import cycles and keeps types neutral. Do it when the pain is real, not preemptively.
76. Always use an unexported custom type for context keys. type contextKey string is the idiomatic pattern. It prevents key collisions across packages without any runtime cost
77. flag: credentials: 'include' is required for any fetch() call that needs to send cookies cross-origin or in non-trivial browser contexts. Without it, cookie-based auth silently fails with no error — just a 401.
78. In Go, the zero value for a pointer is nil. This is helpful with SQLC because we can omit nullable fields in a struct literal, and Go will default them to nil, which SQLC then inserts as a NULL in the database.
79. When do we use URL params vs. Request Bodies? URL params (/api/jobs/{id}) are great for identifying a resource you want to GET or DELETE. Request Bodies are better for POST requests when you are sending data to create a new record.
80. Why do we JOIN the jobs table when fetching saved_jobs? (Because the join table only has IDs, and we need the human-readable data like Title and Company).
81. In Vue SFCs, PascalCase is preferred for components to distinguish them from native HTML elements. `<RouterLink>` and `<router-link>` are functionally identical.
82. API Design: When using POST, the backend expects a JSON body. We define a Go struct with json tags to tell json.NewDecoder how to map the keys.
83. State Management: After an API call modifies data on the server, we must also update the local frontend state (e.g., the Pinia store or a ref array) to keep the UI in sync without a page refresh.
84. SQL Joins for User State: Using LEFT JOIN allows us to combine global data (Jobs) with user-specific data (Saved/Applied) in a single query. IS NOT NULL is a handy way to turn a join result into a boolean is_saved flag.
85. The "Immediate" Watcher: Why immediate: true is necessary when a component needs to react to props as soon as it's born.
86. Hoisting in `<script setup>`: Why constants like formatters must be defined at the top before they are used in watchers or lifecycle hooks.
87. The "Stale State" Problem: Why the parent needs to update its list when a child updates a single item (and how the Spread Operator helps keep joined data intact).
88. Debouncing: Why we use clearTimeout and setTimeout to protect the server from excessive API calls.
89. SQL Data Loss: Why RETURNING \* in an update doesn't return joined columns.
