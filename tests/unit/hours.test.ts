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
