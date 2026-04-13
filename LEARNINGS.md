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
18. Why we used sslmode=disable in our connection string (Hint: it's for local development).
