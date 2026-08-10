---
name: quickbooks-plan
description: QBO integration is planned, not built — pointer to the plan and its 3 open decisions
type: reference
---

QuickBooks Online integration (push an invoice to QBO on "Submit & mark reviewed") is
**planned, not built**. The full plan — OAuth 2.0 refresh-token flow, token rotation gotcha
(each refresh invalidates the old token), sandbox-first testing, and the implementation steps
(`internal/qbo/client.go`, `qbo_customer_id`/`qbo_invoice_id` columns) — lives in
`docs/quickbooks_plan.md`.

Three decisions must be made before implementing: how klapp customers map to QBO `CustomerRef`s,
whether job descriptions map to per-service QBO items or one generic service, and where the
dollar amount comes from (the form captures time and materials, not a total).
