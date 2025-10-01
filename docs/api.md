# Dotfiles Web API Documentation

## Overview

The Dotfiles Web API is a RESTful API for managing dotfiles templates, users, organizations, and reviews. This API provides comprehensive functionality for creating, sharing, and managing dotfiles configurations.

### Base URL
```
http://localhost:8080/api
```

### Authentication
The API uses session-based authentication via GitHub OAuth. Most endpoints require authentication.

### Content Type
All requests and responses use `application/json` content type.

### Error Handling
The API returns consistent error responses with the following structure:

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable error message",
    "details": "Additional details (optional)",
    "status_code": 400
  }
}
```

Common error codes:
- `VALIDATION_ERROR`: Invalid input data
- `NOT_FOUND`: Resource not found
- `UNAUTHORIZED`: Authentication required
- `FORBIDDEN`: Insufficient permissions
- `CONFLICT`: Resource already exists
- `INTERNAL_ERROR`: Server error
- `RATE_LIMIT`: Too many requests

## Authentication Endpoints

### GitHub OAuth Login
```
GET /auth/github
```
Redirects to GitHub OAuth authorization page.

### GitHub OAuth Callback
```
GET /auth/github/callback?code={code}&state={state}
```
Handles GitHub OAuth callback and creates user session.

### Logout
```
POST /auth/logout
```
Destroys user session.

## User Management

### Create User
```
POST /api/users
```

**Request Body:**
```json
{
  "username": "string (required, 3-30 chars, alphanumeric + _ -)",
  "name": "string (required)",
  "email": "string (required, valid email)",
  "avatar_url": "string (optional)",
  "bio": "string (optional)",
  "location": "string (optional)",
  "website": "string (optional, valid URL)",
  "company": "string (optional)"
}
```

**Response:** `201 Created`
```json
{
  "message": "User created successfully"
}
```

### Get User by ID
```
GET /api/users/{id}
```

**Response:** `200 OK`
```json
{
  "id": "string",
  "username": "string",
  "name": "string",
  "email": "string",
  "avatar_url": "string",
  "bio": "string",
  "location": "string",
  "website": "string",
  "company": "string",
  "created_at": "2023-01-01T00:00:00Z",
  "updated_at": "2023-01-01T00:00:00Z"
}
```

### Get User by Username
```
GET /api/users/username/{username}
```

**Response:** Same as Get User by ID

### Update User
```
PUT /api/users/{id}
```

**Request Body:**
```json
{
  "name": "string (optional)",
  "bio": "string (optional)",
  "location": "string (optional)",
  "website": "string (optional, valid URL)",
  "company": "string (optional)"
}
```

**Response:** `200 OK`
```json
{
  "message": "User updated successfully"
}
```

### Delete User
```
DELETE /api/users/{id}
```

**Response:** `200 OK`
```json
{
  "message": "User deleted successfully"
}
```

### List Users
```
GET /api/users?limit={limit}&offset={offset}
```

**Query Parameters:**
- `limit`: Number of users to return (1-100, default: 10)
- `offset`: Number of users to skip (default: 0)

**Response:** `200 OK`
```json
{
  "users": [
    {
      "id": "string",
      "username": "string",
      "name": "string",
      "email": "string",
      "avatar_url": "string",
      "bio": "string",
      "location": "string",
      "website": "string",
      "company": "string",
      "created_at": "2023-01-01T00:00:00Z",
      "updated_at": "2023-01-01T00:00:00Z"
    }
  ],
  "limit": 10,
  "offset": 0,
  "total": 1
}
```

### Add Template to Favorites
```
POST /api/users/{id}/favorites/{templateId}
```

**Response:** `200 OK`
```json
{
  "message": "Template added to favorites"
}
```

### Remove Template from Favorites
```
DELETE /api/users/{id}/favorites/{templateId}
```

**Response:** `200 OK`
```json
{
  "message": "Template removed from favorites"
}
```

### Get User Favorites
```
GET /api/users/{id}/favorites
```

**Response:** `200 OK`
```json
{
  "favorites": ["template_id_1", "template_id_2"]
}
```

## Template Management

### Create Template
```
POST /api/templates
```

**Request Body:**
```json
{
  "taps": ["string"],
  "brews": ["string"],
  "casks": ["string"],
  "stow": ["string"],
  "metadata": {
    "name": "string (required, 3-100 chars)",
    "description": "string (required, 10-500 chars)",
    "author": "string (required)",
    "version": "string (required)",
    "tags": ["string"] // max 10 tags, each max 30 chars
  },
  "extends": "string",
  "overrides": ["string"],
  "add_only": false,
  "public": true,
  "featured": false,
  "organization_id": "string"
}
```

**Response:** `201 Created`
```json
{
  "message": "Template created successfully"
}
```

### Get Template
```
GET /api/templates/{id}
```

**Response:** `200 OK`
```json
{
  "id": "string",
  "taps": ["string"],
  "brews": ["string"],
  "casks": ["string"],
  "stow": ["string"],
  "metadata": {
    "name": "string",
    "description": "string",
    "author": "string",
    "version": "string",
    "tags": ["string"],
    "created_at": "2023-01-01T00:00:00Z",
    "updated_at": "2023-01-01T00:00:00Z"
  },
  "extends": "string",
  "overrides": ["string"],
  "add_only": false,
  "public": true,
  "featured": false,
  "organization_id": "string",
  "downloads": 0,
  "created_at": "2023-01-01T00:00:00Z",
  "updated_at": "2023-01-01T00:00:00Z"
}
```

### Update Template
```
PUT /api/templates/{id}
```

**Request Body:** Same as Create Template, but all fields optional

**Response:** `200 OK`
```json
{
  "message": "Template updated successfully"
}
```

### Delete Template
```
DELETE /api/templates/{id}
```

**Response:** `200 OK`
```json
{
  "message": "Template deleted successfully"
}
```

### List Templates
```
GET /api/templates?author={author}&tags={tag1,tag2}&featured={true|false}&public={true|false}&organization_id={orgId}&sort_by={field}&sort_order={asc|desc}&limit={limit}&offset={offset}
```

**Query Parameters:**
- `author`: Filter by author username
- `tags`: Filter by tags (comma-separated)
- `featured`: Filter by featured status
- `public`: Filter by public status
- `organization_id`: Filter by organization
- `sort_by`: Sort field (default: created_at)
- `sort_order`: Sort order (asc/desc, default: desc)
- `limit`: Number of templates (1-100, default: 10)
- `offset`: Number to skip (default: 0)

**Response:** `200 OK`
```json
{
  "templates": [
    // Array of template objects
  ],
  "limit": 10,
  "offset": 0,
  "total": 1
}
```

### Search Templates
```
GET /api/templates/search?q={query}&limit={limit}&offset={offset}
```

**Query Parameters:**
- `q`: Search query (required)
- `limit`: Number of results (1-100, default: 10)
- `offset`: Number to skip (default: 0)

**Response:** `200 OK`
```json
{
  "templates": [
    // Array of template objects
  ],
  "query": "search query",
  "limit": 10,
  "offset": 0,
  "total": 1
}
```

### Download Template
```
GET /api/templates/{id}/download
```

Downloads the template configuration and increments download counter.

**Response:** `200 OK`
```json
{
  "taps": ["string"],
  "brews": ["string"],
  "casks": ["string"],
  "stow": ["string"],
  "metadata": {
    // metadata object
  },
  "extends": "string",
  "overrides": ["string"],
  "add_only": false,
  "public": true,
  "featured": false,
  "organization_id": "string"
}
```

### Get Template Statistics
```
GET /api/templates/stats
```

**Response:** `200 OK`
```json
{
  "total_templates": 100,
  "featured_templates": 10,
  "total_downloads": 1000,
  "categories": 5
}
```

### Get Template Rating
```
GET /api/templates/{id}/rating
```

**Response:** `200 OK`
```json
{
  "template_id": "string",
  "average_rating": 4.5,
  "total_ratings": 20,
  "distribution": {
    "1": 1,
    "2": 2,
    "3": 3,
    "4": 4,
    "5": 10
  }
}
```

## Organization Management

### Create Organization
```
POST /api/organizations
```

**Request Body:**
```json
{
  "name": "string (required, 3-50 chars)",
  "slug": "string (required, 3-30 chars, lowercase, alphanumeric + hyphens)",
  "description": "string (optional, max 200 chars)",
  "website": "string (optional, valid URL)",
  "public": true
}
```

**Response:** `201 Created`

### Get Organization
```
GET /api/organizations/{id}
```

**Response:** `200 OK`
```json
{
  "id": "string",
  "name": "string",
  "slug": "string",
  "description": "string",
  "website": "string",
  "owner_id": "string",
  "public": true,
  "created_at": "2023-01-01T00:00:00Z",
  "updated_at": "2023-01-01T00:00:00Z",
  "member_count": 5
}
```

### Get Organization by Slug
```
GET /api/organizations/slug/{slug}
```

**Response:** Same as Get Organization

### Update Organization
```
PUT /api/organizations/{id}
```

**Request Body:**
```json
{
  "name": "string (optional)",
  "description": "string (optional)",
  "website": "string (optional)",
  "public": "boolean (optional)"
}
```

### Delete Organization
```
DELETE /api/organizations/{id}
```

### List Organizations
```
GET /api/organizations?limit={limit}&offset={offset}
```

### Search Organizations
```
GET /api/organizations/search?q={query}&limit={limit}&offset={offset}
```

### Get User's Organizations
```
GET /api/users/{userId}/organizations
```

## Organization Membership

### Add Member
```
POST /api/organizations/{id}/members
```

**Request Body:**
```json
{
  "user_id": "string (required)",
  "role": "string (required: owner|admin|member)"
}
```

### Remove Member
```
DELETE /api/organizations/{id}/members/{userId}
```

### Update Member Role
```
PUT /api/organizations/{id}/members/{userId}
```

**Request Body:**
```json
{
  "role": "string (required: owner|admin|member)"
}
```

### Get Organization Members
```
GET /api/organizations/{id}/members
```

**Response:** `200 OK`
```json
{
  "members": [
    {
      "id": "string",
      "organization_id": "string",
      "user_id": "string",
      "username": "string",
      "name": "string",
      "avatar_url": "string",
      "role": "string",
      "joined_at": "2023-01-01T00:00:00Z"
    }
  ]
}
```

## Organization Invitations

### Invite User
```
POST /api/organizations/{id}/invites
```

**Request Body:**
```json
{
  "email": "string (required, valid email)",
  "role": "string (required: owner|admin|member)"
}
```

### Get Organization Invites
```
GET /api/organizations/{id}/invites
```

### Accept Invite
```
POST /api/organizations/invites/accept
```

**Request Body:**
```json
{
  "token": "string (required)"
}
```

### Delete Invite
```
DELETE /api/organizations/invites/{id}
```

## Review Management

### Create Review
```
POST /api/reviews
```

**Request Body:**
```json
{
  "template_id": "string (required)",
  "rating": "number (required, 1-5)",
  "comment": "string (optional, max 1000 chars)"
}
```

### Get Review
```
GET /api/reviews/{id}
```

### Update Review
```
PUT /api/reviews/{id}
```

**Request Body:**
```json
{
  "rating": "number (optional, 1-5)",
  "comment": "string (optional)"
}
```

### Delete Review
```
DELETE /api/reviews/{id}
```

### Get Template Reviews
```
GET /api/templates/{id}/reviews?limit={limit}&offset={offset}
```

### Get User Reviews
```
GET /api/users/{id}/reviews?limit={limit}&offset={offset}
```

### Mark Review Helpful
```
POST /api/reviews/{id}/helpful
```

## Rate Limiting

The API implements rate limiting to prevent abuse:
- Default: 100 requests per hour per IP address
- Rate limit headers are included in responses:
  - `X-RateLimit-Limit`: Maximum requests allowed
  - `X-RateLimit-Remaining`: Remaining requests
  - `X-RateLimit-Reset`: Time when rate limit resets

## Pagination

List endpoints support pagination with query parameters:
- `limit`: Number of items to return (max 100)
- `offset`: Number of items to skip

Responses include pagination metadata:
```json
{
  "data": [...],
  "limit": 10,
  "offset": 0,
  "total": 100
}
```

## Filtering and Sorting

Many list endpoints support filtering and sorting:
- Use query parameters for filtering (e.g., `?public=true&featured=true`)
- Use `sort_by` and `sort_order` for sorting
- Multiple values can be comma-separated (e.g., `?tags=frontend,javascript`)

## Webhooks

The API supports webhooks for real-time notifications:
- Template created/updated/deleted
- Organization membership changes
- Review submissions

Webhook endpoints can be configured per organization with appropriate permissions.

## SDKs and Libraries

Official SDKs are available for:
- JavaScript/TypeScript
- Python
- Go
- Rust

Community SDKs are available for other languages.

## Support

For API support and questions:
- GitHub Issues: [github.com/wsoule/dotfiles-web/issues](https://github.com/wsoule/dotfiles-web/issues)
- Email: support@dotfiles-web.com
- Documentation: [docs.dotfiles-web.com](https://docs.dotfiles-web.com)