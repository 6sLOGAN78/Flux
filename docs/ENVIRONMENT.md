# Environment

## Backend (`apps/backend/.env`)
- `DATABASE_URL`: PostgreSQL connection string (Required)
- `REDIS_URL`: Redis connection string
- `CLICKHOUSE_URL`: ClickHouse connection string
- `JWT_SECRET`: Secret for auth
- `PORT`: HTTP Port (default 8080)

## Frontend (`apps/frontend/.env`)
- `VITE_API_URL`: Backend URL (default `http://localhost:8080`)
- `VITE_CLERK_PUBLISHABLE_KEY`: Clerk public key for the frontend
- `CLERK_SECRET_KEY`: Clerk secret key for the backend to verify JWTs

## Clerk Configurations

### Enabling Organizations (B2B Multi-tenancy)
By default, some Clerk instances do not have Organizations enabled. The API dynamically adapts to fallback "Personal Workspaces" if no active organization is present in the token. 

To enable the target production architecture:
1. Go to your [Clerk Dashboard](https://dashboard.clerk.com).
2. Select your application.
3. Navigate to **Organization Settings**.
4. Enable Organizations.
5. Create default roles (e.g., `org:owner`, `org:admin`, `org:member`).
6. Update your `<OrganizationSwitcher />` and `<CreateOrganization />` settings to allow users to create and switch orgs.


### Redis / Analytics
- `REDIS_URL`: (e.g., `localhost:6379`) Connection URL to the Redis cluster used for rate-limiting, redirect caching, and stream queueing.
- `ANALYTICS_REDIS_STREAM`: (e.g., `analytics:events`) Redis Stream name where the `RedisAnalyticsPublisher` buffers background click events.
- `CLICKHOUSE_URL`: (e.g., `localhost:9000`) Connection URL to the ClickHouse database cluster serving the OLAP backend for analytics.
