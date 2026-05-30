import { PrismaClient, Role } from "@/generated/prisma/client";
import { PrismaPg } from "@prisma/adapter-pg";
import bcrypt from "bcryptjs";

export const testDb = new PrismaClient({
  adapter: new PrismaPg({
    connectionString: process.env.DATABASE_URL_TEST,
  }),
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
    data: {
      email: "admin@t.test",
      name: "Admin",
      role: Role.ADMIN,
      passwordHash: bcrypt.hashSync("x", 4),
    },
  });
  const emp = await testDb.user.create({
    data: {
      email: "emp@t.test",
      name: "Emp",
      role: Role.EMPLOYEE,
      passwordHash: bcrypt.hashSync("x", 4),
    },
  });
  const customer = await testDb.customer.create({ data: { name: "Cust" } });
  const jobType = await testDb.jobType.create({ data: { name: "Mowing" } });
  const material = await testDb.material.create({ data: { name: "Mulch" } });
  return { admin, emp, customer, jobType, material };
}
