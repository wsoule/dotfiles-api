package mongo

import (
	"context"
	"time"

	"dotfiles-web/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// OrganizationRepository implements the OrganizationRepository interface using MongoDB
type OrganizationRepository struct {
	orgCollection     *mongo.Collection
	memberCollection  *mongo.Collection
	inviteCollection  *mongo.Collection
}

// NewOrganizationRepository creates a new organization repository
func NewOrganizationRepository(client *Client) *OrganizationRepository {
	return &OrganizationRepository{
		orgCollection:     client.Collection("organizations"),
		memberCollection:  client.Collection("organization_members"),
		inviteCollection:  client.Collection("organization_invites"),
	}
}

// Create stores a new organization
func (r *OrganizationRepository) Create(ctx context.Context, org *models.Organization) error {
	if org.ID == "" {
		org.ID = primitive.NewObjectID().Hex()
	}
	org.CreatedAt = time.Now()
	org.UpdatedAt = time.Now()

	_, err := r.orgCollection.InsertOne(ctx, org)
	return err
}

// GetByID retrieves an organization by ID
func (r *OrganizationRepository) GetByID(ctx context.Context, id string) (*models.Organization, error) {
	var org models.Organization
	err := r.orgCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&org)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &org, nil
}

// GetBySlug retrieves an organization by slug
func (r *OrganizationRepository) GetBySlug(ctx context.Context, slug string) (*models.Organization, error) {
	var org models.Organization
	err := r.orgCollection.FindOne(ctx, bson.M{"slug": slug}).Decode(&org)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &org, nil
}

// Update updates an existing organization
func (r *OrganizationRepository) Update(ctx context.Context, org *models.Organization) error {
	org.UpdatedAt = time.Now()
	_, err := r.orgCollection.ReplaceOne(ctx, bson.M{"_id": org.ID}, org)
	return err
}

