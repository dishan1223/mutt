# Mutt API

Base: `http://localhost:8080/api/v1`

---

## Health

```
GET /ping                  → {"message":"pong"}
GET /api/v1/ping           → {"message":"pong"}
```

## Auth

| Method | Path | Auth | What |
|--------|------|------|------|
| POST | `/auth/signup` | — | Create account |
| POST | `/auth/login` | — | Login, get cookies |
| POST | `/auth/refresh` | — | Refresh tokens |
| POST | `/auth/logout` | JWT | Logout |
| GET | `/auth/me` | JWT | Current user info |

**Signup body:**
```json
{
    "username":"hridoy",
    "email":"hridoy@test.com",
    "password":"Pass1234!",
    "phone":"1234567890"
    }
```

## Stats

```
GET /stats    JWT required
```

Returns counts for dashboard:
```json
{
    "total_projects":3,
    "total_error_groups":15,
    "total_errors":42,"by_status":{"critical":12,"resolved":2,"recovered":1},"errors_last_24h":8}
```

## Projects

| Method | Path | Auth | What |
|--------|------|------|------|
| GET | `/projects?page=&limit=&q=` | JWT | List (paginated, searchable) |
| POST | `/projects` | JWT | Create |
| GET | `/projects/:id` | JWT | Get one |
| PATCH | `/projects/:id` | JWT | Update |
| DELETE | `/projects/:id` | JWT | Delete |
| POST | `/projects/:id/rotate-key` | JWT | New API key |

**Note:** `q` max 100 chars.

## Error Groups

| Method | Path | Auth | What |
|--------|------|------|------|
| GET | `/projects/:id/errors?page=&limit=&status=&q=` | JWT | List (paginated, filterable, searchable) |
| GET | `/projects/:id/errors/:gid?page=&limit=&severity=` | JWT | Get one + events |
| PATCH | `/projects/:id/errors/:gid` | JWT | Update status |
| DELETE | `/projects/:id/errors/:gid` | JWT | Delete |

**Status values:** `critical` | `resolved` | `recovered`

**Severity values:** `error` | `warning` | `info`

**Note:** `q` max 200 chars.

## Ingestion

```
POST /ingest    API Key required (header: X-Mutt-Key)
```

Rate limit: 100/min per project.

```json
{"title":"TypeError","log":"x is undefined","stack_trace":"at render (app.js:42)","severity":"error"}
```

- `title` required, max 500
- `log` required
- `stack_trace` optional
- `severity` optional, defaults to `error`

Returns `202` with `error_group_id` and `error_id`.

## Backup

| Method | Path | Auth | Rate limit |
|--------|------|------|------------|
| GET | `/backup` | JWT | 5/min |
| POST | `/backup/import` | JWT | 2/min |

Export: JSON (auto-gzip > 10KB). Import: multipart upload, max 5MB / 100K records.

## Pagination

Every list endpoint returns:
```json
{"page":1,"limit":20,"total_count":42,"total_pages":3}
```

Defaults: page=1, limit=20. Limit max 100.
