# zero-climbing-web

React + TypeScript SPA for Zero Climbing. Talks to the Go API under `/api/v1`.

## Stack

- Vite + React + TypeScript
- React Router
- TanStack Query
- Tailwind CSS v4 + [HyperUI](https://hyperui.dev) component patterns (copy-paste, no component library)
- ESLint + Prettier

## Scripts

```sh
npm install
npm run dev        # dev server, proxies /api and /proxy to the Go API
npm run build      # type-check + static bundle into dist/
npm run preview    # preview the production build
npm run lint       # eslint
npm run typecheck  # tsc -b
npm run format     # prettier --write
```

## Configuration

Copy `.env.example` to `.env` if you need to override defaults:

- `VITE_API_BASE` — API base path (default `/api/v1`, same-origin).
- `API_PROXY_TARGET` — dev proxy target (default `http://localhost:8082`).
