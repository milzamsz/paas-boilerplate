package org

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"

	apiErrors "paas-core/apps/api/internal/errors"
	"paas-core/apps/api/internal/model"
)

// Service defines the org service interface.
type Service interface {
	CreateOrg(ctx context.Context, userID uuid.UUID, req CreateOrgRequest) (*OrgResponse, error)
	GetOrg(ctx context.Context, orgID uuid.UUID) (*OrgResponse, error)
	UpdateOrg(ctx context.Context, orgID uuid.UUID, req UpdateOrgRequest) (*OrgResponse, error)
	DeleteOrg(ctx context.Context, orgID uuid.UUID) error
	ListUserOrgs(ctx context.Context, userID uuid.UUID) ([]OrgResponse, error)

	// Members
	ListMembers(ctx context.Context, orgID uuid.UUID) ([]MemberResponse, error)
	UpdateMemberRole(ctx context.Context, membershipID uuid.UUID, req UpdateMemberRoleRequest) (*MemberResponse, error)
	RemoveMember(ctx context.Context, membershipID uuid.UUID) error

	// Invites
	InviteMember(ctx context.Context, orgID, invitedBy uuid.UUID, req InviteMemberRequest) (*InviteResponse, error)
	AcceptInvite(ctx context.Context, token string, userID uuid.UUID) (*MemberResponse, error)
	ListInvites(ctx context.Context, orgID uuid.UUID) ([]InviteResponse, error)
	RevokeInvite(ctx context.Context, inviteID uuid.UUID) error
}

type service struct {
	repo Repository
}

// NewService creates a new org service.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// --- Org CRUD ---

