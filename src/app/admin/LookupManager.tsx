"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";

type Item = { id: string; name: string; active: boolean };
type Actions = {
  create: (input: { name: string }) => Promise<{ ok: boolean; error?: string }>;
  setActive: (id: string, active: boolean) => Promise<{ ok: boolean; error?: string }>;
};

export default function LookupManager({
  title,
  items,
  actions,
}: {
  title: string;
  items: Item[];
  actions: Actions;
}) {
  const router = useRouter();
  const [name, setName] = useState("");
  const [error, setError] = useState("");

  return (
    <div>
      <h1>{title}</h1>
      <form
        onSubmit={async (e) => {
          e.preventDefault();
          setError("");
          const res = await actions.create({ name });
          if (res.ok) {
            setName("");
            router.refresh();
          } else {
            setError(res.error ?? "Error");
          }
        }}
        style={{ display: "flex", gap: 8, marginBottom: 12 }}
      >
        <input
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder={`New ${title.toLowerCase()}`}
        />
        <button type="submit">Add</button>
      </form>
      {error && <p style={{ color: "red" }}>{error}</p>}
      <ul style={{ listStyle: "none", padding: 0 }}>
        {items.map((it) => (
          <li
            key={it.id}
            style={{
              display: "flex",
              gap: 8,
              padding: "4px 0",
              opacity: it.active ? 1 : 0.5,
            }}
          >
            <span style={{ flex: 1 }}>
              {it.name}
              {!it.active && " (inactive)"}
            </span>
            <button
              onClick={async () => {
                await actions.setActive(it.id, !it.active);
                router.refresh();
              }}
            >
              {it.active ? "Deactivate" : "Activate"}
            </button>
          </li>
        ))}
      </ul>
    </div>
  );
}
