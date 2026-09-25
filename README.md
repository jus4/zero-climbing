# Zero Climbing

A map of places where there's no climbing to be found. Mark and share locations that lack any climbing routes, so everyone knows where to avoid (or be disappointed).

## Features

- Interactive map (OpenLayers) with Finnish National Land Survey (Maanmittauslaitos) base tiles, proxied through the app
- Add, view, and describe climbing-less locations as GeoJSON
- User accounts with session-based auth
- Server-side rendering with [templ](https://templ.guide/) + Tailwind CSS / DaisyUI

## Tech stack

- **Backend:** Go, Echo v4, pgx
- **Database:** PostgreSQL (migrations via goose, queries via sqlc)
- **Frontend:** templ, Tailwind CSS, DaisyUI, OpenLayers (TypeScript, bundled with esbuild), htmx
- **Map tiles:** Maanmittauslaitos WMTS proxy (requires an API key)

## Requirements

- Go 1.25+
- PostgreSQL 17
- [goose](https://github.com/pressly/goose), [templ](https://templ.guide/), [sqlc](https://sqlc.dev/), [air](https://github.com/air-verse/air)
- Node.js + npm (for CSS/JS builds)

If you use Nix, `shell.nix` provides all of the above.

## Setup

```sh
# 1. Clone and install Go dependencies
git clone git@github.com:jus4/zero-climbing.git
cd zero-climbing
go mod download

# 2. Install frontend dependencies
npm install --prefix js
npm install --prefix tailwind

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
# 7. Build the map JS and CSS once (or use the npm watch scripts during dev)
npm run build --prefix js
npm run build-css-prod --prefix tailwind

# 8. Start the server
go run ./cmd/main.go
```

The app runs at <http://localhost:8082>.

### Development

- **Hot reload for Go:** `air` (uses `.air.toml`, rebuilds `./cmd/main.go`)
- **Templates:** `templ generate --watch`
- **JS:** `npm run watch --prefix js`
- **CSS:** `npm run watch-css --prefix tailwind`

## Project structure

```
cmd/            entrypoint (main.go)
db/
  migrations/   goose SQL migrations
  queries/      sqlc queries
  db.go         pgx pool setup
handlers/       HTTP handlers and routes
services/       business logic
views/          templ templates
assets/         static files (css, js, images)
js/             OpenLayers map (TypeScript, esbuild)
tailwind/       Tailwind CSS build
```

## License

MIT