package memory

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"dotfiles-api/internal/models"
	"dotfiles-api/internal/repository"
)

type TemplateRepository struct {
	templates map[string]*models.StoredTemplate
	mu        sync.RWMutex
}

func NewTemplateRepository() *TemplateRepository {
	repo := &TemplateRepository{
		templates: make(map[string]*models.StoredTemplate),
	}

	// Initialize with some sample templates
	repo.initSampleTemplates()

	return repo
}

func (r *TemplateRepository) initSampleTemplates() {
	now := time.Now()
	sampleTemplates := []*models.StoredTemplate{
		{
			ID: "web-dev-template",
			Template: models.Template{
				Taps:  []string{"homebrew/cask-fonts"},
				Brews: []string{"git", "node", "yarn", "docker", "docker-compose"},
				Casks: []string{"visual-studio-code", "google-chrome", "iterm2"},
				Stow:  []string{"vim", "zsh", "git"},
				Metadata: models.ShareMetadata{
					Name:        "Full Stack Web Developer",
					Description: "Complete setup for modern web development with Node.js, Docker, and essential tools",
					Author:      "dotfiles-community",
					Version:     "1.0.0",
					Tags:        []string{"web", "frontend", "backend", "docker", "javascript"},
				},
				Public:   true,
				Featured: true,
			},
			Downloads: 1523,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID: "data-science-template",
			Template: models.Template{
				Taps:  []string{},
				Brews: []string{"python", "r", "jupyter", "postgresql"},
				Casks: []string{"rstudio", "tableau"},
				Stow:  []string{"python", "jupyter"},
				Metadata: models.ShareMetadata{
					Name:        "Data Science Toolkit",
					Description: "Python, R, Jupyter notebooks, and analytics tools for data science work",
					Author:      "dotfiles-community",
					Version:     "1.0.0",
					Tags:        []string{"data-science", "python", "r", "analytics", "jupyter"},
				},
				Public:   true,
				Featured: true,
			},
			Downloads: 892,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID: "devops-template",
			Template: models.Template{
				Taps:  []string{},
				Brews: []string{"kubectl", "terraform", "ansible", "docker", "aws-cli"},
				Casks: []string{"docker"},
				Stow:  []string{"kubectl", "terraform"},
				Metadata: models.ShareMetadata{
					Name:        "DevOps Engineer Setup",
					Description: "Infrastructure as code, container orchestration, and cloud tools",
					Author:      "dotfiles-community",
					Version:     "1.0.0",
					Tags:        []string{"devops", "kubernetes", "terraform", "docker", "cloud"},
				},
				Public:   true,
				Featured: true,
			},
			Downloads: 1104,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID: "mobile-dev-template",
			Template: models.Template{
				Taps:  []string{},
				Brews: []string{"cocoapods", "fastlane", "node"},
				Casks: []string{"android-studio"},
				Stow:  []string{"vim", "git"},
				Metadata: models.ShareMetadata{
					Name:        "Mobile Developer Setup",
					Description: "Tools for iOS and Android development with React Native and Flutter",
					Author:      "dotfiles-community",
					Version:     "1.0.0",
					Tags:        []string{"mobile", "ios", "android", "react-native", "flutter"},
				},
				Public:   true,
				Featured: false,
			},
			Downloads: 634,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID: "backend-dev-template",
			Template: models.Template{
				Taps:  []string{},
				Brews: []string{"go", "postgresql", "redis", "nginx"},
				Casks: []string{"postman", "tableplus"},
				Stow:  []string{"vim", "zsh", "git"},
				Metadata: models.ShareMetadata{
					Name:        "Backend Developer Kit",
					Description: "Server-side development with Go, databases, and API tools",
					Author:      "dotfiles-community",
					Version:     "1.0.0",
					Tags:        []string{"backend", "go", "database", "api", "server"},
				},
				Public:   true,
				Featured: false,
			},
			Downloads: 445,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID: "minimal-template",
			Template: models.Template{
				Taps:  []string{},
				Brews: []string{"git", "curl", "wget", "vim"},
				Casks: []string{"visual-studio-code"},
				Stow:  []string{"vim", "git"},
				Metadata: models.ShareMetadata{
					Name:        "Minimal Developer Setup",
					Description: "Essential tools only - perfect starting point for any developer",
					Author:      "dotfiles-community",
					Version:     "1.0.0",
					Tags:        []string{"minimal", "essential", "starter"},
				},
				Public:   true,
				Featured: true,
			},
			Downloads: 2341,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	for _, template := range sampleTemplates {
		r.templates[template.ID] = template
	}
}

func (r *TemplateRepository) Create(ctx context.Context, template *models.StoredTemplate) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if template.ID == "" {
		template.ID = fmt.Sprintf("template-%d", time.Now().UnixNano())
	}

	template.CreatedAt = time.Now()
	template.UpdatedAt = time.Now()

	r.templates[template.ID] = template
	return nil
}

func (r *TemplateRepository) GetByID(ctx context.Context, id string) (*models.StoredTemplate, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	template, exists := r.templates[id]
	if !exists {
		return nil, repository.ErrNotFound
	}

	return template, nil
}

func (r *TemplateRepository) Update(ctx context.Context, template *models.StoredTemplate) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.templates[template.ID]; !exists {
		return repository.ErrNotFound
	}

	template.UpdatedAt = time.Now()
	r.templates[template.ID] = template
	return nil
}

