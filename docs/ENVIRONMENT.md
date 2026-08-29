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
