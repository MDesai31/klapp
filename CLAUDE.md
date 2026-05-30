@AGENTS.md

## Project
Klaus Field Log — Next.js 16.2.6 (App Router) + Prisma 6 + Auth.js v5 beta + PostgreSQL (Docker).

## Database
- Postgres runs in Docker: `docker start klapp-postgres` (or see README to create it fresh)
- Dev DB: `postgresql://klapp:klapp@localhost:5432/klapp`
- Test DB: `postgresql://klapp:klapp@localhost:5432/klapp_test`
- Apply migrations: `npx prisma migrate dev`
- Seed: `npm run db:seed`
- Apply to test DB: `DATABASE_URL="postgresql://klapp:klapp@localhost:5432/klapp_test?schema=public" npx prisma migrate deploy`

## Prisma 6 — breaking changes from training data
- Import client from `@/generated/prisma/client` (NOT `@prisma/client`)
- Generator: `provider = "prisma-client"`, output `"../src/generated/prisma"`
- `PrismaClient` requires a driver adapter — no `datasourceUrl` option:
  `new PrismaClient({ adapter: new PrismaPg({ connectionString: url }) })`
- Install: `@prisma/adapter-pg` + `pg`
- Datasource URL lives in `prisma.config.ts`, not in `schema.prisma`
- Seed scripts: use relative import `../src/generated/prisma/client` + `import "dotenv/config"`

## Next.js 16 — breaking changes from training data
- `searchParams` in page components is `Promise<{...}>` — always `await searchParams`
- Read `node_modules/next/dist/docs/` before writing page/layout code

## Auth.js v5 beta (next-auth@5.0.0-beta.31)
- `next-auth/middleware` is deleted — use `auth` from `@/lib/auth` to wrap middleware
- `npx auth secret` pulls the wrong package — set `AUTH_SECRET` in `.env` manually
- `signIn`/`signOut`/`auth` all import from `@/lib/auth` in server components

## Testing
- Run: `npm test` (unit + integration, needs `.env` with `DATABASE_URL_TEST`)
- Integration tests share test DB — `maxWorkers: 1` in vitest config prevents race conditions
- `tests/setup.ts` loads `dotenv/config` so `.env` values are available in test workers
- E2E (`npm run test:e2e`) requires Playwright — does NOT work on Ubuntu 26.04
