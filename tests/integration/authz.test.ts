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
