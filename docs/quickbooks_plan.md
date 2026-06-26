# QuickBooks Integration Plan

## Goal

When an admin clicks "Submit & mark reviewed" on an invoice, it should also create an invoice in QuickBooks Online (QBO), in addition to sending the email to `mylawncut@aol.com`.

---

## Approach: QuickBooks Online API (OAuth 2.0)

QBO provides a REST API. The integration requires:

1. A **Intuit Developer account** and an app registered at https://developer.intuit.com
2. **OAuth 2.0** authorization (one-time flow) to get a refresh token
3. Storing the refresh token on the server so the app can request access tokens without user interaction

---

## Step-by-step setup

### 1. Create a developer app

- Go to https://developer.intuit.com → My Apps → Create an app → QuickBooks Online and Payments
- Set the redirect URI to something reachable during setup, e.g. `http://localhost:9090/callback` (only needed once for the initial auth)
- Note the **Client ID** and **Client Secret**

### 2. Authorize once (get a refresh token)

QBO uses OAuth 2.0 with refresh tokens that last 100 days (access tokens last 1 hour). Run the one-time authorization flow:

```bash
# A minimal CLI tool or a temporary HTTP server can do this.
# Simplest: use the Intuit OAuth playground at
# https://developer.intuit.com/app/developer/playground
# to get an initial refresh token for the production company.
```

After authorization you will have:
- `refresh_token` (save this — it's long-lived)
- `realm_id` (your QBO company ID, visible in the URL when logged into QBO)

Store these in environment variables or a config file:
```
QBO_CLIENT_ID=...
QBO_CLIENT_SECRET=...
QBO_REFRESH_TOKEN=...
QBO_REALM_ID=...
```

### 3. Token refresh logic

Before each API call, exchange the refresh token for a fresh access token:

```
POST https://oauth.platform.intuit.com/oauth2/v1/tokens/bearer
Content-Type: application/x-www-form-urlencoded
Authorization: Basic base64(client_id:client_secret)

grant_type=refresh_token&refresh_token=<stored_token>
```

Response includes a new `access_token` (1 hour) and a new `refresh_token` (100 days). **Always save the new refresh token** — the old one is immediately invalidated.

Recommend storing the current refresh token in a small DB table or a file on disk so it survives restarts.

### 4. Create an invoice via the API

```
POST https://quickbooks.api.intuit.com/v3/company/{realmId}/invoice
Authorization: Bearer <access_token>
Accept: application/json
Content-Type: application/json

{
  "Line": [
    {
      "DetailType": "SalesItemLineDetail",
      "Amount": <total>,
      "SalesItemLineDetail": {
        "ItemRef": { "value": "<item_id>", "name": "<service_name>" }
      },
      "Description": "<job description text>"
    }
  ],
  "CustomerRef": { "value": "<qbo_customer_id>" },
  "TxnDate": "<YYYY-MM-DD>",
  "DocNumber": "<invoice_id>"
}
```

Key decisions to make before implementing:

- **Customers**: QBO invoices require a `CustomerRef`. You will either need to maintain a mapping between klapp customer IDs and QBO customer IDs, or look them up by name at submission time and create them if missing.
- **Items/Services**: QBO line items reference a `Service` item. Decide whether each job description maps to a specific QBO service item or all invoices use a single generic "Lawn Care" service.
- **Amount**: The API needs a dollar amount. Currently the invoice form captures labor time and materials but not a total. Decide whether the total is computed from `hourly_rate × hours × workers + material costs` or entered manually by the admin before submission.

### 5. Sandbox testing

Before going live:
- Use the **sandbox environment** (`sandbox-quickbooks.api.intuit.com`) with a sandbox company
- The Intuit developer dashboard provides a sandbox company automatically
- All testing should happen in sandbox first; sandbox and production use separate OAuth credentials

---

## Implementation plan (when ready)

1. Add `QBO_*` env vars to the systemd unit
2. Write `internal/qbo/client.go` — token refresh + `CreateInvoice(inv *models.Invoice) error`
3. Add a `qbo_customer_id TEXT` column to the `customers` table (migration 0009)
4. Add a `qbo_invoice_id TEXT` column to the `invoices` table (migration 0010) — set after successful push
5. In `adminInvoiceSubmit` (after email send): call `qbo.CreateInvoice(&inv)`, log errors, store the returned QBO invoice ID
6. Show the QBO invoice ID on the admin invoice view page as a link to the QBO invoice

---

## Useful references

- QBO REST API explorer: https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/invoice
- OAuth 2.0 guide: https://developer.intuit.com/app/developer/qbo/docs/develop/authentication-and-authorization/oauth-2.0
- Go SDK (optional, not required): https://github.com/rwestlund/quickbooks-go or roll raw HTTP calls (simpler for this use case)
