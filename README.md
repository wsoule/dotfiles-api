# Dotfiles Manager - Community Template Platform

A comprehensive Go web application for sharing and discovering development environment templates. Features GitHub authentication, community reviews, and advanced search capabilities. The "awesome" version of machine setup!

## 🌟 Key Features

### 🔐 **User Authentication**
- **GitHub OAuth 2.0 integration** - Secure sign-in with GitHub
- **User profiles** with avatars and metadata
- **Session management** with secure cookies
- **Protected endpoints** for user-specific actions

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

## 🏗️ Architecture

### Backend (Go + Gin)
- **RESTful API** with 20+ endpoints
- **Modular storage interface** supporting MongoDB and in-memory storage
- **Comprehensive data models** for users, templates, reviews, and ratings
- **Authentication middleware** protecting sensitive operations
- **Seeded data** for immediate functionality

### Frontend (HTML + CSS + JavaScript)
- **Vanilla JavaScript** for maximum compatibility
- **Modern CSS** with variables and responsive design
- **Component-based** template rendering
- **Real-time API integration**
- **Progressive enhancement** approach

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
- `GET /api/templates/:id/reviews` - Get template reviews
- `POST /api/templates/:id/reviews` - Create review (auth required)
- `GET /api/templates/:id/rating` - Get template rating

### Users & Favorites
- `GET /api/users/:username` - Get user profile
- `POST /api/users/favorites/:templateId` - Add to favorites (auth required)
- `DELETE /api/users/favorites/:templateId` - Remove from favorites (auth required)

### Reviews
- `PUT /api/reviews/:id` - Update review (auth required)
- `DELETE /api/reviews/:id` - Delete review (auth required)
- `POST /api/reviews/:id/helpful` - Mark review helpful (auth required)

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

## 📁 Project Structure

```
├── main.go                     # Main application with all backend logic
├── static/
│   ├── index.html             # Home page with config upload
│   ├── templates.html         # Template browser with advanced features
│   ├── docs.html              # Documentation page
│   ├── styles.css             # Main stylesheet with dark theme
│   ├── universal-header.css   # Shared header styles
│   └── app.js                 # JavaScript for home page
├── go.mod                     # Go module dependencies
├── go.sum                     # Go module checksums
└── README.md                  # This file
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

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test locally
5. Submit a pull request

## 📄 License

This project is open source and available under the MIT License.