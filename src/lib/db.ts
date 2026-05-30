import { PrismaClient } from "@/generated/prisma/client";
import { PrismaPg } from "@prisma/adapter-pg";

const globalForPrisma = globalThis as unknown as { prisma?: PrismaClient };

const url =
  process.env.NODE_ENV === "test"
    ? process.env.DATABASE_URL_TEST!
    : process.env.DATABASE_URL!;

export const db =
  globalForPrisma.prisma ??
  new PrismaClient({
    adapter: new PrismaPg({ connectionString: url }),
  });

if (process.env.NODE_ENV !== "production") globalForPrisma.prisma = db;
