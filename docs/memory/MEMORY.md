# Memory — repo knowledge base

Canonical, shared, version-controlled reference knowledge for this repo. One fact per file,
indexed below. Schema and the CLAUDE.md-vs-memory boundary rule live in the workspace-os
plugin's `conventions/memory.md`. Add facts with `/ingest`; check integrity with
`/memory-lint`.

## domain
- [worker-auth-model](worker-auth-model.md) — PIN auth has no worker picker; why cookie counters and IP buddy-punch detection were rejected
- [time-model-and-compliance](time-model-and-compliance.md) — pay-period/day bucketing rules; late vs non-compliant punch distinction
- [bilingual-ui](bilingual-ui.md) — per-worker language field (default spanish) drives punch and invoice form language
- [invoice-review-flow](invoice-review-flow.md) — invoice lifecycle; marked reviewed even when the email fails
- [nextjs-era-history](nextjs-era-history.md) — stale-prior guard: klapp was Next.js+Prisma until 2026-08-09; JobLog/no-billing/Postgres facts are dead

## convention
_No facts yet._

## reference
- [quickbooks-plan](quickbooks-plan.md) — QBO integration planned, not built; pointer to the plan and its 3 open decisions
