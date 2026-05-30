import { db } from "@/lib/db";
import { requireUser } from "@/lib/authz";
import JobForm from "@/app/jobs/JobForm";

export default async function NewJobPage() {
  await requireUser();
  const [customers, employees, jobTypes, materials] = await Promise.all([
    db.customer.findMany({ where: { active: true }, orderBy: { name: "asc" } }),
    db.user.findMany({ where: { active: true }, orderBy: { name: "asc" } }),
    db.jobType.findMany({ where: { active: true }, orderBy: { name: "asc" } }),
    db.material.findMany({ where: { active: true }, orderBy: { name: "asc" } }),
  ]);
  return (
    <div>
      <h1>Log a Job</h1>
      <JobForm
        customers={customers.map((c) => ({ id: c.id, name: c.name }))}
        employees={employees.map((u) => ({ id: u.id, name: u.name }))}
        jobTypes={jobTypes.map((j) => ({ id: j.id, name: j.name }))}
        materials={materials.map((m) => ({ id: m.id, name: m.name }))}
      />
    </div>
  );
}
