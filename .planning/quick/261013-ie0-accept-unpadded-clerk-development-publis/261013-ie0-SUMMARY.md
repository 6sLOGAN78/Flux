---
type: quick
status: complete
completed: 2026-10-13
commit: 3a8549b
---

# Clerk publishable-key encoding correction

The provider prerequisite validator accepts Clerk's canonical unpadded base64
encoding as well as its existing padded form. It rejects excessive padding,
malformed encodings and issuer mismatch. Credentials and recipient validation
remain enforced.

The new regression first failed on the unpadded form, then passed after the fix.
All 44 frontend unit tests, frontend typecheck, formatting and lint passed.

This closes a configuration compatibility bug. It does not complete plan 35 or
provide evidence for real inbox verification, OAuth completion or invitation delivery.
