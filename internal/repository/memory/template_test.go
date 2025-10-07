package memory

import (
	"context"
	"testing"
	"time"

	"dotfiles-api/internal/models"
	"dotfiles-api/internal/repository"
)

func TestCreateTemplate(t *testing.T) {
	repo := NewTemplateRepository()
	ctx := context.Background()

	template := &models.StoredTemplate{
		Template: models.Template{
			Taps:  []string{"homebrew/cask"},
			Brews: []string{"git", "node", "golang"},
			Casks: []string{"visual-studio-code"},
			Stow:  []string{"vim", "zsh"},
			Metadata: models.ShareMetadata{
				Name:        "Test Template",
				Description: "This is a test template for verification",
				Author:      "test-user",
				Version:     "1.0.0",
				Tags:        []string{"test", "development"},
			},
			Public:   true,
			Featured: false,
		},
	}

	// Test Create
	err := repo.Create(ctx, template)
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	if template.ID == "" {
		t.Error("Template ID should be generated")
	}

	if template.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}

	if template.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should be set")
	}

	// Test GetByID
	retrieved, err := repo.GetByID(ctx, template.ID)
	if err != nil {
		t.Fatalf("Failed to get template by ID: %v", err)
	}

	if retrieved.ID != template.ID {
		t.Errorf("Expected ID %s, got %s", template.ID, retrieved.ID)
	}

	if retrieved.Template.Metadata.Name != template.Template.Metadata.Name {
		t.Errorf("Expected name %s, got %s", template.Template.Metadata.Name, retrieved.Template.Metadata.Name)
	}

	if len(retrieved.Template.Brews) != len(template.Template.Brews) {
		t.Errorf("Expected %d brews, got %d", len(template.Template.Brews), len(retrieved.Template.Brews))
	}

	t.Logf("✓ Template created successfully with ID: %s", template.ID)
	t.Logf("✓ Template retrieved successfully")
}

func TestCreateTemplateWithCustomID(t *testing.T) {
	repo := NewTemplateRepository()
	ctx := context.Background()

	customID := "my-custom-template-id"
	template := &models.StoredTemplate{
		ID: customID,
		Template: models.Template{
			Brews: []string{"git"},
			Metadata: models.ShareMetadata{
				Name:        "Custom ID Template",
				Description: "Template with custom ID",
				Author:      "test-user",
				Version:     "1.0.0",
			},
		},
	}

	err := repo.Create(ctx, template)
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	if template.ID != customID {
		t.Errorf("Expected ID %s, got %s", customID, template.ID)
	}

	t.Logf("✓ Template created with custom ID: %s", template.ID)
}

func TestListTemplates(t *testing.T) {
	repo := NewTemplateRepository()
	ctx := context.Background()

	// Get initial count (may include sample templates)
	initial, err := repo.List(ctx, repository.TemplateFilters{})
	if err != nil {
		t.Fatalf("Failed to list initial templates: %v", err)
	}
	initialCount := len(initial)

	// Create multiple templates
	templates := []*models.StoredTemplate{
		{
			Template: models.Template{
				Brews: []string{"git"},
				Metadata: models.ShareMetadata{
					Name:        "Template 1",
					Description: "First template",
					Author:      "author1",
					Version:     "1.0.0",
					Tags:        []string{"tag1"},
				},
				Public:   true,
				Featured: true,
			},
		},
		{
			Template: models.Template{
				Brews: []string{"node"},
				Metadata: models.ShareMetadata{
					Name:        "Template 2",
					Description: "Second template",
					Author:      "author2",
					Version:     "1.0.0",
					Tags:        []string{"tag2"},
				},
				Public:   true,
				Featured: false,
			},
		},
	}

	for _, tmpl := range templates {
		if err := repo.Create(ctx, tmpl); err != nil {
			t.Fatalf("Failed to create template: %v", err)
		}
	}

	// Test listing all templates
	allTemplates, err := repo.List(ctx, repository.TemplateFilters{})
	if err != nil {
		t.Fatalf("Failed to list templates: %v", err)
	}

	expectedCount := initialCount + 2
	if len(allTemplates) != expectedCount {
		t.Errorf("Expected %d templates, got %d", expectedCount, len(allTemplates))
	}

	t.Logf("✓ Listed %d templates successfully (%d initial + 2 created)", len(allTemplates), initialCount)
}

func TestUpdateTemplate(t *testing.T) {
	repo := NewTemplateRepository()
	ctx := context.Background()

	template := &models.StoredTemplate{
		Template: models.Template{
			Metadata: models.ShareMetadata{
				Name:        "Original Name",
				Description: "Original description",
				Author:      "test-author",
				Version:     "1.0.0",
			},
		},
	}

	// Create
	if err := repo.Create(ctx, template); err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	createdAt := template.CreatedAt
	time.Sleep(10 * time.Millisecond) // Ensure time difference

	// Update
	template.Template.Metadata.Name = "Updated Name"
	if err := repo.Update(ctx, template); err != nil {
		t.Fatalf("Failed to update template: %v", err)
	}

	// Verify
	updated, err := repo.GetByID(ctx, template.ID)
	if err != nil {
		t.Fatalf("Failed to get updated template: %v", err)
	}

	if updated.Template.Metadata.Name != "Updated Name" {
		t.Errorf("Expected name 'Updated Name', got '%s'", updated.Template.Metadata.Name)
	}

	if !updated.UpdatedAt.After(createdAt) {
		t.Error("UpdatedAt should be after CreatedAt")
	}

	t.Logf("✓ Template updated successfully")
}

func TestDeleteTemplate(t *testing.T) {
	repo := NewTemplateRepository()
	ctx := context.Background()

	template := &models.StoredTemplate{
		Template: models.Template{
			Metadata: models.ShareMetadata{
				Name:        "To Delete",
				Description: "This will be deleted",
				Author:      "test-author",
				Version:     "1.0.0",
			},
		},
	}

	// Create
	if err := repo.Create(ctx, template); err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	// Delete
	if err := repo.Delete(ctx, template.ID); err != nil {
		t.Fatalf("Failed to delete template: %v", err)
	}

	// Verify deletion
	_, err := repo.GetByID(ctx, template.ID)
	if err == nil {
		t.Error("Expected error when getting deleted template")
	}

	t.Logf("✓ Template deleted successfully")
}
