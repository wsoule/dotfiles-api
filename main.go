package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
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
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
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
	Taps           []string         `json:"taps"`
	Brews          []string         `json:"brews"`
	Casks          []string         `json:"casks"`
	Stow           []string         `json:"stow"`
	Metadata       ShareMetadata    `json:"metadata"`
	Extends        string           `json:"extends,omitempty"`
	Overrides      []string         `json:"overrides,omitempty"`
	AddOnly        bool             `json:"addOnly"`
	Public         bool             `json:"public"`
	Featured       bool             `json:"featured"`
	OrganizationID string           `json:"organization_id,omitempty" bson:"organization_id,omitempty"`
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
	OwnerID      string          `json:"owner_id" bson:"owner_id"`
}

// User models
type User struct {
	ID           string    `json:"id" bson:"_id"`
	GitHubID     int       `json:"github_id" bson:"github_id"`
	Username     string    `json:"username" bson:"username"`
	Name         string    `json:"name" bson:"name"`
	Email        string    `json:"email" bson:"email"`
	AvatarURL    string    `json:"avatar_url" bson:"avatar_url"`
	Bio          string    `json:"bio" bson:"bio"`
	Location     string    `json:"location" bson:"location"`
	Website      string    `json:"website" bson:"website"`
	CreatedAt    time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" bson:"updated_at"`
	Favorites    []string  `json:"favorites" bson:"favorites"`
	Collections  []string  `json:"collections" bson:"collections"`
}

type Review struct {
	ID         string    `json:"id" bson:"_id"`
	TemplateID string    `json:"template_id" bson:"template_id"`
	UserID     string    `json:"user_id" bson:"user_id"`
	Username   string    `json:"username" bson:"username"`
	AvatarURL  string    `json:"avatar_url" bson:"avatar_url"`
	Rating     int       `json:"rating" bson:"rating"` // 1-5 stars
	Comment    string    `json:"comment" bson:"comment"`
	CreatedAt  time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" bson:"updated_at"`
	Helpful    int       `json:"helpful" bson:"helpful"` // helpful votes
}

type TemplateRating struct {
	TemplateID   string  `json:"template_id" bson:"template_id"`
	AverageRating float64 `json:"average_rating" bson:"average_rating"`
	TotalRatings int     `json:"total_ratings" bson:"total_ratings"`
	Distribution map[int]int `json:"distribution" bson:"distribution"` // rating -> count
}

// Organization models
type Organization struct {
	ID          string    `json:"id" bson:"_id"`
	Name        string    `json:"name" bson:"name"`
	Slug        string    `json:"slug" bson:"slug"` // URL-friendly name
	Description string    `json:"description" bson:"description"`
	Website     string    `json:"website" bson:"website"`
	AvatarURL   string    `json:"avatar_url" bson:"avatar_url"`
	OwnerID     string    `json:"owner_id" bson:"owner_id"`
	Public      bool      `json:"public" bson:"public"`
	CreatedAt   time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" bson:"updated_at"`
	MemberCount int       `json:"member_count" bson:"member_count"`
}

type OrganizationMember struct {
	ID             string    `json:"id" bson:"_id"`
	OrganizationID string    `json:"organization_id" bson:"organization_id"`
	UserID         string    `json:"user_id" bson:"user_id"`
	Username       string    `json:"username" bson:"username"`
	Role           string    `json:"role" bson:"role"` // owner, admin, member
	JoinedAt       time.Time `json:"joined_at" bson:"joined_at"`
	InvitedBy      string    `json:"invited_by" bson:"invited_by"`
}

type OrganizationInvite struct {
	ID             string    `json:"id" bson:"_id"`
	OrganizationID string    `json:"organization_id" bson:"organization_id"`
	InviterID      string    `json:"inviter_id" bson:"inviter_id"`
	Email          string    `json:"email" bson:"email"`
	Role           string    `json:"role" bson:"role"`
	Token          string    `json:"token" bson:"token"`
	ExpiresAt      time.Time `json:"expires_at" bson:"expires_at"`
	CreatedAt      time.Time `json:"created_at" bson:"created_at"`
	AcceptedAt     *time.Time `json:"accepted_at,omitempty" bson:"accepted_at,omitempty"`
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

// User storage interface
type UserStorage interface {
	StoreUser(user *User) error
	GetUser(id string) (*User, error)
	GetUserByGitHubID(githubID int) (*User, error)
	GetUserByUsername(username string) (*User, error)
	UpdateUser(user *User) error
	AddToFavorites(userID, templateID string) error
	RemoveFromFavorites(userID, templateID string) error
}

// Review storage interface
type ReviewStorage interface {
	StoreReview(review *Review) error
	GetReview(id string) (*Review, error)
	GetReviewsByTemplate(templateID string) ([]*Review, error)
	GetReviewsByUser(userID string) ([]*Review, error)
	UpdateReview(review *Review) error
	DeleteReview(id string) error
	GetTemplateRating(templateID string) (*TemplateRating, error)
	UpdateTemplateRating(templateID string) error
}

// Organization storage interface
type OrganizationStorage interface {
	StoreOrganization(org *Organization) error
	GetOrganization(id string) (*Organization, error)
	GetOrganizationBySlug(slug string) (*Organization, error)
	UpdateOrganization(org *Organization) error
	DeleteOrganization(id string) error
	GetUserOrganizations(userID string) ([]*Organization, error)
	SearchOrganizations(query string, publicOnly bool) ([]*Organization, error)

	// Member management
	AddMember(member *OrganizationMember) error
	RemoveMember(orgID, userID string) error
	UpdateMemberRole(orgID, userID, role string) error
	GetOrganizationMembers(orgID string) ([]*OrganizationMember, error)
	GetUserMemberships(userID string) ([]*OrganizationMember, error)
	IsUserMember(orgID, userID string) (bool, error)

	// Invite management
	CreateInvite(invite *OrganizationInvite) error
	GetInvite(token string) (*OrganizationInvite, error)
	AcceptInvite(token string, userID string) error
	DeleteInvite(token string) error
	GetOrganizationInvites(orgID string) ([]*OrganizationInvite, error)
}

// In-memory storage (fallback)
type MemoryStorage struct {
	configs            map[string]*StoredConfig
	templates          map[string]*StoredTemplate
	users              map[string]*User
	reviews            map[string]*Review
	organizations      map[string]*Organization
	orgMembers         map[string]*OrganizationMember
	orgInvites         map[string]*OrganizationInvite
	orgSlugIndex       map[string]string // slug -> org ID
	userOrgIndex       map[string][]string // user ID -> org IDs
	mu                 sync.RWMutex
}

// OAuth configuration
var (
	oauthConfig *oauth2.Config
	oauthStateString = "randomstate"
)

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		configs:       make(map[string]*StoredConfig),
		templates:     make(map[string]*StoredTemplate),
		users:         make(map[string]*User),
		reviews:       make(map[string]*Review),
		organizations: make(map[string]*Organization),
		orgMembers:    make(map[string]*OrganizationMember),
		orgInvites:    make(map[string]*OrganizationInvite),
		orgSlugIndex:  make(map[string]string),
		userOrgIndex:  make(map[string][]string),
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

// User methods for MemoryStorage
func (m *MemoryStorage) StoreUser(user *User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.users[user.ID] = user
	return nil
}

func (m *MemoryStorage) GetUser(id string) (*User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	user, exists := m.users[id]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

func (m *MemoryStorage) GetUserByGitHubID(githubID int) (*User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, user := range m.users {
		if user.GitHubID == githubID {
			return user, nil
		}
	}
	return nil, fmt.Errorf("user not found")
}

func (m *MemoryStorage) GetUserByUsername(username string) (*User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, user := range m.users {
		if user.Username == username {
			return user, nil
		}
	}
	return nil, fmt.Errorf("user not found")
}

func (m *MemoryStorage) UpdateUser(user *User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.users[user.ID]; exists {
		user.UpdatedAt = time.Now()
		m.users[user.ID] = user
		return nil
	}
	return fmt.Errorf("user not found")
}

func (m *MemoryStorage) AddToFavorites(userID, templateID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if user, exists := m.users[userID]; exists {
		// Check if already favorited
		for _, fav := range user.Favorites {
			if fav == templateID {
				return nil // Already favorited
			}
		}
		user.Favorites = append(user.Favorites, templateID)
		user.UpdatedAt = time.Now()
		return nil
	}
	return fmt.Errorf("user not found")
}

func (m *MemoryStorage) RemoveFromFavorites(userID, templateID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if user, exists := m.users[userID]; exists {
		for i, fav := range user.Favorites {
			if fav == templateID {
				user.Favorites = append(user.Favorites[:i], user.Favorites[i+1:]...)
				user.UpdatedAt = time.Now()
				break
			}
		}
		return nil
	}
	return fmt.Errorf("user not found")
}

// Review methods for MemoryStorage
func (m *MemoryStorage) StoreReview(review *Review) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.reviews[review.ID] = review
	return nil
}

func (m *MemoryStorage) GetReview(id string) (*Review, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	review, exists := m.reviews[id]
	if !exists {
		return nil, fmt.Errorf("review not found")
	}
	return review, nil
}

func (m *MemoryStorage) GetReviewsByTemplate(templateID string) ([]*Review, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var reviews []*Review
	for _, review := range m.reviews {
		if review.TemplateID == templateID {
			reviews = append(reviews, review)
		}
	}
	return reviews, nil
}

func (m *MemoryStorage) GetReviewsByUser(userID string) ([]*Review, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var reviews []*Review
	for _, review := range m.reviews {
		if review.UserID == userID {
			reviews = append(reviews, review)
		}
	}
	return reviews, nil
}

func (m *MemoryStorage) UpdateReview(review *Review) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.reviews[review.ID]; exists {
		review.UpdatedAt = time.Now()
		m.reviews[review.ID] = review
		return nil
	}
	return fmt.Errorf("review not found")
}

