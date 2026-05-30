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
