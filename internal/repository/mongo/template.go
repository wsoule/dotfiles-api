package mongo

import (
	"context"
	"time"

	"dotfiles-api/internal/models"
	"dotfiles-api/internal/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TemplateRepository implements the TemplateRepository interface using MongoDB
type TemplateRepository struct {
	collection *mongo.Collection
}

// NewTemplateRepository creates a new template repository
func NewTemplateRepository(client *Client) *TemplateRepository {
	repo := &TemplateRepository{
		collection: client.Collection("templates"),
	}

	// Seed with default template if collection is empty
	repo.seedDefaultTemplate()

	return repo
}

// seedDefaultTemplate adds the essential developer setup template if no templates exist
func (r *TemplateRepository) seedDefaultTemplate() {
	ctx := context.Background()

	// Check if any templates exist
	count, err := r.collection.CountDocuments(ctx, bson.M{})
	if err != nil || count > 0 {
		return // Don't seed if templates already exist
	}

	now := time.Now()
	defaultTemplate := &models.StoredTemplate{
		ID: "essential-developer-setup",
		Template: models.Template{
			Taps: []string{
				"homebrew/cask-fonts",
			},
			Brews: []string{
				"git", "curl", "wget", "tree", "jq", "stow", "gh",
				"starship", "neovim", "tmux", "fzf", "ripgrep",
				"bat", "eza", "zoxide",
			},
			Casks: []string{
				"visual-studio-code", "ghostty", "raycast",
				"rectangle", "obsidian", "1password",
				"font-jetbrains-mono-nerd-font",
			},
			Stow: []string{"vim", "zsh", "tmux", "starship", "git"},
			Metadata: models.ShareMetadata{
				Name:        "Essential Developer Setup",
				Description: "Complete modern developer setup with CLI tools, shell enhancements, and essential apps with automated post-install configuration",
				Author:      "Dotfiles Manager",
				Version:     "1.0.0",
				Tags:        []string{"essential", "developer", "productivity", "shell", "cli"},
			},
			Public:   true,
			Featured: true,
			Hooks: &models.Hooks{
				PreInstall: []string{
					"brew update",
				},
				PostInstall: []string{
					"echo '✅ Installation complete! Run dotfiles stow to symlink your config files.'",
				},
				PreStow: []string{
					"echo '🔗 Creating symlinks...'",
				},
				PostStow: []string{
					"echo '✅ Dotfiles stowed successfully!'",
				},
			},
			PackageConfigs: map[string]models.PackageConfig{
				"starship": {
					PostInstall: []string{
						"echo 'eval \"$(starship init bash)\"' >> ~/.bashrc",
						"echo 'eval \"$(starship init zsh)\"' >> ~/.zshrc",
					},
				},
				"zoxide": {
					PostInstall: []string{
						"echo 'eval \"$(zoxide init bash)\"' >> ~/.bashrc",
						"echo 'eval \"$(zoxide init zsh)\"' >> ~/.zshrc",
					},
				},
				"fzf": {
					PostInstall: []string{
						"$(brew --prefix)/opt/fzf/install --key-bindings --completion --no-update-rc",
					},
				},
				"neovim": {
					PostInstall: []string{
						"mkdir -p ~/.config/nvim",
						"echo '-- Neovim configuration will be managed via stow' > ~/.config/nvim/init.lua",
					},
				},
				"tmux": {
					PostInstall: []string{
						"git clone https://github.com/tmux-plugins/tpm ~/.tmux/plugins/tpm || echo 'TPM already installed'",
					},
				},
			},
		},
		Downloads: 0,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Insert the default template
	_, _ = r.collection.InsertOne(ctx, defaultTemplate)
}

// Create stores a new template
func (r *TemplateRepository) Create(ctx context.Context, template *models.StoredTemplate) error {
	if template.ID == "" {
		template.ID = primitive.NewObjectID().Hex()
	}
	template.CreatedAt = time.Now()
	template.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, template)
	return err
}

// GetByID retrieves a template by ID
func (r *TemplateRepository) GetByID(ctx context.Context, id string) (*models.StoredTemplate, error) {
	var template models.StoredTemplate
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&template)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &template, nil
}

// Update updates an existing template
func (r *TemplateRepository) Update(ctx context.Context, template *models.StoredTemplate) error {
	template.UpdatedAt = time.Now()
	_, err := r.collection.ReplaceOne(ctx, bson.M{"_id": template.ID}, template)
	return err
}

