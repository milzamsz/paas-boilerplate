package org

import (
	"time"

	"github.com/google/uuid"
)

// CreateOrgRequest is the DTO for creating a new organization.
type CreateOrgRequest struct {
	Name string `json:"name" binding:"required,min=2,max=100"`
	Slug string `json:"slug" binding:"required,min=2,max=50,alphanumunicode"`
}

// UpdateOrgRequest is the DTO for updating an organization.
type UpdateOrgRequest struct {
	Name    string `json:"name" binding:"omitempty,min=2,max=100"`
	LogoURL string `json:"logo_url" binding:"omitempty,url"`
}

// InviteMemberRequest is the DTO for inviting a member to an org.
type InviteMemberRequest struct {
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role" binding:"required,oneof=admin developer viewer"`
}

// UpdateMemberRoleRequest changes a member's role.
type UpdateMemberRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=owner admin developer viewer"`
}

// OrgResponse is the public representation of an organization.
type OrgResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	LogoURL   string    `json:"logo_url,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// MemberResponse is the public representation of a membership.
type MemberResponse struct {
	ID       uuid.UUID    `json:"id"`
	UserID   uuid.UUID    `json:"user_id"`
	OrgID    uuid.UUID    `json:"org_id"`
	Role     string       `json:"role"`
	JoinedAt time.Time    `json:"joined_at"`
	User     *MemberUser  `json:"user,omitempty"`
}

// MemberUser is the nested user info within a membership.
type MemberUser struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	AvatarURL string    `json:"avatar_url,omitempty"`
}

// InviteResponse is the public representation of an invite.
type InviteResponse struct {
	ID        uuid.UUID  `json:"id"`
	OrgID     uuid.UUID  `json:"org_id"`
	Email     string     `json:"email"`
	Role      string     `json:"role"`
	ExpiresAt time.Time  `json:"expires_at"`
	AcceptedAt *time.Time `json:"accepted_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}
