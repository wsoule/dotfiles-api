# Dotfiles Config Sharing API

Simple Go web API for sharing dotfiles configurations. Designed for Railway deployment.

## Features

- 🌐 **Minimal Web Frontend** - Clean interface to browse all configurations
- 📤 **Upload/download** dotfiles configurations
- 🔍 **Search** public configurations with real-time filtering
- ⭐ **Featured** configurations (most downloaded)
- 📊 **Statistics** dashboard with config counts and download metrics
- 💾 **Simple storage** (in-memory for demo, easily replaceable with database)
- 🌍 **CORS enabled** for web frontend integration
- 📋 **Detailed package display** - Shows brews, casks, taps, and stow packages

## API Endpoints

- `POST /api/configs/upload` - Upload a new config
- `GET /api/configs/:id` - Download a config by ID
- `GET /api/configs/search?q=query` - Search public configs
- `GET /api/configs/featured` - Get featured configs
- `GET /api/configs/stats` - Get platform statistics

## Environment Variables

- `PORT` - Server port (default: 8080, automatically set by Railway)
- `MONGODB_URI` - MongoDB connection string (optional, uses in-memory storage if not provided)
- `MONGODB_DATABASE` - MongoDB database name (default: "dotfiles")

## Local Development

```bash
go mod tidy
go run main.go
```


Server will start on http://localhost:8080

**🌐 Open http://localhost:8080 in your browser to see the web interface!**

## Railway Deployment

1. **Connect to Railway**
   - Connect your GitHub repo to Railway
   - Railway will automatically detect this as a Go app

2. **Add MongoDB Database**
   - In Railway dashboard, add MongoDB as a service
   - Copy the MongoDB connection string

3. **Set Environment Variables**
   - `MONGODB_URI` - The MongoDB connection string from step 2
   - `MONGODB_DATABASE` - "dotfiles" (or your preferred database name)
   - `GIN_MODE` - "release" (for production)

4. **Deploy!**
   - Railway will automatically build and deploy your app
   - MongoDB will persist all uploaded configurations

### MongoDB Setup
The app will automatically:
- Connect to MongoDB on startup
- Create collections as needed
- Fall back to in-memory storage if MongoDB is unavailable
- Seed initial demo data if the database is empty

## Testing

```bash
# Upload a config
curl -X POST http://localhost:8080/api/configs/upload \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Config",
    "description": "A test configuration",
    "author": "testuser",
    "tags": ["test", "demo"],
    "config": "{\"brews\":[\"git\",\"curl\"],\"casks\":[\"vscode\"],\"taps\":[],\"stow\":[\"vim\"]}",
    "public": true
  }'

# Search configs
curl http://localhost:8080/api/configs/search?q=test

# Get featured configs
curl http://localhost:8080/api/configs/featured
```
