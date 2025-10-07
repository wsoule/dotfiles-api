# Dotfiles API - TypeScript Types

This document contains all TypeScript type definitions for the Dotfiles API.

## User Types

```typescript
interface User {
  id: string;
  github_id: number;
  username: string;
  name: string;
  email: string;
  avatar_url: string;
  bio: string;
  location: string;
  website: string;
  company: string;
  created_at: string; // ISO 8601 datetime
  updated_at: string; // ISO 8601 datetime
  favorites: string[]; // Template IDs
  collections: string[]; // Collection IDs
}

interface UserProfile {
  user: User;
  reviews: Review[];
  favorite_templates: Template[];
  organizations: Organization[];
  stats: UserStats;
}

interface UserStats {
  template_count: number;
  review_count: number;
  favorite_count: number;
  organization_count: number;
}
```

## Template Types

```typescript
interface ShareMetadata {
  name: string;
  description: string;
  author: string;
  tags: string[];
  created_at: string; // ISO 8601 datetime
  updated_at: string; // ISO 8601 datetime
  version: string;
}

interface Template {
  taps: string[];
  brews: string[];
  casks: string[];
  stow: string[];
  metadata: ShareMetadata;
  extends?: string;
  overrides?: string[];
  addOnly: boolean;
  public: boolean;
  featured: boolean;
  organization_id?: string;
}

interface StoredTemplate {
  id: string;
  template: Template;
  created_at: string; // ISO 8601 datetime
  updated_at: string; // ISO 8601 datetime
  downloads: number;
}

interface TemplateStats {
  total_templates: number;
  featured_templates: number;
  total_downloads: number;
  categories: number;
}

interface TemplateRating {
  template_id: string;
  average_rating: number;
  total_ratings: number;
  distribution: Record<string, number>; // rating -> count
}
```

## Config Types

```typescript
interface BasicConfig {
  brews: string[];
  casks: string[];
  taps: string[];
  stow: string[];
}

interface ShareableConfig extends BasicConfig {
  metadata: ShareMetadata;
}

interface StoredConfig {
  id: string;
  config: ShareableConfig;
  public: boolean;
  created_at: string; // ISO 8601 datetime
  download_count: number;
  owner_id: string;
}

interface ConfigStats {
  total_configs: number;
  public_configs: number;
  total_downloads: number;
}
```

## Review Types

```typescript
interface Review {
  id: string;
  template_id: string;
  user_id: string;
  username: string;
  avatar_url: string;
  rating: number; // 1-5
  comment: string;
  helpful: number; // helpful votes count
  created_at: string; // ISO 8601 datetime
  updated_at: string; // ISO 8601 datetime
}
```

## Organization Types

```typescript
type OrganizationRole = 'owner' | 'admin' | 'member';

interface Organization {
  id: string;
  name: string;
  slug: string;
  description: string;
  website: string;
  owner_id: string;
  public: boolean;
  created_at: string; // ISO 8601 datetime
  updated_at: string; // ISO 8601 datetime
  member_count: number;
}

interface OrganizationMember {
  id: string;
  organization_id: string;
  user_id: string;
  role: OrganizationRole;
  joined_at: string; // ISO 8601 datetime
}

interface OrganizationInvite {
  id: string;
  organization_id: string;
  email: string;
  role: OrganizationRole;
  token: string;
  invited_by: string;
  created_at: string; // ISO 8601 datetime
  expires_at: string; // ISO 8601 datetime
  accepted_at?: string; // ISO 8601 datetime
}
```

## API Response Types

```typescript
interface ApiError {
  code: string;
  message: string;
  status_code: number;
}

interface SuccessResponse<T> {
  data: T;
  message?: string;
}

interface ErrorResponse {
  error: ApiError;
}

interface PaginatedResponse<T> {
  items: T[];
  limit: number;
  offset: number;
  total: number;
}
```

## Auth Types