// Delete removes a template
func (r *TemplateRepository) Delete(ctx context.Context, id string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

// List retrieves templates with filters
func (r *TemplateRepository) List(ctx context.Context, filters repository.TemplateFilters) ([]*models.StoredTemplate, error) {
	filter := bson.M{}

	// Apply filters
	if filters.Author != "" {
		filter["template.metadata.author"] = filters.Author
	}
	if filters.OrganizationID != "" {
		filter["template.organization_id"] = filters.OrganizationID
	}
	if filters.Featured != nil {
		filter["template.featured"] = *filters.Featured
	}
	if filters.Public != nil {
		filter["template.public"] = *filters.Public
	}
	if len(filters.Tags) > 0 {
		filter["template.metadata.tags"] = bson.M{"$in": filters.Tags}
	}

	// Sort options
	sortBy := "created_at"
	if filters.SortBy != "" {
		sortBy = filters.SortBy
	}
	sortOrder := -1 // desc
	if filters.SortOrder == "asc" {
		sortOrder = 1
	}

	opts := &options.FindOptions{
		Sort:  bson.D{{Key: sortBy, Value: sortOrder}},
		Limit: int64ptr(filters.Limit),
		Skip:  int64ptr(filters.Offset),
	}

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var templates []*models.StoredTemplate
	if err = cursor.All(ctx, &templates); err != nil {
		return nil, err
	}
	return templates, nil
}

// Search searches templates by query
func (r *TemplateRepository) Search(ctx context.Context, query string, limit, offset int) ([]*models.StoredTemplate, error) {
	filter := bson.M{
		"$text": bson.M{"$search": query},
	}

	opts := &options.FindOptions{
		Sort:  bson.D{{Key: "score", Value: bson.M{"$meta": "textScore"}}},
		Limit: int64ptr(limit),
		Skip:  int64ptr(offset),
	}

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var templates []*models.StoredTemplate
	if err = cursor.All(ctx, &templates); err != nil {
		return nil, err
	}
	return templates, nil
}

// GetByAuthor retrieves templates by author
func (r *TemplateRepository) GetByAuthor(ctx context.Context, authorID string, limit, offset int) ([]*models.StoredTemplate, error) {
	filter := bson.M{"template.metadata.author": authorID}

	opts := &options.FindOptions{
		Sort:  bson.D{{Key: "created_at", Value: -1}},
		Limit: int64ptr(limit),
		Skip:  int64ptr(offset),
	}

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var templates []*models.StoredTemplate
	if err = cursor.All(ctx, &templates); err != nil {
		return nil, err
	}
	return templates, nil
}

// GetByOrganization retrieves templates by organization
func (r *TemplateRepository) GetByOrganization(ctx context.Context, orgID string, limit, offset int) ([]*models.StoredTemplate, error) {
	filter := bson.M{"template.organization_id": orgID}

	opts := &options.FindOptions{
		Sort:  bson.D{{Key: "created_at", Value: -1}},
		Limit: int64ptr(limit),
		Skip:  int64ptr(offset),
	}

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var templates []*models.StoredTemplate
	if err = cursor.All(ctx, &templates); err != nil {
		return nil, err
	}
	return templates, nil
}

// GetFeatured retrieves featured templates
func (r *TemplateRepository) GetFeatured(ctx context.Context, limit int) ([]*models.StoredTemplate, error) {
	filter := bson.M{"template.featured": true, "template.public": true}

	opts := &options.FindOptions{
		Sort:  bson.D{{Key: "downloads", Value: -1}},
		Limit: int64ptr(limit),
	}

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var templates []*models.StoredTemplate
	if err = cursor.All(ctx, &templates); err != nil {
		return nil, err
	}
	return templates, nil
}

// IncrementDownloads increments the download count for a template
func (r *TemplateRepository) IncrementDownloads(ctx context.Context, id string) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$inc": bson.M{"downloads": 1}},
	)
	return err
}

// GetStats returns template statistics
func (r *TemplateRepository) GetStats(ctx context.Context) (*models.TemplateStats, error) {
	total, err := r.collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	featured, err := r.collection.CountDocuments(ctx, bson.M{"template.featured": true})
	if err != nil {
		return nil, err
	}

	// Calculate total downloads
	pipeline := []bson.M{
		{"$group": bson.M{
			"_id":   nil,
			"total": bson.M{"$sum": "$downloads"},
		}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var result struct {
		Total int `bson:"total"`
	}
	totalDownloads := 0
	if cursor.Next(ctx) {
		cursor.Decode(&result)
		totalDownloads = result.Total
	}

	// Count unique tags as categories
	pipeline = []bson.M{
		{"$unwind": "$template.metadata.tags"},
		{"$group": bson.M{"_id": "$template.metadata.tags"}},
		{"$count": "categories"},
	}

	cursor, err = r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var categoryResult struct {
		Categories int `bson:"categories"`
	}
	categories := 0
	if cursor.Next(ctx) {
		cursor.Decode(&categoryResult)
		categories = categoryResult.Categories
	}

	return &models.TemplateStats{
		TotalTemplates:    int(total),
		FeaturedTemplates: int(featured),
		TotalDownloads:    totalDownloads,
		Categories:        categories,
	}, nil
}

// GetRating returns template rating information
func (r *TemplateRepository) GetRating(ctx context.Context, templateID string) (*models.TemplateRating, error) {
	// This would typically come from a reviews collection
	// For now, return a placeholder
	return &models.TemplateRating{
		TemplateID:     templateID,
		AverageRating:  0.0,
		TotalRatings:   0,
		Distribution:   make(map[string]int),
	}, nil
}