# Klaus Landscaping — Field Job Logging & Assignment MVP — Design Spec

**Date:** 2026-05-30
**Status:** Approved design (assignment workflow added 2026-05-30), in build
**Supersedes:** the existing `klauslandscapping` React (CRA) + Flask + MySQL prototype

---

## 1. Summary

An internal field-operations tool for a landscaping company with two roles,
**EMPLOYEE** and **ADMIN**.

- **Employees** log the jobs they work (customer, time on site, co-workers, job
  types, materials, notes), view their own hours, and handle jobs **assigned** to
  them (accept, then fill in on completion).
- **Admins** view all jobs and all employee hours, manage reference data, and
  **create & assign jobs to specific employees**.

There is **no money/billing in the MVP**. "Invoice" in the old prototype actually
meant a *job ticket / work order*; this MVP names it **`JobLog`**. The data model
leaves clean room to add pricing and a **QuickBooks** integration later.

### MVP priorities
1. Employees can reliably log a job from phone or desktop (online).
2. Employees can see their own hours.
3. Admins can see all jobs and all hours, and do full CRUD on lookups.
4. Admins can assign jobs to employees; employees accept and complete them.

### Out of scope for MVP
- Billing / pricing / customer invoices / payroll export
- QuickBooks integration (designed-for, not built)
- Offline logging / sync
- Separate clock-in/out timesheets (hours derive from job logs only)
- Per-assignee partial acceptance (acceptance is at the job/crew level — see §3)

---

## 2. Architecture & Stack

A single **Next.js (App Router) + TypeScript** application. One codebase, one deploy.

| Concern | Choice |
|---|---|
| Framework | Next.js (App Router), TypeScript |
| UI | React Server Components + Server Actions; responsive (phone + desktop), online only |
| Auth | Auth.js (NextAuth v5), credentials provider; session carries `role` |
| ORM / DB | Prisma over PostgreSQL (installed locally) |
| Passwords | bcrypt hashes |
| Hosting | Vercel + managed Postgres (later) |
| Config | Secrets in env vars; `.env` git-ignored (repo is PUBLIC) |

**Data flow:** pages are Server Components that read via Prisma. Mutations are
**Server Actions** that re-check session role server-side, write through Prisma, and
revalidate the affected route. The client never queries the DB directly and is never
trusted for authorization.

**Repo:** lives in `klapp/`, GitHub `MDesai31/klapp` (public, branch `main`). The
old `klauslandscapping/` prototype is left untouched; only the domain model carries
over.

**Security fixes vs. prototype:** bcrypt passwords (no plaintext), no hardcoded DB
creds (env vars), enforced server-side authorization on every mutation/admin view.

---

## 3. Data Model (Prisma)

The old split `users` + `employee` tables collapse into one `User`; the dead
`employee_hours` table is dropped (hours are computed).

`JobLog` represents a job in **any lifecycle state** — whether self-logged by an
employee or created and assigned by an admin. A `status` field distinguishes them.

### Enums
- `Role`: `EMPLOYEE | ADMIN`
- `JobStatus`: `ASSIGNED | ACCEPTED | COMPLETED`

### Entities

**User** — one row per person.
- `id`, `email` (unique), `name`, `passwordHash`
- `role` (`Role`, default `EMPLOYEE`), `active` (bool, default true)
- timestamps

**Customer** — `id`, `name`, `active`, timestamps. (room later for address/contact/`quickbooksId`)

**JobType** (was `jobs`) — `id`, `name`, `active`, timestamps.

**Material** (was `materials`) — `id`, `name`, `active`, timestamps.

