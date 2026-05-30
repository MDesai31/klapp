# Klaus Landscaping — Field Job Logging MVP — Design Spec

**Date:** 2026-05-30
**Status:** Approved design, pending spec review
**Supersedes:** the existing `klauslandscapping` React (CRA) + Flask + MySQL prototype

---

## 1. Summary

An internal field-operations tool for a landscaping company. Employees log the
jobs they work (customer, time on site, co-workers, job types, materials, notes).
Admins view all job logs and all employee hours across the company, and manage the
reference data.

There is **no money/billing in the MVP**. "Invoice" in the old prototype actually
meant a *job ticket / work order*; this MVP names it **`JobLog`**. The data model
leaves clean room to add pricing and a **QuickBooks** integration in a later phase.

### MVP priorities
1. Employees can reliably log a job from phone or desktop (online).
2. Employees can see their own hours.
3. Admins can see all jobs and all hours, and do full CRUD on lookups.

### Explicitly out of scope for MVP
- Billing / pricing / customer invoices / payroll export
- QuickBooks integration (designed-for, not built)
- Offline logging / sync
- Separate clock-in/out timesheets (hours derive from job logs only)

---

## 2. Architecture & Stack

A single **Next.js (App Router) + TypeScript** application. One codebase, one deploy.

| Concern | Choice |
|---|---|
| Framework | Next.js (App Router), TypeScript |
| UI | React Server Components + Server Actions; responsive (phone + desktop), online only |
| Auth | Auth.js (NextAuth), credentials provider; session carries `role` |
| ORM / DB | Prisma over PostgreSQL |
| Passwords | bcrypt hashes |
| Hosting | Vercel + managed Postgres (Neon or Supabase) |
| Config | All secrets (DB URL, auth secret) in env vars; nothing committed |

**Data flow:** pages are Server Components that read via Prisma. Mutations (log a
job, CRUD) are **Server Actions** that re-check the session role server-side, write
through Prisma, and revalidate the affected route. The client never queries the DB
directly and is never trusted for authorization.

**Relationship to existing code:** the old `klauslandscapping` React/Flask/MySQL
code is **replaced, not extended**. We carry over the *domain model* only. The new
app is scaffolded in a fresh sibling directory, `klaus-fieldlog/` (the old
`klauslandscapping/` project is left untouched).

**Security fixes by construction** (vs. the prototype): no plaintext passwords
(bcrypt), no hardcoded DB credentials (env vars), enforced server-side
authorization on every mutation and admin view.

---

## 3. Data Model (Prisma)

Postgres/Prisma-native version of the existing relational design. The dead
`employee_hours` table is dropped (hours are computed). The old split
`users` + `employee` tables collapse into one `User`.

### Entities

**User** — one row per person (employee or admin).
- `id`, `email` (unique), `name`, `passwordHash`
- `role`: enum `EMPLOYEE | ADMIN`
- `active`: boolean (soft-delete / deactivate)
- `createdAt`, `updatedAt`

**Customer**
- `id`, `name`, `active`, `createdAt`, `updatedAt`
- (room later for address/contact/`quickbooksId`)

**JobType** (was `jobs`)
- `id`, `name`, `active`, timestamps

**Material** (was `materials`)
- `id`, `name`, `active`, timestamps

**JobLog** — the core record (was `invoice`).
- `id`, `customerId` (FK)
- `date` (date), `timeArrived` (timestamp), `timeLeft` (timestamp)
- `notes` (text, optional)
- `createdById` (FK → User; who logged it)
- `createdAt`, `updatedAt`
- *(old `num_workers` dropped — it is just the count of linked workers)*

### Link tables (many-to-many)

**JobLogWorker** — `jobLogId` + `userId`. Which employees worked the job. Drives hours.
**JobLogJobType** — `jobLogId` + `jobTypeId`.
**JobLogMaterial** — `jobLogId` + `materialId`.

### Key decisions

1. **Soft-delete via `active` flags** on User/Customer/JobType/Material. You can't
   hard-delete a record referenced by historical job logs. UI "CRUD" =
   create / edit / activate-deactivate.
