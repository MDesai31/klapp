---
name: nextjs-era-history
description: stale-prior guard — klapp was a Next.js+Prisma app until 2026-08-09; JobLog/no-billing/Postgres facts are dead
type: domain
---

Until 2026-08-09, `main` held a completely different app: Next.js 16 (App Router) + Prisma 6 +
Auth.js v5 beta + Postgres, built around a `JobLog` work record with an explicit "no money or
billing in the MVP" rule and hours derived in `src/lib/hours.ts`. That code is preserved on the
`legacy/nextjs` branch; main was fast-forwarded to Thomas's Go + SQLite rewrite
([[D-20260809-promote-go-rewrite-to-main]], rationale in [[D-20260809-go-rewrite]] /
`docs/design/refactor.md`).

Treat any prior knowledge from that era as stale: the Go app is punch/timesheet-centric, **has**
invoices with billing intent (QBO planned — [[quickbooks-plan]]), uses SQLite not Postgres, and
session auth exists only on the admin and invoice sites. The era's brainstorm doc
(`docs/ideas.md`) was removed at adoption — its still-relevant entries live in
`docs/project-tracking/ideas.md`; the file itself survives on `legacy/nextjs`.
