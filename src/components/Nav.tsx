import Link from "next/link";
import { auth, signOut } from "@/lib/auth";

export default async function Nav() {
  const session = await auth();
  if (!session?.user) return null;
  const role = (session.user as { role?: string }).role;
  return (
    <nav style={{ display: "flex", gap: 16, padding: 12, borderBottom: "1px solid #ddd" }}>
      <Link href="/jobs">My Jobs</Link>
      <Link href="/jobs/new">Log a Job</Link>
      <Link href="/hours">My Hours</Link>
      {role === "ADMIN" && (
        <>
          <Link href="/admin">All Jobs</Link>
          <Link href="/admin/hours">All Hours</Link>
          <Link href="/admin/customers">Customers</Link>
          <Link href="/admin/users">Employees</Link>
          <Link href="/admin/job-types">Job Types</Link>
          <Link href="/admin/materials">Materials</Link>
        </>
      )}
      <form
        action={async () => {
          "use server";
          await signOut({ redirectTo: "/login" });
        }}
        style={{ marginLeft: "auto" }}
      >
        <button type="submit">Logout</button>
      </form>
    </nav>
  );
}
