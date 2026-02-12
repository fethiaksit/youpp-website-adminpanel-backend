# Provisioning Flow (Postman Examples)

## 1) Request Provisioning Code

**POST** `/api/provision/request-code`

Headers:
- `Content-Type: application/json`
- `X-API-Key: {{PROVISION_API_KEY}}`

Body:
```json
{
  "siteName": "Acme Clinic",
  "siteSlug": "acme-clinic"
}
```

## 2) Setup Login

**POST** `/api/auth/setup-login`

Headers:
- `Content-Type: application/json`

Body:
```json
{
  "code": "ABCD-EFGH"
}
```

## 3) Setup Register

**POST** `/api/auth/setup-register`

Headers:
- `Content-Type: application/json`
- `Authorization: Bearer {{setupToken}}`

Body:
```json
{
  "email": "owner@acme.example",
  "password": "StrongPass!234",
  "name": "Acme Owner"
}
```
