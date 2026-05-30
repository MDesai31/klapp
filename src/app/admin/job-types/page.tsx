import { db } from "@/lib/db";
import { requireAdmin } from "@/lib/authz";
import LookupManager from "@/app/admin/LookupManager";
import { createJobType, setJobTypeActive } from "@/app/admin/actions";

export default async function JobTypesPage() {
  await requireAdmin();
  const items = await db.jobType.findMany({ orderBy: { name: "asc" } });
  return (
    <LookupManager
      title="Job Types"
      items={items.map((i) => ({ id: i.id, name: i.name, active: i.active }))}
      actions={{ create: createJobType, setActive: setJobTypeActive }}
    />
  );
}
