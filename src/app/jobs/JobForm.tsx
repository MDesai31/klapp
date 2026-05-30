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
      <label>
        Workers
        <select name="workerIds" multiple required>
          {employees.map((e) => (
            <option key={e.id} value={e.id}>{e.name}</option>
          ))}
        </select>
      </label>
      <label>
        Job types
        <select name="jobTypeIds" multiple required>
          {jobTypes.map((j) => (
            <option key={j.id} value={j.id}>{j.name}</option>
          ))}
        </select>
      </label>
      <label>
        Materials
        <select name="materialIds" multiple>
          {materials.map((m) => (
            <option key={m.id} value={m.id}>{m.name}</option>
          ))}
        </select>
      </label>
      <label>Notes <textarea name="notes" rows={3} /></label>
      <button type="submit" disabled={saving}>
        {saving ? "Saving..." : "Submit Job"}
      </button>
    </form>
  );
}