**JobLog** — the core work record.
- `id`, `customerId` (FK)
- `status` (`JobStatus`, default `COMPLETED`)
- `date` (Date) — work date for a self-log; scheduled date for an assignment
- `timeArrived` (DateTime, **nullable** — filled on completion)
- `timeLeft` (DateTime, **nullable** — filled on completion)
- `instructions` (text, optional) — admin's notes when assigning
- `notes` (text, optional) — employee's notes on completion / self-log
- `createdById` (FK → User) — creator (admin for assignments, employee for self-logs)
- `acceptedById` (FK → User, nullable), `acceptedAt` (DateTime, nullable)
- `completedById` (FK → User, nullable), `completedAt` (DateTime, nullable)
- timestamps

### Link tables (many-to-many)
- **JobLogWorker** — `jobLogId` + `userId`. The employees on the job (assignees for
  an assignment; participants for a self-log). Drives hours.
- **JobLogJobType** — `jobLogId` + `jobTypeId`.
- **JobLogMaterial** — `jobLogId` + `materialId`.

### Lifecycle

```
self-log:    (employee) ──create──▶ COMPLETED        (times + types set at creation)
assignment:  (admin) ──assign──▶ ASSIGNED
                                   │ employee accepts
                                   ▼
                                ACCEPTED
                                   │ employee completes (fills times/types/materials/notes)
                                   ▼
                                COMPLETED
```

### Key decisions

1. **One entity, status lifecycle.** Assigned jobs and self-logged jobs are the same
   `JobLog`; `status` + nullable times distinguish an open assignment from a finished
   job. One model → one hours calculation, one admin dashboard.
2. **Acceptance is job/crew-level, not per-assignee.** A job assigned to a crew is
   accepted once (by any assignee) — `acceptedById`/`acceptedAt` record who/when.
   This matches how a crew lead takes a ticket; per-person acceptance is out of scope.
3. **Hours come only from COMPLETED jobs** with both `timeArrived` and `timeLeft`.
   The calculator ignores entries missing either time, so open assignments never
   inflate hours. Duration = `timeLeft − timeArrived` (full timestamps → correct
   cross-midnight math), summed over an employee's `JobLogWorker` links.
