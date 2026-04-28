# Jobradar

Jobradar is a full-stack job aggregator with AI matching, notifications, and an integrated job application tracker. It is built to streamline the job search process by centralizing listings and tracking application statuses.

## Tech Stack

- **Backend:** Go, Chi router, SQLC, pgx
- **Database:** PostgreSQL (Core data), Redis (Sessions & Queues)
- **Frontend:** Vue 3, TypeScript, Vite, Pinia, Vue Router
- **Infrastructure:** Docker Compose

## Features (Current & Planned)

- **Authentication:** GitHub OAuth with Redis-backed sessions and HttpOnly cookies.
- **Job Aggregation:** Automated fetching from sources like RemoteOK.
- **Application Tracking:** Built-in tracker for managing job applications.
- **AI Matching:** (Planned) Match jobs against user profiles.

## Local Development Setup

### Prerequisites
- Docker & Docker Compose
- Go 1.22+
- Node.js 20+

### Steps to Run

1. **Clone the repository:**
   ```bash
   git clone https:/`/github.com/yourusername/jobradar.git`
   cd jobradar
2. Environment Variables:
   Copy `.env.example` to `.env` (you will need GitHub OAuth credentials).
   ```bash
   cp .env.example .env
3. Start Infrastructure (Postgres & Redis):

    ```bash
    make up
4. Database Migrations: (Add your migration command here, e.g.,
    ```bash
    make migrate-up
5. **Run the Backend (Go):**
    ```bash
    make run
6. **Run the Frontend (Vue 3):**
    ```bash
    cd frontend
    npm install
    npm run dev

The backend will run on http://localhost:8080 (or your configured port) and the frontend on http://localhost:5173.