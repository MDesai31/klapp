import { db } from "@/lib/db";
import { requireAdmin } from "@/lib/authz";
import LookupManager from "@/app/admin/LookupManager";
import { createCustomer, setCustomerActive } from "@/app/admin/actions";

export default async function CustomersPage() {
  await requireAdmin();
  const items = await db.customer.findMany({ orderBy: { name: "asc" } });
  return (
    <LookupManager
      title="Customers"
      items={items.map((i) => ({ id: i.id, name: i.name, active: i.active }))}
      actions={{ create: createCustomer, setActive: setCustomerActive }}
    />
  );
}
