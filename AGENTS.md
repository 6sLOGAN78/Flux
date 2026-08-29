# Flux AI Engineering Rules

## Source of Truth
The repository and maintained documentation are authoritative. 
Do not trust previous AI conversation context.

## Before Changing Code
Always:
1. Inspect relevant files.
2. Read relevant documentation.
3. Identify dependencies.
4. Identify affected features.
5. Identify API/data contracts.
6. Understand existing behavior.
7. Make the smallest safe change.

## After Changing Code
Always:
1. Run appropriate tests/checks.
2. Verify the change.
3. Update affected documentation.
4. Update feature status.
5. Update bug status if applicable.
6. Update changelog.
7. Record architectural decisions when necessary.

## Never
* blindly rewrite working code
* silently change API contracts
* introduce unnecessary dependencies
* delete functionality without checking dependencies
* mark code as fixed without verification
* hide known failures
* assume old conversation context is correct
* perform unrelated refactors
* change architecture silently
