# Zero Climbing

A map of places where there's no climbing to be found. Mark and share locations that lack any climbing routes, so everyone knows where to avoid (or be disappointed).

## Features

- Interactive map (OpenLayers) with Finnish National Land Survey (Maanmittauslaitos) base tiles, proxied and cached server-side
- Import location data from a **GPX file** (tracks → lines) or a **SQLite database** (`place` table → points), with a review step to pick which items to import
- All point locations are shown as dots on the map; click a dot to see its title, comments, and uploader
- Owners can **edit** the title/comments and **delete** their own locations right from the popup
- User accounts with session-based auth (Google OAuth); the map is public, write endpoints are protected by a same-site session cookie and CSRF tokens

## Tech stack

- **Backend:** Go, Echo v4, pgx
- **Database:** PostgreSQL (migrations via goose, queries via sqlc); SQLite files are read during import via `modernc.org/sqlite`
- **Frontend:** React + TypeScript SPA (`web/`) built with Vite, React Router, TanStack Query, Tailwind CSS, OpenLayers
- **Map tiles:** Maanmittauslaitos WMTS proxy (requires an API key), cached in memory with `go-cache`
- **Deploy:** one Go service serves the API and the built SPA (same origin, no nginx); containerized with a multi-stage `Dockerfile.api`

## Requirements

