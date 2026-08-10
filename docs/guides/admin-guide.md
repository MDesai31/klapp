# Admin Guide — Klaus Field Log

This guide covers the day-to-day use of the admin site at
`http://<server-ip>:8082/admin` (only reachable on the local network or over
WireGuard).

---

## Logging in

Go to `/admin/login` and enter your username and password. Your session stays
active for 30 days. Click **Log out** in the top navigation when you're done.

---

## Dashboard

**Tab: Dashboard**

Shows every active worker's punch status for today:

| What you see | Means |
|---|---|
| *Not in* (grayed out) | Worker has not punched in today |
| *In since 7:30 AM (2h 15m)* | Worker is currently clocked in |
| *Out at 4:00 PM (8h 30m worked)* | Worker punched out, total time shown |

This page is read-only — use the Timesheet tab to make any corrections.

---

## Timesheet

**Tab: Timesheet**

Shows every punch entry for a pay period (14-day blocks). Use the links at the
top to switch between pay periods.

### Row colors

| Color | Means |
|---|---|
| White | Normal entry |
| Yellow | Worker used the late punch-out link (did not punch out in the field) |
| **Red / bold** | Non-compliant — worker was still clocked in at 9 PM and was automatically punched out |

Non-compliant entries always appear at the top of the table.

### Editing an entry

Click **edit** on any row to open the edit form. You can correct the start time,
end time, or both. The date is part of the time field — changing it to a
different day (e.g. fixing a Saturday entry that should have been Friday) will
move the record to the correct day and pay period automatically. Leave the end
time blank to leave the punch open (worker still clocked in).

### Adding an entry

Click **+ Add entry** above the table. Select the worker, enter the start time,
and optionally an end time. The record will be filed under whichever pay period
the start time falls in.

> Use this when a worker forgot to punch in or you need to enter hours manually.

### Deleting an entry

Click **delete** on any row. You will be asked to confirm before anything is
removed.

---

## Summary

**Tab: Summary**

Shows a grid of hours worked per day for every worker in a pay period, with
pay-period totals and an estimated pre-tax salary (based on each worker's hourly
rate).

- Click a worker's daily hours cell to jump directly to an edit form for that
  punch.
- Switch pay periods with the links at the top.
- Workers with no hourly rate set show `—` in the Salary column. Set their rate
  on the Workers tab.
- The salary figure is informational only — it does not account for taxes,
  overtime, or deductions.

---

## Workers

**Tab: Workers**

Manage who can punch in and what their details are.

### Adding a worker

Fill in the **Add worker** form at the bottom of the page:

- **Name** — how the worker appears on the dashboard and timesheet.
- **PIN** — the numeric code the worker enters on their phone to punch in. Must
  be unique across all workers.
- **Phone** — optional, used for future text notifications.
- **Hourly Rate** — used to calculate the salary estimate on the Summary tab.
- **Language** — controls whether the worker's punch page is shown in Spanish
  or English.

### Editing a worker

Click **Edit** next to any worker to change their name, PIN, phone, rate, or
language.

### Deactivating / reactivating a worker

Click **Deactivate** to hide a worker from the dashboard and prevent them from
punching in. Their historical punch records are kept. Click **Reactivate** to
restore access.

> Deactivating is preferred over deleting — it preserves time history.

To see deactivated workers, click **Show deactivated workers** at the top of the
page.

---

## Invoices

**Tab: Invoices**

Workers submit job invoices from the invoice site. Unreviewed invoices are
highlighted in yellow.

### Reviewing an invoice

Click **View** on any invoice to see the full details:

- Date, worker, customer, house number
- Arrival and departure times
- Number of workers on the job
- Jobs performed and materials used
- Any comments left by the worker

### Submitting an invoice

Once you have reviewed an invoice, click **Submit & mark reviewed**. This
emails a copy to `mylawncut@aol.com` and marks the invoice as reviewed. The
button is disabled after submission — it cannot be submitted twice.

> If the button does nothing or an error appears, check that msmtp is
> configured correctly on the server (see the README).

---

## Customers

**Tab: Customers**

A directory of customers that workers can select when submitting invoices.

### Finding a customer

Use the search box to filter by name, house number, or address.

### Adding a customer

Fill in the **Add customer** form at the bottom of the page. House number is
required (workers select jobs by house number on the invoice form). Address and
phone are optional.

### Editing a customer

Click **Edit** next to any customer to update their details.

### Viewing a customer's history

Click **View** to see a customer's profile and every invoice associated with
them.

---

## Catalog

**Tab: Catalog**

Controls the options workers see when filling out an invoice — the list of jobs
and materials to choose from.

### Job descriptions

A list of standard job types (e.g. *Lawn mowing*, *Hedge trimming*). Workers
pick from this list when submitting an invoice.

- **Add** — type a description and click Add.
- **Delete** — removes the description from future invoices. Existing invoices
  that used it are not affected.

### Materials

A list of materials that may be used on a job, each with an optional unit and
price.

- **Add** — enter the name, unit (e.g. *bag*, *lb*, *gallon*), and price, then
  click Add.
- **Delete** — removes the material from future invoices.

Keep this list tidy — workers see everything in it when filling out an invoice.

---

## Common tasks

### A worker forgot to punch out

1. Go to **Timesheet**.
2. Find the open entry (Out column shows `—`).
3. Click **edit** and enter the correct end time.

### A worker punched in on the wrong day

1. Go to **Timesheet**.
2. Click **edit** on the wrong entry.
3. Change the start (and end) time to the correct date and time.
4. Save — the record will move to the correct day automatically.

### A worker has no hours for a day they worked

1. Go to **Timesheet**.
2. Click **+ Add entry**.
3. Select the worker, enter the correct start and end times, and save.

### Need to remove a duplicate or erroneous entry

1. Go to **Timesheet**.
2. Click **delete** on the row and confirm.

### Check what a worker earned this pay period

1. Go to **Summary**.
2. Find the worker's row — the **Total** and **Salary** columns show their
   hours and estimated pay.
3. Make sure their hourly rate is set correctly on the **Workers** tab if the
   Salary column shows `—`.