func (m *MemoryStorage) DeleteReview(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.reviews, id)
	return nil
}

func (m *MemoryStorage) GetTemplateRating(templateID string) (*TemplateRating, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var totalRating float64
	var count int
	distribution := make(map[int]int)

	for _, review := range m.reviews {
		if review.TemplateID == templateID {
			totalRating += float64(review.Rating)
			count++
			distribution[review.Rating]++
		}
	}

	if count == 0 {
		return &TemplateRating{
			TemplateID:    templateID,
			AverageRating: 0,
			TotalRatings:  0,
			Distribution:  distribution,
		}, nil
	}

	return &TemplateRating{
		TemplateID:    templateID,
		AverageRating: totalRating / float64(count),
		TotalRatings:  count,
		Distribution:  distribution,
	}, nil
}

func (m *MemoryStorage) UpdateTemplateRating(templateID string) error {
	// For memory storage, ratings are calculated on-the-fly
	return nil
}

// Organization methods for MemoryStorage
func (m *MemoryStorage) StoreOrganization(org *Organization) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.organizations[org.ID] = org
	m.orgSlugIndex[org.Slug] = org.ID

	// Add owner as first member
	member := &OrganizationMember{
		ID:             uuid.New().String(),
		OrganizationID: org.ID,
		UserID:         org.OwnerID,
		Role:           "owner",
		JoinedAt:       org.CreatedAt,
		InvitedBy:      org.OwnerID,
	}
	m.orgMembers[member.ID] = member

	// Update user org index
	if userOrgs, exists := m.userOrgIndex[org.OwnerID]; exists {
		m.userOrgIndex[org.OwnerID] = append(userOrgs, org.ID)
	} else {
		m.userOrgIndex[org.OwnerID] = []string{org.ID}
	}

	return nil
}

func (m *MemoryStorage) GetOrganization(id string) (*Organization, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	org, exists := m.organizations[id]
	if !exists {
		return nil, fmt.Errorf("organization not found")
	}
	return org, nil
}

func (m *MemoryStorage) GetOrganizationBySlug(slug string) (*Organization, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	orgID, exists := m.orgSlugIndex[slug]
	if !exists {
		return nil, fmt.Errorf("organization not found")
	}

	org, exists := m.organizations[orgID]
	if !exists {
		return nil, fmt.Errorf("organization not found")
	}
	return org, nil
}

func (m *MemoryStorage) UpdateOrganization(org *Organization) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.organizations[org.ID]; !exists {
		return fmt.Errorf("organization not found")
	}

	// Update slug index if changed
	if existing := m.organizations[org.ID]; existing.Slug != org.Slug {
		delete(m.orgSlugIndex, existing.Slug)
		m.orgSlugIndex[org.Slug] = org.ID
	}

	org.UpdatedAt = time.Now()
	m.organizations[org.ID] = org
	return nil
}

func (m *MemoryStorage) DeleteOrganization(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	org, exists := m.organizations[id]
	if !exists {
		return fmt.Errorf("organization not found")
	}

	// Remove from slug index
	delete(m.orgSlugIndex, org.Slug)

	// Remove all members
	for memberID, member := range m.orgMembers {
		if member.OrganizationID == id {
			delete(m.orgMembers, memberID)
		}
	}

	// Remove from user org indexes
	for userID, orgList := range m.userOrgIndex {
		for i, orgID := range orgList {
			if orgID == id {
				m.userOrgIndex[userID] = append(orgList[:i], orgList[i+1:]...)
				break
			}
		}
	}

	delete(m.organizations, id)
	return nil
}

func (m *MemoryStorage) GetUserOrganizations(userID string) ([]*Organization, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	orgIDs, exists := m.userOrgIndex[userID]
	if !exists {
		return []*Organization{}, nil
	}

	var orgs []*Organization
	for _, orgID := range orgIDs {
		if org, exists := m.organizations[orgID]; exists {
			orgs = append(orgs, org)
		}
	}

	return orgs, nil
}

func (m *MemoryStorage) SearchOrganizations(query string, publicOnly bool) ([]*Organization, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var results []*Organization
	queryLower := strings.ToLower(query)

	for _, org := range m.organizations {
		if publicOnly && !org.Public {
			continue
		}

		searchText := strings.ToLower(org.Name + " " + org.Description)
		if query == "" || strings.Contains(searchText, queryLower) {
			results = append(results, org)
		}
	}

	return results, nil
}

