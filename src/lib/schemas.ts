import { z } from "zod";

export const jobLogSchema = z
  .object({
    customerId: z.string().min(1, "Customer is required"),
    date: z.string().min(1, "Date is required"),
    timeArrived: z.string().min(1, "Arrival time is required"),
    timeLeft: z.string().min(1, "Leave time is required"),
    workerIds: z.array(z.string()).min(1, "Select at least one worker"),
    jobTypeIds: z.array(z.string()).min(1, "Select at least one job type"),
    materialIds: z.array(z.string()),
    notes: z.string().max(2000).optional().or(z.literal("")),
  })
  .refine((d) => new Date(d.timeLeft) > new Date(d.timeArrived), {
    message: "Leave time must be after arrival time",
    path: ["timeLeft"],
  });

export const lookupSchema = z.object({
  name: z.string().min(1, "Name is required").max(100),
});

export const userSchema = z.object({
  email: z.string().email("Valid email required"),
  name: z.string().min(1, "Name is required"),
  role: z.enum(["EMPLOYEE", "ADMIN"]),
  password: z.string().min(8, "Password must be at least 8 characters"),
});

export type JobLogInput = z.infer<typeof jobLogSchema>;
export type LookupInput = z.infer<typeof lookupSchema>;
export type UserInput = z.infer<typeof userSchema>;
