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
