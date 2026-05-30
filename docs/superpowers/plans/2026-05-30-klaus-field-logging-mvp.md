# Klaus Field Logging MVP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build an internal field-operations web app where employees log the jobs they work and view their own hours, and admins view all jobs/hours and manage reference data.

**Architecture:** A single Next.js (App Router) + TypeScript application. Pages are React Server Components reading via Prisma; all mutations are Server Actions that re-check session role server-side. Auth.js (credentials) provides login and role-bearing sessions. PostgreSQL via Prisma. Deployed on Vercel + managed Postgres (later).

**Tech Stack:** Next.js (App Router), TypeScript, Prisma, PostgreSQL, Auth.js (NextAuth v5), bcrypt, Zod, Vitest, Playwright.

**Project root:** `/home/manthan01/Documents/Codebase/klapp/` (already a git repo; remote `origin` = public `https://github.com/MDesai31/klapp`, default branch `main`, currently holds README + `docs/`).

**Note on git:** Commit after each task (steps below) and push to `origin main` at sensible checkpoints. The repo is **public** — never commit `.env` or secrets; `.gitignore` already excludes them.

---

## File Structure

```
klaus-fieldlog/
  prisma/
    schema.prisma            # data model + enums
    seed.ts                  # seed admin/employees/lookups
  src/
    lib/
      db.ts                  # Prisma client singleton
      hours.ts              # duration/hours calculation (pure)
      schemas.ts            # Zod schemas (shared client/server)
      auth.ts               # Auth.js config (providers, callbacks, role in session)
      authz.ts              # requireUser()/requireAdmin() server guards
    middleware.ts            # route protection by auth/role
    app/
      layout.tsx             # root layout + nav
      login/page.tsx         # login form
      jobs/page.tsx          # employee: My Jobs
      jobs/new/page.tsx      # employee: Log a Job (form)
      jobs/actions.ts        # createJobLog server action
      hours/page.tsx         # employee: My Hours
      admin/page.tsx         # admin: all job logs
      admin/hours/page.tsx   # admin: all hours
      admin/customers/page.tsx
      admin/users/page.tsx
      admin/job-types/page.tsx
      admin/materials/page.tsx
      admin/actions.ts       # admin CRUD server actions
  tests/
    unit/hours.test.ts
    unit/schemas.test.ts
    integration/jobs.test.ts
    integration/authz.test.ts
    integration/admin-crud.test.ts
    e2e/employee.spec.ts
    e2e/admin.spec.ts
  .env.example
  .gitignore
  package.json
  vitest.config.ts
  playwright.config.ts
```

---

## Task 1: Scaffold the Next.js project

**Files:**
- Create: Next app files in the existing `klapp/` repo. `.gitignore` and `.env.example` already exist (created during git setup) — leave `.gitignore` as is.

- [ ] **Step 1: Scaffold the app into the existing repo**

The `klapp/` folder already exists (git repo with README + `docs/` + `.gitignore`). Scaffold into the current directory rather than a new folder.
Run from `/home/manthan01/Documents/Codebase/klapp/`:
```bash
npx create-next-app@latest . \
  --typescript --app --eslint --src-dir --no-tailwind \
  --import-alias "@/*" --use-npm
```
When prompted that the directory is not empty, allow it to proceed (it will not overwrite `README.md`, `docs/`, or `.gitignore`). Expected: `src/app/` created alongside the existing `docs/`.

- [ ] **Step 2: Verify the dev server boots**

Run: `cd klaus-fieldlog && npm run dev`
Expected: server starts on http://localhost:3000 with no errors. Stop it with Ctrl-C.

- [ ] **Step 3: Add `.env.example` and ensure `.env` is git-ignored**

Create `klaus-fieldlog/.env.example`:
```
DATABASE_URL="postgresql://USER:PASSWORD@localhost:5432/klaus_fieldlog?schema=public"
DATABASE_URL_TEST="postgresql://USER:PASSWORD@localhost:5432/klaus_fieldlog_test?schema=public"
AUTH_SECRET="generate-with: npx auth secret"
```
Confirm `.gitignore` (created by create-next-app) contains `.env*`. If not, append `.env` and `.env.local`.

- [ ] **Step 4: Commit** *(skip per user git preference)*

```bash
git add -A && git commit -m "chore: scaffold next.js app"
```

---

## Task 2: Install dependencies and tooling

**Files:**
- Modify: `package.json`
- Create: `vitest.config.ts`

- [ ] **Step 1: Install runtime + dev dependencies**

Run in `klaus-fieldlog/`:
```bash
npm install prisma @prisma/client next-auth@beta bcryptjs zod
npm install -D vitest @vitejs/plugin-react tsx @types/bcryptjs \
  @playwright/test
```

- [ ] **Step 2: Initialize Prisma**

Run: `npx prisma init --datasource-provider postgresql`
Expected: creates `prisma/schema.prisma` and adds `DATABASE_URL` to `.env`.

- [ ] **Step 3: Create `vitest.config.ts`**

```ts
import { defineConfig } from "vitest/config";
import path from "node:path";

export default defineConfig({
  test: {
    environment: "node",
    include: ["tests/unit/**/*.test.ts", "tests/integration/**/*.test.ts"],
  },
  resolve: {
    alias: { "@": path.resolve(__dirname, "src") },
  },
});
```

- [ ] **Step 4: Add scripts to `package.json`**

In the `"scripts"` block add:
```json
"test": "vitest run",
"test:watch": "vitest",
"test:e2e": "playwright test",
"db:seed": "tsx prisma/seed.ts",
"db:reset": "prisma migrate reset --force"
```

- [ ] **Step 5: Verify the test runner works**

Run: `npm test`
Expected: vitest runs and reports "No test files found" (exit 0). That confirms config loads.

- [ ] **Step 6: Commit** *(skip per user git preference)*

```bash
git add -A && git commit -m "chore: add prisma, auth, zod, vitest, playwright"
```

---

## Task 3: Define the Prisma schema

**Files:**
- Modify: `prisma/schema.prisma`

- [ ] **Step 1: Write the schema**

Replace the model section of `prisma/schema.prisma` with:
```prisma
generator client {
  provider = "prisma-client-js"
}

datasource db {
  provider = "postgresql"
  url      = env("DATABASE_URL")
}

enum Role {
  EMPLOYEE
  ADMIN
}

model User {
  id           String   @id @default(cuid())
  email        String   @unique
  name         String
  passwordHash String
  role         Role     @default(EMPLOYEE)
  active       Boolean  @default(true)
  createdAt    DateTime @default(now())
  updatedAt    DateTime @updatedAt

  createdJobLogs JobLog[]        @relation("CreatedBy")
  workedJobs     JobLogWorker[]
}

model Customer {
  id        String   @id @default(cuid())
  name      String
  active    Boolean  @default(true)
  createdAt DateTime @default(now())
  updatedAt DateTime @updatedAt
  jobLogs   JobLog[]
}

model JobType {
  id        String   @id @default(cuid())
  name      String
  active    Boolean  @default(true)
  createdAt DateTime @default(now())
  updatedAt DateTime @updatedAt
  jobLogs   JobLogJobType[]
}

model Material {
  id        String   @id @default(cuid())
  name      String
  active    Boolean  @default(true)
  createdAt DateTime @default(now())
  updatedAt DateTime @updatedAt
  jobLogs   JobLogMaterial[]
}

model JobLog {
  id          String   @id @default(cuid())
  customerId  String
  customer    Customer @relation(fields: [customerId], references: [id])
  date        DateTime @db.Date
  timeArrived DateTime
  timeLeft    DateTime
  notes       String?
  createdById String
  createdBy   User     @relation("CreatedBy", fields: [createdById], references: [id])
  createdAt   DateTime @default(now())
  updatedAt   DateTime @updatedAt

  workers   JobLogWorker[]
  jobTypes  JobLogJobType[]
  materials JobLogMaterial[]
}

model JobLogWorker {
  jobLogId String
  jobLog   JobLog @relation(fields: [jobLogId], references: [id], onDelete: Cascade)
  userId   String
  user     User   @relation(fields: [userId], references: [id])
  @@id([jobLogId, userId])
}

model JobLogJobType {
  jobLogId  String
  jobLog    JobLog  @relation(fields: [jobLogId], references: [id], onDelete: Cascade)
  jobTypeId String
  jobType   JobType @relation(fields: [jobTypeId], references: [id])
  @@id([jobLogId, jobTypeId])
}

model JobLogMaterial {
  jobLogId   String
  jobLog     JobLog   @relation(fields: [jobLogId], references: [id], onDelete: Cascade)
  materialId String
  material   Material @relation(fields: [materialId], references: [id])
  @@id([jobLogId, materialId])
}
```