2. **"Other job" free text** goes into `JobLog.notes`. The prototype silently
   created JobType rows from free text; the MVP does not. Admins add real JobTypes
   via lookup CRUD to keep the list clean.
3. **Hours are computed, not stored.** Per-JobLog duration = `timeLeft − timeArrived`,
   summed over an employee's `JobLogWorker` links. Times are full timestamps so
   cross-midnight math is correct. Leaves room for a future timesheet/adjustments
   model without migration pain.
4. **Money-ready but money-free.** No rate/price fields yet; Customer and JobLog
   have obvious places to add them and a QuickBooks id later.

---

## 4. Screens & Flows

Responsive; everything behind login; UI adapts to `role`. Role-aware nav.
Unauthorized route access is bounced by middleware.

### Shared
- **Login** (`/login`) — email + password (Auth.js). Redirect by role:
  employee → `/jobs`, admin → `/admin`. Logout in nav.

### Employee
- **My Jobs** (`/jobs`) — job logs this employee worked, newest first
  (customer, date, duration). Empty state prompts logging the first job.
- **Log a Job** (`/jobs/new`) — core form, mobile-first: customer, date,
  time arrived/left, co-workers, one or more job types, materials, notes.
  Submits via server action. Inline validation.
- **My Hours** (`/hours`) — this employee's hours filtered by date / month / year,
  per-entry duration + total. Computed from job-log links.

### Admin (everything employees see, plus)
- **Admin Dashboard** (`/admin`) — all job logs company-wide; filter by employee,
  customer, date range; shows who logged each.
- **All Hours** (`/admin/hours`) — hours per employee for a chosen period
  (payroll-review view).
- **Manage lookups** — full CRUD (create / edit / activate-deactivate):
  - Customers (`/admin/customers`)
  - Employees/Users (`/admin/users`) — add person, set role, deactivate,
    set/reset password
  - Job Types (`/admin/job-types`)
  - Materials (`/admin/materials`)

---

## 5. Error Handling

- **Validation (Zod, shared):** one schema per form validates on client (instant
  feedback) and server (the real gate). Covers required fields,
  `timeLeft > timeArrived`, valid foreign keys, non-empty job-type selection.
- **Auth/role failures:** middleware redirects unauthenticated users to `/login`;
  server actions throw on missing/insufficient role (403, not a silent no-op).
- **Server action results:** typed `{ ok: true } | { ok: false, error }`. UI shows
  inline field errors or a form-level message — never a raw stack trace.
  Unexpected exceptions are caught, logged server-side, surfaced as a generic
  retry message.
- **DB integrity:** the multi-row job-log write (JobLog + worker/jobtype/material
  links) runs in a **single Prisma transaction** — all or nothing. FKs enforce
  referential integrity; soft-delete preserves history.
- **Not-found / empty states:** missing records render a clean 404; empty lists
  render purposeful empty states.

---

## 6. Testing

TDD throughout — tests written before implementation for logic and server actions.
CI-friendly.

- **Unit (Vitest):** Zod schemas; the **hours/duration calculator** (cross-midnight,
  summing across multiple job logs) — highest-value, most bug-prone logic, gets the
  most coverage.
- **Integration (Vitest + test Postgres):** server actions against a real test DB —
  logging a job writes all link rows; CRUD + soft-delete; **authorization tests**
  proving an employee cannot invoke admin actions or view others' hours.
- **E2E (Playwright, smoke only):** employee logs in → logs a job → sees it in My
  Jobs and My Hours; admin logs in → sees all jobs → adds a customer.

---

## 7. Seed Data

Seed script creates: one admin user, a couple of employee users, a few customers,
the existing job types (Lawn Mowing, Weeding, Tree Pruning, Hedge Trimming, Garden
Bed Installation) and materials (Topsoil, Mulch Brown, Mulch Red, Gravel). All
passwords bcrypt-hashed. No real credentials committed.

---

## 8. Future Phases (designed-for, not built)

- Pricing on JobType/Material + per-customer rates → customer invoices
- QuickBooks API integration (push invoices / payroll) via an API route + `quickbooksId` fields
- Offline logging + sync for poor-signal job sites
- Separate timesheet / hours adjustments (travel, shop time)
