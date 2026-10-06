# Deferred items

- Found in plan 01-01 real PostgreSQL helper smoke: `apps/backend/internal/database/migrations/001_setup.sql` contains no forward SQL. Tern rejects it with `loading database migrations: no sql in forward migration step`; `SetupTestDB` intentionally continues surfacing that error. The plan owning migrations must make the scaffold executable and verify migrating fixtures. `SetupTestPostgres` provides raw dependency integration infrastructure independently of application schema.
- The email workspace has no build script, so plan 01-01 verified it with the frozen root `tsc --project packages/emails/tsconfig.json` plus its real export command. Root quality-script coverage belongs to its assigned later plan.
