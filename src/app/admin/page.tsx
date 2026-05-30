import { db } from "@/lib/db";
import { requireAdmin } from "@/lib/authz";
import { durationHours } from "@/lib/hours";

export default async function AdminPage() {
  await requireAdmin();
  const logs = await db.jobLog.findMany({
    include: {
      customer: true,
      createdBy: true,
      workers: { include: { user: true } },
    },
    orderBy: { date: "desc" },
    take: 200,
  });
  return (
    <div>
      <h1>All Jobs</h1>
      <table style={{ width: "100%", borderCollapse: "collapse" }}>
        <thead>
          <tr>
            <th align="left">Date</th>
            <th align="left">Customer</th>
            <th align="left">Logged by</th>
            <th align="left">Workers</th>
            <th align="left">Hours</th>
          </tr>
        </thead>
        <tbody>
          {logs.map((l) => (
            <tr key={l.id}>
              <td>{l.date.toISOString().slice(0, 10)}</td>
              <td>{l.customer.name}</td>
              <td>{l.createdBy.name}</td>
              <td>{l.workers.map((w) => w.user.name).join(", ")}</td>
              <td>{durationHours(l.timeArrived, l.timeLeft).toFixed(2)}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
