# Dotfiles Web - Community Template Platform

A modern, scalable Go web application for sharing and discovering development environment templates. Built with clean architecture principles, featuring GitHub authentication, organization management, community reviews, and comprehensive API documentation.

## 🎯 **Recently Refactored with Clean Architecture**

This application has been completely refactored from a monolithic structure into a maintainable, well-organized codebase following Go best practices. Perfect foundation for modern frontend frameworks!

## 🌟 Key Features

### 🔐 **User Authentication**
- **GitHub OAuth 2.0 integration** - Secure sign-in with GitHub
- **User profiles** with avatars and metadata
- **Session management** with secure cookies
- **Protected endpoints** for user-specific actions

### 🏢 **Organizations & Teams**
- **Organization management** with role-based permissions (owner, admin, member)
- **Template ownership** by organizations or individuals
- **Invitation system** with secure token-based invites
- **Member management** with different access levels
- **Organization profiles** with public/private visibility

### ⭐ **Template Ratings & Reviews**
- **5-star rating system** with aggregate calculations
- **Community reviews** with comments and helpful voting
- **Rating distributions** showing review breakdowns
- **Review management** (create, edit, delete your own reviews)
- **Helpful vote tracking** for community feedback

### 🔍 **Advanced Search & Filtering**
- **Real-time search** by name, description, and technologies
- **Tag-based filtering** with category support
- **Featured vs. community template filtering**
- **Multiple sorting options** (downloads, name, date, author)
- **Category browsing** with visual category cards
- **Grid and list view** toggles for different browsing preferences

### 🎨 **Modern Web Interface**
- **Dark theme** responsive design
- **Template browser** with detailed modal views
- **Interactive search** and filtering
- **User dashboard** with favorites management
- **Documentation pages** with navigation sidebar

### 📦 **Template Management**
- **6 pre-built templates** covering major development stacks:
  - Full Stack Web Developer
  - Data Science Toolkit
  - DevOps Engineer Setup
  - Mobile Developer Setup
  - Backend Developer Kit
  - Minimal Developer Setup
- **Package categorization** (Homebrew, Casks, Taps, Dotfiles)
- **Download tracking** and statistics
- **Template versioning** and metadata

## 🏗️ Clean Architecture

This application follows clean architecture principles with clear separation of concerns:

### 📁 **Project Structure**
```
├── cmd/server/              # Application entry point
├── internal/
│   ├── config/              # Configuration management
│   ├── dto/                 # Data Transfer Objects + validation
│   ├── handlers/            # HTTP request handlers
│   ├── middleware/          # HTTP middleware (auth, CORS, rate limiting)
│   ├── models/              # Domain models
│   ├── repository/          # Data access layer interfaces
│   └── services/            # Business logic layer
├── pkg/errors/              # Custom error handling
├── docs/                    # API and architecture documentation
└── static/                  # Frontend assets (legacy)
```

### 🎯 **Architecture Benefits**
- **Clean separation** of concerns across layers
- **Repository pattern** for easy database switching
- **Comprehensive validation** with DTOs
- **Structured error handling** with custom types
- **Middleware-based** cross-cutting concerns
- **Interface-driven** design for testability
- **Environment-based** configuration management

### Backend (Go + Gin)
- **RESTful API** with 30+ documented endpoints
- **Repository pattern** supporting MongoDB and in-memory storage
- **Comprehensive data models** for users, templates, organizations, and reviews
- **Role-based authentication** and authorization middleware
- **Rate limiting** and security features
- **Comprehensive API documentation**

### Frontend (Ready for Modern Frameworks)
- **Well-documented REST API** for easy integration
- **Consistent error responses** for proper error handling
- **Comprehensive validation** at API boundaries
- **Static assets** as reference implementation
- **CORS support** for frontend frameworks

## 🚀 API Endpoints

### Authentication
- `GET /auth/github` - Initiate GitHub OAuth
- `GET /auth/github/callback` - OAuth callback
- `GET /auth/logout` - Sign out
- `GET /auth/user` - Get current user

### Templates
- `GET /api/templates` - List templates with search/filter
- `GET /api/templates/:id` - Get template details
- `GET /api/templates/:id/download` - Download template
- `POST /api/templates` - Create new template
- `PUT /api/templates/:id` - Update template
- `DELETE /api/templates/:id` - Delete template
- `GET /api/templates/search` - Search templates
- `GET /api/templates/stats` - Get template statistics
- `GET /api/templates/:id/rating` - Get template rating

