package memory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"dotfiles-api/internal/models"
	"dotfiles-api/internal/repository"
)

type ReviewRepository struct {
	reviews map[string]*models.Review
	mu      sync.RWMutex
}

func NewReviewRepository() *ReviewRepository {
	return &ReviewRepository{
		reviews: make(map[string]*models.Review),
	}
}

func (r *ReviewRepository) Create(ctx context.Context, review *models.Review) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if review.ID == "" {
		review.ID = fmt.Sprintf("review-%d", time.Now().UnixNano())
	}

	review.CreatedAt = time.Now()
	review.UpdatedAt = time.Now()

	r.reviews[review.ID] = review
	return nil
}

func (r *ReviewRepository) GetByID(ctx context.Context, id string) (*models.Review, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	review, exists := r.reviews[id]
	if !exists {
		return nil, repository.ErrNotFound
	}

	return review, nil
}

func (r *ReviewRepository) Update(ctx context.Context, review *models.Review) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.reviews[review.ID]; !exists {
		return repository.ErrNotFound
	}

	review.UpdatedAt = time.Now()
	r.reviews[review.ID] = review
	return nil
}

func (r *ReviewRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.reviews[id]; !exists {
		return repository.ErrNotFound
	}

	delete(r.reviews, id)
	return nil
}

func (r *ReviewRepository) GetByTemplate(ctx context.Context, templateID string, limit, offset int) ([]*models.Review, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*models.Review

	for _, review := range r.reviews {
		if review.TemplateID == templateID {
			result = append(result, review)
		}
	}

	// Apply offset and limit
	if offset > 0 && offset < len(result) {
		result = result[offset:]
	} else if offset >= len(result) {
		result = []*models.Review{}
	}

	if limit > 0 && limit < len(result) {
		result = result[:limit]
	}

	return result, nil
}

func (r *ReviewRepository) GetByUser(ctx context.Context, userID string, limit, offset int) ([]*models.Review, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*models.Review

	for _, review := range r.reviews {
		if review.UserID == userID {
			result = append(result, review)
		}
	}

	// Apply offset and limit
	if offset > 0 && offset < len(result) {
		result = result[offset:]
	} else if offset >= len(result) {
		result = []*models.Review{}
	}

	if limit > 0 && limit < len(result) {
		result = result[:limit]
	}

	return result, nil
}

func (r *ReviewRepository) GetUserReviewForTemplate(ctx context.Context, userID, templateID string) (*models.Review, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, review := range r.reviews {
		if review.UserID == userID && review.TemplateID == templateID {
			return review, nil
		}
	}

	return nil, nil // No review found
}

func (r *ReviewRepository) IncrementHelpful(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	review, exists := r.reviews[id]
	if !exists {
		return repository.ErrNotFound
	}

	review.Helpful++
	return nil
}

func (r *ReviewRepository) CalculateTemplateRating(ctx context.Context, templateID string) (*models.TemplateRating, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rating := &models.TemplateRating{
		TemplateID:   templateID,
		Distribution: make(map[string]int),
	}

	var totalRating int
	var count int

	for _, review := range r.reviews {
		if review.TemplateID == templateID {
			totalRating += review.Rating
			count++
			rating.Distribution[fmt.Sprintf("%d", review.Rating)]++
		}
	}

	rating.TotalRatings = count
	if count > 0 {
		rating.AverageRating = float64(totalRating) / float64(count)
	}

	return rating, nil
}
