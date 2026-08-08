# SECURITY.md — Protask Security Architecture & Policies

> **Source**: Complete Protask Architecture & Exhaustive Technology Reference specification.

## 1. Authentication & Session Management
- **Clerk Auth**: Frontend identity and token issuance managed via `@clerk/clerk-react ^5.38.1` & `@clerk/themes`.
- **Backend JWT Verification**: Echo v4 backend verifies Clerk JWT Bearer tokens using `github.com/clerk/clerk-sdk-go/v2` in `internal/middleware/auth.go`.
- **Token Lifespan & Security**: Access token lifespan: 15 minutes. Refresh token lifespan: 7 days.
- **Password Hashing**: Delegated securely to Clerk Auth / Argon2id / bcrypt.

## 2. Enterprise SSO & Directory Sync
- **SAML 2.0 / OIDC**: Enterprise identity provider integration (Okta, Azure AD, Ping Identity).
- **SCIM 2.0**: Automated user provisioning and de-provisioning (`/scim/v2/Users`, `/scim/v2/Groups`).

## 3. IP Allowlists & Anti-Abuse
- CIDR IP allowlisting middleware for admin dashboards and API endpoints.
- Automated phishing & malware scanning via Google Safe Browsing / VirusTotal APIs.

<!-- KNOWLEDGE_GRAPH_NAVIGATION_START -->
---
### 🧭 Knowledge Graph & Navigation
| Dimension | Link / Reference |
| :--- | :--- |
| **Parent** | [docs/INDEX.md](file:///home/logan78/Desktop/flux/docs/INDEX.md) |
| **Previous** | [docs/ARCHITECTURE.md](file:///home/logan78/Desktop/flux/docs/ARCHITECTURE.md) |
| **Next** | [ai/PERFORMANCE.md](file:///home/logan78/Desktop/flux/ai/PERFORMANCE.md) |
| **Children** | None |
| **Dependencies** | [docs/INDEX.md](file:///home/logan78/Desktop/flux/docs/INDEX.md) |
| **Related Documents** | [AGENTS.md](file:///home/logan78/Desktop/flux/AGENTS.md) |
| **Navigation Hub** | [docs/INDEX.md](file:///home/logan78/Desktop/flux/docs/INDEX.md) \| [AGENTS.md](file:///home/logan78/Desktop/flux/AGENTS.md) |
---
<!-- KNOWLEDGE_GRAPH_NAVIGATION_END -->
