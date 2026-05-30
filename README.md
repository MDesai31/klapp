# Klaus Field Log

Internal field-operations tool: employees log jobs and view their hours; admins
view all jobs/hours and manage customers, employees, job types, and materials.

## Stack

Next.js (App Router) + TypeScript, Prisma 6 + PostgreSQL, Auth.js v5, Zod, Vitest, Playwright.

## Setup

1. `npm install`
2. Copy `.env.example` to `.env` and fill `DATABASE_URL`, `DATABASE_URL_TEST`, `AUTH_SECRET`
   - Generate AUTH_SECRET: `npx auth secret`
3. Start PostgreSQL (Docker): `docker run -d --name klapp-postgres -e POSTGRES_USER=klapp -e POSTGRES_PASSWORD=klapp -e POSTGRES_DB=klapp -p 5432:5432 postgres:16-alpine`
4. Create test DB: `docker exec klapp-postgres psql -U klapp -c "CREATE DATABASE klapp_test;"`
5. `npx prisma migrate dev` — apply schema
6. `npm run db:seed` — seed admin/employees/lookups
7. `npm run dev` — http://localhost:3000

## Seed logins

| Role     | Email                  | Password      |
|----------|------------------------|---------------|
| Admin    | admin@klaus.test       | admin12345    |
| Employee | manthan@klaus.test     | employee12345 |
| Employee | thomas@klaus.test      | employee12345 |

## Tests

```bash
npm test            # unit + integration (needs DATABASE_URL_TEST in .env)
npm run test:e2e    # Playwright smoke tests (needs seeded dev DB)
```

## Roadmap (post-MVP)

- Pricing + customer invoices
- QuickBooks integration
- Offline logging/sync
