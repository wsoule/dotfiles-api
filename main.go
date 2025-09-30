package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Config struct {
	Brews []string `json:"brews"`
	Casks []string `json:"casks"`
	Taps  []string `json:"taps"`
	Stow  []string `json:"stow"`
}

type ShareableConfig struct {
	Config   `json:",inline"`
	Metadata ShareMetadata `json:"metadata"`
}

type ShareMetadata struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Author      string    `json:"author"`
	Tags        []string  `json:"tags"`
	CreatedAt   time.Time `json:"created_at"`
	Version     string    `json:"version"`
}

type Template struct {
	Taps      []string         `json:"taps"`
	Brews     []string         `json:"brews"`
	Casks     []string         `json:"casks"`
	Stow      []string         `json:"stow"`
	Metadata  ShareMetadata    `json:"metadata"`
	Extends   string           `json:"extends,omitempty"`
	Overrides []string         `json:"overrides,omitempty"`
	AddOnly   bool             `json:"addOnly"`
	Public    bool             `json:"public"`
	Featured  bool             `json:"featured"`
}

type StoredTemplate struct {
	ID           string        `json:"id" bson:"_id"`
	Template     Template      `json:"template" bson:"template"`
	CreatedAt    time.Time     `json:"created_at" bson:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at" bson:"updated_at"`
	Downloads    int           `json:"downloads" bson:"downloads"`
}

type StoredConfig struct {
	ID           string          `json:"id" bson:"_id"`
	Config       ShareableConfig `json:"config" bson:"config"`
	Public       bool            `json:"public" bson:"public"`
	CreatedAt    time.Time       `json:"created_at" bson:"created_at"`
	DownloadCount int            `json:"download_count" bson:"download_count"`
}

// Storage interface
type ConfigStorage interface {
	Store(config *StoredConfig) error
	Get(id string) (*StoredConfig, error)
	Search(query string, publicOnly bool) ([]*StoredConfig, error)
	GetStats() (total, public, downloads int, error error)
	IncrementDownloads(id string) error
}

// Template storage interface
type TemplateStorage interface {
	StoreTemplate(template *StoredTemplate) error
	GetTemplate(id string) (*StoredTemplate, error)
	SearchTemplates(search, tags string, featured *bool) ([]*StoredTemplate, error)
	IncrementTemplateDownloads(id string) error
}

// In-memory storage (fallback)
type MemoryStorage struct {
	configs   map[string]*StoredConfig
	templates map[string]*StoredTemplate
	mu        sync.RWMutex
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		configs:   make(map[string]*StoredConfig),
		templates: make(map[string]*StoredTemplate),
	}
}

func (m *MemoryStorage) Store(config *StoredConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.configs[config.ID] = config
	return nil
}

func (m *MemoryStorage) Get(id string) (*StoredConfig, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	config, exists := m.configs[id]
	if !exists {
		return nil, fmt.Errorf("config not found")
	}
	return config, nil
}

func (m *MemoryStorage) Search(query string, publicOnly bool) ([]*StoredConfig, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var results []*StoredConfig
	queryLower := strings.ToLower(query)

	for _, stored := range m.configs {
		if publicOnly && !stored.Public {
			continue
		}

		searchText := strings.ToLower(stored.Config.Metadata.Name + " " +
			stored.Config.Metadata.Description + " " +
			strings.Join(stored.Config.Metadata.Tags, " "))

		if query == "" || strings.Contains(searchText, queryLower) {
			results = append(results, stored)
		}
	}
	return results, nil
}

func (m *MemoryStorage) GetStats() (int, int, int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	total := len(m.configs)
	public := 0
	downloads := 0

	for _, stored := range m.configs {
		if stored.Public {
			public++
		}
		downloads += stored.DownloadCount
	}

	return total, public, downloads, nil
}

func (m *MemoryStorage) IncrementDownloads(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if config, exists := m.configs[id]; exists {
		config.DownloadCount++
	}
	return nil
}

// Template methods for MemoryStorage
func (m *MemoryStorage) StoreTemplate(template *StoredTemplate) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.templates[template.ID] = template
	return nil
}

func (m *MemoryStorage) GetTemplate(id string) (*StoredTemplate, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	template, exists := m.templates[id]
	if !exists {
		return nil, fmt.Errorf("template not found")
	}
	return template, nil
}