func (m *MemoryStorage) AddMember(member *OrganizationMember) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.orgMembers[member.ID] = member

	// Update user org index
	if userOrgs, exists := m.userOrgIndex[member.UserID]; exists {
		// Check if already exists
		for _, orgID := range userOrgs {
			if orgID == member.OrganizationID {
				return nil // Already a member
			}
		}
		m.userOrgIndex[member.UserID] = append(userOrgs, member.OrganizationID)
	} else {
		m.userOrgIndex[member.UserID] = []string{member.OrganizationID}
	}

	// Update organization member count
	if org, exists := m.organizations[member.OrganizationID]; exists {
		org.MemberCount++
		m.organizations[member.OrganizationID] = org
	}

	return nil
}

func (m *MemoryStorage) RemoveMember(orgID, userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Find and remove member
	var memberID string
	for id, member := range m.orgMembers {
		if member.OrganizationID == orgID && member.UserID == userID {
			memberID = id
			break
		}
	}

	if memberID == "" {
		return fmt.Errorf("member not found")
	}

	delete(m.orgMembers, memberID)

	// Update user org index
	if userOrgs, exists := m.userOrgIndex[userID]; exists {
		for i, id := range userOrgs {
			if id == orgID {
				m.userOrgIndex[userID] = append(userOrgs[:i], userOrgs[i+1:]...)
				break
			}
		}
	}

	// Update organization member count
	if org, exists := m.organizations[orgID]; exists {
		org.MemberCount--
		m.organizations[orgID] = org
	}

	return nil
}

func (m *MemoryStorage) UpdateMemberRole(orgID, userID, role string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, member := range m.orgMembers {
		if member.OrganizationID == orgID && member.UserID == userID {
			member.Role = role
			m.orgMembers[id] = member
			return nil
		}
	}

	return fmt.Errorf("member not found")
}

func (m *MemoryStorage) GetOrganizationMembers(orgID string) ([]*OrganizationMember, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var members []*OrganizationMember
	for _, member := range m.orgMembers {
		if member.OrganizationID == orgID {
			members = append(members, member)
		}
	}

	return members, nil
}

func (m *MemoryStorage) GetUserMemberships(userID string) ([]*OrganizationMember, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var memberships []*OrganizationMember
	for _, member := range m.orgMembers {
		if member.UserID == userID {
			memberships = append(memberships, member)
		}
	}

	return memberships, nil
}

func (m *MemoryStorage) IsUserMember(orgID, userID string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, member := range m.orgMembers {
		if member.OrganizationID == orgID && member.UserID == userID {
			return true, nil
		}
	}

	return false, nil
}

func (m *MemoryStorage) CreateInvite(invite *OrganizationInvite) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.orgInvites[invite.Token] = invite
	return nil
}

func (m *MemoryStorage) GetInvite(token string) (*OrganizationInvite, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	invite, exists := m.orgInvites[token]
	if !exists {
		return nil, fmt.Errorf("invite not found")
	}

	if time.Now().After(invite.ExpiresAt) {
		return nil, fmt.Errorf("invite expired")
	}

	return invite, nil
}

func (m *MemoryStorage) AcceptInvite(token string, userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	invite, exists := m.orgInvites[token]
	if !exists {
		return fmt.Errorf("invite not found")
	}

	if time.Now().After(invite.ExpiresAt) {
		return fmt.Errorf("invite expired")
	}

	if invite.AcceptedAt != nil {
		return fmt.Errorf("invite already accepted")
	}

	// Add as member
	member := &OrganizationMember{
		ID:             uuid.New().String(),
		OrganizationID: invite.OrganizationID,
		UserID:         userID,
		Role:           invite.Role,
		JoinedAt:       time.Now(),
		InvitedBy:      invite.InviterID,
	}

	if err := m.AddMember(member); err != nil {
		return err
	}

	// Mark invite as accepted
	now := time.Now()
	invite.AcceptedAt = &now
	m.orgInvites[token] = invite

	return nil
}

func (m *MemoryStorage) DeleteInvite(token string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.orgInvites, token)
	return nil
}

