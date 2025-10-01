package models

import "time"

// Organization represents an organization that can own templates
type Organization struct {
	ID          string    `json:"id" bson:"_id"`
	Name        string    `json:"name" bson:"name"`
	Slug        string    `json:"slug" bson:"slug"`
	Description string    `json:"description" bson:"description"`
	Website     string    `json:"website" bson:"website"`
	OwnerID     string    `json:"owner_id" bson:"owner_id"`
	Public      bool      `json:"public" bson:"public"`
	CreatedAt   time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" bson:"updated_at"`
	MemberCount int       `json:"member_count" bson:"member_count"`
}

// OrganizationMember represents a user's membership in an organization
type OrganizationMember struct {
	ID             string    `json:"id" bson:"_id"`
	OrganizationID string    `json:"organization_id" bson:"organization_id"`
	UserID         string    `json:"user_id" bson:"user_id"`
	Role           string    `json:"role" bson:"role"` // owner, admin, member
	JoinedAt       time.Time `json:"joined_at" bson:"joined_at"`
}

// OrganizationInvite represents an invitation to join an organization
type OrganizationInvite struct {
	ID             string    `json:"id" bson:"_id"`
	OrganizationID string    `json:"organization_id" bson:"organization_id"`
	Email          string    `json:"email" bson:"email"`
	Role           string    `json:"role" bson:"role"`
	Token          string    `json:"token" bson:"token"`
	InvitedBy      string    `json:"invited_by" bson:"invited_by"`
	CreatedAt      time.Time `json:"created_at" bson:"created_at"`
	ExpiresAt      time.Time `json:"expires_at" bson:"expires_at"`
	AcceptedAt     *time.Time `json:"accepted_at,omitempty" bson:"accepted_at,omitempty"`
}

// OrganizationRole constants
const (
	RoleOwner  = "owner"
	RoleAdmin  = "admin"
	RoleMember = "member"
)

// IsValid checks if a role is valid
func (r OrganizationMember) IsValidRole() bool {
	return r.Role == RoleOwner || r.Role == RoleAdmin || r.Role == RoleMember
}

// CanManageMembers checks if the role can manage organization members
func (r OrganizationMember) CanManageMembers() bool {
	return r.Role == RoleOwner || r.Role == RoleAdmin
}

// CanDeleteOrganization checks if the role can delete the organization
func (r OrganizationMember) CanDeleteOrganization() bool {
	return r.Role == RoleOwner
}