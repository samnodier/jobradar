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
