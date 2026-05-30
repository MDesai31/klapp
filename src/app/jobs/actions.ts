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