func (m *MemoryStorage) SearchTemplates(search, tags string, featured *bool) ([]*StoredTemplate, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var results []*StoredTemplate
	searchLower := strings.ToLower(search)
	var tagsList []string
	if tags != "" {
		tagsList = strings.Split(strings.ToLower(tags), ",")
	}

	for _, template := range m.templates {
		// Filter by public status
		if !template.Template.Public {
			continue
		}

		// Filter by featured if specified
		if featured != nil && template.Template.Featured != *featured {
			continue
		}

		// Search in name and description
		if search != "" {
			searchText := strings.ToLower(template.Template.Metadata.Name + " " + template.Template.Metadata.Description)
			if !strings.Contains(searchText, searchLower) {
				continue
			}
		}

		// Filter by tags if specified
		if len(tagsList) > 0 {
			tagMatch := false
			for _, templateTag := range template.Template.Metadata.Tags {
				for _, searchTag := range tagsList {
					if strings.Contains(strings.ToLower(templateTag), strings.TrimSpace(searchTag)) {
						tagMatch = true
						break
					}
				}
				if tagMatch {
					break
				}
			}
			if !tagMatch {
				continue
			}
		}

		results = append(results, template)
	}
	return results, nil
}

func (m *MemoryStorage) IncrementTemplateDownloads(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if template, exists := m.templates[id]; exists {
		template.Downloads++
	}
	return nil
}

// MongoDB storage
type MongoStorage struct {
	collection         *mongo.Collection
	templateCollection *mongo.Collection
}

func NewMongoStorage(mongoURI, dbName string) (*MongoStorage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		return nil, err
	}

	// Test connection
	if err = client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	collection := client.Database(dbName).Collection("configs")
	templateCollection := client.Database(dbName).Collection("templates")
	return &MongoStorage{
		collection:         collection,
		templateCollection: templateCollection,
	}, nil
}

func (m *MongoStorage) Store(config *StoredConfig) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := m.collection.InsertOne(ctx, config)
	return err
}

func (m *MongoStorage) Get(id string) (*StoredConfig, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var config StoredConfig
	err := m.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (m *MongoStorage) Search(query string, publicOnly bool) ([]*StoredConfig, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{}
	if publicOnly {
		filter["public"] = true
	}

	if query != "" {
		filter["$or"] = bson.A{
			bson.M{"config.metadata.name": bson.M{"$regex": query, "$options": "i"}},
			bson.M{"config.metadata.description": bson.M{"$regex": query, "$options": "i"}},
			bson.M{"config.metadata.tags": bson.M{"$regex": query, "$options": "i"}},
		}
	}

	cursor, err := m.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []*StoredConfig
	for cursor.Next(ctx) {
		var config StoredConfig
		if err := cursor.Decode(&config); err != nil {
			continue
		}
		results = append(results, &config)
	}

	return results, nil
}

func (m *MongoStorage) GetStats() (int, int, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	total, err := m.collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return 0, 0, 0, err
	}

	publicCount, err := m.collection.CountDocuments(ctx, bson.M{"public": true})
	if err != nil {
		return 0, 0, 0, err
	}

	// Calculate total downloads
	pipeline := []bson.M{
		{"$group": bson.M{"_id": nil, "total": bson.M{"$sum": "$download_count"}}},
	}

	cursor, err := m.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return int(total), int(publicCount), 0, nil
	}
	defer cursor.Close(ctx)

	var result struct {
		Total int `bson:"total"`
	}
	downloads := 0
	if cursor.Next(ctx) {
		cursor.Decode(&result)
		downloads = result.Total
	}

	return int(total), int(publicCount), downloads, nil
}

func (m *MongoStorage) IncrementDownloads(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := m.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$inc": bson.M{"download_count": 1}},
	)
	return err
}

// Template methods for MongoStorage
func (m *MongoStorage) StoreTemplate(template *StoredTemplate) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := m.templateCollection.InsertOne(ctx, template)
	return err
}

func (m *MongoStorage) GetTemplate(id string) (*StoredTemplate, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var template StoredTemplate
	err := m.templateCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&template)
	if err != nil {
		return nil, err
	}
	return &template, nil
}

