# Repository Guidelines

## Project Structure & Module Organization

Runtime entry points live in `apps/`: `api`, `worker`, `mock-payment`, and database utilities are Go commands; `apps/web` is the Next.js application. Shared backend packages belong in `internal/`, with HTTP transport, configuration, database, platform wiring, and generated contracts kept separate. Define the REST contract in `api/openapi/openapi.yaml`. Store ordered schema changes in `migrations/` and deterministic fixtures in `seeds/`. Deployment assets live under `deploy/` and `infra/`; architectural and product decisions live in `docs/`.

## Build, Test, and Development Commands

- `make setup`: install pnpm and Go dependencies, then generate REST and gRPC types.
- `make compose-up`: start PostgreSQL, Redis, RabbitMQ, and MinIO.
- `make migrate && make seed`: prepare local data after copying `.env.example` to `.env`.
- `make dev`: run web, API, worker, search, and mock-payment processes concurrently.
- `make test`: run Go tests, frontend TypeScript checking, and Vitest.
- `make web-e2e`: run Playwright browser tests against mocked API boundaries.
- `make lint`: verify Go formatting/vetting, ESLint, and OpenAPI validity.
- `make build`: compile all Go commands and create a production Next.js build.
- `make check`: regenerate contracts, lint, run unit/type tests, and build.

## Coding Style & Naming Conventions

Format Go with `gofmt`; use standard Go package names, exported `PascalCase` identifiers, and local `camelCase` identifiers. Name Go tests `*_test.go` and test functions `TestXxx`. Frontend code uses strict TypeScript, Next.js App Router conventions, two-space indentation, and ESLint configuration from `apps/web/eslint.config.mjs`. Keep route folders lowercase and React components `PascalCase`. Do not edit `internal/contract/openapi.gen.go` or `apps/web/src/lib/api/schema.d.ts` manually; update OpenAPI and run `make generate`. Follow additional rules in `apps/web/AGENTS.md` for web changes.

For visual frontend changes, audit existing styles first and follow the [design system](docs/ui-overhaul/design-system.md).

## Testing Guidelines

Place Go tests beside implementation and prefer table-driven cases, `httptest`, and small fakes at package boundaries. Add regression coverage for changed behavior. Frontend unit/component tests use Vitest and Testing Library; browser tests use Playwright. No numeric coverage threshold exists. `make test`, `make web-e2e`, and `make lint` are mandatory for relevant changes. Run `make check` before opening a pull request.

## Commit & Pull Request Guidelines

History follows Conventional Commit style: `feat(api): ...`, `chore(web): ...`, `docs: ...`, `ci: ...`. Use imperative, focused subjects and a scope when useful. Pull requests should explain intent, summarize implementation, link relevant issues, and list verification commands. Include screenshots for UI changes and call out migrations, environment changes, or OpenAPI contract changes. Keep generated artifacts synchronized; CI rejects drift.

## Security & Configuration

Never commit `.env` or credentials. Add safe placeholders to `.env.example`. Treat local secrets as development-only, validate untrusted input at boundaries, and preserve authorization checks in both route and domain layers.