func (s *service) CreateOrg(ctx context.Context, userID uuid.UUID, req CreateOrgRequest) (*OrgResponse, error) {
	existing, err := s.repo.FindBySlug(ctx, req.Slug)
	if err != nil {
		return nil, apiErrors.InternalServerError(err)
	}
	if existing != nil {
		return nil, apiErrors.Conflict("Organization slug already taken")
	}

	org := &model.Org{
		Name: req.Name,
		Slug: req.Slug,
	}

	err = s.repo.Transaction(ctx, func(txCtx context.Context) error {
		if err := s.repo.Create(txCtx, org); err != nil {
			return fmt.Errorf("create org: %w", err)
		}
		membership := &model.Membership{
			UserID: userID,
			OrgID:  org.ID,
			Role:   model.RoleOwner,
		}
		if err := s.repo.CreateMembership(txCtx, membership); err != nil {
			return fmt.Errorf("create membership: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, apiErrors.InternalServerError(err)
	}

	return toOrgResponse(org), nil
}

func (s *service) GetOrg(ctx context.Context, orgID uuid.UUID) (*OrgResponse, error) {
	org, err := s.repo.FindByID(ctx, orgID)
	if err != nil {
		return nil, apiErrors.InternalServerError(err)
	}
	if org == nil {
		return nil, apiErrors.NotFound("Organization not found")
	}
	return toOrgResponse(org), nil
}

func (s *service) UpdateOrg(ctx context.Context, orgID uuid.UUID, req UpdateOrgRequest) (*OrgResponse, error) {
	org, err := s.repo.FindByID(ctx, orgID)
	if err != nil {
		return nil, apiErrors.InternalServerError(err)
	}
	if org == nil {
		return nil, apiErrors.NotFound("Organization not found")
	}

	if req.Name != "" {
		org.Name = req.Name
	}
	if req.LogoURL != "" {
		org.LogoURL = req.LogoURL
	}

	if err := s.repo.Update(ctx, org); err != nil {
		return nil, apiErrors.InternalServerError(err)
	}
	return toOrgResponse(org), nil
}

func (s *service) DeleteOrg(ctx context.Context, orgID uuid.UUID) error {
	if err := s.repo.Delete(ctx, orgID); err != nil {
		return apiErrors.InternalServerError(err)
	}
	return nil
}

func (s *service) ListUserOrgs(ctx context.Context, userID uuid.UUID) ([]OrgResponse, error) {
	orgs, err := s.repo.ListByUser(ctx, userID)
	if err != nil {
		return nil, apiErrors.InternalServerError(err)
	}
	var responses []OrgResponse
	for _, o := range orgs {
		responses = append(responses, *toOrgResponse(&o))
	}
	return responses, nil
}

// --- Members ---

func (s *service) ListMembers(ctx context.Context, orgID uuid.UUID) ([]MemberResponse, error) {
	members, err := s.repo.ListMembers(ctx, orgID)
	if err != nil {
		return nil, apiErrors.InternalServerError(err)
	}
	var responses []MemberResponse
	for _, m := range members {
		responses = append(responses, *toMemberResponse(&m))
	}
	return responses, nil
}

func (s *service) UpdateMemberRole(ctx context.Context, membershipID uuid.UUID, req UpdateMemberRoleRequest) (*MemberResponse, error) {
	var m model.Membership
	m.ID = membershipID
	m.Role = req.Role

	if err := s.repo.UpdateMembership(ctx, &m); err != nil {
		return nil, apiErrors.InternalServerError(err)
	}
	return &MemberResponse{ID: m.ID, Role: m.Role}, nil
}

func (s *service) RemoveMember(ctx context.Context, membershipID uuid.UUID) error {
	return s.repo.DeleteMembership(ctx, membershipID)
}

// --- Invites ---

func (s *service) InviteMember(ctx context.Context, orgID, invitedBy uuid.UUID, req InviteMemberRequest) (*InviteResponse, error) {
	// Check for existing pending invite
	existing, err := s.repo.FindInviteByEmail(ctx, orgID, req.Email)
	if err != nil {
		return nil, apiErrors.InternalServerError(err)
	}
	if existing != nil {
		return nil, apiErrors.Conflict("An invitation for this email already exists")
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, apiErrors.InternalServerError(fmt.Errorf("generate token: %w", err))
	}

	invite := &model.OrgInvite{
		OrgID:     orgID,
		Email:     req.Email,
		Role:      req.Role,
		Token:     hex.EncodeToString(tokenBytes),
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // 7 days
		InvitedBy: invitedBy,
	}

	if err := s.repo.CreateInvite(ctx, invite); err != nil {
		return nil, apiErrors.InternalServerError(err)
	}

	return toInviteResponse(invite), nil
}

func (s *service) AcceptInvite(ctx context.Context, token string, userID uuid.UUID) (*MemberResponse, error) {
	invite, err := s.repo.FindInviteByToken(ctx, token)
	if err != nil {
		return nil, apiErrors.InternalServerError(err)
	}
	if invite == nil {
		return nil, apiErrors.NotFound("Invitation not found")
	}
	if invite.AcceptedAt != nil {
		return nil, apiErrors.Conflict("Invitation already accepted")
	}
	if time.Now().After(invite.ExpiresAt) {
		return nil, apiErrors.BadRequest("Invitation has expired")
	}

	var membership *model.Membership
	err = s.repo.Transaction(ctx, func(txCtx context.Context) error {
		now := time.Now()
		invite.AcceptedAt = &now
		if err := s.repo.UpdateInvite(txCtx, invite); err != nil {
			return err
		}

		membership = &model.Membership{
			UserID: userID,
			OrgID:  invite.OrgID,
			Role:   invite.Role,
		}
		return s.repo.CreateMembership(txCtx, membership)
	})
	if err != nil {
		return nil, apiErrors.InternalServerError(err)
	}

	return &MemberResponse{
		ID:       membership.ID,
		UserID:   membership.UserID,
		OrgID:    membership.OrgID,
		Role:     membership.Role,
		JoinedAt: membership.JoinedAt,
	}, nil
}

func (s *service) ListInvites(ctx context.Context, orgID uuid.UUID) ([]InviteResponse, error) {
	invites, err := s.repo.ListInvites(ctx, orgID)
	if err != nil {
		return nil, apiErrors.InternalServerError(err)
	}
	var responses []InviteResponse
	for _, inv := range invites {
		responses = append(responses, *toInviteResponse(&inv))
	}
	return responses, nil
}

func (s *service) RevokeInvite(ctx context.Context, inviteID uuid.UUID) error {
	return s.repo.DeleteInvite(ctx, inviteID)
}

// --- Helpers ---

func toOrgResponse(o *model.Org) *OrgResponse {
	return &OrgResponse{
		ID:        o.ID,
		Name:      o.Name,
		Slug:      o.Slug,
		LogoURL:   o.LogoURL,
		CreatedAt: o.CreatedAt,
	}
}

func toMemberResponse(m *model.Membership) *MemberResponse {
	resp := &MemberResponse{
		ID:       m.ID,
		UserID:   m.UserID,
		OrgID:    m.OrgID,
		Role:     m.Role,
		JoinedAt: m.JoinedAt,
	}
	if m.User.ID != uuid.Nil {
		resp.User = &MemberUser{
			ID:        m.User.ID,
			Name:      m.User.Name,
			Email:     m.User.Email,
			AvatarURL: m.User.AvatarURL,
		}
	}
	return resp
}

func toInviteResponse(inv *model.OrgInvite) *InviteResponse {
	return &InviteResponse{
		ID:         inv.ID,
		OrgID:      inv.OrgID,
		Email:      inv.Email,
		Role:       inv.Role,
		ExpiresAt:  inv.ExpiresAt,
		AcceptedAt: inv.AcceptedAt,
		CreatedAt:  inv.CreatedAt,
	}
}
