---
name: bilingual-ui
description: per-worker language field (default spanish) drives punch and invoice form language
type: domain
---

Worker-facing pages are bilingual per worker, not per browser: `workers.language`
(migration 0004, `TEXT NOT NULL DEFAULT 'spanish'`) controls whether the punch page and the
invoice form render in Spanish or English. Admins set it on the Workers tab. The PIN entry
error is shown in **both** languages because the worker is unidentified at that point.

This closed the Next.js-era "Spanish language support" idea wholesale — see
[[nextjs-era-history]].
