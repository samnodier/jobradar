Review the code Sam just wrote. This applies to both Go backend code and Vue/TypeScript frontend code.

Structure the review in exactly this order:

---

## What's Good

List what is correct, idiomatic, or well-structured about the code. Be specific — name the line or pattern. This is not filler. Sam needs to know what to keep doing.

---

## What Would Fail in Production

Go through this checklist and call out every issue found:

**Go Correctness**
- [ ] Are all errors handled? (`err != nil` checked after every fallible call)
- [ ] Are database transactions rolled back on error? (`defer tx.Rollback()` before any early return)
- [ ] Are contexts passed all the way through? (not `context.Background()` where `r.Context()` should be used)
- [ ] Are goroutines supervised? (nothing launched with `go func()` without a WaitGroup or error channel)

**Go Production Readiness**
- [ ] Are there nil pointer dereferences waiting to happen? (pointer fields accessed without nil check)
- [ ] Are there missing DB indexes for the query patterns used?
- [ ] Are there potential data races? (shared state accessed from multiple goroutines without a mutex)
- [ ] Is anything sensitive being logged? (tokens, passwords, emails in log lines)

**Go Idioms**
- [ ] Are custom error types used where the caller needs to distinguish error cases?
- [ ] Are interfaces the right size? (accept interfaces, return concrete types)
- [ ] Does each function/struct have a single clear responsibility?

**JobRadar-Specific Patterns**
- [ ] Does the handler use the `apiConfig` pattern correctly? (method on `*apiConfig`, not a standalone function)
- [ ] Are SQLC nullable fields handled with pointer checks? (`pgtype.Text`, `*string`, etc.)
- [ ] Are Redis keys namespaced by feature? (`auth:pending_signup:`, `queue:jobs`, not bare keys)
- [ ] Do Chi routes follow the project convention? (`{jobID}` not `:jobID`, `chi.URLParam` not `r.PathValue`)
- [ ] Are destructive operations (DELETE, UPDATE) scoped to BOTH resource ID AND authenticated user ID?

**Vue/TypeScript (if applicable)**
- [ ] Are TypeScript interfaces defined for all API response shapes?
- [ ] Does the component fetch data in `onMounted`, not in the template?
- [ ] Is state mutation happening through the Pinia store, not directly on props?
- [ ] Are all `fetch()` calls to the backend using `credentials: 'include'`?
- [ ] Are external URLs normalized to include `https://` before rendering as `<a href>`?

---

## What to Change

For each issue found above: describe what needs to change in plain English. **Do not write the fixed code.** Sam writes the fix.

Format each item as:
> **[Issue]** — [What is wrong and why it matters]. What to do: [direction without the solution].

---

## What Belongs in LEARNINGS.md

After reviewing, identify 1–3 concepts that came up in this review that Sam should document. Say:

> "Write this in LEARNINGS.md: [concept and one-sentence explanation of why it matters]."
