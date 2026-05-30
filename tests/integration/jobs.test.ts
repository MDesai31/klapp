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
      timeLeft: "2026-05-30T08:00",
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
        customerId: "x",
        date: "2026-05-30",
        timeArrived: "2026-05-30T08:00",
        timeLeft: "2026-05-30T10:00",
        workerIds: ["x"],
        jobTypeIds: ["x"],
        materialIds: [],
        notes: "",
      }),
    ).rejects.toThrow("UNAUTHENTICATED");
  });
});