- [ ] **Step 2: Create the database and first migration**

Ensure a local Postgres is running and `DATABASE_URL` in `.env` points at it.
Run: `npx prisma migrate dev --name init`
Expected: migration applied; `@prisma/client` generated.

- [ ] **Step 3: Verify schema is valid**

Run: `npx prisma validate`
Expected: "The schema is valid."

- [ ] **Step 4: Commit** *(skip per user git preference)*

```bash
git add -A && git commit -m "feat: add prisma schema for users, customers, job logs"
```

---

## Task 4: Prisma client singleton

**Files:**
- Create: `src/lib/db.ts`

- [ ] **Step 1: Write the singleton**

```ts
import { PrismaClient } from "@prisma/client";

const globalForPrisma = globalThis as unknown as { prisma?: PrismaClient };

export const db =
  globalForPrisma.prisma ??
  new PrismaClient({
    datasourceUrl:
      process.env.NODE_ENV === "test"
        ? process.env.DATABASE_URL_TEST
        : process.env.DATABASE_URL,
  });

if (process.env.NODE_ENV !== "production") globalForPrisma.prisma = db;
```

- [ ] **Step 2: Commit** *(skip per user git preference)*

```bash
git add -A && git commit -m "feat: add prisma client singleton"
```

---

## Task 5: Hours calculation (pure logic, TDD)

**Files:**
- Create: `src/lib/hours.ts`
- Test: `tests/unit/hours.test.ts`

- [ ] **Step 1: Write the failing tests**

```ts
import { describe, it, expect } from "vitest";
import { durationHours, sumHours } from "@/lib/hours";

describe("durationHours", () => {
  it("computes a simple same-day duration", () => {
    const a = new Date("2026-05-30T08:00:00");
    const b = new Date("2026-05-30T11:30:00");
    expect(durationHours(a, b)).toBeCloseTo(3.5, 5);
  });

  it("handles a cross-midnight shift (timeLeft next day)", () => {
    const a = new Date("2026-05-30T23:00:00");
    const b = new Date("2026-05-31T01:30:00");
    expect(durationHours(a, b)).toBeCloseTo(2.5, 5);
  });

  it("returns 0 when arrival equals leave", () => {
    const a = new Date("2026-05-30T09:00:00");
    expect(durationHours(a, a)).toBe(0);
  });
});

describe("sumHours", () => {
  it("sums durations across multiple entries", () => {
    const entries = [
      { timeArrived: new Date("2026-05-30T08:00:00"), timeLeft: new Date("2026-05-30T10:00:00") },
      { timeArrived: new Date("2026-05-31T08:00:00"), timeLeft: new Date("2026-05-31T12:00:00") },
    ];
    expect(sumHours(entries)).toBeCloseTo(6, 5);
  });

  it("returns 0 for an empty list", () => {
    expect(sumHours([])).toBe(0);
  });
});
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `npx vitest run tests/unit/hours.test.ts`
Expected: FAIL — "Failed to resolve import '@/lib/hours'".

- [ ] **Step 3: Implement minimal code**

```ts
export function durationHours(timeArrived: Date, timeLeft: Date): number {
  const ms = timeLeft.getTime() - timeArrived.getTime();
  return ms / (1000 * 60 * 60);
}

