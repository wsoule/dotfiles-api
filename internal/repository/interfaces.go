package repository

import (
	"context"
	"errors"

	"dotfiles-web/internal/models"
)

// Common repository errors
var (
	ErrNotFound      = errors.New("resource not found")
	ErrAlreadyExists = errors.New("resource already exists")
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id string) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByGitHubID(ctx context.Context, githubID int) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]*models.User, error)
	AddFavorite(ctx context.Context, userID, templateID string) error
	RemoveFavorite(ctx context.Context, userID, templateID string) error
	GetFavorites(ctx context.Context, userID string) ([]string, error)
}

type TemplateRepository interface {
	Create(ctx context.Context, template *models.StoredTemplate) error
	GetByID(ctx context.Context, id string) (*models.StoredTemplate, error)
	Update(ctx context.Context, template *models.StoredTemplate) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filters TemplateFilters) ([]*models.StoredTemplate, error)
	Search(ctx context.Context, query string, limit, offset int) ([]*models.StoredTemplate, error)
	GetByAuthor(ctx context.Context, authorID string, limit, offset int) ([]*models.StoredTemplate, error)
	GetByOrganization(ctx context.Context, orgID string, limit, offset int) ([]*models.StoredTemplate, error)
	GetFeatured(ctx context.Context, limit int) ([]*models.StoredTemplate, error)
	IncrementDownloads(ctx context.Context, id string) error
	GetStats(ctx context.Context) (*models.TemplateStats, error)
	GetRating(ctx context.Context, templateID string) (*models.TemplateRating, error)
}

type OrganizationRepository interface {
	Create(ctx context.Context, org *models.Organization) error
	GetByID(ctx context.Context, id string) (*models.Organization, error)
	GetBySlug(ctx context.Context, slug string) (*models.Organization, error)
	Update(ctx context.Context, org *models.Organization) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]*models.Organization, error)
	Search(ctx context.Context, query string, limit, offset int) ([]*models.Organization, error)
	GetByOwner(ctx context.Context, ownerID string) ([]*models.Organization, error)
	GetUserOrganizations(ctx context.Context, userID string) ([]*models.Organization, error)

	AddMember(ctx context.Context, member *models.OrganizationMember) error
	RemoveMember(ctx context.Context, orgID, userID string) error
	UpdateMemberRole(ctx context.Context, orgID, userID, role string) error
	GetMembers(ctx context.Context, orgID string) ([]*models.OrganizationMember, error)
	GetMember(ctx context.Context, orgID, userID string) (*models.OrganizationMember, error)
	IsMember(ctx context.Context, orgID, userID string) (bool, error)

	CreateInvite(ctx context.Context, invite *models.OrganizationInvite) error
	GetInvite(ctx context.Context, token string) (*models.OrganizationInvite, error)
	GetInvitesByOrganization(ctx context.Context, orgID string) ([]*models.OrganizationInvite, error)
	AcceptInvite(ctx context.Context, token string, userID string) error
	DeleteInvite(ctx context.Context, id string) error
	CleanupExpiredInvites(ctx context.Context) error
}

type ReviewRepository interface {
	Create(ctx context.Context, review *models.Review) error
	GetByID(ctx context.Context, id string) (*models.Review, error)
	Update(ctx context.Context, review *models.Review) error
	Delete(ctx context.Context, id string) error
	GetByTemplate(ctx context.Context, templateID string, limit, offset int) ([]*models.Review, error)
	GetByUser(ctx context.Context, userID string, limit, offset int) ([]*models.Review, error)
	GetUserReviewForTemplate(ctx context.Context, userID, templateID string) (*models.Review, error)
	IncrementHelpful(ctx context.Context, id string) error
	CalculateTemplateRating(ctx context.Context, templateID string) (*models.TemplateRating, error)
}

type ConfigRepository interface {
	Create(ctx context.Context, config *models.StoredConfig) error
	GetByID(ctx context.Context, id string) (*models.StoredConfig, error)
	Update(ctx context.Context, config *models.StoredConfig) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]*models.StoredConfig, error)
	GetStats(ctx context.Context) (*models.ConfigStats, error)
	IncrementDownloads(ctx context.Context, id string) error
}

type TemplateFilters struct {
	Author         string
	Tags           []string
	Featured       *bool
	Public         *bool
	OrganizationID string
	Limit          int
	Offset         int
	SortBy         string
	SortOrder      string
}

type Repositories struct {
	Users         UserRepository
	Templates     TemplateRepository
	Organizations OrganizationRepository
	Reviews       ReviewRepository
	Configs       ConfigRepository
}