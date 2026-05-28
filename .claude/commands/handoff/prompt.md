Generate an end-of-session handoff document and write it to `.context/handoff.md`.

Overwrite any existing `.context/handoff.md` — this is always the most recent session's handoff.

## What to Include

Populate each section below with accurate, specific information from this session. Do not use placeholders — if something is unknown, say so explicitly.

**Date** — today's date in YYYY-MM-DD format

**Branch** — current git branch (run `git branch --show-current` to confirm)

**Accomplished** — bullet list of what was actually completed this session. Be specific: file names, function names, route names. Not "worked on the matcher" — "implemented `MatchJob` in `internal/matcher/matcher.go`, added `UpdateJobMatchingResult` SQLC query."

**Current State** — one paragraph describing where the code stands right now. Is it working? Is there a known bug? What's the last stable state?

**What Didn't Work** — what was tried that failed or was abandoned, and why. This is critical for the next session — it prevents re-exploring dead ends.

**Next Steps** — ordered list, specific enough to pick up immediately without reading through the whole conversation. Each step should be one action: "Add `ai_summary` column to jobs migration", not "continue Phase 5."

**Open Questions** — things that still need a decision, things that are unclear, things to research before the next session.

**Files Changed** — list every file that was modified or created this session. Run `git diff --name-only HEAD` to confirm.

---

Write the output to `.context/handoff.md` using the template structure in that file.

After writing the file, confirm with: "Handoff written to `.context/handoff.md`. Review it before closing the session."