```typescript
interface AuthResponse {
  message: string;
  user: {
    id: string;
    username: string;
    name: string;
    email: string;
    avatar_url: string;
  };
}

interface CurrentUserResponse {
  user: {
    id: string;
    username: string;
    name: string;
    email: string;
    avatar_url: string;
    bio: string;
    location: string;
    website: string;
    created_at: string;
  };
  configured: boolean;
}
```

## Request Types

```typescript
// Review Requests
interface CreateReviewRequest {
  rating: number; // 1-5
  comment: string;
}

interface UpdateReviewRequest {
  rating: number; // 1-5
  comment: string;
}

// Organization Requests
interface CreateOrganizationRequest {
  name: string;
  slug: string;
  description: string;
  website: string;
  public: boolean;
}

// Template Requests
interface CreateTemplateRequest {
  template: Template;
}

// Config Requests
interface UploadConfigRequest extends ShareableConfig {}
```

## API Endpoints

### Authentication
- `GET /auth/github` - Initiate GitHub OAuth
- `GET /auth/github/callback` - GitHub OAuth callback
- `GET /auth/logout` - Logout
- `GET /auth/user` - Get current user

### Templates
- `POST /api/templates` - Create template
- `GET /api/templates` - List templates
- `GET /api/templates/:id` - Get template
- `GET /api/templates/:id/download` - Download template
- `GET /api/templates/:id/reviews` - Get template reviews
- `POST /api/templates/:id/reviews` - Create review (auth required)
- `GET /api/templates/:id/rating` - Get template rating

### Users
- `GET /api/users/:username` - Get user profile
- `POST /api/users/favorites/:templateId` - Add favorite (auth required)
- `DELETE /api/users/favorites/:templateId` - Remove favorite (auth required)
- `GET /api/users/:username/organizations` - Get user organizations

### Reviews
- `PUT /api/reviews/:id` - Update review (auth required)
- `DELETE /api/reviews/:id` - Delete review (auth required)
- `POST /api/reviews/:id/helpful` - Mark review helpful (auth required)

### Organizations
- `POST /api/organizations` - Create organization (auth required)
- `GET /api/organizations` - List organizations
- `GET /api/organizations/:slug` - Get organization
- `PUT /api/organizations/:slug` - Update organization (auth required)
- `DELETE /api/organizations/:slug` - Delete organization (auth required)
- `GET /api/organizations/:slug/members` - Get members
- `POST /api/organizations/:slug/members` - Invite member (auth required)
- `DELETE /api/organizations/:slug/members/:username` - Remove member (auth required)
- `PUT /api/organizations/:slug/members/:username` - Update member role (auth required)
- `GET /api/organizations/:slug/invites` - Get invites (auth required)
- `POST /api/invites/:token/accept` - Accept invite (auth required)

### Configs
- `POST /api/configs/upload` - Upload config
- `GET /api/configs/:id` - Get config
- `GET /api/configs/:id/download` - Download config
- `GET /api/configs/search?q=query` - Search configs
- `GET /api/configs/featured` - Get featured configs
- `GET /api/configs/stats` - Get statistics

## Error Codes

```typescript
type ErrorCode =
  | 'VALIDATION_ERROR' // 400
  | 'BAD_REQUEST' // 400
  | 'UNAUTHORIZED' // 401
  | 'FORBIDDEN' // 403
  | 'NOT_FOUND' // 404
  | 'CONFLICT' // 409
  | 'INTERNAL_ERROR'; // 500
```

## Query Parameters

```typescript
interface PaginationParams {
  limit?: number; // Default: 10, Max: 100
  offset?: number; // Default: 0
}

interface SearchParams extends PaginationParams {
  q: string; // Search query
}
```

## Notes

- All datetime fields are in ISO 8601 format (e.g., `"2025-10-06T20:55:17Z"`)
- Authentication uses session cookies (cookie name: `session_id`)
- The API uses `HttpOnly` cookies for security
- Some features require MongoDB to be configured (configs, reviews, organizations)
- Templates and users work with in-memory storage when MongoDB is not available
