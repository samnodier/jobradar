---
description: Instructs Claude to prompt Sam to document concepts in LEARNINGS.md whenever something important comes up
---

## LEARNINGS.md

When something important comes up — a concept, a tradeoff, a pattern, a mistake — say: **"Write this in LEARNINGS.md."**

Topics Sam must deeply understand by the end of this project:

- Go context and how middleware passes values to handlers
- Redis list as a queue — LPUSH, BRPOP, blocking vs polling
- Goroutines vs OS threads — why Go's concurrency model is efficient
- Context cancellation for graceful goroutine shutdown
- Exponential backoff — the math and why linear backoff fails under load
- Dead letter queues and when to use them
- Thundering herd problem
- Middleware pattern in Go HTTP handlers
- How to prompt LLMs for reliable structured JSON output
- Token bucket rate limiting algorithm
- Structured logging vs printf debugging
- Observability — knowing what your system is doing without a debugger
- Open/closed principle — extending the system without breaking it
- SPA fallback routing — why the Go server needs to handle Vue Router routes
- Vue 3 Composition API — why it exists and how it differs from Options API
- Pinia — why centralized state beats per-component fetching
- TypeScript interfaces and discriminated unions
- Idempotency in notification systems
- Cache-Control headers and why browsers need them
- Sharing connections vs duplicating them in Go
- INNER JOIN vs LEFT JOIN and when to use each
- Composite primary keys vs UUID primary keys — tradeoffs
- Partial unique indexes vs UNIQUE constraints
- .PHONY in Makefiles and why it matters