### Organizations
- `GET /api/organizations` - List organizations
- `POST /api/organizations` - Create organization
- `GET /api/organizations/:id` - Get organization details
- `PUT /api/organizations/:id` - Update organization
- `DELETE /api/organizations/:id` - Delete organization
- `GET /api/organizations/:id/members` - Get organization members
- `POST /api/organizations/:id/members` - Add member
- `PUT /api/organizations/:id/members/:userId` - Update member role
- `DELETE /api/organizations/:id/members/:userId` - Remove member
- `POST /api/organizations/:id/invites` - Create invitation
- `GET /api/organizations/:id/invites` - List invitations
- `POST /api/organizations/invites/accept` - Accept invitation

### Users & Profiles
- `GET /api/users/:id` - Get user by ID
- `GET /api/users/username/:username` - Get user by username
- `POST /api/users` - Create user
- `PUT /api/users/:id` - Update user
- `DELETE /api/users/:id` - Delete user
- `GET /api/users` - List users
- `POST /api/users/:id/favorites/:templateId` - Add to favorites
- `DELETE /api/users/:id/favorites/:templateId` - Remove from favorites
- `GET /api/users/:id/favorites` - Get user favorites

### Reviews & Ratings
- `POST /api/reviews` - Create review
- `GET /api/reviews/:id` - Get review
- `PUT /api/reviews/:id` - Update review
- `DELETE /api/reviews/:id` - Delete review
- `GET /api/templates/:id/reviews` - Get template reviews
- `GET /api/users/:id/reviews` - Get user reviews
- `POST /api/reviews/:id/helpful` - Mark review helpful

### Legacy Config API
- `POST /api/configs/upload` - Upload a config
- `GET /api/configs/:id` - Get config by ID
- `GET /api/configs/search` - Search configs
- `GET /api/configs/featured` - Get featured configs
- `GET /api/configs/stats` - Get platform statistics

## 🔧 Environment Variables

- `PORT` - Server port (default: 8080, automatically set by Railway)
- `MONGODB_URI` - MongoDB connection string (optional, uses in-memory storage if not provided)
- `MONGODB_DATABASE` - MongoDB database name (default: "dotfiles")
- `GITHUB_CLIENT_ID` - GitHub OAuth app client ID
- `GITHUB_CLIENT_SECRET` - GitHub OAuth app client secret
- `OAUTH_REDIRECT_URL` - OAuth callback URL (e.g., `http://localhost:8080/auth/github/callback`)

## 🏃 Local Development

### Prerequisites
- Go 1.19+ installed
- (Optional) MongoDB for persistent storage
- GitHub OAuth app configured

### Setup GitHub OAuth (Optional)
1. Go to GitHub Settings > Developer settings > OAuth Apps
2. Create a new OAuth app with:
   - Homepage URL: `http://localhost:8080`
   - Authorization callback URL: `http://localhost:8080/auth/github/callback`
3. Copy the Client ID and Client Secret

### Run the Application
```bash
# Install dependencies
go mod tidy

# Set environment variables (optional)
export GITHUB_CLIENT_ID="your_github_client_id"
export GITHUB_CLIENT_SECRET="your_github_client_secret"
export OAUTH_REDIRECT_URL="http://localhost:8080/auth/github/callback"

# Run the server
go run main.go
```

Server will start on http://localhost:8080

**🌐 Open http://localhost:8080 in your browser to see the web interface!**

### Pages Available
- `/` - Home page with config upload
- `/templates` - Template browser with search and filtering
- `/docs` - Documentation with sidebar navigation
- `/template/:id` - Individual template detail pages

## 🚢 Railway Deployment

1. **Connect to Railway**
   - Connect your GitHub repo to Railway
   - Railway will automatically detect this as a Go app

2. **Add MongoDB Database (Optional)**
   - In Railway dashboard, add MongoDB as a service
   - Copy the MongoDB connection string

3. **Set Environment Variables**
   - `MONGODB_URI` - The MongoDB connection string (optional)
   - `MONGODB_DATABASE` - "dotfiles" (or your preferred database name)
   - `GIN_MODE` - "release" (for production)
   - `GITHUB_CLIENT_ID` - Your GitHub OAuth app client ID
   - `GITHUB_CLIENT_SECRET` - Your GitHub OAuth app client secret
   - `OAUTH_REDIRECT_URL` - Your production callback URL