func (m *MemoryStorage) GetOrganizationInvites(orgID string) ([]*OrganizationInvite, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var invites []*OrganizationInvite
	for _, invite := range m.orgInvites {
		if invite.OrganizationID == orgID && invite.AcceptedAt == nil {
			invites = append(invites, invite)
		}
	}

	return invites, nil
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

// User methods for MongoStorage - stub implementations
func (m *MongoStorage) StoreUser(user *User) error {
	// TODO: Implement MongoDB user storage
	return fmt.Errorf("user storage not implemented for MongoDB")
}

func (m *MongoStorage) GetUser(id string) (*User, error) {
	// TODO: Implement MongoDB user retrieval
	return nil, fmt.Errorf("user storage not implemented for MongoDB")
}

func (m *MongoStorage) GetUserByGitHubID(githubID int) (*User, error) {
	// TODO: Implement MongoDB user retrieval by GitHub ID
	return nil, fmt.Errorf("user storage not implemented for MongoDB")
}

func (m *MongoStorage) GetUserByUsername(username string) (*User, error) {
	// TODO: Implement MongoDB user retrieval by username
	return nil, fmt.Errorf("user storage not implemented for MongoDB")
}

func (m *MongoStorage) UpdateUser(user *User) error {
	// TODO: Implement MongoDB user update
	return fmt.Errorf("user storage not implemented for MongoDB")
}

func (m *MongoStorage) AddToFavorites(userID, templateID string) error {
	// TODO: Implement MongoDB favorites functionality
	return fmt.Errorf("favorites not implemented for MongoDB")
}

func (m *MongoStorage) RemoveFromFavorites(userID, templateID string) error {
	// TODO: Implement MongoDB favorites functionality
	return fmt.Errorf("favorites not implemented for MongoDB")
}

// Review methods for MongoStorage - stub implementations
func (m *MongoStorage) StoreReview(review *Review) error {
	// TODO: Implement MongoDB review storage
	return fmt.Errorf("review storage not implemented for MongoDB")
}

func (m *MongoStorage) GetReview(id string) (*Review, error) {
	// TODO: Implement MongoDB review retrieval
	return nil, fmt.Errorf("review storage not implemented for MongoDB")
}

func (m *MongoStorage) GetReviewsByTemplate(templateID string) ([]*Review, error) {
	// TODO: Implement MongoDB review retrieval by template
	return nil, fmt.Errorf("review storage not implemented for MongoDB")
}

func (m *MongoStorage) GetReviewsByUser(userID string) ([]*Review, error) {
	// TODO: Implement MongoDB review retrieval by user
	return nil, fmt.Errorf("review storage not implemented for MongoDB")
}

func (m *MongoStorage) UpdateReview(review *Review) error {
	// TODO: Implement MongoDB review update
	return fmt.Errorf("review storage not implemented for MongoDB")
}

func (m *MongoStorage) DeleteReview(id string) error {
	// TODO: Implement MongoDB review deletion
	return fmt.Errorf("review storage not implemented for MongoDB")
}

func (m *MongoStorage) GetTemplateRating(templateID string) (*TemplateRating, error) {
	// TODO: Implement MongoDB template rating retrieval
	return nil, fmt.Errorf("rating storage not implemented for MongoDB")
}

func (m *MongoStorage) UpdateTemplateRating(templateID string) error {
	// TODO: Implement MongoDB template rating update
	return fmt.Errorf("rating storage not implemented for MongoDB")
}

// Organization methods for MongoStorage - stub implementations
func (m *MongoStorage) StoreOrganization(org *Organization) error {
	// TODO: Implement MongoDB organization storage
	return fmt.Errorf("organization storage not implemented for MongoDB")
}

func (m *MongoStorage) GetOrganization(id string) (*Organization, error) {
	// TODO: Implement MongoDB organization retrieval
	return nil, fmt.Errorf("organization storage not implemented for MongoDB")
}

func (m *MongoStorage) GetOrganizationBySlug(slug string) (*Organization, error) {
	// TODO: Implement MongoDB organization retrieval by slug
	return nil, fmt.Errorf("organization storage not implemented for MongoDB")
}

func (m *MongoStorage) UpdateOrganization(org *Organization) error {
	// TODO: Implement MongoDB organization update
	return fmt.Errorf("organization storage not implemented for MongoDB")
}

func (m *MongoStorage) DeleteOrganization(id string) error {
	// TODO: Implement MongoDB organization deletion
	return fmt.Errorf("organization storage not implemented for MongoDB")
}

func (m *MongoStorage) GetUserOrganizations(userID string) ([]*Organization, error) {
	// TODO: Implement MongoDB user organizations retrieval
	return nil, fmt.Errorf("organization storage not implemented for MongoDB")
}

func (m *MongoStorage) SearchOrganizations(query string, publicOnly bool) ([]*Organization, error) {
	// TODO: Implement MongoDB organization search
	return nil, fmt.Errorf("organization storage not implemented for MongoDB")
}

func (m *MongoStorage) AddMember(member *OrganizationMember) error {
	// TODO: Implement MongoDB member addition
	return fmt.Errorf("organization storage not implemented for MongoDB")
}

func (m *MongoStorage) RemoveMember(orgID, userID string) error {
	// TODO: Implement MongoDB member removal
	return fmt.Errorf("organization storage not implemented for MongoDB")
}

func (m *MongoStorage) UpdateMemberRole(orgID, userID, role string) error {
	// TODO: Implement MongoDB member role update
	return fmt.Errorf("organization storage not implemented for MongoDB")
}

func (m *MongoStorage) GetOrganizationMembers(orgID string) ([]*OrganizationMember, error) {
	// TODO: Implement MongoDB organization members retrieval
	return nil, fmt.Errorf("organization storage not implemented for MongoDB")
}

func (m *MongoStorage) GetUserMemberships(userID string) ([]*OrganizationMember, error) {
	// TODO: Implement MongoDB user memberships retrieval
	return nil, fmt.Errorf("organization storage not implemented for MongoDB")
}

func (m *MongoStorage) IsUserMember(orgID, userID string) (bool, error) {
	// TODO: Implement MongoDB user membership check
	return false, fmt.Errorf("organization storage not implemented for MongoDB")
}

func (m *MongoStorage) CreateInvite(invite *OrganizationInvite) error {
	// TODO: Implement MongoDB invite creation
	return fmt.Errorf("organization storage not implemented for MongoDB")
}

func (m *MongoStorage) GetInvite(token string) (*OrganizationInvite, error) {
	// TODO: Implement MongoDB invite retrieval
	return nil, fmt.Errorf("organization storage not implemented for MongoDB")
}

func (m *MongoStorage) AcceptInvite(token string, userID string) error {
	// TODO: Implement MongoDB invite acceptance
	return fmt.Errorf("organization storage not implemented for MongoDB")
}

func (m *MongoStorage) DeleteInvite(token string) error {
	// TODO: Implement MongoDB invite deletion
	return fmt.Errorf("organization storage not implemented for MongoDB")
}

func (m *MongoStorage) GetOrganizationInvites(orgID string) ([]*OrganizationInvite, error) {
	// TODO: Implement MongoDB organization invites retrieval
	return nil, fmt.Errorf("organization storage not implemented for MongoDB")
}

var storage ConfigStorage
var templateStorage TemplateStorage
var userStorage UserStorage
var reviewStorage ReviewStorage
var organizationStorage OrganizationStorage

func seedTemplates() {
	// Check if we already have templates
	templates, err := templateStorage.SearchTemplates("", "", nil)
	if err == nil && len(templates) > 0 {
		return // Already have templates
	}

	seedTemplates := []Template{
		{
			Taps:  []string{"homebrew/cask-fonts"},
			Brews: []string{"git", "curl", "wget", "tree", "jq", "node", "npm", "yarn", "python3", "docker", "postgresql"},
			Casks: []string{"visual-studio-code", "google-chrome", "firefox", "iterm2", "rectangle", "figma", "slack", "postman"},
			Stow:  []string{"git", "zsh", "vim", "vscode", "tmux"},
			Metadata: ShareMetadata{
				Name:        "Full Stack Web Developer",
				Description: "Complete setup for modern web development with Node.js, Python, Docker, and essential development tools. Perfect for frontend, backend, and full-stack developers.",
				Author:      "webdev_pro",
				Tags:        []string{"web-dev", "javascript", "python", "docker", "frontend", "backend", "full-stack"},
				CreatedAt:   time.Now().AddDate(0, 0, -14),
				Version:     "2.1.0",
			},
			Public:   true,
			Featured: true,
			AddOnly:  false,
		},
		{
			Brews: []string{"git", "python3", "r", "jupyter", "postgresql", "sqlite", "graphviz", "pandoc"},
			Casks: []string{"visual-studio-code", "rstudio", "tableau-public", "docker", "anaconda"},
			Stow:  []string{"git", "zsh", "vim", "python", "jupyter", "r"},
			Metadata: ShareMetadata{
				Name:        "Data Science Toolkit",
				Description: "Comprehensive Python, R, and Jupyter environment for data scientists, researchers, and analysts. Includes visualization tools and database connections.",
				Author:      "data_scientist",
				Tags:        []string{"data-science", "python", "r", "jupyter", "analytics", "ml", "statistics"},
				CreatedAt:   time.Now().AddDate(0, 0, -10),
				Version:     "1.8.0",
			},
			Public:   true,
			Featured: true,
			AddOnly:  false,
		},
		{
			Taps:  []string{"hashicorp/tap", "kubernetes/tap"},
			Brews: []string{"git", "curl", "kubectl", "terraform", "ansible", "aws-cli", "docker", "helm", "jq", "yq"},
			Casks: []string{"visual-studio-code", "iterm2", "lens", "postman", "docker"},
			Stow:  []string{"git", "zsh", "kubectl", "terraform", "aws"},
			Metadata: ShareMetadata{
				Name:        "DevOps Engineer Setup",
				Description: "Infrastructure, containerization, and cloud tools for DevOps workflows. Includes Kubernetes, Terraform, AWS CLI, and monitoring tools.",
				Author:      "devops_master",
				Tags:        []string{"devops", "kubernetes", "terraform", "aws", "docker", "infrastructure", "cloud"},
				CreatedAt:   time.Now().AddDate(0, 0, -5),
				Version:     "3.0.0",
			},
			Public:   true,
			Featured: true,
			AddOnly:  false,
		},
		{
			Brews: []string{"git", "node", "watchman", "cocoapods", "ruby", "android-platform-tools"},
			Casks: []string{"visual-studio-code", "android-studio", "xcode", "simulator", "flipper"},
			Stow:  []string{"git", "zsh", "vim", "react-native"},
			Metadata: ShareMetadata{
				Name:        "Mobile Developer Setup",
				Description: "iOS and Android development environment with React Native, Flutter, and native toolchains. Includes simulators and debugging tools.",
				Author:      "mobile_dev",
				Tags:        []string{"mobile-dev", "ios", "android", "react-native", "flutter", "xcode"},
				CreatedAt:   time.Now().AddDate(0, 0, -8),
				Version:     "1.5.0",
			},
			Public:   true,
			Featured: false,
			AddOnly:  false,
		},
		{
			Brews: []string{"git", "node", "python3", "go", "rust", "postgresql", "redis", "nginx"},
			Casks: []string{"visual-studio-code", "iterm2", "postman", "tableplus"},
			Stow:  []string{"git", "zsh", "vim", "go", "rust"},
			Metadata: ShareMetadata{
				Name:        "Backend Developer Kit",
				Description: "Server-side development with multiple languages: Go, Rust, Python, Node.js. Includes databases and API development tools.",
				Author:      "backend_guru",
				Tags:        []string{"backend", "go", "rust", "python", "nodejs", "api", "database"},
				CreatedAt:   time.Now().AddDate(0, 0, -12),
				Version:     "1.3.0",
			},
			Public:   true,
			Featured: false,
			AddOnly:  false,
		},
		{
			Brews: []string{"git", "curl", "tree", "vim", "tmux"},
			Casks: []string{"iterm2", "visual-studio-code"},
			Stow:  []string{"git", "vim", "tmux"},
			Metadata: ShareMetadata{
				Name:        "Minimal Developer Setup",
				Description: "Essential tools for any developer. Lightweight setup with just the basics you need to get started with development.",
				Author:      "minimalist",
				Tags:        []string{"minimal", "essential", "basic", "lightweight"},
				CreatedAt:   time.Now().AddDate(0, 0, -20),
				Version:     "1.0.0",
			},
			Public:   true,
			Featured: false,
			AddOnly:  false,
		},
	}

	for i, template := range seedTemplates {
		id := uuid.New().String()
		stored := &StoredTemplate{
			ID:        id,
			Template:  template,
			CreatedAt: template.Metadata.CreatedAt,
			UpdatedAt: template.Metadata.CreatedAt,
			Downloads: int(time.Since(template.Metadata.CreatedAt).Hours()/24) * (i + 1), // Simulate downloads
		}
		templateStorage.StoreTemplate(stored)
	}
}

func seedData() {
	// Check if we already have data
	total, _, _, err := storage.GetStats()
	if err == nil && total > 0 {
		return // Already have data
	}

	// Seed templates first
	seedTemplates()

	// Then seed configs

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
	// Initialize OAuth configuration
	oauthConfig = &oauth2.Config{
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("OAUTH_REDIRECT_URL"),
		Scopes:       []string{"user:email"},
		Endpoint:     github.Endpoint,
	}

	// Generate secure random state string
	b := make([]byte, 32)
	rand.Read(b)
	oauthStateString = base64.URLEncoding.EncodeToString(b)

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
			userStorage = memStorage
			reviewStorage = memStorage
		} else {
			storage = mongoStorage
			templateStorage = mongoStorage
			userStorage = mongoStorage
			reviewStorage = mongoStorage
			organizationStorage = mongoStorage
			log.Println("Connected to MongoDB successfully")
		}
	} else {
		memStorage := NewMemoryStorage()
		storage = memStorage
		templateStorage = memStorage
		userStorage = memStorage
		reviewStorage = memStorage
		organizationStorage = memStorage
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

	// Authentication routes
	auth := r.Group("/auth")
	{
		auth.GET("/github", githubLogin)
		auth.GET("/github/callback", githubCallback)
		auth.GET("/logout", logout)
		auth.GET("/user", getCurrentUser)
	}

	// API routes
	api := r.Group("/api")
	{
		// Upload config
		api.POST("/configs/upload", uploadConfig)

		// Get config by ID
		api.GET("/configs/:id", getConfig)
		api.GET("/configs/:id/download", downloadConfig)

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
		api.GET("/templates/:id/download", downloadTemplate)
		api.GET("/templates/:id/reviews", getTemplateReviews)
		api.POST("/templates/:id/reviews", authRequired(), createReview)
		api.GET("/templates/:id/rating", getTemplateRating)

		// User endpoints
		api.GET("/users/:username", getUserProfile)
		api.POST("/users/favorites/:templateId", authRequired(), addToFavorites)
		api.DELETE("/users/favorites/:templateId", authRequired(), removeFromFavorites)

		// Review endpoints
		api.PUT("/reviews/:id", authRequired(), updateReview)
		api.DELETE("/reviews/:id", authRequired(), deleteReview)
		api.POST("/reviews/:id/helpful", authRequired(), markReviewHelpful)

		// Organization endpoints
		api.POST("/organizations", authRequired(), createOrganization)
		api.GET("/organizations", getOrganizations)
		api.GET("/organizations/:slug", getOrganizationBySlug)
		api.PUT("/organizations/:slug", authRequired(), updateOrganization)
		api.DELETE("/organizations/:slug", authRequired(), deleteOrganization)
		api.GET("/organizations/:slug/members", getOrganizationMembers)
		api.POST("/organizations/:slug/members", authRequired(), inviteMember)
		api.DELETE("/organizations/:slug/members/:username", authRequired(), removeMember)
		api.PUT("/organizations/:slug/members/:username", authRequired(), updateMemberRole)
		api.GET("/organizations/:slug/invites", authRequired(), getOrganizationInvites)
		api.POST("/invites/:token/accept", authRequired(), acceptInvite)
		api.GET("/users/:username/organizations", getUserOrganizations)
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

	// Templates page route
	r.GET("/templates", func(c *gin.Context) {
		c.File("./static/templates.html")
	})

	// Documentation page route
	r.GET("/docs", func(c *gin.Context) {
		c.File("./static/docs.html")
	})

	// Profile page routes (support both patterns)
	r.GET("/profile/:username", func(c *gin.Context) {
		c.File("./static/profile.html")
	})
	r.GET("/users/:username", func(c *gin.Context) {
		c.File("./static/profile.html")
	})

	// Organizations page route
	r.GET("/organizations", func(c *gin.Context) {
		c.File("./static/organizations.html")
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

	// Don't increment download count - this is just for viewing
	c.JSON(200, stored.Config)
}

func downloadConfig(c *gin.Context) {
	id := c.Param("id")

	stored, err := storage.Get(id)
	if err != nil {
		c.JSON(404, gin.H{"error": "Config not found"})
		return
	}

	// Increment download count for actual downloads
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

	// Get authenticated user if template is being assigned to an organization
	var user *User
	if template.OrganizationID != "" {
		userID, err := c.Cookie("user_id")
		if err != nil {
			c.JSON(401, gin.H{"error": "Authentication required to assign template to organization"})
			return
		}

		user, err = userStorage.GetUser(userID)
		if err != nil {
			c.JSON(401, gin.H{"error": "Invalid authentication"})
			return
		}

		// Check if user is a member of the organization
		isMember, err := organizationStorage.IsUserMember(template.OrganizationID, user.ID)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to verify organization membership"})
			return
		}

		if !isMember {
			c.JSON(403, gin.H{"error": "You must be a member of the organization to assign templates to it"})
			return
		}

		// Set the template author to the authenticated user
		template.Metadata.Author = user.Username
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
		"url": fmt.Sprintf("https://dotfiles.wyat.me/templates/%s", id),
	})
}

func getTemplate(c *gin.Context) {
	id := c.Param("id")

	stored, err := templateStorage.GetTemplate(id)
	if err != nil {
		c.JSON(404, gin.H{"error": "Template not found"})
		return
	}

	// Don't increment download count - this is just for viewing
	c.JSON(200, stored.Template)
}

func downloadTemplate(c *gin.Context) {
	id := c.Param("id")

	stored, err := templateStorage.GetTemplate(id)
	if err != nil {
		c.JSON(404, gin.H{"error": "Template not found"})
		return
	}

	// Increment download count for actual downloads
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
			"brews":       stored.Template.Brews,
			"casks":       stored.Template.Casks,
			"taps":        stored.Template.Taps,
			"stow":        stored.Template.Stow,
		})
	}

	c.JSON(200, gin.H{
		"templates": results,
		"total":     len(results),
	})
}