func (m *MongoStorage) SearchTemplates(search, tags string, featured *bool) ([]*StoredTemplate, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"template.public": true}

	// Add featured filter if specified
	if featured != nil {
		filter["template.featured"] = *featured
	}

	// Add search filter
	if search != "" {
		filter["$or"] = bson.A{
			bson.M{"template.metadata.name": bson.M{"$regex": search, "$options": "i"}},
			bson.M{"template.metadata.description": bson.M{"$regex": search, "$options": "i"}},
		}
	}

	// Add tags filter
	if tags != "" {
		tagsList := strings.Split(tags, ",")
		for i, tag := range tagsList {
			tagsList[i] = strings.TrimSpace(tag)
		}
		filter["template.metadata.tags"] = bson.M{"$in": tagsList}
	}

	cursor, err := m.templateCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []*StoredTemplate
	for cursor.Next(ctx) {
		var template StoredTemplate
		if err := cursor.Decode(&template); err != nil {
			continue
		}
		results = append(results, &template)
	}

	return results, nil
}

func (m *MongoStorage) IncrementTemplateDownloads(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := m.templateCollection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$inc": bson.M{"downloads": 1}},
	)
	return err
}

var storage ConfigStorage
var templateStorage TemplateStorage

func seedData() {
	// Check if we already have data
	total, _, _, err := storage.GetStats()
	if err == nil && total > 0 {
		return // Already have data
	}

	// Add some seed data
	seedConfigs := []struct {
		config ShareableConfig
		public bool
	}{
		{
			config: ShareableConfig{
				Config: Config{
					Brews: []string{"git", "curl", "wget", "tree", "jq", "node", "npm", "python3", "docker"},
					Casks: []string{"visual-studio-code", "google-chrome", "iterm2", "rectangle", "figma"},
					Taps:  []string{"homebrew/cask-fonts"},
					Stow:  []string{"git", "zsh", "vim", "vscode"},
				},
				Metadata: ShareMetadata{
					Name:        "Full Stack Web Developer",
					Description: "Complete setup for modern web development with Node.js, Python, and essential tools",
					Author:      "webdev_pro",
					Tags:        []string{"web-dev", "javascript", "python", "docker", "frontend"},
					CreatedAt:   time.Now().AddDate(0, 0, -7),
					Version:     "1.0.0",
				},
			},
			public: true,
		},
		{
			config: ShareableConfig{
				Config: Config{
					Brews: []string{"git", "python3", "r", "jupyter", "postgresql", "sqlite"},
					Casks: []string{"visual-studio-code", "rstudio", "tableau-public", "docker"},
					Taps:  []string{"homebrew/cask-fonts"},
					Stow:  []string{"git", "zsh", "vim", "python", "jupyter"},
				},
				Metadata: ShareMetadata{
					Name:        "Data Science Toolkit",
					Description: "Python, R, Jupyter, and analytics tools for data scientists and researchers",
					Author:      "data_scientist",
					Tags:        []string{"data-science", "python", "r", "jupyter", "analytics", "ml"},
					CreatedAt:   time.Now().AddDate(0, 0, -3),
					Version:     "1.0.0",
				},
			},
			public: true,
		},
		{
			config: ShareableConfig{
				Config: Config{
					Brews: []string{"git", "curl", "kubectl", "terraform", "ansible", "aws-cli", "docker"},
					Casks: []string{"visual-studio-code", "iterm2", "lens", "postman"},
					Taps:  []string{"hashicorp/tap"},
					Stow:  []string{"git", "zsh", "kubectl", "terraform"},
				},
				Metadata: ShareMetadata{
					Name:        "DevOps Engineer Setup",
					Description: "Infrastructure, containerization, and cloud tools for DevOps workflows",
					Author:      "devops_master",
					Tags:        []string{"devops", "kubernetes", "terraform", "aws", "docker", "infrastructure"},
					CreatedAt:   time.Now().AddDate(0, 0, -1),
					Version:     "1.0.0",
				},
			},
			public: true,
		},
	}

	for _, seed := range seedConfigs {
		id := uuid.New().String()
		storedConfig := &StoredConfig{
			ID:           id,
			Config:       seed.config,
			Public:       seed.public,
			CreatedAt:    seed.config.Metadata.CreatedAt,
			DownloadCount: int(time.Since(seed.config.Metadata.CreatedAt).Hours() / 24), // Simulate downloads
		}
		storage.Store(storedConfig)
	}
}

