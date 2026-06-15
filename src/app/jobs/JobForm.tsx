"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { createJobLog } from "@/app/jobs/actions";

type Opt = { id: string; name: string };

export default function JobForm({
  customers,
  employees,
  jobTypes,
  materials,
}: {
  customers: Opt[];
  employees: Opt[];
  jobTypes: Opt[];
  materials: Opt[];
}) {
  const router = useRouter();
  const [error, setError] = useState("");
  const [saving, setSaving] = useState(false);

  return (
    <form
      onSubmit={async (e) => {
        e.preventDefault();
        setError("");
        setSaving(true);
        const fd = new FormData(e.currentTarget);
        const multi = (k: string) => fd.getAll(k).map(String);
        const res = await createJobLog({
          customerId: String(fd.get("customerId") || ""),
          date: String(fd.get("date") || ""),
          timeArrived: String(fd.get("timeArrived") || ""),
          timeLeft: String(fd.get("timeLeft") || ""),
          workerIds: multi("workerIds"),
          jobTypeIds: multi("jobTypeIds"),
          materialIds: multi("materialIds"),
          notes: String(fd.get("notes") || ""),
        });
        setSaving(false);
        if (res.ok) router.push("/jobs");
        else setError(res.error);
      }}
      style={{ display: "flex", flexDirection: "column", gap: 12 }}
    >
      {error && <p style={{ color: "red" }}>{error}</p>}
      <label>
        Customer
        <select name="customerId" required defaultValue="">
          <option value="" disabled>Select customer</option>
          {customers.map((c) => (
            <option key={c.id} value={c.id}>{c.name}</option>
          ))}
        </select>
      </label>
      <label>Date <input name="date" type="date" required /></label>
      <label>Time arrived <input name="timeArrived" type="datetime-local" required /></label>
      <label>Time left <input name="timeLeft" type="datetime-local" required /></label>
      <fieldset style={{ border: "1px solid #ccc", borderRadius: 4, padding: "8px 12px" }}>
        <legend>Workers <span style={{ color: "red" }}>*</span></legend>
        <div style={{ display: "flex", flexDirection: "column", gap: 6, marginTop: 4 }}>
          {employees.map((emp) => (
            <label key={emp.id} style={{ display: "flex", alignItems: "center", gap: 8, cursor: "pointer" }}>
              <input type="checkbox" name="workerIds" value={emp.id} />
              {emp.name}
            </label>
          ))}
        </div>
      </fieldset>
      <fieldset style={{ border: "1px solid #ccc", borderRadius: 4, padding: "8px 12px" }}>
        <legend>Job types <span style={{ color: "red" }}>*</span></legend>
        <div style={{ display: "flex", flexDirection: "column", gap: 6, marginTop: 4 }}>
          {jobTypes.map((j) => (
            <label key={j.id} style={{ display: "flex", alignItems: "center", gap: 8, cursor: "pointer" }}>
              <input type="checkbox" name="jobTypeIds" value={j.id} />
              {j.name}
            </label>
          ))}
        </div>
      </fieldset>
      <fieldset style={{ border: "1px solid #ccc", borderRadius: 4, padding: "8px 12px" }}>
        <legend>Materials</legend>
        <div style={{ display: "flex", flexDirection: "column", gap: 6, marginTop: 4 }}>
          {materials.map((m) => (
            <label key={m.id} style={{ display: "flex", alignItems: "center", gap: 8, cursor: "pointer" }}>
              <input type="checkbox" name="materialIds" value={m.id} />
              {m.name}
            </label>
          ))}
        </div>
      </fieldset>
      <label>Notes <textarea name="notes" rows={3} /></label>
      <button type="submit" disabled={saving}>
        {saving ? "Saving..." : "Submit Job"}
      </button>
    </form>
  );
}
