package memory

import (
	"context"
	"testing"

	"dotfiles-api/internal/models"
)

func TestCreateReview(t *testing.T) {
	repo := NewReviewRepository()
	ctx := context.Background()

	review := &models.Review{
		TemplateID: "template-1",
		UserID:     "user-1",
		Username:   "testuser",
		Rating:     5,
		Comment:    "Great template!",
	}

	err := repo.Create(ctx, review)
	if err != nil {
		t.Fatalf("Failed to create review: %v", err)
	}

	if review.ID == "" {
		t.Error("Review ID should be generated")
	}

	if review.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}

	if review.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should be set")
	}

	t.Logf("✓ Review created successfully with ID: %s", review.ID)
}

func TestGetReviewByID(t *testing.T) {
	repo := NewReviewRepository()
	ctx := context.Background()

	review := &models.Review{
		TemplateID: "template-1",
		UserID:     "user-1",
		Rating:     4,
		Comment:    "Good template",
	}

	if err := repo.Create(ctx, review); err != nil {
		t.Fatalf("Failed to create review: %v", err)
	}

	retrieved, err := repo.GetByID(ctx, review.ID)
	if err != nil {
		t.Fatalf("Failed to get review: %v", err)
	}

	if retrieved.ID != review.ID {
		t.Errorf("Expected ID %s, got %s", review.ID, retrieved.ID)
	}

	if retrieved.Rating != review.Rating {
		t.Errorf("Expected rating %d, got %d", review.Rating, retrieved.Rating)
	}

	t.Logf("✓ Review retrieved successfully")
}

func TestGetByTemplate(t *testing.T) {
	repo := NewReviewRepository()
	ctx := context.Background()

	templateID := "template-test-xyz"

	// Create multiple reviews for the same template
	reviews := []*models.Review{
		{TemplateID: templateID, UserID: "user-1", Rating: 5, Comment: "Excellent"},
		{TemplateID: templateID, UserID: "user-2", Rating: 4, Comment: "Good"},
		{TemplateID: "other-template-abc", UserID: "user-3", Rating: 3, Comment: "Ok"},
	}

	for i, r := range reviews {
		if err := repo.Create(ctx, r); err != nil {
			t.Fatalf("Failed to create review %d: %v", i, err)
		}
		t.Logf("Created review %d: ID=%s, TemplateID=%s", i, r.ID, r.TemplateID)
	}

	templateReviews, err := repo.GetByTemplate(ctx, templateID, 10, 0)
	if err != nil {
		t.Fatalf("Failed to get template reviews: %v", err)
	}

	for i, r := range templateReviews {
		t.Logf("Found review %d: ID=%s, TemplateID=%s", i, r.ID, r.TemplateID)
	}

	if len(templateReviews) != 2 {
		t.Errorf("Expected 2 reviews for template, got %d", len(templateReviews))
	}

	t.Logf("✓ Retrieved %d reviews for template", len(templateReviews))
}

func TestGetUserReviewForTemplate(t *testing.T) {
	repo := NewReviewRepository()
	ctx := context.Background()

	userID := "user-test"
	templateID := "template-test"

	review := &models.Review{
		TemplateID: templateID,
		UserID:     userID,
		Rating:     5,
		Comment:    "My review",
	}

	if err := repo.Create(ctx, review); err != nil {
		t.Fatalf("Failed to create review: %v", err)
	}

	found, err := repo.GetUserReviewForTemplate(ctx, userID, templateID)
	if err != nil {
		t.Fatalf("Failed to get user review: %v", err)
	}

	if found == nil {
		t.Error("Expected to find review, got nil")
	}

	if found.UserID != userID {
		t.Errorf("Expected user ID %s, got %s", userID, found.UserID)
	}

	// Test non-existent review
	notFound, err := repo.GetUserReviewForTemplate(ctx, "non-existent-user", templateID)
	if err != nil {
		t.Fatalf("Error getting non-existent review: %v", err)
	}

	if notFound != nil {
		t.Error("Expected nil for non-existent review")
	}

	t.Logf("✓ User review lookup working correctly")
}