// Authentication handlers
func githubLogin(c *gin.Context) {
	// Check if OAuth is configured
	if oauthConfig.ClientID == "" {
		c.JSON(400, gin.H{
			"error": "GitHub OAuth not configured",
			"message": "Please set GITHUB_CLIENT_ID, GITHUB_CLIENT_SECRET, and OAUTH_REDIRECT_URL environment variables to enable GitHub authentication."})
		return
	}

	url := oauthConfig.AuthCodeURL(oauthStateString, oauth2.AccessTypeOffline)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func githubCallback(c *gin.Context) {
	state := c.Query("state")
	if state != oauthStateString {
		c.JSON(400, gin.H{"error": "Invalid OAuth state"})
		return
	}

	code := c.Query("code")
	token, err := oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		c.JSON(400, gin.H{"error": "Failed to exchange OAuth code"})
		return
	}

	// Get user info from GitHub
	client := oauthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to get user info"})
		return
	}
	defer resp.Body.Close()

	var githubUser struct {
		ID        int    `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
		Bio       string `json:"bio"`
		Location  string `json:"location"`
		Blog      string `json:"blog"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&githubUser); err != nil {
		c.JSON(500, gin.H{"error": "Failed to decode user info"})
		return
	}

	// Check if user exists
	user, err := userStorage.GetUserByGitHubID(githubUser.ID)
	if err != nil {
		// Create new user
		user = &User{
			ID:        uuid.New().String(),
			GitHubID:  githubUser.ID,
			Username:  githubUser.Login,
			Name:      githubUser.Name,
			Email:     githubUser.Email,
			AvatarURL: githubUser.AvatarURL,
			Bio:       githubUser.Bio,
			Location:  githubUser.Location,
			Website:   githubUser.Blog,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Favorites: []string{},
		}
		userStorage.StoreUser(user)
	} else {
		// Update existing user
		user.Name = githubUser.Name
		user.Email = githubUser.Email
		user.AvatarURL = githubUser.AvatarURL
		user.Bio = githubUser.Bio
		user.Location = githubUser.Location
		user.Website = githubUser.Blog
		user.UpdatedAt = time.Now()
		userStorage.UpdateUser(user)
	}

	// Set session (simple approach with secure cookies)
	c.SetCookie("user_id", user.ID, 3600*24*30, "/", "", false, true) // 30 days

	c.Redirect(http.StatusTemporaryRedirect, "/")
}

