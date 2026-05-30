import { db } from "@/lib/db";
import { requireAdmin } from "@/lib/authz";
import { createUser, setUserActive } from "@/app/admin/actions";

export default async function UsersPage() {
  await requireAdmin();
  const users = await db.user.findMany({ orderBy: { name: "asc" } });
  return (
    <div>
      <h1>Employees</h1>
      <form
        action={async (fd: FormData) => {
          "use server";
          await createUser({
            email: String(fd.get("email")),
            name: String(fd.get("name")),
            role: String(fd.get("role")) as "EMPLOYEE" | "ADMIN",
            password: String(fd.get("password")),
          });
        }}
        style={{ display: "flex", gap: 8, marginBottom: 12, flexWrap: "wrap" }}
      >
        <input name="name" placeholder="Name" required />
        <input name="email" type="email" placeholder="Email" required />
        <select name="role" defaultValue="EMPLOYEE">
          <option value="EMPLOYEE">Employee</option>
          <option value="ADMIN">Admin</option>
        </select>
        <input name="password" type="password" placeholder="Temp password (min 8)" required />
        <button type="submit">Add</button>
      </form>
      <ul style={{ listStyle: "none", padding: 0 }}>
        {users.map((u) => (
          <li
            key={u.id}
            style={{ display: "flex", gap: 8, padding: "4px 0", opacity: u.active ? 1 : 0.5 }}
          >
            <span style={{ flex: 1 }}>
              {u.name} — {u.email} ({u.role}){!u.active && " (inactive)"}
            </span>
            <form
              action={async () => {
                "use server";
                await setUserActive(u.id, !u.active);
              }}
            >
              <button type="submit">{u.active ? "Deactivate" : "Activate"}</button>
            </form>
          </li>
        ))}
      </ul>
    </div>
  );
}