// Delete removes an organization
func (r *OrganizationRepository) Delete(ctx context.Context, id string) error {
	// Also cleanup members and invites
	_, err := r.memberCollection.DeleteMany(ctx, bson.M{"organization_id": id})
	if err != nil {
		return err
	}

	_, err = r.inviteCollection.DeleteMany(ctx, bson.M{"organization_id": id})
	if err != nil {
		return err
	}

	_, err = r.orgCollection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

// List retrieves organizations with pagination
func (r *OrganizationRepository) List(ctx context.Context, limit, offset int) ([]*models.Organization, error) {
	opts := &options.FindOptions{
		Sort:  bson.D{{Key: "created_at", Value: -1}},
		Limit: int64ptr(limit),
		Skip:  int64ptr(offset),
	}

	cursor, err := r.orgCollection.Find(ctx, bson.M{"public": true}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var orgs []*models.Organization
	if err = cursor.All(ctx, &orgs); err != nil {
		return nil, err
	}
	return orgs, nil
}

// Search searches organizations by query
func (r *OrganizationRepository) Search(ctx context.Context, query string, limit, offset int) ([]*models.Organization, error) {
	filter := bson.M{
		"public": true,
		"$or": []bson.M{
			{"name": bson.M{"$regex": query, "$options": "i"}},
			{"description": bson.M{"$regex": query, "$options": "i"}},
		},
	}

	opts := &options.FindOptions{
		Sort:  bson.D{{Key: "created_at", Value: -1}},
		Limit: int64ptr(limit),
		Skip:  int64ptr(offset),
	}

	cursor, err := r.orgCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var orgs []*models.Organization
	if err = cursor.All(ctx, &orgs); err != nil {
		return nil, err
	}
	return orgs, nil
}

// GetByOwner retrieves organizations owned by a user
func (r *OrganizationRepository) GetByOwner(ctx context.Context, ownerID string) ([]*models.Organization, error) {
	filter := bson.M{"owner_id": ownerID}

	cursor, err := r.orgCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var orgs []*models.Organization
	if err = cursor.All(ctx, &orgs); err != nil {
		return nil, err
	}
	return orgs, nil
}

// GetUserOrganizations retrieves organizations where user is a member
func (r *OrganizationRepository) GetUserOrganizations(ctx context.Context, userID string) ([]*models.Organization, error) {
	// Find all organization IDs where user is a member
	cursor, err := r.memberCollection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var members []models.OrganizationMember
	if err = cursor.All(ctx, &members); err != nil {
		return nil, err
	}

	if len(members) == 0 {
		return []*models.Organization{}, nil
	}

	// Extract organization IDs
	var orgIDs []string
	for _, member := range members {
		orgIDs = append(orgIDs, member.OrganizationID)
	}

	// Find organizations
	cursor, err = r.orgCollection.Find(ctx, bson.M{"_id": bson.M{"$in": orgIDs}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var orgs []*models.Organization
	if err = cursor.All(ctx, &orgs); err != nil {
		return nil, err
	}
	return orgs, nil
}

// AddMember adds a member to an organization
func (r *OrganizationRepository) AddMember(ctx context.Context, member *models.OrganizationMember) error {
	if member.ID == "" {
		member.ID = primitive.NewObjectID().Hex()
	}
	member.JoinedAt = time.Now()

	_, err := r.memberCollection.InsertOne(ctx, member)
	if err != nil {
		return err
	}

	// Update member count
	_, err = r.orgCollection.UpdateOne(
		ctx,
		bson.M{"_id": member.OrganizationID},
		bson.M{"$inc": bson.M{"member_count": 1}},
	)
	return err
}

// RemoveMember removes a member from an organization
func (r *OrganizationRepository) RemoveMember(ctx context.Context, orgID, userID string) error {
	_, err := r.memberCollection.DeleteOne(ctx, bson.M{
		"organization_id": orgID,
		"user_id":         userID,
	})
	if err != nil {
		return err
	}

	// Update member count
	_, err = r.orgCollection.UpdateOne(
		ctx,
		bson.M{"_id": orgID},
		bson.M{"$inc": bson.M{"member_count": -1}},
	)
	return err
}

// UpdateMemberRole updates a member's role
func (r *OrganizationRepository) UpdateMemberRole(ctx context.Context, orgID, userID, role string) error {
	_, err := r.memberCollection.UpdateOne(
		ctx,
		bson.M{
			"organization_id": orgID,
			"user_id":         userID,
		},
		bson.M{"$set": bson.M{"role": role}},
	)
	return err
}

// GetMembers retrieves all members of an organization
func (r *OrganizationRepository) GetMembers(ctx context.Context, orgID string) ([]*models.OrganizationMember, error) {
	cursor, err := r.memberCollection.Find(ctx, bson.M{"organization_id": orgID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var members []*models.OrganizationMember
	if err = cursor.All(ctx, &members); err != nil {
		return nil, err
	}
	return members, nil
}

// GetMember retrieves a specific member
func (r *OrganizationRepository) GetMember(ctx context.Context, orgID, userID string) (*models.OrganizationMember, error) {
	var member models.OrganizationMember
	err := r.memberCollection.FindOne(ctx, bson.M{
		"organization_id": orgID,
		"user_id":         userID,
	}).Decode(&member)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &member, nil
}

// IsMember checks if a user is a member of an organization
func (r *OrganizationRepository) IsMember(ctx context.Context, orgID, userID string) (bool, error) {
	count, err := r.memberCollection.CountDocuments(ctx, bson.M{
		"organization_id": orgID,
		"user_id":         userID,
	})
	return count > 0, err
}

// CreateInvite creates an organization invite
func (r *OrganizationRepository) CreateInvite(ctx context.Context, invite *models.OrganizationInvite) error {
	if invite.ID == "" {
		invite.ID = primitive.NewObjectID().Hex()
	}
	invite.CreatedAt = time.Now()

	_, err := r.inviteCollection.InsertOne(ctx, invite)
	return err
}

// GetInvite retrieves an invite by token
func (r *OrganizationRepository) GetInvite(ctx context.Context, token string) (*models.OrganizationInvite, error) {
	var invite models.OrganizationInvite
	err := r.inviteCollection.FindOne(ctx, bson.M{"token": token}).Decode(&invite)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &invite, nil
}

// GetInvitesByOrganization retrieves all invites for an organization
func (r *OrganizationRepository) GetInvitesByOrganization(ctx context.Context, orgID string) ([]*models.OrganizationInvite, error) {
	cursor, err := r.inviteCollection.Find(ctx, bson.M{"organization_id": orgID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var invites []*models.OrganizationInvite
	if err = cursor.All(ctx, &invites); err != nil {
		return nil, err
	}
	return invites, nil
}

// AcceptInvite marks an invite as accepted
func (r *OrganizationRepository) AcceptInvite(ctx context.Context, token string, userID string) error {
	now := time.Now()
	_, err := r.inviteCollection.UpdateOne(
		ctx,
		bson.M{"token": token},
		bson.M{"$set": bson.M{"accepted_at": &now}},
	)
	return err
}

// DeleteInvite removes an invite
func (r *OrganizationRepository) DeleteInvite(ctx context.Context, id string) error {
	_, err := r.inviteCollection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

// CleanupExpiredInvites removes expired invites
func (r *OrganizationRepository) CleanupExpiredInvites(ctx context.Context) error {
	_, err := r.inviteCollection.DeleteMany(ctx, bson.M{
		"expires_at": bson.M{"$lt": time.Now()},
		"accepted_at": nil,
	})
	return err
}