4. **Deploy!**
   - Railway will automatically build and deploy your app
   - The app works with or without MongoDB
   - In-memory storage is used as fallback

## 🧪 Testing

### API Testing
```bash
# Get all templates
curl http://localhost:8080/api/templates

# Search templates
curl "http://localhost:8080/api/templates?search=web&featured=true"

# Filter by tags
curl "http://localhost:8080/api/templates?tags=devops,docker"

# Get template rating
curl http://localhost:8080/api/templates/TEMPLATE_ID/rating

# Get template reviews
curl http://localhost:8080/api/templates/TEMPLATE_ID/reviews

# Test authentication status
curl http://localhost:8080/auth/user
```

### Frontend Testing
1. Open http://localhost:8080 in your browser
2. Navigate to the Templates page
3. Try searching and filtering templates
4. Click on templates to see detailed views
5. Test GitHub authentication (if configured)
6. Try rating and reviewing templates (requires auth)

## 📚 Documentation

- **[API Documentation](docs/api.md)** - Complete REST API reference with examples
- **[Architecture Guide](docs/architecture.md)** - Detailed architecture documentation
- **[GitHub Repository](https://github.com/wsoule/dotfiles-web)** - Source code and issues

## 📁 Current Project Structure

```
├── main.go                          # Legacy monolithic file (being migrated)
├── cmd/server/                      # Application entry point (new architecture)
├── internal/                        # Private application code
│   ├── config/                      # Configuration management
│   ├── dto/                         # Data Transfer Objects + validation
│   ├── handlers/                    # HTTP request handlers
│   ├── middleware/                  # HTTP middleware
│   ├── models/                      # Domain models
│   ├── repository/                  # Data access layer
│   └── services/                    # Business logic
├── pkg/errors/                      # Custom error handling
├── docs/                            # API and architecture documentation
├── static/                          # Frontend assets (legacy)
│   ├── index.html                   # Home page
│   ├── templates.html               # Template browser
│   ├── profile.html                 # User profiles
│   ├── organizations.html           # Organization management
│   ├── docs.html                    # Documentation page
│   ├── styles.css                   # Main stylesheet
│   ├── universal-header.css         # Shared header styles
│   └── app.js                       # JavaScript
├── go.mod                           # Go module dependencies
├── go.sum                           # Go module checksums
└── README.md                        # This file
```

## 🎯 Key Features in Detail

### Template Categories
- **Web Development** - Frontend, backend, full-stack setups
- **Data Science** - Python, R, Jupyter, analytics tools
- **DevOps** - Kubernetes, Docker, infrastructure tools
- **Mobile Development** - iOS, Android, React Native
- **Backend Development** - Server frameworks and databases
- **Minimal Setups** - Essential tools for any developer

### Rating System
- 5-star ratings with half-star precision
- Aggregate ratings with distribution charts
- Review comments with helpful voting
- User-specific review management

### Search & Discovery
- Full-text search across template metadata
- Tag-based filtering with autocomplete
- Category browsing with counts
- Featured template promotion
- Sorting by popularity, date, name, author

### User Experience
- Responsive design for all devices
- Dark theme optimized for developers
- Grid and list view options
- Modal dialogs for detailed views
- Toast notifications for actions
- Progressive loading with fallbacks

## 🚀 Migration Status

This application is currently being migrated from a monolithic structure to clean architecture:

### ✅ **Completed**
- Domain models extraction and separation
- Error handling package with custom types
- DTOs with comprehensive validation
- Configuration management system
- Middleware package (auth, CORS, logging, rate limiting)
- Repository pattern interfaces
- Handler package structure
- Comprehensive API documentation

### 🔄 **In Progress**
- Service layer implementation
- Complete repository implementations
- Migration from monolithic main.go
- Testing framework setup

### 📋 **Planned**
- Database integration (MongoDB/PostgreSQL)
- Advanced caching strategies
- Microservices decomposition
- Event-driven architecture

## 🤝 Contributing

1. Fork the repository at [github.com/wsoule/dotfiles-web](https://github.com/wsoule/dotfiles-web)
2. Create a feature branch
3. Make your changes following the clean architecture patterns
4. Test locally with both legacy and new endpoints
5. Submit a pull request

## 📄 License

This project is open source and available under the MIT License.

---

**Ready for modern frontend frameworks with a solid, well-documented backend foundation!** 🎉