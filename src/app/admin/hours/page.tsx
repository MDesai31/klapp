import { db } from "@/lib/db";
import { requireAdmin } from "@/lib/authz";
import { sumHours } from "@/lib/hours";

export default async function AdminHoursPage({
  searchParams,
}: {
  searchParams: Promise<{ year?: string; month?: string }>;
}) {
  await requireAdmin();
  const params = await searchParams;
  const now = new Date();
  const year = Number(params.year ?? now.getFullYear());
  const month = params.month ? Number(params.month) : null;
  const start = new Date(year, month ? month - 1 : 0, 1);
  const end = new Date(year, month ? month : 12, month ? 1 : 1);

  const employees = await db.user.findMany({
    where: { role: "EMPLOYEE" },
    orderBy: { name: "asc" },
  });

  const rows = await Promise.all(
    employees.map(async (e) => {
      const logs = await db.jobLog.findMany({
        where: {
          workers: { some: { userId: e.id } },
          date: { gte: start, lt: end },
        },
        select: { timeArrived: true, timeLeft: true },
      });
      return { name: e.name, total: sumHours(logs) };
    }),
  );

  return (
    <div>
      <h1>All Hours</h1>
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
            <th align="left">Employee</th>
            <th align="left">Total Hours</th>
          </tr>
        </thead>
        <tbody>
          {rows.map((r) => (
            <tr key={r.name}>
              <td>{r.name}</td>
              <td>{r.total.toFixed(2)}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