func TestCalculateTemplateRating(t *testing.T) {
	repo := NewReviewRepository()
	ctx := context.Background()

	templateID := "template-rating-test-unique-12345"

	// Create reviews with different ratings
	reviews := []*models.Review{
		{TemplateID: templateID, UserID: "user-rating-1", Rating: 5},
		{TemplateID: templateID, UserID: "user-rating-2", Rating: 4},
		{TemplateID: templateID, UserID: "user-rating-3", Rating: 5},
		{TemplateID: templateID, UserID: "user-rating-4", Rating: 3},
	}

	for i, r := range reviews {
		if err := repo.Create(ctx, r); err != nil {
			t.Fatalf("Failed to create review %d: %v", i, err)
		}
		t.Logf("Created review %d: ID=%s, Rating=%d", i, r.ID, r.Rating)
	}

	// Verify all reviews were created
	allReviews, _ := repo.GetByTemplate(ctx, templateID, 100, 0)
	t.Logf("Total reviews found for template: %d", len(allReviews))
	for i, r := range allReviews {
		t.Logf("  Review %d: ID=%s, Rating=%d", i, r.ID, r.Rating)
	}

	rating, err := repo.CalculateTemplateRating(ctx, templateID)
	if err != nil {
		t.Fatalf("Failed to calculate rating: %v", err)
	}

	if rating.TotalRatings != 4 {
		t.Errorf("Expected 4 total ratings, got %d", rating.TotalRatings)
	}

	expectedAvg := (5.0 + 4.0 + 5.0 + 3.0) / 4.0
	if rating.AverageRating != expectedAvg {
		t.Errorf("Expected average rating %.2f, got %.2f", expectedAvg, rating.AverageRating)
	}

	if rating.Distribution["5"] != 2 {
		t.Errorf("Expected 2 five-star ratings, got %d", rating.Distribution["5"])
	}

	if rating.Distribution["4"] != 1 {
		t.Errorf("Expected 1 four-star rating, got %d", rating.Distribution["4"])
	}

	if rating.Distribution["3"] != 1 {
		t.Errorf("Expected 1 three-star rating, got %d", rating.Distribution["3"])
	}

	t.Logf("✓ Rating calculation correct: %.2f average from %d reviews", rating.AverageRating, rating.TotalRatings)
}

func TestIncrementHelpful(t *testing.T) {
	repo := NewReviewRepository()
	ctx := context.Background()

	review := &models.Review{
		TemplateID: "template-1",
		UserID:     "user-1",
		Rating:     5,
		Helpful:    0,
	}

	if err := repo.Create(ctx, review); err != nil {
		t.Fatalf("Failed to create review: %v", err)
	}

	// Increment helpful count
	if err := repo.IncrementHelpful(ctx, review.ID); err != nil {
		t.Fatalf("Failed to increment helpful: %v", err)
	}

	retrieved, err := repo.GetByID(ctx, review.ID)
	if err != nil {
		t.Fatalf("Failed to get review: %v", err)
	}

	if retrieved.Helpful != 1 {
		t.Errorf("Expected helpful count 1, got %d", retrieved.Helpful)
	}

	t.Logf("✓ Helpful count incremented successfully")
}

func TestUpdateReview(t *testing.T) {
	repo := NewReviewRepository()
	ctx := context.Background()

	review := &models.Review{
		TemplateID: "template-1",
		UserID:     "user-1",
		Rating:     3,
		Comment:    "Original comment",
	}

	if err := repo.Create(ctx, review); err != nil {
		t.Fatalf("Failed to create review: %v", err)
	}

	// Update review
	review.Rating = 5
	review.Comment = "Updated comment"

	if err := repo.Update(ctx, review); err != nil {
		t.Fatalf("Failed to update review: %v", err)
	}

	retrieved, err := repo.GetByID(ctx, review.ID)
	if err != nil {
		t.Fatalf("Failed to get review: %v", err)
	}

	if retrieved.Rating != 5 {
		t.Errorf("Expected rating 5, got %d", retrieved.Rating)
	}

	if retrieved.Comment != "Updated comment" {
		t.Errorf("Expected updated comment, got %s", retrieved.Comment)
	}

	t.Logf("✓ Review updated successfully")
}

func TestDeleteReview(t *testing.T) {
	repo := NewReviewRepository()
	ctx := context.Background()

	review := &models.Review{
		TemplateID: "template-1",
		UserID:     "user-1",
		Rating:     5,
	}

	if err := repo.Create(ctx, review); err != nil {
		t.Fatalf("Failed to create review: %v", err)
	}

	if err := repo.Delete(ctx, review.ID); err != nil {
		t.Fatalf("Failed to delete review: %v", err)
	}

	_, err := repo.GetByID(ctx, review.ID)
	if err == nil {
		t.Error("Expected error when getting deleted review")
	}

	t.Logf("✓ Review deleted successfully")
}