func main() {
	// Initialize storage (MongoDB or fallback to memory)
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI != "" {
		dbName := os.Getenv("MONGODB_DATABASE")
		if dbName == "" {
			dbName = "dotfiles"
		}

		mongoStorage, err := NewMongoStorage(mongoURI, dbName)
		if err != nil {
			log.Printf("Failed to connect to MongoDB: %v, falling back to memory storage", err)
			memStorage := NewMemoryStorage()
			storage = memStorage
			templateStorage = memStorage
		} else {
			storage = mongoStorage
			templateStorage = mongoStorage
			log.Println("Connected to MongoDB successfully")
		}
	} else {
		memStorage := NewMemoryStorage()
		storage = memStorage
		templateStorage = memStorage
		log.Println("Using in-memory storage (set MONGODB_URI for persistent storage)")
	}

	// Seed initial data
	seedData()

	// Get port from environment (Railway sets PORT)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Initialize Gin router
	r := gin.Default()

	// Add CORS middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type,Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Serve static files
	r.Static("/static", "./static")

	// Serve frontend
	r.GET("/", func(c *gin.Context) {
		c.File("./static/index.html")
	})

	// API routes
	api := r.Group("/api")
	{
		// Upload config
		api.POST("/configs/upload", uploadConfig)

		// Get config by ID
		api.GET("/configs/:id", getConfig)

		// Search configs
		api.GET("/configs/search", searchConfigs)

		// Get featured configs
		api.GET("/configs/featured", getFeaturedConfigs)

		// Get stats
		api.GET("/configs/stats", getStats)

		// Template endpoints
		api.POST("/templates", createTemplate)
		api.GET("/templates", getTemplates)
		api.GET("/templates/:id", getTemplate)
	}

	// Web interface routes
	r.GET("/config/:id", func(c *gin.Context) {
		id := c.Param("id")
		c.Redirect(302, "/api/configs/"+id)
	})

	// Template detail page route
	r.GET("/template/:id", func(c *gin.Context) {
		c.File("./static/template.html")
	})

	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func uploadConfig(c *gin.Context) {
	var req struct {
		Name        string   `json:"name" binding:"required"`
		Description string   `json:"description"`
		Author      string   `json:"author"`
		Tags        []string `json:"tags"`
		Config      string   `json:"config" binding:"required"`
		Public      bool     `json:"public"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	// Parse the config JSON
	var shareableConfig ShareableConfig
	if err := json.Unmarshal([]byte(req.Config), &shareableConfig); err != nil {
		c.JSON(400, gin.H{"error": "Invalid config JSON", "details": err.Error()})
		return
	}

	// Create stored config
	id := uuid.New().String()
	stored := &StoredConfig{
		ID:           id,
		Config:       shareableConfig,
		Public:       req.Public,
		CreatedAt:    time.Now(),
		DownloadCount: 0,
	}

	// Store in database
	if err := storage.Store(stored); err != nil {
		c.JSON(500, gin.H{"error": "Failed to store config", "details": err.Error()})
		return
	}

	log.Printf("Config uploaded: %s (%s)", req.Name, id)

	c.JSON(201, gin.H{
		"id":  id,
		"url": fmt.Sprintf("/config/%s", id),
	})
}

func getConfig(c *gin.Context) {
	id := c.Param("id")

	stored, err := storage.Get(id)
	if err != nil {
		c.JSON(404, gin.H{"error": "Config not found"})
		return
	}

	// Increment download count
	storage.IncrementDownloads(id)

	c.JSON(200, stored.Config)
}

func searchConfigs(c *gin.Context) {
	query := c.Query("q")

	configs, err := storage.Search(query, true) // Only public configs
	if err != nil {
		c.JSON(500, gin.H{"error": "Search failed", "details": err.Error()})
		return
	}

	var results []gin.H
	for _, stored := range configs {
		results = append(results, gin.H{
			"id":          stored.ID,
			"html_url":    fmt.Sprintf("/config/%s", stored.ID),
			"description": fmt.Sprintf("Dotfiles Config: %s", stored.Config.Metadata.Name),
			"public":      stored.Public,
			"created_at":  stored.CreatedAt,
			"updated_at":  stored.CreatedAt,
			"files": map[string]interface{}{
				"dotfiles-config.json": gin.H{
					"content": "config content",
				},
			},
			"owner": gin.H{
				"login":      stored.Config.Metadata.Author,
				"avatar_url": "",
			},
		})
	}

	c.JSON(200, gin.H{
		"total_count":         len(results),
		"incomplete_results": false,
		"items":              results,
	})
}

func getFeaturedConfigs(c *gin.Context) {
	// Get public configs and sort by downloads
	configs, err := storage.Search("", true) // Get all public configs
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to get featured configs", "details": err.Error()})
		return
	}

	var featured []gin.H
	for _, stored := range configs {
		featured = append(featured, gin.H{
			"name":        stored.Config.Metadata.Name,
			"description": stored.Config.Metadata.Description,
			"author":      stored.Config.Metadata.Author,
			"url":         fmt.Sprintf("/config/%s", stored.ID),
			"tags":        stored.Config.Metadata.Tags,
			"downloads":   stored.DownloadCount,
		})
	}

	// Sort by download count (simple bubble sort for demo)
	for i := 0; i < len(featured); i++ {
		for j := i + 1; j < len(featured); j++ {
			if featured[i]["downloads"].(int) < featured[j]["downloads"].(int) {
				featured[i], featured[j] = featured[j], featured[i]
			}
		}
	}

	// Limit to top 10
	if len(featured) > 10 {
		featured = featured[:10]
	}

	c.JSON(200, featured)
}

func getStats(c *gin.Context) {
	total, public, downloads, err := storage.GetStats()
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to get stats", "details": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"total_configs":    total,
		"public_configs":   public,
		"total_downloads":  downloads,
	})
}

func createTemplate(c *gin.Context) {
	var template Template
	if err := c.ShouldBindJSON(&template); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	// Validation
	if template.Metadata.Name == "" {
		c.JSON(400, gin.H{"error": "Template name is required"})
		return
	}

	// Create stored template
	id := uuid.New().String()
	now := time.Now()
	stored := &StoredTemplate{
		ID:        id,
		Template:  template,
		CreatedAt: now,
		UpdatedAt: now,
		Downloads: 0,
	}

	// Set metadata created_at if not provided
	if stored.Template.Metadata.CreatedAt.IsZero() {
		stored.Template.Metadata.CreatedAt = now
	}

	// Store in database
	if err := templateStorage.StoreTemplate(stored); err != nil {
		c.JSON(500, gin.H{"error": "Failed to store template", "details": err.Error()})
		return
	}

	log.Printf("Template created: %s (%s)", template.Metadata.Name, id)

	c.JSON(201, gin.H{
		"id":  id,
		"url": fmt.Sprintf("https://new-dotfiles-production.up.railway.app/templates/%s", id),
	})
}

func getTemplate(c *gin.Context) {
	id := c.Param("id")

	stored, err := templateStorage.GetTemplate(id)
	if err != nil {
		c.JSON(404, gin.H{"error": "Template not found"})
		return
	}

	// Increment download count
	templateStorage.IncrementTemplateDownloads(id)

	c.JSON(200, stored.Template)
}

func getTemplates(c *gin.Context) {
	// Parse query parameters
	search := c.Query("search")
	tags := c.Query("tags")
	featuredParam := c.Query("featured")

	var featured *bool
	if featuredParam != "" {
		if featuredParam == "true" {
			f := true
			featured = &f
		} else if featuredParam == "false" {
			f := false
			featured = &f
		}
	}

	// Search templates
	templates, err := templateStorage.SearchTemplates(search, tags, featured)
	if err != nil {
		c.JSON(500, gin.H{"error": "Search failed", "details": err.Error()})
		return
	}

	// Format response
	var results []gin.H
	for _, stored := range templates {
		results = append(results, gin.H{
			"id":          stored.ID,
			"name":        stored.Template.Metadata.Name,
			"description": stored.Template.Metadata.Description,
			"author":      stored.Template.Metadata.Author,
			"tags":        stored.Template.Metadata.Tags,
			"featured":    stored.Template.Featured,
			"downloads":   stored.Downloads,
			"updated_at":  stored.UpdatedAt,
		})
	}

	c.JSON(200, gin.H{
		"templates": results,
		"total":     len(results),
	})
}
