# Youpp Adminpanel Backend

Professional multi-tenant SaaS admin API backend built with Go, Gin, MongoDB Atlas, and JWT.

## Environment

Create a `.env` file or export the following variables:

```bash
MONGO_URI="mongodb+srv://<user>:<pass>@cluster.mongodb.net"
MONGO_DB="youpp_admin"
JWT_SECRET="super-secret"
JWT_REFRESH_SECRET="super-refresh-secret"
ACCESS_TTL_MIN="15"
REFRESH_TTL_DAYS="30"
```

## Run

```bash
go run ./cmd/api
```

## API Overview

### Auth

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"password123"}'
```

```bash
curl -X POST http://localhost:8080/api/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refreshToken":"<refresh-token>"}'
```

```bash
curl -X GET http://localhost:8080/api/me \
  -H "Authorization: Bearer <access-token>"
```

### Sites

```bash
curl -X GET http://localhost:8080/api/sites \
  -H "Authorization: Bearer <access-token>"
```

```bash
curl -X POST http://localhost:8080/api/sites \
  -H "Authorization: Bearer <access-token>" \
  -H "Content-Type: application/json" \
  -d '{"name":"My Site","slug":"my-site"}'
```

```bash
curl -X GET http://localhost:8080/api/sites/<site-id> \
  -H "Authorization: Bearer <access-token>"
```

```bash
curl -X PUT http://localhost:8080/api/sites/<site-id>/content \
  -H "Authorization: Bearer <access-token>" \
  -H "Content-Type: application/json" \
  -d '{"content":{"blocks":[],"meta":{}}}'
```

```bash
curl -X POST http://localhost:8080/api/sites/<site-id>/publish \
  -H "Authorization: Bearer <access-token>"
```

```bash
curl -X POST http://localhost:8080/api/sites/<site-id>/unpublish \
  -H "Authorization: Bearer <access-token>"
```

### Public

```bash
curl -X GET http://localhost:8080/s/<slug>
```

## Notes

- Tokens embed userId, orgId, and role to enforce tenant scoping on every query.
- All authenticated queries filter by `orgId` to ensure tenant isolation.
