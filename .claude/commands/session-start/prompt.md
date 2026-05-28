Start a new Jobradar working session.

## Step 1 — Load Project State

Read `.context/PLAN.md` and identify:
- Current phase and its status
- What was last completed
- What is next

## Step 2 — Ask Sam These Questions (ask all four, wait for answers)

1. **Branch** — What branch are you on right now? (`git branch` if unsure)
2. **Last completed** — What was the last thing you finished? (Even if it was last session)
3. **Goal today** — What are you trying to accomplish in this session? Be specific — one task or one feature, not "work on Phase 5"
4. **Already tried** — If you already started on this: what have you attempted so far? What happened?

## Step 3 — Load Relevant Context

- If the task touches the API or DB routes: read `.context/ROUTES.md` and `.context/SCHEMA.md`
- If the task involves the queue or worker: read `internal/queue/` and `cmd/jobradar/worker_match.go`
- If the task involves the matcher: read `internal/matcher/`

## Step 4 — Confirm Phase Fit

Check whether the session goal fits the current phase in `.context/PLAN.md`. If it doesn't fit (e.g., jumping to Phase 7 while Phase 5 is not started), flag it:

> "This task belongs to Phase N, but the current phase is Phase M. Do you want to continue anyway, or should we focus on Phase M first?"

## Step 5 — Agree on a Session Goal

Restate the session goal in one specific sentence, confirm with Sam, and begin.

---

**Session summary format** (output at the start after Step 5 is complete):

> Session goal: [one sentence]. Branch: [branch name]. Phase: [current phase].
