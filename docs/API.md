# Key Agent HTTP API Reference

## Overview

Key Agent provides a RESTful HTTP API for managing key-value data and secrets.

- **Base URL**: `http://127.0.0.1:8080`
- **Authentication**: Bearer Token (Authorization: Bearer `<token>`)
- **Content-Type**: `application/json`

## Authentication

All API endpoints (except `/health`) require authentication via Bearer token:

```http
Authorization: Bearer ka_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

## Endpoints

### Health

#### GET /health

Check server health status. No authentication required.

**Response**

```json
{
  "status": "healthy",
  "version": "1.0.0"
}
```

**Status Codes**

| Code | Description |
|------|-------------|
| 200 | Server is healthy |

---

### KV Operations

#### GET /api/v1/kv

List all KV keys with optional prefix filter.

**Query Parameters**

| Parameter | Type | Description |
|-----------|------|-------------|
| prefix | string | Filter keys by prefix |

**Response**

```json
{
  "keys": ["app/database/host", "app/database/port", "app/config"]
}
```

**Status Codes**

| Code | Description |
|------|-------------|
| 200 | Success |
| 401 | Unauthorized |

---

#### GET /api/v1/kv/:key

Get a KV entry by key.

**Path Parameters**

| Parameter | Type | Description |
|-----------|------|-------------|
| key | string | The key to retrieve |

**Response**

```json
{
  "key": "app/database/host",
  "value": "localhost",
  "metadata": {
    "description": "Database host"
  },
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z",
  "version": 1
}
```

**Status Codes**

| Code | Description |
|------|-------------|
| 200 | Success |
| 401 | Unauthorized |
| 404 | Key not found |

---

#### PUT /api/v1/kv/:key

Create or update a KV entry.

**Path Parameters**

| Parameter | Type | Description |
|-----------|------|-------------|
| key | string | The key to set |

**Request Body**

```json
{
  "value": "localhost",
  "metadata": {
    "description": "Database host"
  }
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| value | string | Yes | The value to store |
| metadata | object | No | Optional metadata |

**Response**

```json
{
  "key": "app/database/host",
  "value": "localhost",
  "metadata": {
    "description": "Database host"
  },
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z",
  "version": 1
}
```

**Status Codes**

| Code | Description |
|------|-------------|
| 200 | Success |
| 400 | Invalid request |
| 401 | Unauthorized |

---

#### DELETE /api/v1/kv/:key

Delete a KV entry.

**Path Parameters**

| Parameter | Type | Description |
|-----------|------|-------------|
| key | string | The key to delete |

**Response**

Empty body.

**Status Codes**

| Code | Description |
|------|-------------|
| 204 | Success (no content) |
| 401 | Unauthorized |
| 404 | Key not found |

---

### Secret Operations

#### GET /api/v1/secrets

List all secret keys with optional prefix filter.

**Query Parameters**

| Parameter | Type | Description |
|-----------|------|-------------|
| prefix | string | Filter keys by prefix |

**Response**

```json
{
  "keys": ["db/postgres/password", "openai/api_key"]
}
```

**Status Codes**

| Code | Description |
|------|-------------|
| 200 | Success |
| 401 | Unauthorized |

---

#### GET /api/v1/secrets/:key

Get a secret entry by key.

**Path Parameters**

| Parameter | Type | Description |
|-----------|------|-------------|
| key | string | The secret key to retrieve |

**Response**

```json
{
  "key": "db/postgres/password",
  "value": "super-secret-password",
  "type": "password",
  "metadata": {
    "description": "PostgreSQL password"
  },
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z",
  "version": 1
}
```

**Status Codes**

| Code | Description |
|------|-------------|
| 200 | Success |
| 401 | Unauthorized |
| 404 | Secret not found |

---

#### PUT /api/v1/secrets/:key

Create or update a secret entry.

**Path Parameters**

| Parameter | Type | Description |
|-----------|------|-------------|
| key | string | The secret key to set |

**Request Body**

```json
{
  "value": "super-secret-password",
  "type": "password",
  "metadata": {
    "description": "PostgreSQL password"
  }
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| value | string | Yes | The secret value |
| type | string | Yes | Secret type: `password`, `api_key`, `certificate`, `private_key`, `token`, `other` |
| metadata | object | No | Optional metadata |

**Response**

```json
{
  "key": "db/postgres/password",
  "value": "super-secret-password",
  "type": "password",
  "metadata": {
    "description": "PostgreSQL password"
  },
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z",
  "version": 1
}
```

**Status Codes**

| Code | Description |
|------|-------------|
| 200 | Success |
| 400 | Invalid request (e.g., invalid type) |
| 401 | Unauthorized |

---

#### DELETE /api/v1/secrets/:key

Delete a secret entry.

**Path Parameters**

| Parameter | Type | Description |
|-----------|------|-------------|
| key | string | The secret key to delete |

**Response**

Empty body.

**Status Codes**

| Code | Description |
|------|-------------|
| 204 | Success (no content) |
| 401 | Unauthorized |
| 404 | Secret not found |

---

### Token Operations

#### POST /api/v1/token

Create a new access token.

**Request Body**

```json
{
  "name": "my-application",
  "type": "client",
  "expires_in": "24h"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| name | string | Yes | Token name for identification |
| type | string | No | Token type: `client` (default) or `mcp` |
| expires_in | string | No | Duration string: `1h`, `24h`, `30d`, etc. |

**Response**

```json
{
  "token": "ka_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
  "name": "my-application",
  "type": "client",
  "created_at": "2024-01-15T10:30:00Z",
  "expires_at": "2024-01-16T10:30:00Z"
}
```

**Status Codes**

| Code | Description |
|------|-------------|
| 201 | Token created |
| 400 | Invalid request |
| 401 | Unauthorized |

---

## Error Responses

All error responses follow this format:

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable error message"
  }
}
```

### Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `UNAUTHORIZED` | 401 | Missing or invalid token |
| `FORBIDDEN` | 403 | Token lacks required permissions |
| `NOT_FOUND` | 404 | Resource not found |
| `INVALID_ARGUMENT` | 400 | Invalid request parameters |
| `METHOD_NOT_ALLOWED` | 405 | HTTP method not supported |
| `INTERNAL_ERROR` | 500 | Internal server error |

---

## Examples

### cURL

```bash
# Health check
curl http://localhost:8080/health

