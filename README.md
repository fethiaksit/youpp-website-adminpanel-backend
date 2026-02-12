# Youpp Adminpanel Backend (Single-tenant / Tenantless Runtime Auth)

This backend now runs in a tenantless authorization model.

## Environment

```bash
MONGO_URI="mongodb+srv://<user>:<pass>@cluster.mongodb.net"
MONGO_DB="youpp_admin"
JWT_SECRET="super-secret"
JWT_REFRESH_SECRET="super-refresh-secret"
PROVISION_API_KEY="change-me"
ACCESS_TTL_MIN="15"
REFRESH_TTL_DAYS="30"
SUPERADMIN_EMAIL="admin@example.com"
SUPERADMIN_PASSWORD="change-me"
DEMO_EMAIL="demo@example.com"
DEMO_PASSWORD="change-me"
DEMO_SITE_SLUG="demo-site"
PORT="8080"
```

## Run

```bash
go run ./cmd/api
```

## Seed

```bash
go run ./cmd/api seed
```

Seed is idempotent and uses env vars above.

## Core APIs

- `POST /api/auth/login`
- `POST /api/auth/refresh`
- `GET /api/me`
- `GET /api/sites`
- `POST /api/sites` (superadmin only)
- `GET /api/sites/:id`
- `PUT /api/sites/:id/content`
- `POST /api/sites/:id/publish`
- `POST /api/sites/:id/unpublish`

Admin APIs (superadmin only):

- `GET /api/admin/sites`
- `POST /api/admin/sites`
- `POST /api/admin/sites/:id/grant`
- `GET /api/admin/sites/:id/users`
- `POST /api/admin/users`
- `GET /api/admin/users`

Provisioning:

- `POST /api/provision/bootstrap` (requires `X-API-Key: PROVISION_API_KEY`)

Public site:

- `GET /s/:slug`
