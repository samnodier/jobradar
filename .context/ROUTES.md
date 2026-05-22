## Current API Routes

| Method | Route                             | Auth  | Status                  |
| ------ | --------------------------------- | ----- | ----------------------- |
| GET    | `/api/health`                     | No    | ✅ Done                 |
| GET    | `/api/jobs`                       | Yes   | 🔧 Enhancing (SQL Join) |
| GET    | `/api/jobs/stats`                 | Yes   | ✅ Done                 |
| GET    | `/api/jobs/{jobID}`               | Yes   | ✅ Done                 |
| GET    | `/auth/users/me`                  | Yes   | ✅ Done                 |
| DELETE | `/api/users/me`                   | Yes   | ✅ Done                 |
| GET    | `/auth/github/login`              | No    | ✅ Done                 |
| GET    | `/auth/github/callback`           | No    | ✅ Done                 |
| GET    | `/auth/onboarding`                | Token | ✅ Done                 |
| POST   | `/auth/onboarding`                | Token | ✅ Done                 |
| POST   | `/auth/logout`                    | Yes   | ✅ Done                 |
| GET    | `/api/saved-jobs`                 | Yes   | ❌ Not built            |
| POST   | `/api/saved_jobs`                 | Yes   | ✅ Done                 |
| DELETE | `/api/saved-jobs/{id}`            | Yes   | ❌ Not built            |
| GET    | `/api/applications`               | Yes   | ✅ Done                 |
| POST   | `/api/applications`               | Yes   | ✅ Done                 |
| GET    | `/api/applications/{id}`          | Yes   | ✅ Done                 |
| PUT    | `/api/applications/{id}/status`   | Yes   | ❌ Not built            |
| PUT    | `/api/applications/{id}/notes`    | Yes   | ❌ Not built            |
| PUT    | `/api/applications/{id}/followup` | Yes   | ❌ Not built            |
| DELETE | `/api/applications/{id}`          | Yes   | ❌ Not built            |
