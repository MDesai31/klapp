---
name: invoice-review-flow
description: invoice lifecycle — worker submits via cmd/invoice, admin reviews, email via msmtp; reviewed even on email failure
type: domain
---

Invoices flow: worker submits on the `cmd/invoice` site (`:8083`, VPN/LAN-only, PIN + 2h scs
session) → appears unreviewed (yellow) on the admin Invoices tab → admin clicks
**Submit & mark reviewed**, which emails a copy to `mylawncut@aol.com` via `msmtp` and sets
`reviewed = TRUE` (button disabled after).

Failure semantics were deliberately settled after flip-flopping: the invoice **is marked
reviewed even when the email fails**, but the admin is told so via a flash (commits f9188594
then e4c8ebe6). msmtp must be configured on the server first (`README.md` § msmtp setup).

QuickBooks push on review is planned but unbuilt — see [[quickbooks-plan]].