func logout(c *gin.Context) {
	c.SetCookie("user_id", "", -1, "/", "", false, true)
	c.Redirect(http.StatusTemporaryRedirect, "/")
}

func getCurrentUser(c *gin.Context) {
	// Check if OAuth is configured
	if oauthConfig.ClientID == "" {
		c.JSON(401, gin.H{
			"error": "GitHub OAuth not configured",
			"configured": false,
			"message": "Authentication is not available. Please configure GitHub OAuth to enable user features."})
		return
	}

	userID, err := c.Cookie("user_id")
	if err != nil {
		c.JSON(401, gin.H{"error": "Not authenticated", "configured": true})
		return
	}

	user, err := userStorage.GetUser(userID)
	if err != nil {
		c.JSON(404, gin.H{"error": "User not found", "configured": true})
		return
	}

	c.JSON(200, user)
}

// Authentication middleware
func authRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := c.Cookie("user_id")
		if err != nil {
			c.JSON(401, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		user, err := userStorage.GetUser(userID)
		if err != nil {
			c.JSON(401, gin.H{"error": "Invalid authentication"})
			c.Abort()
			return
		}

		c.Set("user", user)
		c.Next()
	}
}

// Review handlers
func getTemplateReviews(c *gin.Context) {
	templateID := c.Param("id")
	reviews, err := reviewStorage.GetReviewsByTemplate(templateID)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to get reviews"})
		return
	}

	c.JSON(200, gin.H{"reviews": reviews})
}

