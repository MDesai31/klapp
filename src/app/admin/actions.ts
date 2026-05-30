"use server";

import { db } from "@/lib/db";
import { requireAdmin } from "@/lib/authz";
import {
  lookupSchema,
  userSchema,
  type LookupInput,
  type UserInput,
} from "@/lib/schemas";
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
  await db.user.update({
    where: { id },
    data: { passwordHash: bcrypt.hashSync(password, 10) },
  });
  revalidatePath("/admin/users");
  return { ok: true };
}
