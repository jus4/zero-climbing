# Zero Climbing

A map of places where there's no climbing to be found. Mark and share locations that lack any climbing routes, so everyone knows where to avoid (or be disappointed).

## Features

- Interactive map (OpenLayers) with Finnish National Land Survey (Maanmittauslaitos) base tiles, proxied and cached server-side
- Import location data from a **GPX file** (tracks → lines) or a **SQLite database** (`place` table → points), with a review step to pick which items to import
- All point locations are shown as dots on the map; click a dot to see its title, comments, and uploader
- Owners can **edit** the title/comments and **delete** their own locations right from the popup
- User accounts with session-based auth; map data endpoint protected by a same-site session cookie, and write endpoints additionally protected with CSRF tokens
- Server-side rendering with [templ](https://templ.guide/) + Tailwind CSS / DaisyUI

## Tech stack

- **Backend:** Go, Echo v4, pgx
- **Database:** PostgreSQL (migrations via goose, queries via sqlc); SQLite files are read during import via `modernc.org/sqlite`
- **Frontend:** templ, Tailwind CSS, DaisyUI, OpenLayers (TypeScript, bundled with esbuild), htmx, SweetAlert2
- **Map tiles:** Maanmittauslaitos WMTS proxy (requires an API key), cached in memory with `go-cache`

## Requirements

- Go 1.25+
- PostgreSQL 17
- [goose](https://github.com/pressly/goose), [templ](https://templ.guide/), [sqlc](https://sqlc.dev/), [air](https://github.com/air-verse/air)
- Node.js + npm (for the map JS and CSS builds under `frontend/`)

If you use Nix, `shell.nix` provides all of the above.

## Setup

```sh
# 1. Clone and install Go dependencies
git clone git@github.com:jus4/zero-climbing.git
cd zero-climbing
go mod download

# 2. Install frontend dependencies
npm install --prefix frontend

# 3. Configure environment
cp .env.example .env   # or create by hand, see below
```

`.env` needs:

```
DATABASE_URL="postgres://USER:PASSWORD@localhost:5432/zero_climbing?sslmode=disable"
MAP_API_KEY="your-maanmittauslaitos-api-key"
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
# 6. Generate templ and sqlc code
templ generate
sqlc generate
```

## Running

```sh
# 7. Build the map JS and CSS once (or use the npm watch script during dev)
npm run build --prefix frontend

# 8. Start the server
go run ./cmd/main.go
```

The app runs at <http://localhost:8082>.

### Development

- **Hot reload for Go:** `air` (uses `.air.toml`, rebuilds `./cmd/main.go`)
- **Templates:** `templ generate --watch`
- **JS + CSS:** `npm run watch --prefix frontend` (esbuild and tailwind together)

## API

All endpoints expect a session cookie from a successful login (or, for `GET /api/locations`, any session, so everyone sees the map). Write endpoints also require the `X-CSRF-Token` header (the `csrf_token` cookie is issued by `GET /api/locations`).

| Method   | Path                     | Description                                            | Auth         |
|----------|--------------------------|--------------------------------------------------------|--------------|
| `GET`    | `/api/locations`         | GeoJSON FeatureCollection of all point locations       | session      |
| `GET`    | `/api/me`                | Current user `{id, alias}`                             | login        |
| `POST`   | `/api/locations/:id/edit`| Update a location's title and description              | login + CSRF |
| `DELETE` | `/api/locations/:id`     | Delete a location                                      | login + CSRF |

Edit and delete are restricted to the location's owner (enforced in the SQL query).

## Project structure

```
cmd/            entrypoint (main.go)
db/
  migrations/   goose SQL migrations
  queries/      sqlc queries
  db.go         pgx pool setup
handlers/       HTTP handlers and routes (auth, import, locations, errors)
services/       business logic (users, locations, import parsing)
views/          templ templates (layout, nav, auth, import)
assets/         static files (css, js, images)
frontend/       map JS (TypeScript, esbuild) + Tailwind CSS build
  src/          OpenLayers map source (map.ts, modules/)
  base.css      Tailwind entry stylesheet
  tailwind.config.cjs
```

## License

MIT