func createReview(c *gin.Context) {
	templateID := c.Param("id")
	user := c.MustGet("user").(*User)

	var req struct {
		Rating  int    `json:"rating" binding:"required,min=1,max=5"`
		Comment string `json:"comment"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	// Check if user already reviewed this template
	userReviews, err := reviewStorage.GetReviewsByUser(user.ID)
	if err == nil {
		for _, review := range userReviews {
			if review.TemplateID == templateID {
				c.JSON(400, gin.H{"error": "You have already reviewed this template"})
				return
			}
		}
	}

	review := &Review{
		ID:         uuid.New().String(),
		TemplateID: templateID,
		UserID:     user.ID,
		Username:   user.Username,
		AvatarURL:  user.AvatarURL,
		Rating:     req.Rating,
		Comment:    req.Comment,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Helpful:    0,
	}

	if err := reviewStorage.StoreReview(review); err != nil {
		c.JSON(500, gin.H{"error": "Failed to create review"})
		return
	}

	// Update template rating
	reviewStorage.UpdateTemplateRating(templateID)

	c.JSON(201, review)
}

func getTemplateRating(c *gin.Context) {
	templateID := c.Param("id")
	rating, err := reviewStorage.GetTemplateRating(templateID)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to get rating"})
		return
	}

	c.JSON(200, rating)
}

func updateReview(c *gin.Context) {
	reviewID := c.Param("id")
	user := c.MustGet("user").(*User)

	review, err := reviewStorage.GetReview(reviewID)
	if err != nil {
		c.JSON(404, gin.H{"error": "Review not found"})
		return
	}

	if review.UserID != user.ID {
		c.JSON(403, gin.H{"error": "Not authorized to update this review"})
		return
	}

	var req struct {
		Rating  int    `json:"rating" binding:"required,min=1,max=5"`
		Comment string `json:"comment"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	review.Rating = req.Rating
	review.Comment = req.Comment
	review.UpdatedAt = time.Now()

	if err := reviewStorage.UpdateReview(review); err != nil {
		c.JSON(500, gin.H{"error": "Failed to update review"})
		return
	}

	// Update template rating
	reviewStorage.UpdateTemplateRating(review.TemplateID)

	c.JSON(200, review)
}

func deleteReview(c *gin.Context) {
	reviewID := c.Param("id")
	user := c.MustGet("user").(*User)

	review, err := reviewStorage.GetReview(reviewID)
	if err != nil {
		c.JSON(404, gin.H{"error": "Review not found"})
		return
	}

	if review.UserID != user.ID {
		c.JSON(403, gin.H{"error": "Not authorized to delete this review"})
		return
	}

	if err := reviewStorage.DeleteReview(reviewID); err != nil {
		c.JSON(500, gin.H{"error": "Failed to delete review"})
		return
	}

	// Update template rating
	reviewStorage.UpdateTemplateRating(review.TemplateID)

	c.JSON(200, gin.H{"message": "Review deleted successfully"})
}

func markReviewHelpful(c *gin.Context) {
	reviewID := c.Param("id")

	review, err := reviewStorage.GetReview(reviewID)
	if err != nil {
		c.JSON(404, gin.H{"error": "Review not found"})
		return
	}

	review.Helpful++
	if err := reviewStorage.UpdateReview(review); err != nil {
		c.JSON(500, gin.H{"error": "Failed to update review"})
		return
	}

	c.JSON(200, gin.H{"helpful": review.Helpful})
}

// User profile handlers
func getUserProfile(c *gin.Context) {
	username := c.Param("username")
	user, err := userStorage.GetUserByUsername(username)
	if err != nil {
		c.JSON(404, gin.H{"error": "User not found"})
		return
	}

	// Get user's reviews
	reviews, _ := reviewStorage.GetReviewsByUser(user.ID)

	// Get user's favorite templates
	var favoriteTemplates []*StoredTemplate
	for _, templateID := range user.Favorites {
		if template, err := templateStorage.GetTemplate(templateID); err == nil {
			favoriteTemplates = append(favoriteTemplates, template)
		}
	}

	c.JSON(200, gin.H{
		"user":              user,
		"reviews":           reviews,
		"favoriteTemplates": favoriteTemplates,
	})
}

func addToFavorites(c *gin.Context) {
	templateID := c.Param("templateId")
	user := c.MustGet("user").(*User)

	// Verify template exists
	if _, err := templateStorage.GetTemplate(templateID); err != nil {
		c.JSON(404, gin.H{"error": "Template not found"})
		return
	}

	if err := userStorage.AddToFavorites(user.ID, templateID); err != nil {
		c.JSON(500, gin.H{"error": "Failed to add to favorites"})
		return
	}

	c.JSON(200, gin.H{"message": "Added to favorites"})
}

func removeFromFavorites(c *gin.Context) {
	templateID := c.Param("templateId")
	user := c.MustGet("user").(*User)

	if err := userStorage.RemoveFromFavorites(user.ID, templateID); err != nil {
		c.JSON(500, gin.H{"error": "Failed to remove from favorites"})
		return
	}

	c.JSON(200, gin.H{"message": "Removed from favorites"})
}

// Organization handlers
func createOrganization(c *gin.Context) {
	user := c.MustGet("user").(*User)

	var req struct {
		Name        string `json:"name" binding:"required"`
		Slug        string `json:"slug" binding:"required"`
		Description string `json:"description"`
		Website     string `json:"website"`
		Public      bool   `json:"public"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	// Check if slug is already taken
	if _, err := organizationStorage.GetOrganizationBySlug(req.Slug); err == nil {
		c.JSON(400, gin.H{"error": "Organization slug already exists"})
		return
	}

	org := &Organization{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		Website:     req.Website,
		OwnerID:     user.ID,
		Public:      req.Public,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		MemberCount: 1,
	}

	if err := organizationStorage.StoreOrganization(org); err != nil {
		c.JSON(500, gin.H{"error": "Failed to create organization"})
		return
	}

	// Add the creator as the owner member
	member := &OrganizationMember{
		ID:             uuid.New().String(),
		OrganizationID: org.ID,
		UserID:         user.ID,
		Role:           "owner",
		JoinedAt:       time.Now(),
	}

	if err := organizationStorage.AddMember(member); err != nil {
		c.JSON(500, gin.H{"error": "Failed to add creator as member"})
		return
	}

	c.JSON(201, org)
}

func getOrganizations(c *gin.Context) {
	query := c.Query("search")
	publicOnly := c.Query("public") != "false"

	orgs, err := organizationStorage.SearchOrganizations(query, publicOnly)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to get organizations"})
		return
	}

	c.JSON(200, gin.H{"organizations": orgs})
}

func getUserOrganizations(c *gin.Context) {
	username := c.Param("username")

	// Get user by username
	user, err := userStorage.GetUserByUsername(username)
	if err != nil {
		c.JSON(404, gin.H{"error": "User not found"})
		return
	}

	// Get user's organizations
	orgs, err := organizationStorage.GetUserOrganizations(user.ID)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to get user organizations"})
		return
	}

	c.JSON(200, gin.H{"organizations": orgs})
}

func getOrganizationBySlug(c *gin.Context) {
	slug := c.Param("slug")

	org, err := organizationStorage.GetOrganizationBySlug(slug)
	if err != nil {
		c.JSON(404, gin.H{"error": "Organization not found"})
		return
	}

	// Get members
	members, err := organizationStorage.GetOrganizationMembers(org.ID)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to get organization members"})
		return
	}

	c.JSON(200, gin.H{
		"organization": org,
		"members":      members,
	})
}

func updateOrganization(c *gin.Context) {
	slug := c.Param("slug")
	user := c.MustGet("user").(*User)

	org, err := organizationStorage.GetOrganizationBySlug(slug)
	if err != nil {
		c.JSON(404, gin.H{"error": "Organization not found"})
		return
	}

	// Check permissions (only owner can update)
	if org.OwnerID != user.ID {
		c.JSON(403, gin.H{"error": "Not authorized to update this organization"})
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Website     string `json:"website"`
		Public      bool   `json:"public"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	org.Name = req.Name
	org.Description = req.Description
	org.Website = req.Website
	org.Public = req.Public

	if err := organizationStorage.UpdateOrganization(org); err != nil {
		c.JSON(500, gin.H{"error": "Failed to update organization"})
		return
	}

	c.JSON(200, org)
}

func deleteOrganization(c *gin.Context) {
	slug := c.Param("slug")
	user := c.MustGet("user").(*User)

	org, err := organizationStorage.GetOrganizationBySlug(slug)
	if err != nil {
		c.JSON(404, gin.H{"error": "Organization not found"})
		return
	}

	// Check permissions (only owner can delete)
	if org.OwnerID != user.ID {
		c.JSON(403, gin.H{"error": "Not authorized to delete this organization"})
		return
	}

	if err := organizationStorage.DeleteOrganization(org.ID); err != nil {
		c.JSON(500, gin.H{"error": "Failed to delete organization"})
		return
	}

	c.JSON(200, gin.H{"message": "Organization deleted successfully"})
}

func getOrganizationMembers(c *gin.Context) {
	slug := c.Param("slug")

	org, err := organizationStorage.GetOrganizationBySlug(slug)
	if err != nil {
		c.JSON(404, gin.H{"error": "Organization not found"})
		return
	}

	members, err := organizationStorage.GetOrganizationMembers(org.ID)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to get members"})
		return
	}

	c.JSON(200, gin.H{"members": members})
}

func inviteMember(c *gin.Context) {
	slug := c.Param("slug")
	user := c.MustGet("user").(*User)

	org, err := organizationStorage.GetOrganizationBySlug(slug)
	if err != nil {
		c.JSON(404, gin.H{"error": "Organization not found"})
		return
	}

	// Check permissions (owner or admin can invite)
	members, err := organizationStorage.GetOrganizationMembers(org.ID)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to check permissions"})
		return
	}

	canInvite := false
	for _, member := range members {
		if member.UserID == user.ID && (member.Role == "owner" || member.Role == "admin") {
			canInvite = true
			break
		}
	}

	if !canInvite {
		c.JSON(403, gin.H{"error": "Not authorized to invite members"})
		return
	}

	var req struct {
		Email string `json:"email" binding:"required"`
		Role  string `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	// Generate invite token
	tokenBytes := make([]byte, 32)
	rand.Read(tokenBytes)
	token := base64.URLEncoding.EncodeToString(tokenBytes)

	invite := &OrganizationInvite{
		ID:             uuid.New().String(),
		OrganizationID: org.ID,
		InviterID:      user.ID,
		Email:          req.Email,
		Role:           req.Role,
		Token:          token,
		ExpiresAt:      time.Now().Add(7 * 24 * time.Hour), // 7 days
		CreatedAt:      time.Now(),
	}

	if err := organizationStorage.CreateInvite(invite); err != nil {
		c.JSON(500, gin.H{"error": "Failed to create invite"})
		return
	}

	c.JSON(201, gin.H{
		"invite": invite,
		"invite_url": fmt.Sprintf("/invites/%s", token),
	})
}

func removeMember(c *gin.Context) {
	slug := c.Param("slug")
	username := c.Param("username")
	user := c.MustGet("user").(*User)

	org, err := organizationStorage.GetOrganizationBySlug(slug)
	if err != nil {
		c.JSON(404, gin.H{"error": "Organization not found"})
		return
	}

	// Get target user
	targetUser, err := userStorage.GetUserByUsername(username)
	if err != nil {
		c.JSON(404, gin.H{"error": "User not found"})
		return
	}

	// Check permissions
	members, err := organizationStorage.GetOrganizationMembers(org.ID)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to check permissions"})
		return
	}

	canRemove := false
	isOwner := false
	targetRole := ""

	for _, member := range members {
		if member.UserID == user.ID {
			if member.Role == "owner" || member.Role == "admin" {
				canRemove = true
			}
			if member.Role == "owner" {
				isOwner = true
			}
		}
		if member.UserID == targetUser.ID {
			targetRole = member.Role
		}
	}

	if !canRemove {
		c.JSON(403, gin.H{"error": "Not authorized to remove members"})
		return
	}

	// Owners cannot be removed, and only owners can remove admins
	if targetRole == "owner" {
		c.JSON(400, gin.H{"error": "Cannot remove organization owner"})
		return
	}

	if targetRole == "admin" && !isOwner {
		c.JSON(403, gin.H{"error": "Only owners can remove admins"})
		return
	}

	if err := organizationStorage.RemoveMember(org.ID, targetUser.ID); err != nil {
		c.JSON(500, gin.H{"error": "Failed to remove member"})
		return
	}

	c.JSON(200, gin.H{"message": "Member removed successfully"})
}

func updateMemberRole(c *gin.Context) {
	slug := c.Param("slug")
	username := c.Param("username")
	user := c.MustGet("user").(*User)

	org, err := organizationStorage.GetOrganizationBySlug(slug)
	if err != nil {
		c.JSON(404, gin.H{"error": "Organization not found"})
		return
	}

	// Get target user
	targetUser, err := userStorage.GetUserByUsername(username)
	if err != nil {
		c.JSON(404, gin.H{"error": "User not found"})
		return
	}

	var req struct {
		Role string `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	// Check permissions (only owner can change roles)
	if org.OwnerID != user.ID {
		c.JSON(403, gin.H{"error": "Only organization owner can change member roles"})
		return
	}

	// Cannot change owner role
	if req.Role == "owner" {
		c.JSON(400, gin.H{"error": "Cannot assign owner role"})
		return
	}

	if err := organizationStorage.UpdateMemberRole(org.ID, targetUser.ID, req.Role); err != nil {
		c.JSON(500, gin.H{"error": "Failed to update member role"})
		return
	}

	c.JSON(200, gin.H{"message": "Member role updated successfully"})
}

func getOrganizationInvites(c *gin.Context) {
	slug := c.Param("slug")
	user := c.MustGet("user").(*User)

	org, err := organizationStorage.GetOrganizationBySlug(slug)
	if err != nil {
		c.JSON(404, gin.H{"error": "Organization not found"})
		return
	}

	// Check permissions
	members, err := organizationStorage.GetOrganizationMembers(org.ID)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to check permissions"})
		return
	}

	canView := false
	for _, member := range members {
		if member.UserID == user.ID && (member.Role == "owner" || member.Role == "admin") {
			canView = true
			break
		}
	}

	if !canView {
		c.JSON(403, gin.H{"error": "Not authorized to view invites"})
		return
	}

	invites, err := organizationStorage.GetOrganizationInvites(org.ID)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to get invites"})
		return
	}

	c.JSON(200, gin.H{"invites": invites})
}

func acceptInvite(c *gin.Context) {
	token := c.Param("token")
	user := c.MustGet("user").(*User)

	invite, err := organizationStorage.GetInvite(token)
	if err != nil {
		c.JSON(404, gin.H{"error": "Invalid or expired invite"})
		return
	}

	if err := organizationStorage.AcceptInvite(token, user.ID); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	org, _ := organizationStorage.GetOrganization(invite.OrganizationID)

	c.JSON(200, gin.H{
		"message": "Invite accepted successfully",
		"organization": org,
	})
}
