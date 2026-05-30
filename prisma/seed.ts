import "dotenv/config";
import { PrismaClient, Role } from "../src/generated/prisma/client";
import { PrismaPg } from "@prisma/adapter-pg";
import bcrypt from "bcryptjs";

const db = new PrismaClient({
  adapter: new PrismaPg({ connectionString: process.env.DATABASE_URL! }),
});

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

  const customerCount = await db.customer.count();
  if (customerCount === 0) {
    for (const name of ["Customer 1", "Customer 2", "Customer 3"]) {
      await db.customer.create({ data: { name } });
    }
  }

  const jobTypeCount = await db.jobType.count();
  if (jobTypeCount === 0) {
    for (const name of [
      "Lawn Mowing",
      "Weeding",
      "Tree Pruning",
      "Hedge Trimming",
      "Garden Bed Installation",
    ]) {
      await db.jobType.create({ data: { name } });
    }
  }

  const materialCount = await db.material.count();
  if (materialCount === 0) {
    for (const name of ["Topsoil", "Mulch (Brown)", "Mulch (Red)", "Gravel"]) {
      await db.material.create({ data: { name } });
    }
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