4. **Soft-delete via `active` flags** on User/Customer/JobType/Material. UI "CRUD" =
   create / edit / activate-deactivate (can't hard-delete records referenced by jobs).
5. **"Other job" free text** goes into `JobLog.notes`; admins add real JobTypes via
   lookup CRUD to keep the list clean.
6. **Money-ready but money-free.** No price fields yet; obvious places to add them
   plus a QuickBooks id later.

---

## 4. Screens & Flows

Responsive; everything behind login; UI adapts to `role`. Role-aware nav;
unauthorized route access is bounced by middleware.

### Shared
- **Login** (`/login`) — email + password (Auth.js). Redirect by role:
  employee → `/jobs`, admin → `/admin`. Logout in nav.

### Employee
- **My Jobs** (`/jobs`) — completed jobs this employee worked, newest first
  (customer, date, duration). Empty state prompts logging the first job.
- **Log a Job** (`/jobs/new`) — self-log form: customer, date, time arrived/left,
  co-workers, job types, materials, notes. Creates a `COMPLETED` JobLog. Inline validation.
- **My Assignments** (`/assignments`) — jobs assigned to me that are `ASSIGNED` or
  `ACCEPTED`: customer, scheduled date, instructions, status.
  - **Accept** button on an `ASSIGNED` job → `ACCEPTED`.
  - **Complete** button → completion form.
- **Complete Assignment** (`/assignments/[id]/complete`) — fill time arrived/left,
  job types, materials, notes; moves job to `COMPLETED`. Then it appears in My Jobs/Hours.
- **My Hours** (`/hours`) — this employee's hours filtered by date / month / year,
  per-entry duration + total. Computed from completed job-log links.

### Admin (everything employees see, plus)
- **Admin Dashboard** (`/admin`) — all job logs company-wide with a **status** column;
  filter by employee, customer, date range, and status. Shows who logged/was assigned.
- **Assign a Job** (`/admin/assign`) — create an assignment: customer, scheduled date,
  assignee(s), optional job types, optional instructions. Creates an `ASSIGNED` JobLog.
- **All Hours** (`/admin/hours`) — hours per employee for a chosen period (payroll review).
- **Manage lookups** — full CRUD (create / edit / activate-deactivate):
  - Customers (`/admin/customers`)
  - Employees/Users (`/admin/users`) — add person, set role, deactivate, set/reset password
  - Job Types (`/admin/job-types`)
  - Materials (`/admin/materials`)

### Server actions (authoritative list)
- `createJobLog` (employee) — self-log → `COMPLETED`.
- `assignJob` (admin) — create `ASSIGNED` job with assignees.
- `acceptJob` (assignee) — `ASSIGNED` → `ACCEPTED`.
- `completeJob` (assignee) — `ACCEPTED`/`ASSIGNED` → `COMPLETED`, fills times/types/materials/notes.
- Admin CRUD: `createCustomer`/`updateCustomer`/`setCustomerActive`,
  `createJobType`/`setJobTypeActive`, `createMaterial`/`setMaterialActive`,
  `createUser`/`setUserActive`/`resetUserPassword`.

---

## 5. Error Handling

- **Validation (Zod, shared):** one schema per form validates on client (instant
  feedback) and server (the real gate). Covers required fields, `timeLeft >
  timeArrived` (when both present), valid foreign keys, non-empty job-type selection
  on completion/self-log, and required assignee(s) on assignment.
- **Auth/role failures:** middleware redirects unauthenticated users to `/login`;
  server actions throw on missing/insufficient role (403). Accept/complete actions
  additionally verify the caller is an assignee of that job.
- **State transitions:** actions reject illegal transitions (e.g. accepting an
  already-completed job, completing a job you're not on) rather than silently no-op.
- **Server action results:** typed `{ ok: true … } | { ok: false, error }`. UI shows
  inline field errors or a form-level message — never a raw stack trace. Unexpected
  exceptions are caught, logged server-side, surfaced as a generic retry message.
- **DB integrity:** multi-row writes (JobLog + worker/jobtype/material links) run in a
  single Prisma transaction — all or nothing. FKs enforce referential integrity;
  soft-delete preserves history.
- **Not-found / empty states:** missing records render a clean 404; empty lists render
  purposeful empty states.

---

## 6. Testing

TDD throughout — tests written before implementation for logic and server actions.
CI-friendly.

- **Unit (Vitest):** Zod schemas; the **hours/duration calculator** (cross-midnight,
  summing, and ignoring entries with missing times) — highest-value, most bug-prone
  logic, most coverage.
- **Integration (Vitest + test Postgres):** server actions against a real test DB —
  self-log writes all link rows; **assignment lifecycle** (assign → accept → complete)
  with status transitions and audit fields; CRUD + soft-delete; **authorization tests**
  proving an employee cannot invoke admin actions, cannot accept/complete jobs they're
  not assigned to, and cannot view others' hours.
- **E2E (Playwright, smoke only):** employee logs in → logs a job → sees it in My Jobs
  and My Hours; admin logs in → assigns a job → employee accepts and completes it →
  it appears in the admin dashboard as COMPLETED.

---

## 7. Seed Data

Seed script creates: one admin, a couple of employees, a few customers, the existing
job types (Lawn Mowing, Weeding, Tree Pruning, Hedge Trimming, Garden Bed Installation)
and materials (Topsoil, Mulch Brown, Mulch Red, Gravel). All passwords bcrypt-hashed.
No real credentials committed.

---

## 8. Future Phases (designed-for, not built)

- Pricing on JobType/Material + per-customer rates → customer invoices
- QuickBooks API integration (push invoices / payroll) via an API route + `quickbooksId`
- Offline logging + sync for poor-signal job sites
- Separate timesheet / hours adjustments (travel, shop time)
- Per-assignee acceptance & richer scheduling (calendar view)
