# klapp — Feature Ideas

Brainstormed 2026-05-30. Not prioritized or scoped yet — pick any to spec out.

---

## Field Worker UX

| Idea | What it does |
|---|---|
| **Edit / delete jobs** | Employees can correct mistakes on their own logs; admin can edit any log |
| **Job detail page** | Click into a job to see all fields (workers, materials, job types, notes) — currently no detail view exists |
| **Clock in / out** | Instead of manually typing arrival/departure, tap in and out in real time |
| **Standalone punch in / out** | Punch in/out for a shift without tying it to a specific job — track raw on-the-clock time independently |
| **Photo attachments** | Upload before/after photos per job (stored in S3/Cloudflare R2) |
| **Offline PWA** | Log jobs without cell signal; sync when back online |
| **Search & filter on My Jobs** | Filter by customer, date range, job type |

---

## Admin / Reporting

| Idea | What it does |
|---|---|
| **Per-customer job history** | Admin sees all jobs for a specific customer with totals |
| **Per-employee job history** | Admin drills into one employee's full log, not just hour totals |
| **Material usage report** | Which materials were used most, by whom, for which customers |
| **CSV / PDF export** | Export hours or job logs for payroll, billing, or records |
| **Dashboard** | Landing page with KPIs — total hours this week, jobs logged today, active employees |
| **Audit log** | Track who created/edited/deleted what and when |
| **Job assignment from admin** | Admin schedules a job and assigns workers, rather than employees self-logging |

---

## Business / Billing

| Idea | What it does |
|---|---|
| **Pricing per job type** | Attach a rate to each job type; auto-calculate job cost |
| **Material quantities + unit cost** | Track how many units of each material were used and at what cost |
| **Customer invoices** | Generate an invoice from a date range of jobs for a customer |
| **Invoice PDF export** | Render invoices as printable/emailable PDFs |
| **QuickBooks sync** | Push invoices or hours to QuickBooks (OAuth integration) |

---

## UX / Polish

| Idea | What it does |
|---|---|
| **Mobile-responsive layout** | Better styling for phone-sized screens |
| **Pagination** | Jobs list gets unwieldy without it |
| **Inline admin tables** | Edit customer/user names directly in the table instead of a modal |
| **Password change** | Let employees change their own password |
| **Email notifications** | Notify admin when a job is logged, or employee when assigned |
| **Text notifications for login/logout** | Send an SMS when a user logs in or out — visibility into who's active and when |
