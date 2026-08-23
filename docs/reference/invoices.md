# Invoice System

## Overview

Three-part addition to the klapp binary suite:

1. **`cmd/invoice`** — a new worker-facing binary on port 8083 (VPN/LAN only) where workers submit job invoices via a PIN-authenticated form.
2. **Admin invoice tab** — admin site (`cmd/admin`, port 8082) gains Invoices, Customers, and Catalog tabs for reviewing and approving submitted invoices.
3. **4 new DB migrations** (0005–0008) adding the customers, job_descriptions, materials, invoices, invoice_jobs, and invoice_materials_used tables.

---

## New binary: `cmd/invoice` (port 8083)

Run alongside the existing binary:

```
go run ./cmd/invoice
# or with custom flags:
go run ./cmd/invoice -addr=:8083 -dsn="file:db/klapp.db?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)"
```

Runs migrations on startup (same `db/migrations/` embed as `cmd/punch`/`cmd/admin`; goose is idempotent so all binaries can run simultaneously).

### Flow

1. Worker hits `/` → enters PIN → identified → redirected to `/form`
2. Session (scs, 2-hour lifetime) holds the worker ID between PIN and form pages
3. Worker fills out and submits the form → `/success`

### Invoice form fields

| Field | Required | Notes |
|---|---|---|
| Date | Yes | Date picker, defaults to today |
| House number | Yes | Numeric text; triggers live customer lookup |
| Customer name | No | Auto-populated from house number; worker can override or leave blank |
| Number of workers | Yes | Integer ≥ 1 |
| Time arrived | Yes | Time picker |
| Time left | Yes | Time picker |
| Job descriptions | No | 3 text boxes by default; "+ Add" adds more; autocomplete from catalog |
| Materials used | No | Same pattern as job descriptions |
| Comments | No | Large free-text box |

### Language

Bilingual: if the authenticated worker's `language` field is `"english"`, the entire form renders in English. Otherwise Spanish. PIN error shown in both languages since the worker is unknown at that point.

### Customer lookup (house number)

As the worker types the house number, a debounced fetch hits `/api/customers?house_number=<value>` (JSON). If there are matches, a dropdown appears with:
- **"No sé" / "I don't know"** at the top (leaves customer name blank)
- One entry per matching customer (name + address)

Selecting an entry fills the customer name field and a hidden `customer_id`. If no match, the worker types the name manually and it's stored as free text on the invoice.

### Autocomplete (jobs and materials)

As the worker types in a job or material box, suggestions are fetched from `/api/jobs?q=` or `/api/materials?q=`. Clicking a suggestion fills that box. The catalog is managed by the admin under the Catalog tab.

---

## Database schema (new tables)

```sql
customers        (id, name, phone, house_number, address)
job_descriptions (id, description UNIQUE)
materials        (id, name UNIQUE, unit, price)

invoices (
    id, submitted_by → workers.id,
    date, house_number, customer_name,
    customer_id → customers.id (nullable),
    time_arrived, time_left, no_of_workers,
    comments, reviewed, created_at
)
invoice_jobs           (id, invoice_id → invoices.id, description)
invoice_materials_used (id, invoice_id → invoices.id, material)
```

---

## Admin additions (`cmd/admin`)

Three new tabs in the admin nav:

### Invoices tab (`/admin/invoices`)

- Paginated list (25 per page), newest first
- Unreviewed rows highlighted yellow
- Columns: ID, date, house number, customer, submitted by, status, View link

**Invoice view** (`/admin/invoices/{id}`):

- Read-only display of all fields including job list and material list
- **Submit & mark reviewed** button → sends email via msmtp to `mylawncut@aol.com` + sets `reviewed = TRUE`
- Button grays out (disabled) once already reviewed
- QuickBooks integration: planned — see `docs/design/quickbooks_plan.md`

### Customers tab (`/admin/customers`)

- Searchable list (by name, house number, or address)
- Add new customer form (name required, house number required)
- Click a customer → see their full profile + all linked invoices
- Edit customer from the profile page

### Catalog tab (`/admin/catalog`)

Manages the autocomplete lists used on the invoice form:

- **Job descriptions** — add / delete free-text descriptions
- **Materials** — add / delete with optional unit and price fields

---

## Email delivery (msmtp)

The Submit button calls `msmtp --account=default mylawncut@aol.com`. msmtp must be installed and configured on the server before this works — see the msmtp setup section in `README.md`. The email failure is logged but does not block marking the invoice as reviewed.

---

## Models

| File | Types |
|---|---|
| `internal/models/customers.go` | `Customer`, `CustomerModel` — List, Search, Get, GetByHouseNumber, Create, Update |
| `internal/models/catalog.go` | `JobDescription`, `Material`, `CatalogModel` — ListJobs/Materials, SearchJobs/Materials, Create/Delete for each |
| `internal/models/invoices.go` | `Invoice`, `InvoiceModel` — Create, List (paginated), Get (with jobs+materials), SetReviewed, ListByCustomer |

---

## Not yet built

- **QuickBooks integration** — see `docs/design/quickbooks_plan.md`
- `cmd/invoice` is not yet in the systemd unit — add a second `ExecStart` line or a separate service file pointing to the invoice binary
