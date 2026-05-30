import { db } from "@/lib/db";
import { requireAdmin } from "@/lib/authz";
import LookupManager from "@/app/admin/LookupManager";
import { createMaterial, setMaterialActive } from "@/app/admin/actions";

export default async function MaterialsPage() {
  await requireAdmin();
  const items = await db.material.findMany({ orderBy: { name: "asc" } });
  return (
    <LookupManager
      title="Materials"
      items={items.map((i) => ({ id: i.id, name: i.name, active: i.active }))}
      actions={{ create: createMaterial, setActive: setMaterialActive }}
    />
  );
}