export function sumHours(
  entries: { timeArrived: Date; timeLeft: Date }[],
): number {
  return entries.reduce(
    (total, e) => total + durationHours(e.timeArrived, e.timeLeft),
    0,
  );
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `npx vitest run tests/unit/hours.test.ts`
Expected: PASS (5 tests).

- [ ] **Step 5: Commit** *(skip per user git preference)*

```bash
git add -A && git commit -m "feat: add hours calculation with tests"
```

---

## Task 6: Zod schemas (TDD)

**Files:**
- Create: `src/lib/schemas.ts`
- Test: `tests/unit/schemas.test.ts`

- [ ] **Step 1: Write the failing tests**

```ts
import { describe, it, expect } from "vitest";
import { jobLogSchema, lookupSchema, userSchema } from "@/lib/schemas";

describe("jobLogSchema", () => {
  const base = {
    customerId: "c1",
    date: "2026-05-30",
    timeArrived: "2026-05-30T08:00",
    timeLeft: "2026-05-30T10:00",
    workerIds: ["u1"],
    jobTypeIds: ["j1"],
    materialIds: [],
    notes: "",
  };

  it("accepts a valid job log", () => {
    expect(jobLogSchema.safeParse(base).success).toBe(true);
  });

  it("rejects when timeLeft is not after timeArrived", () => {
    const bad = { ...base, timeLeft: "2026-05-30T08:00" };
    expect(jobLogSchema.safeParse(bad).success).toBe(false);
  });

  it("rejects when no job type is selected", () => {
    const bad = { ...base, jobTypeIds: [] };
    expect(jobLogSchema.safeParse(bad).success).toBe(false);
  });

  it("rejects a missing customer", () => {
    const bad = { ...base, customerId: "" };
    expect(jobLogSchema.safeParse(bad).success).toBe(false);
  });
});

describe("lookupSchema", () => {
  it("accepts a non-empty name", () => {
    expect(lookupSchema.safeParse({ name: "Topsoil" }).success).toBe(true);
  });
  it("rejects an empty name", () => {
    expect(lookupSchema.safeParse({ name: "" }).success).toBe(false);
  });
});

describe("userSchema", () => {
  it("accepts a valid user", () => {
    expect(
      userSchema.safeParse({
        email: "a@b.com",
        name: "A",
        role: "EMPLOYEE",
        password: "secret12",
      }).success,
    ).toBe(true);
  });
  it("rejects a bad email", () => {
    expect(
      userSchema.safeParse({
        email: "nope",
        name: "A",
        role: "EMPLOYEE",
        password: "secret12",
      }).success,
    ).toBe(false);
  });
});
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `npx vitest run tests/unit/schemas.test.ts`
Expected: FAIL — cannot resolve `@/lib/schemas`.

- [ ] **Step 3: Implement the schemas**

```ts
import { z } from "zod";

export const jobLogSchema = z
  .object({
    customerId: z.string().min(1, "Customer is required"),
    date: z.string().min(1, "Date is required"),
    timeArrived: z.string().min(1, "Arrival time is required"),
    timeLeft: z.string().min(1, "Leave time is required"),
    workerIds: z.array(z.string()).min(1, "Select at least one worker"),
    jobTypeIds: z.array(z.string()).min(1, "Select at least one job type"),
    materialIds: z.array(z.string()),
    notes: z.string().max(2000).optional().or(z.literal("")),
  })
  .refine((d) => new Date(d.timeLeft) > new Date(d.timeArrived), {
    message: "Leave time must be after arrival time",
    path: ["timeLeft"],
  });

export const lookupSchema = z.object({
  name: z.string().min(1, "Name is required").max(100),
});

export const userSchema = z.object({
  email: z.string().email("Valid email required"),
  name: z.string().min(1, "Name is required"),
  role: z.enum(["EMPLOYEE", "ADMIN"]),
  password: z.string().min(8, "Password must be at least 8 characters"),
});

export type JobLogInput = z.infer<typeof jobLogSchema>;
export type LookupInput = z.infer<typeof lookupSchema>;
export type UserInput = z.infer<typeof userSchema>;
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `npx vitest run tests/unit/schemas.test.ts`
Expected: PASS.

- [ ] **Step 5: Commit** *(skip per user git preference)*

```bash
git add -A && git commit -m "feat: add zod validation schemas with tests"
```

---

## Task 7: Seed script

**Files:**
- Create: `prisma/seed.ts`

- [ ] **Step 1: Write the seed**

```ts
import { PrismaClient, Role } from "@prisma/client";
import bcrypt from "bcryptjs";

const db = new PrismaClient();

async function main() {
  const hash = (pw: string) => bcrypt.hashSync(pw, 10);

  const admin = await db.user.upsert({
    where: { email: "admin@klaus.test" },
    update: {},
    create: {
      email: "admin@klaus.test",
      name: "Klaus Admin",
      role: Role.ADMIN,
      passwordHash: hash("admin12345"),
    },
  });

  await db.user.upsert({
    where: { email: "manthan@klaus.test" },
    update: {},
    create: {
      email: "manthan@klaus.test",
      name: "Manthan",
      role: Role.EMPLOYEE,
      passwordHash: hash("employee12345"),
    },
  });
  await db.user.upsert({
    where: { email: "thomas@klaus.test" },
    update: {},
    create: {
      email: "thomas@klaus.test",
      name: "Thomas",
      role: Role.EMPLOYEE,
      passwordHash: hash("employee12345"),
    },
  });

  for (const name of ["Customer 1", "Customer 2", "Customer 3"]) {
    await db.customer.create({ data: { name } });
  }
  for (const name of [
    "Lawn Mowing",
    "Weeding",
    "Tree Pruning",
    "Hedge Trimming",
    "Garden Bed Installation",
  ]) {
    await db.jobType.create({ data: { name } });
  }
  for (const name of ["Topsoil", "Mulch (Brown)", "Mulch (Red)", "Gravel"]) {
    await db.material.create({ data: { name } });
  }

  console.log("Seeded. Admin id:", admin.id);
}

main()
  .then(() => db.$disconnect())
  .catch(async (e) => {
    console.error(e);
    await db.$disconnect();
    process.exit(1);
  });
```

- [ ] **Step 2: Run the seed**

Run: `npm run db:seed`
Expected: "Seeded. Admin id: ..." and no errors.

- [ ] **Step 3: Verify rows exist**

Run: `npx prisma studio` (opens browser) OR
`npx prisma db execute --stdin <<< "SELECT count(*) FROM \"User\";"`
Expected: 3 users.

- [ ] **Step 4: Commit** *(skip per user git preference)*

```bash
git add -A && git commit -m "feat: add database seed script"
```

---

## Task 8: Auth.js configuration

**Files:**
- Create: `src/lib/auth.ts`, `src/app/api/auth/[...nextauth]/route.ts`

- [ ] **Step 1: Write the auth config**

`src/lib/auth.ts`:
```ts
import NextAuth from "next-auth";
import Credentials from "next-auth/providers/credentials";
import bcrypt from "bcryptjs";
import { db } from "@/lib/db";

export const { handlers, auth, signIn, signOut } = NextAuth({
  session: { strategy: "jwt" },
  pages: { signIn: "/login" },
  providers: [
    Credentials({
      credentials: { email: {}, password: {} },
      authorize: async (creds) => {
        const email = creds?.email as string;
        const password = creds?.password as string;
        if (!email || !password) return null;
        const user = await db.user.findUnique({ where: { email } });
        if (!user || !user.active) return null;
        if (!bcrypt.compareSync(password, user.passwordHash)) return null;
        return { id: user.id, name: user.name, email: user.email, role: user.role };
      },
    }),
  ],
  callbacks: {
    jwt: ({ token, user }) => {
      if (user) {
        token.role = (user as { role: string }).role;
        token.id = (user as { id: string }).id;
      }
      return token;
    },
    session: ({ session, token }) => {
      if (session.user) {
        (session.user as { role?: string }).role = token.role as string;
        (session.user as { id?: string }).id = token.id as string;
      }
      return session;
    },
  },
});
```

- [ ] **Step 2: Wire the route handler**

`src/app/api/auth/[...nextauth]/route.ts`:
```ts
import { handlers } from "@/lib/auth";
export const { GET, POST } = handlers;
```

- [ ] **Step 3: Add a type augmentation for `role`**

Create `src/types/next-auth.d.ts`:
```ts
import type { DefaultSession } from "next-auth";

declare module "next-auth" {
  interface Session {
    user: { id: string; role: string } & DefaultSession["user"];
  }
}
```

- [ ] **Step 4: Generate `AUTH_SECRET` and confirm typecheck**

Run: `npx auth secret` (writes `AUTH_SECRET` to `.env`), then `npx tsc --noEmit`
Expected: no type errors.

- [ ] **Step 5: Commit** *(skip per user git preference)*

```bash
git add -A && git commit -m "feat: add auth.js credentials auth with role in session"
```

---

## Task 9: Server-side authorization guards

**Files:**
- Create: `src/lib/authz.ts`

- [ ] **Step 1: Write the guards**

```ts
import { auth } from "@/lib/auth";

export async function requireUser() {
  const session = await auth();
  if (!session?.user) throw new Error("UNAUTHENTICATED");
  return session.user;
}

export async function requireAdmin() {
  const user = await requireUser();
  if (user.role !== "ADMIN") throw new Error("FORBIDDEN");
  return user;
}
```

- [ ] **Step 2: Typecheck**

Run: `npx tsc --noEmit`
Expected: no errors.

- [ ] **Step 3: Commit** *(skip per user git preference)*

```bash
git add -A && git commit -m "feat: add requireUser/requireAdmin guards"
```

---

## Task 10: Route protection middleware

**Files:**
- Create: `src/middleware.ts`

- [ ] **Step 1: Write the middleware**

```ts
import { auth } from "@/lib/auth";
import { NextResponse } from "next/server";

export default auth((req) => {
  const { nextUrl } = req;
  const isLoggedIn = !!req.auth?.user;
  const role = (req.auth?.user as { role?: string } | undefined)?.role;
  const path = nextUrl.pathname;

  const isLogin = path === "/login";
  const isAdminArea = path.startsWith("/admin");

  if (!isLoggedIn && !isLogin) {
    return NextResponse.redirect(new URL("/login", nextUrl));
  }
  if (isLoggedIn && isLogin) {
    return NextResponse.redirect(new URL(role === "ADMIN" ? "/admin" : "/jobs", nextUrl));
  }
  if (isAdminArea && role !== "ADMIN") {
    return NextResponse.redirect(new URL("/jobs", nextUrl));
  }
  return NextResponse.next();
});

export const config = {
  matcher: ["/((?!api|_next/static|_next/image|favicon.ico).*)"],
};
```

- [ ] **Step 2: Typecheck**

Run: `npx tsc --noEmit`
Expected: no errors.

- [ ] **Step 3: Commit** *(skip per user git preference)*

```bash
git add -A && git commit -m "feat: protect routes by auth and role"
```

---

## Task 11: Test DB harness + authorization integration tests (TDD)

**Files:**
- Create: `tests/integration/helpers.ts`
- Test: `tests/integration/authz.test.ts`

- [ ] **Step 1: Create the test DB and apply schema**

Run:
```bash
createdb klaus_fieldlog_test 2>/dev/null || true
DATABASE_URL="$DATABASE_URL_TEST" npx prisma migrate deploy
```
Expected: migrations applied to the test DB.

- [ ] **Step 2: Write a reset/seed helper**

`tests/integration/helpers.ts`:
```ts
import { PrismaClient, Role } from "@prisma/client";
import bcrypt from "bcryptjs";

export const testDb = new PrismaClient({
  datasourceUrl: process.env.DATABASE_URL_TEST,
});

export async function resetDb() {
  await testDb.jobLogWorker.deleteMany();
  await testDb.jobLogJobType.deleteMany();
  await testDb.jobLogMaterial.deleteMany();
  await testDb.jobLog.deleteMany();
  await testDb.customer.deleteMany();
  await testDb.jobType.deleteMany();
  await testDb.material.deleteMany();
  await testDb.user.deleteMany();
}

export async function makeBaseData() {
  const admin = await testDb.user.create({
    data: { email: "admin@t.test", name: "Admin", role: Role.ADMIN, passwordHash: bcrypt.hashSync("x", 4) },
  });
  const emp = await testDb.user.create({
    data: { email: "emp@t.test", name: "Emp", role: Role.EMPLOYEE, passwordHash: bcrypt.hashSync("x", 4) },
  });
  const customer = await testDb.customer.create({ data: { name: "Cust" } });
  const jobType = await testDb.jobType.create({ data: { name: "Mowing" } });
  const material = await testDb.material.create({ data: { name: "Mulch" } });
  return { admin, emp, customer, jobType, material };
}
```

- [ ] **Step 3: Write the failing authz test**

`tests/integration/authz.test.ts`:
```ts
import { describe, it, expect, vi, beforeEach } from "vitest";

const mockUser = vi.fn();
vi.mock("@/lib/auth", () => ({ auth: () => mockUser() }));

import { requireAdmin, requireUser } from "@/lib/authz";

describe("authz guards", () => {
  beforeEach(() => mockUser.mockReset());

  it("requireUser throws when unauthenticated", async () => {
    mockUser.mockResolvedValue(null);
    await expect(requireUser()).rejects.toThrow("UNAUTHENTICATED");
  });

  it("requireAdmin throws for an employee", async () => {
    mockUser.mockResolvedValue({ user: { id: "1", role: "EMPLOYEE" } });
    await expect(requireAdmin()).rejects.toThrow("FORBIDDEN");
  });

  it("requireAdmin passes for an admin", async () => {
    mockUser.mockResolvedValue({ user: { id: "1", role: "ADMIN" } });
    await expect(requireAdmin()).resolves.toMatchObject({ role: "ADMIN" });
  });
});
```

- [ ] **Step 4: Run to verify pass**

Run: `npx vitest run tests/integration/authz.test.ts`
Expected: PASS (the guards from Task 9 already exist).

- [ ] **Step 5: Commit** *(skip per user git preference)*

```bash
git add -A && git commit -m "test: add test db harness and authz tests"
```

---

## Task 12: Create-job-log server action (TDD)

**Files:**
- Create: `src/app/jobs/actions.ts`
- Test: `tests/integration/jobs.test.ts`

- [ ] **Step 1: Write the failing test**

`tests/integration/jobs.test.ts`:
```ts
import { describe, it, expect, beforeEach, vi } from "vitest";
import { resetDb, makeBaseData, testDb } from "./helpers";

const mockAuth = vi.fn();
vi.mock("@/lib/auth", () => ({ auth: () => mockAuth() }));
vi.mock("@/lib/db", () => ({ db: testDb }));
vi.mock("next/cache", () => ({ revalidatePath: vi.fn() }));

import { createJobLog } from "@/app/jobs/actions";

describe("createJobLog", () => {
  beforeEach(async () => {
    await resetDb();
  });

  it("writes a job log with all link rows in one transaction", async () => {
    const { emp, customer, jobType, material } = await makeBaseData();
    mockAuth.mockResolvedValue({ user: { id: emp.id, role: "EMPLOYEE" } });

    const res = await createJobLog({
      customerId: customer.id,
      date: "2026-05-30",
      timeArrived: "2026-05-30T08:00",
      timeLeft: "2026-05-30T10:00",
      workerIds: [emp.id],
      jobTypeIds: [jobType.id],
      materialIds: [material.id],
      notes: "test",
    });

    expect(res.ok).toBe(true);
    const logs = await testDb.jobLog.findMany({
      include: { workers: true, jobTypes: true, materials: true },
    });
    expect(logs).toHaveLength(1);
    expect(logs[0].createdById).toBe(emp.id);
    expect(logs[0].workers).toHaveLength(1);
    expect(logs[0].jobTypes).toHaveLength(1);
    expect(logs[0].materials).toHaveLength(1);
  });

  it("rejects invalid input without writing", async () => {
    const { emp, customer, jobType } = await makeBaseData();
    mockAuth.mockResolvedValue({ user: { id: emp.id, role: "EMPLOYEE" } });

    const res = await createJobLog({
      customerId: customer.id,
      date: "2026-05-30",
      timeArrived: "2026-05-30T10:00",
      timeLeft: "2026-05-30T08:00", // before arrival
      workerIds: [emp.id],
      jobTypeIds: [jobType.id],
      materialIds: [],
      notes: "",
    });

    expect(res.ok).toBe(false);
    expect(await testDb.jobLog.count()).toBe(0);
  });

  it("throws when unauthenticated", async () => {
    mockAuth.mockResolvedValue(null);
    await expect(
      createJobLog({
        customerId: "x", date: "2026-05-30",
        timeArrived: "2026-05-30T08:00", timeLeft: "2026-05-30T10:00",
        workerIds: ["x"], jobTypeIds: ["x"], materialIds: [], notes: "",
      }),
    ).rejects.toThrow("UNAUTHENTICATED");
  });
});
```

- [ ] **Step 2: Run to verify it fails**

Run: `npx vitest run tests/integration/jobs.test.ts`
Expected: FAIL — cannot resolve `@/app/jobs/actions`.

- [ ] **Step 3: Implement the action**

`src/app/jobs/actions.ts`:
```ts
"use server";

import { db } from "@/lib/db";
import { requireUser } from "@/lib/authz";
import { jobLogSchema, type JobLogInput } from "@/lib/schemas";
import { revalidatePath } from "next/cache";

type Result = { ok: true; id: string } | { ok: false; error: string };

export async function createJobLog(input: JobLogInput): Promise<Result> {
  const user = await requireUser();

  const parsed = jobLogSchema.safeParse(input);
  if (!parsed.success) {
    return { ok: false, error: parsed.error.issues[0]?.message ?? "Invalid input" };
  }
  const d = parsed.data;

  try {
    const created = await db.$transaction(async (tx) => {
      const log = await tx.jobLog.create({
        data: {
          customerId: d.customerId,
          date: new Date(d.date),
          timeArrived: new Date(d.timeArrived),
          timeLeft: new Date(d.timeLeft),
          notes: d.notes || null,
          createdById: user.id,
          workers: { create: d.workerIds.map((userId) => ({ userId })) },
          jobTypes: { create: d.jobTypeIds.map((jobTypeId) => ({ jobTypeId })) },
          materials: { create: d.materialIds.map((materialId) => ({ materialId })) },
        },
      });
      return log;
    });
    revalidatePath("/jobs");
    return { ok: true, id: created.id };
  } catch {
    return { ok: false, error: "Could not save the job. Please try again." };
  }
}
```

- [ ] **Step 4: Run to verify it passes**

Run: `npx vitest run tests/integration/jobs.test.ts`
Expected: PASS (3 tests).

- [ ] **Step 5: Commit** *(skip per user git preference)*

```bash
git add -A && git commit -m "feat: add createJobLog server action with tests"
```

---

## Task 13: Admin CRUD server actions (TDD)

**Files:**
- Create: `src/app/admin/actions.ts`
- Test: `tests/integration/admin-crud.test.ts`

- [ ] **Step 1: Write the failing test**

`tests/integration/admin-crud.test.ts`:
```ts
import { describe, it, expect, beforeEach, vi } from "vitest";
import { resetDb, makeBaseData, testDb } from "./helpers";

const mockAuth = vi.fn();
vi.mock("@/lib/auth", () => ({ auth: () => mockAuth() }));
vi.mock("@/lib/db", () => ({ db: testDb }));
vi.mock("next/cache", () => ({ revalidatePath: vi.fn() }));

import { createCustomer, setCustomerActive } from "@/app/admin/actions";

describe("admin customer CRUD", () => {
  beforeEach(async () => await resetDb());

  it("admin can create a customer", async () => {
    const { admin } = await makeBaseData();
    mockAuth.mockResolvedValue({ user: { id: admin.id, role: "ADMIN" } });
    const res = await createCustomer({ name: "New Co" });
    expect(res.ok).toBe(true);
    expect(await testDb.customer.count({ where: { name: "New Co" } })).toBe(1);
  });

  it("employee cannot create a customer", async () => {
    const { emp } = await makeBaseData();
    mockAuth.mockResolvedValue({ user: { id: emp.id, role: "EMPLOYEE" } });
    await expect(createCustomer({ name: "Nope" })).rejects.toThrow("FORBIDDEN");
  });

  it("deactivate soft-deletes (sets active=false)", async () => {
    const { admin, customer } = await makeBaseData();
    mockAuth.mockResolvedValue({ user: { id: admin.id, role: "ADMIN" } });
    const res = await setCustomerActive(customer.id, false);
    expect(res.ok).toBe(true);
    const c = await testDb.customer.findUnique({ where: { id: customer.id } });
    expect(c?.active).toBe(false);
  });
});
```

- [ ] **Step 2: Run to verify it fails**

Run: `npx vitest run tests/integration/admin-crud.test.ts`
Expected: FAIL — cannot resolve `@/app/admin/actions`.

- [ ] **Step 3: Implement the actions**

`src/app/admin/actions.ts`:
```ts
"use server";

import { db } from "@/lib/db";
import { requireAdmin } from "@/lib/authz";
import { lookupSchema, userSchema, type LookupInput, type UserInput } from "@/lib/schemas";
import { revalidatePath } from "next/cache";
import bcrypt from "bcryptjs";

type Result = { ok: true } | { ok: false; error: string };

// --- Customers ---
export async function createCustomer(input: LookupInput): Promise<Result> {
  await requireAdmin();
  const p = lookupSchema.safeParse(input);
  if (!p.success) return { ok: false, error: p.error.issues[0].message };
  await db.customer.create({ data: { name: p.data.name } });
  revalidatePath("/admin/customers");
  return { ok: true };
}

export async function updateCustomer(id: string, input: LookupInput): Promise<Result> {
  await requireAdmin();
  const p = lookupSchema.safeParse(input);
  if (!p.success) return { ok: false, error: p.error.issues[0].message };
  await db.customer.update({ where: { id }, data: { name: p.data.name } });
  revalidatePath("/admin/customers");
  return { ok: true };
}

export async function setCustomerActive(id: string, active: boolean): Promise<Result> {
  await requireAdmin();
  await db.customer.update({ where: { id }, data: { active } });
  revalidatePath("/admin/customers");
  return { ok: true };
}

// --- Job Types ---
export async function createJobType(input: LookupInput): Promise<Result> {
  await requireAdmin();
  const p = lookupSchema.safeParse(input);
  if (!p.success) return { ok: false, error: p.error.issues[0].message };
  await db.jobType.create({ data: { name: p.data.name } });
  revalidatePath("/admin/job-types");
  return { ok: true };
}

export async function setJobTypeActive(id: string, active: boolean): Promise<Result> {
  await requireAdmin();
  await db.jobType.update({ where: { id }, data: { active } });
  revalidatePath("/admin/job-types");
  return { ok: true };
}

// --- Materials ---
export async function createMaterial(input: LookupInput): Promise<Result> {
  await requireAdmin();
  const p = lookupSchema.safeParse(input);
  if (!p.success) return { ok: false, error: p.error.issues[0].message };
  await db.material.create({ data: { name: p.data.name } });
  revalidatePath("/admin/materials");
  return { ok: true };
}

export async function setMaterialActive(id: string, active: boolean): Promise<Result> {
  await requireAdmin();
  await db.material.update({ where: { id }, data: { active } });
  revalidatePath("/admin/materials");
  return { ok: true };
}

// --- Users ---
export async function createUser(input: UserInput): Promise<Result> {
  await requireAdmin();
  const p = userSchema.safeParse(input);
  if (!p.success) return { ok: false, error: p.error.issues[0].message };
  await db.user.create({
    data: {
      email: p.data.email,
      name: p.data.name,
      role: p.data.role,
      passwordHash: bcrypt.hashSync(p.data.password, 10),
    },
  });
  revalidatePath("/admin/users");
  return { ok: true };
}

export async function setUserActive(id: string, active: boolean): Promise<Result> {
  await requireAdmin();
  await db.user.update({ where: { id }, data: { active } });
  revalidatePath("/admin/users");
  return { ok: true };
}

export async function resetUserPassword(id: string, password: string): Promise<Result> {
  await requireAdmin();
  if (password.length < 8) return { ok: false, error: "Password must be at least 8 characters" };
  await db.user.update({ where: { id }, data: { passwordHash: bcrypt.hashSync(password, 10) } });
  revalidatePath("/admin/users");
  return { ok: true };
}
```

- [ ] **Step 4: Run to verify it passes**

Run: `npx vitest run tests/integration/admin-crud.test.ts`
Expected: PASS (3 tests).

- [ ] **Step 5: Run the full unit+integration suite**

Run: `npm test`
Expected: all tests PASS.

- [ ] **Step 6: Commit** *(skip per user git preference)*

```bash
git add -A && git commit -m "feat: add admin CRUD server actions with tests"
```

---

## Task 14: Root layout, nav, and login page

**Files:**
- Modify: `src/app/layout.tsx`
- Create: `src/app/login/page.tsx`, `src/components/Nav.tsx`

- [ ] **Step 1: Write the nav**

`src/components/Nav.tsx`:
```tsx
import Link from "next/link";
import { auth, signOut } from "@/lib/auth";

export default async function Nav() {
  const session = await auth();
  if (!session?.user) return null;
  const role = (session.user as { role?: string }).role;
  return (
    <nav style={{ display: "flex", gap: 16, padding: 12, borderBottom: "1px solid #ddd" }}>
      <Link href="/jobs">My Jobs</Link>
      <Link href="/jobs/new">Log a Job</Link>
      <Link href="/hours">My Hours</Link>
      {role === "ADMIN" && (
        <>
          <Link href="/admin">All Jobs</Link>
          <Link href="/admin/hours">All Hours</Link>
          <Link href="/admin/customers">Customers</Link>
          <Link href="/admin/users">Employees</Link>
          <Link href="/admin/job-types">Job Types</Link>
          <Link href="/admin/materials">Materials</Link>
        </>
      )}
      <form action={async () => { "use server"; await signOut({ redirectTo: "/login" }); }} style={{ marginLeft: "auto" }}>
        <button type="submit">Logout</button>
      </form>
    </nav>
  );
}
```

- [ ] **Step 2: Update the root layout**

`src/app/layout.tsx`:
```tsx
import type { ReactNode } from "react";
import Nav from "@/components/Nav";

export const metadata = { title: "Klaus Field Log" };

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="en">
      <body style={{ fontFamily: "system-ui, sans-serif", margin: 0 }}>
        <Nav />
        <main style={{ maxWidth: 720, margin: "0 auto", padding: 16 }}>{children}</main>
      </body>
    </html>
  );
}
```

- [ ] **Step 3: Write the login page**

`src/app/login/page.tsx`:
```tsx
import { signIn } from "@/lib/auth";

export default function LoginPage({ searchParams }: { searchParams: { error?: string } }) {
  return (
    <div style={{ maxWidth: 360, margin: "10vh auto" }}>
      <h1>Klaus Field Log</h1>
      {searchParams.error && <p style={{ color: "red" }}>Invalid email or password.</p>}
      <form
        action={async (formData) => {
          "use server";
          await signIn("credentials", {
            email: formData.get("email"),
            password: formData.get("password"),
            redirectTo: "/jobs",
          });
        }}
        style={{ display: "flex", flexDirection: "column", gap: 10 }}
      >
        <input name="email" type="email" placeholder="Email" required />
        <input name="password" type="password" placeholder="Password" required />
        <button type="submit">Log in</button>
      </form>
    </div>
  );
}
```

- [ ] **Step 4: Manually verify login works**

Run: `npm run dev`, open http://localhost:3000 → redirected to `/login`. Log in as `admin@klaus.test` / `admin12345`. Expected: redirected to `/jobs`, nav shows admin links.

- [ ] **Step 5: Commit** *(skip per user git preference)*

```bash
git add -A && git commit -m "feat: add layout, nav, and login page"
```

---

## Task 15: Employee pages — My Jobs, Log a Job, My Hours

**Files:**
- Create: `src/app/jobs/page.tsx`, `src/app/jobs/new/page.tsx`, `src/app/jobs/JobForm.tsx`, `src/app/hours/page.tsx`

- [ ] **Step 1: My Jobs list page**

`src/app/jobs/page.tsx`:
```tsx
import { db } from "@/lib/db";
import { requireUser } from "@/lib/authz";
import { durationHours } from "@/lib/hours";
import Link from "next/link";

export default async function JobsPage() {
  const user = await requireUser();
  const logs = await db.jobLog.findMany({
    where: { workers: { some: { userId: user.id } } },
    include: { customer: true },
    orderBy: { date: "desc" },
  });

  if (logs.length === 0) {
    return (
      <div>
        <h1>My Jobs</h1>
        <p>No jobs logged yet. <Link href="/jobs/new">Log your first job</Link>.</p>
      </div>
    );
  }

  return (
    <div>
      <h1>My Jobs</h1>
      <table style={{ width: "100%", borderCollapse: "collapse" }}>
        <thead><tr><th align="left">Date</th><th align="left">Customer</th><th align="left">Hours</th></tr></thead>
        <tbody>
          {logs.map((l) => (
            <tr key={l.id}>
              <td>{l.date.toISOString().slice(0, 10)}</td>
              <td>{l.customer.name}</td>
              <td>{durationHours(l.timeArrived, l.timeLeft).toFixed(2)}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
```

- [ ] **Step 2: Job form (client component)**

`src/app/jobs/JobForm.tsx`:
```tsx
"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { createJobLog } from "@/app/jobs/actions";

type Opt = { id: string; name: string };

export default function JobForm({
  customers, employees, jobTypes, materials,
}: { customers: Opt[]; employees: Opt[]; jobTypes: Opt[]; materials: Opt[] }) {
  const router = useRouter();
  const [error, setError] = useState("");
  const [saving, setSaving] = useState(false);

  return (
    <form
      onSubmit={async (e) => {
        e.preventDefault();
        setError(""); setSaving(true);
        const fd = new FormData(e.currentTarget);
        const multi = (k: string) => fd.getAll(k).map(String);
        const res = await createJobLog({
          customerId: String(fd.get("customerId") || ""),
          date: String(fd.get("date") || ""),
          timeArrived: String(fd.get("timeArrived") || ""),
          timeLeft: String(fd.get("timeLeft") || ""),
          workerIds: multi("workerIds"),
          jobTypeIds: multi("jobTypeIds"),
          materialIds: multi("materialIds"),
          notes: String(fd.get("notes") || ""),
        });
        setSaving(false);
        if (res.ok) router.push("/jobs");
        else setError(res.error);
      }}
      style={{ display: "flex", flexDirection: "column", gap: 12 }}
    >
      {error && <p style={{ color: "red" }}>{error}</p>}
      <label>Customer
        <select name="customerId" required defaultValue="">
          <option value="" disabled>Select customer</option>
          {customers.map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}
        </select>
      </label>
      <label>Date <input name="date" type="date" required /></label>
      <label>Time arrived <input name="timeArrived" type="datetime-local" required /></label>
      <label>Time left <input name="timeLeft" type="datetime-local" required /></label>
      <label>Workers
        <select name="workerIds" multiple required>
          {employees.map((e) => <option key={e.id} value={e.id}>{e.name}</option>)}
        </select>
      </label>
      <label>Job types
        <select name="jobTypeIds" multiple required>
          {jobTypes.map((j) => <option key={j.id} value={j.id}>{j.name}</option>)}
        </select>
      </label>
      <label>Materials
        <select name="materialIds" multiple>
          {materials.map((m) => <option key={m.id} value={m.id}>{m.name}</option>)}
        </select>
      </label>
      <label>Notes <textarea name="notes" rows={3} /></label>
      <button type="submit" disabled={saving}>{saving ? "Saving..." : "Submit Job"}</button>
    </form>
  );
}
```

- [ ] **Step 3: Log a Job page (loads options, renders form)**

`src/app/jobs/new/page.tsx`:
```tsx
import { db } from "@/lib/db";
import { requireUser } from "@/lib/authz";
import JobForm from "@/app/jobs/JobForm";

export default async function NewJobPage() {
  await requireUser();
  const [customers, employees, jobTypes, materials] = await Promise.all([
    db.customer.findMany({ where: { active: true }, orderBy: { name: "asc" } }),
    db.user.findMany({ where: { active: true }, orderBy: { name: "asc" } }),
    db.jobType.findMany({ where: { active: true }, orderBy: { name: "asc" } }),
    db.material.findMany({ where: { active: true }, orderBy: { name: "asc" } }),
  ]);
  return (
    <div>
      <h1>Log a Job</h1>
      <JobForm
        customers={customers.map((c) => ({ id: c.id, name: c.name }))}
        employees={employees.map((u) => ({ id: u.id, name: u.name }))}
        jobTypes={jobTypes.map((j) => ({ id: j.id, name: j.name }))}
        materials={materials.map((m) => ({ id: m.id, name: m.name }))}
      />
    </div>
  );
}
```

- [ ] **Step 4: My Hours page**

`src/app/hours/page.tsx`:
```tsx
import { db } from "@/lib/db";
import { requireUser } from "@/lib/authz";
import { durationHours, sumHours } from "@/lib/hours";

export default async function HoursPage({ searchParams }: { searchParams: { year?: string; month?: string } }) {
  const user = await requireUser();
  const now = new Date();
  const year = Number(searchParams.year ?? now.getFullYear());
  const month = searchParams.month ? Number(searchParams.month) : null;

  const start = new Date(year, month ? month - 1 : 0, 1);
  const end = new Date(year, month ? month : 12, month ? 1 : 1);

  const logs = await db.jobLog.findMany({
    where: { workers: { some: { userId: user.id } }, date: { gte: start, lt: end } },
    include: { customer: true },
    orderBy: { date: "asc" },
  });

  const total = sumHours(logs);

  return (
    <div>
      <h1>My Hours</h1>
      <form method="get" style={{ display: "flex", gap: 8, marginBottom: 12 }}>
        <input name="year" type="number" defaultValue={year} />
        <input name="month" type="number" min={1} max={12} defaultValue={month ?? ""} placeholder="month (optional)" />
        <button type="submit">Filter</button>
      </form>
      <table style={{ width: "100%", borderCollapse: "collapse" }}>
        <thead><tr><th align="left">Date</th><th align="left">Customer</th><th align="left">Hours</th></tr></thead>
        <tbody>
          {logs.map((l) => (
            <tr key={l.id}>
              <td>{l.date.toISOString().slice(0, 10)}</td>
              <td>{l.customer.name}</td>
              <td>{durationHours(l.timeArrived, l.timeLeft).toFixed(2)}</td>
            </tr>
          ))}
        </tbody>
      </table>
      <h3 style={{ textAlign: "right" }}>Total: {total.toFixed(2)} h</h3>
    </div>
  );
}
```

- [ ] **Step 5: Manually verify the employee flow**

Run: `npm run dev`. Log in as `manthan@klaus.test` / `employee12345`. Log a job, confirm it appears in My Jobs and contributes to My Hours.

- [ ] **Step 6: Commit** *(skip per user git preference)*

```bash
git add -A && git commit -m "feat: add employee jobs and hours pages"
```

---

## Task 16: Admin pages — all jobs, all hours, lookup CRUD

**Files:**
- Create: `src/app/admin/page.tsx`, `src/app/admin/hours/page.tsx`,
  `src/app/admin/customers/page.tsx`, `src/app/admin/users/page.tsx`,
  `src/app/admin/job-types/page.tsx`, `src/app/admin/materials/page.tsx`,
  `src/app/admin/LookupManager.tsx`

- [ ] **Step 1: Admin all-jobs dashboard**

`src/app/admin/page.tsx`:
```tsx
import { db } from "@/lib/db";
import { requireAdmin } from "@/lib/authz";
import { durationHours } from "@/lib/hours";

export default async function AdminPage() {
  await requireAdmin();
  const logs = await db.jobLog.findMany({
    include: { customer: true, createdBy: true, workers: { include: { user: true } } },
    orderBy: { date: "desc" },
    take: 200,
  });
  return (
    <div>
      <h1>All Jobs</h1>
      <table style={{ width: "100%", borderCollapse: "collapse" }}>
        <thead><tr><th align="left">Date</th><th align="left">Customer</th><th align="left">Logged by</th><th align="left">Workers</th><th align="left">Hours</th></tr></thead>
        <tbody>
          {logs.map((l) => (
            <tr key={l.id}>
              <td>{l.date.toISOString().slice(0, 10)}</td>
              <td>{l.customer.name}</td>
              <td>{l.createdBy.name}</td>
              <td>{l.workers.map((w) => w.user.name).join(", ")}</td>
              <td>{durationHours(l.timeArrived, l.timeLeft).toFixed(2)}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
```

- [ ] **Step 2: Admin all-hours page (per employee)**

`src/app/admin/hours/page.tsx`:
```tsx
import { db } from "@/lib/db";
import { requireAdmin } from "@/lib/authz";
import { sumHours } from "@/lib/hours";

export default async function AdminHoursPage({ searchParams }: { searchParams: { year?: string; month?: string } }) {
  await requireAdmin();
  const now = new Date();
  const year = Number(searchParams.year ?? now.getFullYear());
  const month = searchParams.month ? Number(searchParams.month) : null;
  const start = new Date(year, month ? month - 1 : 0, 1);
  const end = new Date(year, month ? month : 12, month ? 1 : 1);

  const employees = await db.user.findMany({ where: { role: "EMPLOYEE" }, orderBy: { name: "asc" } });
  const rows = await Promise.all(
    employees.map(async (e) => {
      const logs = await db.jobLog.findMany({
        where: { workers: { some: { userId: e.id } }, date: { gte: start, lt: end } },
        select: { timeArrived: true, timeLeft: true },
      });
      return { name: e.name, total: sumHours(logs) };
    }),
  );

  return (
    <div>
      <h1>All Hours</h1>
      <form method="get" style={{ display: "flex", gap: 8, marginBottom: 12 }}>
        <input name="year" type="number" defaultValue={year} />
        <input name="month" type="number" min={1} max={12} defaultValue={month ?? ""} placeholder="month (optional)" />
        <button type="submit">Filter</button>
      </form>
      <table style={{ width: "100%", borderCollapse: "collapse" }}>
        <thead><tr><th align="left">Employee</th><th align="left">Total Hours</th></tr></thead>
        <tbody>
          {rows.map((r) => (<tr key={r.name}><td>{r.name}</td><td>{r.total.toFixed(2)}</td></tr>))}
        </tbody>
      </table>
    </div>
  );
}
```

- [ ] **Step 3: Reusable lookup manager (client component)**

`src/app/admin/LookupManager.tsx`:
```tsx
"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";

type Item = { id: string; name: string; active: boolean };
type Actions = {
  create: (input: { name: string }) => Promise<{ ok: boolean; error?: string }>;
  setActive: (id: string, active: boolean) => Promise<{ ok: boolean; error?: string }>;
};

export default function LookupManager({ title, items, actions }: { title: string; items: Item[]; actions: Actions }) {
  const router = useRouter();
  const [name, setName] = useState("");
  const [error, setError] = useState("");

  return (
    <div>
      <h1>{title}</h1>
      <form
        onSubmit={async (e) => {
          e.preventDefault();
          setError("");
          const res = await actions.create({ name });
          if (res.ok) { setName(""); router.refresh(); }
          else setError(res.error ?? "Error");
        }}
        style={{ display: "flex", gap: 8, marginBottom: 12 }}
      >
        <input value={name} onChange={(e) => setName(e.target.value)} placeholder={`New ${title.toLowerCase()}`} />
        <button type="submit">Add</button>
      </form>
      {error && <p style={{ color: "red" }}>{error}</p>}
      <ul style={{ listStyle: "none", padding: 0 }}>
        {items.map((it) => (
          <li key={it.id} style={{ display: "flex", gap: 8, padding: "4px 0", opacity: it.active ? 1 : 0.5 }}>
            <span style={{ flex: 1 }}>{it.name}{!it.active && " (inactive)"}</span>
            <button onClick={async () => { await actions.setActive(it.id, !it.active); router.refresh(); }}>
              {it.active ? "Deactivate" : "Activate"}
            </button>
          </li>
        ))}
      </ul>
    </div>
  );
}
```

- [ ] **Step 4: Customers page (server wrapper passing bound actions)**

`src/app/admin/customers/page.tsx`:
```tsx
import { db } from "@/lib/db";
import { requireAdmin } from "@/lib/authz";
import LookupManager from "@/app/admin/LookupManager";
import { createCustomer, setCustomerActive } from "@/app/admin/actions";

export default async function CustomersPage() {
  await requireAdmin();
  const items = await db.customer.findMany({ orderBy: { name: "asc" } });
  return (
    <LookupManager
      title="Customers"
      items={items.map((i) => ({ id: i.id, name: i.name, active: i.active }))}
      actions={{ create: createCustomer, setActive: setCustomerActive }}
    />
  );
}
```

- [ ] **Step 5: Job Types and Materials pages**

`src/app/admin/job-types/page.tsx`:
```tsx
import { db } from "@/lib/db";
import { requireAdmin } from "@/lib/authz";
import LookupManager from "@/app/admin/LookupManager";
import { createJobType, setJobTypeActive } from "@/app/admin/actions";

export default async function JobTypesPage() {
  await requireAdmin();
  const items = await db.jobType.findMany({ orderBy: { name: "asc" } });
  return (
    <LookupManager
      title="Job Types"
      items={items.map((i) => ({ id: i.id, name: i.name, active: i.active }))}
      actions={{ create: createJobType, setActive: setJobTypeActive }}
    />
  );
}
```

`src/app/admin/materials/page.tsx`:
```tsx
import { db } from "@/lib/db";
import { requireAdmin } from "@/lib/authz";
import LookupManager from "@/app/admin/LookupManager";
import { createMaterial, setMaterialActive } from "@/app/admin/actions";

export default async function MaterialsPage() {
  await requireAdmin();
  const items = await db.material.findMany({ orderBy: { name: "asc" } });
  return (
    <LookupManager
      title="Materials"
      items={items.map((i) => ({ id: i.id, name: i.name, active: i.active }))}
      actions={{ create: createMaterial, setActive: setMaterialActive }}
    />
  );
}
```

- [ ] **Step 6: Users (employees) page**

`src/app/admin/users/page.tsx`:
```tsx
import { db } from "@/lib/db";
import { requireAdmin } from "@/lib/authz";
import { createUser, setUserActive } from "@/app/admin/actions";

export default async function UsersPage() {
  await requireAdmin();
  const users = await db.user.findMany({ orderBy: { name: "asc" } });
  return (
    <div>
      <h1>Employees</h1>
      <form action={async (fd) => {
        "use server";
        await createUser({
          email: String(fd.get("email")), name: String(fd.get("name")),
          role: (String(fd.get("role")) as "EMPLOYEE" | "ADMIN"),
          password: String(fd.get("password")),
        });
      }} style={{ display: "flex", gap: 8, marginBottom: 12, flexWrap: "wrap" }}>
        <input name="name" placeholder="Name" required />
        <input name="email" type="email" placeholder="Email" required />
        <select name="role" defaultValue="EMPLOYEE"><option value="EMPLOYEE">Employee</option><option value="ADMIN">Admin</option></select>
        <input name="password" type="password" placeholder="Temp password (min 8)" required />
        <button type="submit">Add</button>
      </form>
      <ul style={{ listStyle: "none", padding: 0 }}>
        {users.map((u) => (
          <li key={u.id} style={{ display: "flex", gap: 8, padding: "4px 0", opacity: u.active ? 1 : 0.5 }}>
            <span style={{ flex: 1 }}>{u.name} — {u.email} ({u.role}){!u.active && " (inactive)"}</span>
            <form action={async () => { "use server"; await setUserActive(u.id, !u.active); }}>
              <button type="submit">{u.active ? "Deactivate" : "Activate"}</button>
            </form>
          </li>
        ))}
      </ul>
    </div>
  );
}
```

- [ ] **Step 7: Manually verify the admin flow**

Run: `npm run dev`, log in as admin. Confirm: All Jobs lists logged jobs; All Hours shows per-employee totals; adding a customer/job-type/material/user works and deactivate toggles.

- [ ] **Step 8: Typecheck + full test suite**

Run: `npx tsc --noEmit && npm test`
Expected: no type errors; all unit + integration tests PASS.

- [ ] **Step 9: Commit** *(skip per user git preference)*

```bash
git add -A && git commit -m "feat: add admin dashboard, hours, and lookup CRUD pages"
```

---

## Task 17: E2E smoke tests (Playwright)

**Files:**
- Create: `playwright.config.ts`, `tests/e2e/employee.spec.ts`, `tests/e2e/admin.spec.ts`

- [ ] **Step 1: Playwright config**

`playwright.config.ts`:
```ts
import { defineConfig } from "@playwright/test";

export default defineConfig({
  testDir: "tests/e2e",
  use: { baseURL: "http://localhost:3000" },
  webServer: {
    command: "npm run dev",
    url: "http://localhost:3000",
    reuseExistingServer: true,
    timeout: 120_000,
  },
});
```

- [ ] **Step 2: Install browsers**

Run: `npx playwright install chromium`
Expected: Chromium downloaded.

- [ ] **Step 3: Employee smoke test**

`tests/e2e/employee.spec.ts`:
```ts
import { test, expect } from "@playwright/test";

test("employee logs in and sees My Jobs", async ({ page }) => {
  await page.goto("/login");
  await page.fill('input[name="email"]', "manthan@klaus.test");
  await page.fill('input[name="password"]', "employee12345");
  await page.click('button[type="submit"]');
  await expect(page).toHaveURL(/\/jobs/);
  await expect(page.getByRole("heading", { name: "My Jobs" })).toBeVisible();
});
```

- [ ] **Step 4: Admin smoke test**

`tests/e2e/admin.spec.ts`:
```ts
import { test, expect } from "@playwright/test";

test("admin logs in and can open All Jobs", async ({ page }) => {
  await page.goto("/login");
  await page.fill('input[name="email"]', "admin@klaus.test");
  await page.fill('input[name="password"]', "admin12345");
  await page.click('button[type="submit"]');
  await expect(page).toHaveURL(/\/admin|\/jobs/);
  await page.goto("/admin");
  await expect(page.getByRole("heading", { name: "All Jobs" })).toBeVisible();
});
```

- [ ] **Step 5: Run E2E (requires seeded dev DB)**

Run: `npm run db:seed` (if not already), then `npm run test:e2e`
Expected: both specs PASS.

- [ ] **Step 6: Commit** *(skip per user git preference)*

```bash
git add -A && git commit -m "test: add playwright e2e smoke tests"
```

---

## Task 18: Project README

**Files:**
- Create: `README.md`

- [ ] **Step 1: Write the README**

```markdown
# Klaus Field Log

Internal field-operations tool: employees log jobs and view their hours; admins
view all jobs/hours and manage customers, employees, job types, and materials.

## Stack
Next.js (App Router) + TypeScript, Prisma + PostgreSQL, Auth.js, Zod, Vitest, Playwright.

## Setup
1. `npm install`
2. Copy `.env.example` to `.env` and fill `DATABASE_URL`, `DATABASE_URL_TEST`, `AUTH_SECRET`.
3. `npx prisma migrate dev` — create schema
4. `npm run db:seed` — seed admin/employees/lookups
5. `npm run dev` — http://localhost:3000

## Seed logins
- Admin: `admin@klaus.test` / `admin12345`
- Employee: `manthan@klaus.test` / `employee12345`

## Tests
- `npm test` — unit + integration (needs `DATABASE_URL_TEST`)
- `npm run test:e2e` — Playwright smoke tests

## Roadmap (post-MVP)
- Pricing + customer invoices
- QuickBooks integration
- Offline logging/sync
```

- [ ] **Step 2: Commit** *(skip per user git preference)*

```bash
git add -A && git commit -m "docs: add project README"
```

---

## Self-Review Notes

- **Spec coverage:** Architecture (T1–T2, T8–T10), data model (T3–T4), hours logic (T5), validation (T6), seed (T7), auth+roles (T8–T10), employee screens (T15), admin screens incl. full lookup CRUD (T16), error handling via Zod + typed results + transaction (T6, T12, T13), testing unit/integration/e2e (T5, T6, T11, T12, T13, T17). All spec sections map to tasks.
- **Money-free/QuickBooks-ready:** no price fields added; roadmap noted in README + spec.
- **Type consistency:** server actions return `{ ok: true ... } | { ok: false, error }`; `LookupManager`/`JobForm` consume that shape. `createJobLog` returns `{ ok, id }`; admin actions return `{ ok }`. Guards `requireUser`/`requireAdmin` used consistently.
- **Hours model:** derived from job logs only (per spec), with room for adjustments later.
```