- Go 1.25+
- PostgreSQL 17
- [goose](https://github.com/pressly/goose), [sqlc](https://sqlc.dev/), [air](https://github.com/air-verse/air)
- Node.js + npm (for the `web/` SPA)

If you use Nix, `shell.nix` provides all of the above.

## Setup

```sh
# 1. Clone and install Go dependencies
git clone git@github.com:jus4/zero-climbing.git
cd zero-climbing
go mod download

# 2. Install SPA dependencies
npm install --prefix web

# 3. Configure environment
cp .env.example .env   # or create by hand, see below
```

`.env` needs:

```
DATABASE_URL="postgres://USER:PASSWORD@localhost:5432/zero_climbing?sslmode=disable"
MAP_API_KEY="your-maanmittauslaitos-api-key"
SESSION_SECRET="generate-with-openssl-rand-hex-32"
GOOGLE_CLIENT_ID="your-google-oauth-client-id"
GOOGLE_CLIENT_SECRET="your-google-oauth-client-secret"
GOOGLE_REDIRECT_URL="http://localhost:8082/auth/google/callback"
GOOGLE_API_REDIRECT_URL="http://localhost:8082/api/v1/auth/google/callback"
FRONTEND_URL="http://localhost:5173"
```

Get a free Maanmittauslaitos API key from <https://www.maanmittauslaitos.fi/en/rajapinnat/avoin-karttakuva-api-user-guide>.

### Database

```sh
# 4. Create the database
createdb zero_climbing

# 5. Run migrations
DATABASE_URL="postgres://USER:PASSWORD@localhost:5432/zero_climbing?sslmode=disable" \
  goose -dir db/migrations postgres "$DATABASE_URL" up
```

### Generate code

```sh
# 6. Generate sqlc code
sqlc generate
```

## Running

Run the API and the SPA dev server in separate terminals:

```sh
# API (http://localhost:8082)
go run ./cmd/main.go

# SPA (http://localhost:5173, proxies /api and /proxy to the API)
npm run dev --prefix web
```

Open <http://localhost:5173>.

Alternatively, run the whole stack with Docker:

```sh
docker compose up --build   # http://localhost:8082
```

### Development

- **Hot reload for Go:** `air` (uses `.air.toml`, rebuilds `./cmd/main.go`)
- **SPA:** `npm run dev --prefix web` (Vite HMR)

## API

All endpoints live under `/api/v1`. Anonymous requests may read `GET /locations`; everything else requires a session cookie. Write endpoints also require the `X-CSRF-Token` header (the `csrf_token` cookie is issued by any request to the API).

| Method   | Path                        | Description                                   | Auth         |
|----------|-----------------------------|-----------------------------------------------|--------------|
| `GET`    | `/healthz`                  | Liveness                                      | none         |
| `GET`    | `/api/v1/auth/me`           | Current user `{id, alias, email, onboardingRequired}` | optional |
| `GET`    | `/api/v1/auth/google`       | Start Google sign-in (302)                    | none         |
| `GET`    | `/api/v1/auth/google/callback` | OAuth callback (302 to `FRONTEND_URL`)     | none         |
| `POST`   | `/api/v1/auth/username`     | Complete onboarding                           | pending      |
| `POST`   | `/api/v1/auth/logout`       | End the session                               | none         |
| `GET`    | `/api/v1/locations`         | GeoJSON FeatureCollection of point locations  | none         |
| `PATCH`  | `/api/v1/locations/:id`     | Update a location's title and description     | owner + CSRF |
| `DELETE` | `/api/v1/locations/:id`     | Delete a location                             | owner + CSRF |
| `POST`   | `/api/v1/import`            | Upload a GPX/SQLite file → `{token, candidates}` | login + CSRF |
| `POST`   | `/api/v1/import/confirm`    | Import selected candidates → `{imported}`     | login + CSRF |
| `GET`    | `/proxy/wmts/*`             | Maanmittauslaitos tile proxy                  | none         |

Edit and delete are restricted to the location's owner (enforced in the SQL query).

## Project structure

```
cmd/            entrypoint (main.go)
db/
  migrations/   goose SQL migrations
  queries/      sqlc queries
  sqlc/         generated sqlc code (committed)
  db.go         pgx pool setup
handlers/       HTTP handlers and routes (/api/v1 auth, locations, import, static, errors)
services/       business logic (users, locations, import parsing)
web/            React + TypeScript SPA (Vite, TanStack Query, Tailwind)
Dockerfile.api            Go API image; also builds and serves the SPA
docker-compose.yml        local dev stack (postgres, api)
docker-compose.koyeb.yml  legacy Koyeb docker-compose (kept for reference)
```

## Deployment

The API and the SPA are served by one Go binary: `Dockerfile.api` builds `web/`
and the Go binary, and the server serves the SPA from `/` with an `index.html`
fallback (`handlers/static.go`). Everything is same-origin, so session and CSRF
cookies work without CORS.

### Local (docker compose)

```sh
cp .env.example .env   # set SESSION_SECRET, MAP_API_KEY, Google OAuth values
docker compose up --build
```

This starts Postgres, then the API (which runs goose migrations on startup) and
serves the SPA at <http://localhost:8082>.

- SPA: <http://localhost:8082>
- Health: <http://localhost:8082/healthz>
- API: <http://localhost:8082/api/v1/...>

### Koyeb (single native service)

1. Create a service from this GitHub repo.
2. Builder: **Dockerfile** → location `Dockerfile.api`. **Privileged: off.**
3. Expose `PORT` (e.g. `8000`) and set the health check to HTTP `/healthz`.
4. Set environment/secrets (below). `DATABASE_URL` comes from Koyeb Managed Postgres.

Koyeb builds the image (Go + Node + goose) and the instance runs the API, applies
migrations, and serves the SPA — no Docker-in-Docker.

#### Environment

| Variable | Required | Notes |
|---|---|---|
| `DATABASE_URL` | yes | Koyeb Managed Postgres connection string (`?sslmode=require`) |
| `SESSION_SECRET` | yes | e.g. `openssl rand -hex 32`; session + CSRF cookies |
| `GOOGLE_CLIENT_ID` / `GOOGLE_CLIENT_SECRET` | yes | Google OAuth client |
| `GOOGLE_API_REDIRECT_URL` | yes | `https://<domain>/api/v1/auth/google/callback` |
| `FRONTEND_URL` | yes | public origin, e.g. `https://<domain>` |
| `MAP_API_KEY` | yes | Maanmittauslaitos API key |
| `PORT` | injected | set by Koyeb |
| `STATIC_DIR` | no | path to the built SPA; baked as `/app/dist` |
| `COOKIE_SECURE` | no | `true` in production |
| `COOKIE_DOMAIN` | no | set for a custom domain |
| `COOKIE_SAMESITE` | no | `lax` (default), `strict`, or `none` |

Add the redirect URI above to the Google OAuth client's authorized redirect URIs.

> `docker-compose.koyeb.yml` is a legacy docker-compose variant kept for
> reference; the supported deployment is the native `Dockerfile.api` service above.

## License

MIT
