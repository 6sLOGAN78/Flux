# Clean-install dependency compatibility repair

Plan 22's fresh full-history clone and frozen Bun install exposed a dependency incompatibility at full-check stage 25: OpenAPI tests passed 2/10 because both installed Zod OpenAPI adapters expected the CommonJS default export removed by ts-deepmerge v8. The existing checkout had retained nested v6.2.1 resolution, masking the error despite the authoritative v8 override.

Retain exact ts-deepmerge 8.0.0 and its registry integrity. Add a Bun-managed patch that exports the existing secure `merge` function as `default` in the CommonJS entrypoint. Its implementation, unsafe-key set, options and named export are unchanged. No new dependency, downgrade, advisory exception or scanner suppression is introduced. Both existing adapter versions remain intact.

`package.json`, `bun.lock` and `patches/ts-deepmerge@8.0.0.patch` persist the patch for frozen installations. Bun preparation ran before editing, keeping its global package cache intact. A new regression in the existing OpenAPI suite resolves the merge dependency from both adapters, proves function identity, preserves nested merging, rejects dangerous keys and checks that ordinary string conversion and Object.prototype remain safe.

Verification after forced frozen installation refreshed the stale checkout: OpenAPI 11/11 tests, typecheck and Biome passed. Existing Node and Bun contract-loading and Node generation tests passed. Canonical generation and real dependency scanning were run separately; final clean-install/full gate evidence belongs in 01-22-SUMMARY.md rather than being inferred from these scoped checks.

Primary references: [upstream ts-deepmerge API](https://github.com/voodoocreation/ts-deepmerge), [reviewed advisory requiring v8](https://github.com/advisories/GHSA-87mf-gv2c-c62c), [Bun's persistent patch workflow](https://bun.com/docs/pm/cli/patch). The local installed registry sources confirm the adapters' default calls and v8's named function.
