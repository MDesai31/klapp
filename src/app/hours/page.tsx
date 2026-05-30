import { db } from "@/lib/db";
import { requireUser } from "@/lib/authz";
import { durationHours, sumHours } from "@/lib/hours";

export default async function HoursPage({
  searchParams,
}: {
  searchParams: Promise<{ year?: string; month?: string }>;
}) {
  const user = await requireUser();
  const params = await searchParams;
  const now = new Date();
  const year = Number(params.year ?? now.getFullYear());
  const month = params.month ? Number(params.month) : null;

  const start = new Date(year, month ? month - 1 : 0, 1);
  const end = new Date(year, month ? month : 12, month ? 1 : 1);

  const logs = await db.jobLog.findMany({
    where: {
      workers: { some: { userId: user.id } },
      date: { gte: start, lt: end },
    },
    include: { customer: true },
    orderBy: { date: "asc" },
  });

  const total = sumHours(logs);

  return (
    <div>
      <h1>My Hours</h1>
      <form method="get" style={{ display: "flex", gap: 8, marginBottom: 12 }}>
        <input name="year" type="number" defaultValue={year} />
        <input
          name="month"
          type="number"
          min={1}
          max={12}
          defaultValue={month ?? ""}
          placeholder="month (optional)"
        />
        <button type="submit">Filter</button>
      </form>
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
      <h3 style={{ textAlign: "right" }}>Total: {total.toFixed(2)} h</h3>
    </div>
  );
}
