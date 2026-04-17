# Learnings

## Day 1
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

## Day 2
12. Explain the difference between an Image and a Container in your own words (the Class vs. Instance analogy is great).
13. Explain why your Go code uses localhost as the host even though the DB is in Docker (Port Mapping).
14. What the -d flag does in docker compose up.
15. Explain the purpose of the internal/ directory in Go and how it helps with encapsulation.
16. Explain why sql.Open doesn't mean you are actually connected yet.
17. What is a "Driver" in the context of Go's database/sql?
18. Explain what go mod tidy does and why it's good practice.
19. Why we use a capital letter for Client and NewClient (Exporting).
20. Explain why sql.Open needs a driver to be imported with a _.Why we used sslmode=disable in our connection string (Hint: it's for local development).
21. The difference between %v and %w.
22. Why we return nil instead of an empty struct when an error occurs.
23. The command to run a Go program located in a sub-folder.

## Day 3
24. What is a UNIQUE constraint and how does it help with Idempotency?
25. Why we use TIMESTAMP WITH TIME ZONE instead of just a local timestamp.
26. The difference between a Primary Key (for the DB) and an External ID (from the API).
26. What is a Makefile and why is it used in Go projects to manage tasks like migrations and builds?

## Day 4
27. Explain the difference between json.Unmarshal (using a byte slice) and json.NewDecoder (using a stream).
28. Why do we need to check res.StatusCode before trying to decode the body? (What happens if the API sends a 404?)
29. What is an HTML Entity (like &amp;) and why do we need to decode it?
30. What does ON CONFLICT DO NOTHING do in a Postgres INSERT statement?
31. Why is it a bad idea to let your fetcher package talk directly to your database package? (Hint: Think about "Dependency Cycles").
32. What is Type Safety in the context of a database?
33. Why does SQLC need access to your Goose migration files to work?

## Day 5
34. What is a Router and why do we use one instead of just if/else statements?
35. What is Middleware in the context of a web server?
36. Why do we use json.NewEncoder(w).Encode() instead of fmt.Fprintf(w, ...)?
37. Why do we use a respondWithJSON helper instead of encoding JSON inside every handler?
38. Why is r.Context() passed to the database query instead of context.Background()? (Hint: Think about what happens if the user closes their browser tab while the query is running).
39. Goroutine: A way to run a function in the background without blocking the rest of the program.
40. Ticker: A tool to trigger events at regular time intervals.
41. Blocking: When a line of code (like <-ticker.C) makes the program wait before moving to the next line.
42. What is CORS and why does it exist? (Security against malicious websites).
43. Why do we use chi.URLParam instead of reading from the query string for IDs?
44. What is the difference between LIKE and ILIKE in Postgres?

## Day 6
45. The difference between Authentication (Who are you?) and Authorization (What are you allowed to do?).
46. The concept of Delegated Identity (using Google/GitHub to handle the "Who are you?" part).
47. JWT Rule: Never put sensitive data (like passwords or secrets) inside a JWT payload because it is readable by anyone who has the token. It is only used to verify integrity, not to hide data.
48. `access_token` and `refresh_token` should be encrypted at rest in production. That's the right call to make when you have the core system working.
49. Git Stash: How to move uncommitted work between branches.
50. Orphaned Rows: Why Foreign Keys with ON DELETE CASCADE are necessary in normalized schemas.
51. The Blacklist Tradeoff: Balancing the speed of JWTs with the security of instant revocation.
52. `CITEXT` is a Postgres extension that makes text comparisons case-insensitive. Important for emails because `Sam@gmail.com` and `sam@gmail.com` should be treated as the same user. 
53. the `state` parameter prevents CSRF attacks (Cross Site Request Forgery), not XSS. Here's the idea: your app generates a random string, sends it to GitHub, and when GitHub redirects back you verify the same string came back. This prevents a malicious site from tricking your server into thinking it initiated the login.
54. Request the minimum OAuth scopes needed; for GitHub profile plus email access, read:user and user:email are usually more appropriate than the broader user scope.”
55. Redis is good for temporary onboarding/session state because it supports TTL and shared access across instances, but permanent account truth belongs in Postgres.
56. Authentication systems often contain multiple independent state artifacts: OAuth state, onboarding session, email verification token, and application session; confusing them leads to bad design.
57. For sensitive changes like updating an email address, the new email should be verified before it replaces the currently trusted account email.
58. The OAuth callback URL is just the route in my app that receives the provider redirect after authorization; it is not special by itself, but it must exactly match what is registered with the provider.
59. golang.org/x/oauth2 handles OAuth client flow in Go: AuthCodeURL builds the provider redirect URL and Exchange swaps the callback code for a token.
60. OAuth state can be stored in a cookie for simple implementations, but production systems typically combine client-side storage (cookie) with server-side storage (e.g., Redis) to ensure one-time use, prevent replay attacks, and support distributed systems.
61. `Secure: true` on cookies means they only travel over HTTPS; always make this configurable or environment-aware so local development still works."
62. Namespace Redis keys consistently by feature prefix so related keys are easy to identify, scan, and manage together — e.g. `auth:pending_signup:` not just `pending:`
63. Use strconv.FormatInt(id, 10) to convert int64 to string when passing numeric IDs as text to a database query.
