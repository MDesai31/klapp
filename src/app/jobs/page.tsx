import { db } from "@/lib/db";
import { requireUser } from "@/lib/authz";
import { durationHours } from "@/lib/hours";
import Link from "next/link";

export default async function JobsPage() {
  const user = await requireUser();
  const logs = await db.jobLog.findMany({
    where: { workers: { some: { userId: user.id } } },
    include: { customer: true },
    orderBy: { date: "desc" },
  });

  if (logs.length === 0) {
    return (
      <div>
        <h1>My Jobs</h1>
        <p>No jobs logged yet. <Link href="/jobs/new">Log your first job</Link>.</p>
      </div>
    );
  }

  return (
    <div>
      <h1>My Jobs</h1>
      <table style={{ width: "100%", borderCollapse: "collapse" }}>
        <thead>
          <tr>
            <th align="left">Date</th>
            <th align="left">Customer</th>
            <th align="left">Hours</th>
          </tr>
        </thead>
        <tbody>
          {logs.map((l) => (
            <tr key={l.id}>
              <td>{l.date.toISOString().slice(0, 10)}</td>
              <td>{l.customer.name}</td>
              <td>{durationHours(l.timeArrived, l.timeLeft).toFixed(2)}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