func (r *TemplateRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.templates[id]; !exists {
		return repository.ErrNotFound
	}

	delete(r.templates, id)
	return nil
}

func (r *TemplateRepository) List(ctx context.Context, filters repository.TemplateFilters) ([]*models.StoredTemplate, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*models.StoredTemplate

	for _, template := range r.templates {
		// Apply filters
		if filters.Public != nil && template.Template.Public != *filters.Public {
			continue
		}

		if filters.Featured != nil && template.Template.Featured != *filters.Featured {
			continue
		}

		if filters.Author != "" && template.Template.Metadata.Author != filters.Author {
			continue
		}

		if filters.OrganizationID != "" && template.Template.OrganizationID != filters.OrganizationID {
			continue
		}

		if len(filters.Tags) > 0 {
			hasAllTags := true
			for _, filterTag := range filters.Tags {
				found := false
				for _, templateTag := range template.Template.Metadata.Tags {
					if templateTag == filterTag {
						found = true
						break
					}
				}
				if !found {
					hasAllTags = false
					break
				}
			}
			if !hasAllTags {
				continue
			}
		}

		result = append(result, template)
	}

	// Apply limit and offset
	if filters.Offset > 0 && filters.Offset < len(result) {
		result = result[filters.Offset:]
	} else if filters.Offset >= len(result) {
		result = []*models.StoredTemplate{}
	}

	if filters.Limit > 0 && filters.Limit < len(result) {
		result = result[:filters.Limit]
	}

	return result, nil
}

func (r *TemplateRepository) Search(ctx context.Context, query string, limit, offset int) ([]*models.StoredTemplate, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*models.StoredTemplate
	lowerQuery := strings.ToLower(query)

	for _, template := range r.templates {
		// Simple search in name and description
		if strings.Contains(strings.ToLower(template.Template.Metadata.Name), lowerQuery) ||
			strings.Contains(strings.ToLower(template.Template.Metadata.Description), lowerQuery) ||
			strings.Contains(strings.ToLower(template.Template.Metadata.Author), lowerQuery) {
			result = append(result, template)
		}
	}

	// Apply offset and limit
	if offset > 0 && offset < len(result) {
		result = result[offset:]
	} else if offset >= len(result) {
		result = []*models.StoredTemplate{}
	}

	if limit > 0 && limit < len(result) {
		result = result[:limit]
	}

	return result, nil
}

func (r *TemplateRepository) GetByAuthor(ctx context.Context, authorID string, limit, offset int) ([]*models.StoredTemplate, error) {
	filters := repository.TemplateFilters{
		Author: authorID,
		Limit:  limit,
		Offset: offset,
	}
	return r.List(ctx, filters)
}

func (r *TemplateRepository) GetByOrganization(ctx context.Context, orgID string, limit, offset int) ([]*models.StoredTemplate, error) {
	filters := repository.TemplateFilters{
		OrganizationID: orgID,
		Limit:          limit,
		Offset:         offset,
	}
	return r.List(ctx, filters)
}

func (r *TemplateRepository) GetFeatured(ctx context.Context, limit int) ([]*models.StoredTemplate, error) {
	featured := true
	filters := repository.TemplateFilters{
		Featured: &featured,
		Limit:    limit,
	}
	return r.List(ctx, filters)
}

func (r *TemplateRepository) IncrementDownloads(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	template, exists := r.templates[id]
	if !exists {
		return repository.ErrNotFound
	}

	template.Downloads++
	return nil
}

func (r *TemplateRepository) GetStats(ctx context.Context) (*models.TemplateStats, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stats := &models.TemplateStats{
		TotalTemplates: len(r.templates),
	}

	for _, template := range r.templates {
		if template.Template.Featured {
			stats.FeaturedTemplates++
		}
		stats.TotalDownloads += template.Downloads
	}

	// Count unique tags as categories
	tagSet := make(map[string]bool)
	for _, template := range r.templates {
		for _, tag := range template.Template.Metadata.Tags {
			tagSet[tag] = true
		}
	}
	stats.Categories = len(tagSet)

	return stats, nil
}

func (r *TemplateRepository) GetRating(ctx context.Context, templateID string) (*models.TemplateRating, error) {
	// For in-memory repository, return empty rating
	// This would need a review repository integration in a full implementation
	return &models.TemplateRating{
		TemplateID:     templateID,
		AverageRating:  0,
		TotalRatings:   0,
		Distribution:   make(map[string]int),
	}, nil
}
