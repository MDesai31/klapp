# Project Tracking

The orientation index for this repo's ongoing work, maintained by the `workspace-os` skills
(`/project-log`, `/project-plan`). One file per purpose.

## Files

| File | Holds |
|---|---|
| `action-items.md` | open action records |
| `ideas.md` | future intents (not started) |
| `decisions-log.md` | decisions (append-only) |
| `resolved.md` | completed records (the record) |

## Workstreams

Every record is tagged with one workstream from this list:

- time-tracking
- invoicing
- admin
- security
- ops

## Conventions

Records use date-slug IDs (`A-YYYYMMDD-slug`, `D-YYYYMMDD-slug`), assigned at creation and never
renumbered. Completing an action moves it to `resolved.md`. The append-heavy files are
`merge=union` (see `.gitattributes`). Full schema: the `workspace-os` plugin's
`conventions/project-tracking.md`.