# Set KV
curl -X PUT http://localhost:8080/api/v1/kv/my-key \
  -H "Authorization: Bearer ka_xxx" \
  -H "Content-Type: application/json" \
  -d '{"value": "my-value"}'

# Get KV
curl http://localhost:8080/api/v1/kv/my-key \
  -H "Authorization: Bearer ka_xxx"

# Set secret
curl -X PUT http://localhost:8080/api/v1/secrets/db-password \
  -H "Authorization: Bearer ka_xxx" \
  -H "Content-Type: application/json" \
  -d '{"value": "secret123", "type": "password"}'

# Create token
curl -X POST http://localhost:8080/api/v1/token \
  -H "Authorization: Bearer ka_xxx" \
  -H "Content-Type: application/json" \
  -d '{"name": "new-app", "type": "client", "expires_in": "24h"}'
```

### Python

```python
import requests

BASE_URL = "http://localhost:8080"
TOKEN = "ka_xxx"

headers = {"Authorization": f"Bearer {TOKEN}"}

# Set KV
response = requests.put(
    f"{BASE_URL}/api/v1/kv/my-key",
    headers=headers,
    json={"value": "my-value"}
)
print(response.json())

# Get KV
response = requests.get(f"{BASE_URL}/api/v1/kv/my-key", headers=headers)
print(response.json())

# Set secret
response = requests.put(
    f"{BASE_URL}/api/v1/secrets/api-key",
    headers=headers,
    json={"value": "sk-xxx", "type": "api_key"}
)
print(response.json())
```
