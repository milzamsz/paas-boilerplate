package org

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	authPkg "paas-core/apps/api/internal/auth"
	apiErrors "paas-core/apps/api/internal/errors"
)

// Handler handles org-related HTTP requests.
type Handler struct {
	orgService Service
}

// NewHandler creates a new org handler.
func NewHandler(orgService Service) *Handler {
	return &Handler{orgService: orgService}
}

// --- Org CRUD ---

// CreateOrg godoc
// @Summary Create a new organization
// @Tags orgs
// @Security BearerAuth
// @Param request body CreateOrgRequest true "Create org"
// @Success 200 {object} errors.Response{data=OrgResponse}
// @Router /api/v1/orgs [post]
func (h *Handler) CreateOrg(c *gin.Context) {
	claims := c.MustGet("claims").(*authPkg.Claims)

	var req CreateOrgRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apiErrors.FromGinValidation(err))
		return
	}

	org, err := h.orgService.CreateOrg(c.Request.Context(), claims.UserID, req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, apiErrors.Success(org))
}

// ListOrgs godoc
// @Summary List organizations the user belongs to
// @Tags orgs
// @Security BearerAuth
// @Success 200 {object} errors.Response{data=[]OrgResponse}
// @Router /api/v1/orgs [get]
func (h *Handler) ListOrgs(c *gin.Context) {
	claims := c.MustGet("claims").(*authPkg.Claims)

	orgs, err := h.orgService.ListUserOrgs(c.Request.Context(), claims.UserID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(orgs))
}

// GetOrg godoc
// @Summary Get organization details
// @Tags orgs
// @Security BearerAuth
// @Param orgId path string true "Organization ID"
// @Success 200 {object} errors.Response{data=OrgResponse}
// @Router /api/v1/orgs/{orgId} [get]
func (h *Handler) GetOrg(c *gin.Context) {
	orgID := c.MustGet("org_id").(uuid.UUID)

	org, err := h.orgService.GetOrg(c.Request.Context(), orgID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(org))
}

// UpdateOrg godoc
// @Summary Update organization
// @Tags orgs
// @Security BearerAuth
// @Param orgId path string true "Organization ID"
// @Param request body UpdateOrgRequest true "Update org"
// @Success 200 {object} errors.Response{data=OrgResponse}
// @Router /api/v1/orgs/{orgId} [put]
func (h *Handler) UpdateOrg(c *gin.Context) {
	orgID := c.MustGet("org_id").(uuid.UUID)

	var req UpdateOrgRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apiErrors.FromGinValidation(err))
		return
	}

	org, err := h.orgService.UpdateOrg(c.Request.Context(), orgID, req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(org))
}

// DeleteOrg godoc
// @Summary Delete organization
// @Tags orgs
// @Security BearerAuth
// @Param orgId path string true "Organization ID"
// @Success 200 {object} errors.Response
// @Router /api/v1/orgs/{orgId} [delete]
func (h *Handler) DeleteOrg(c *gin.Context) {
	orgID := c.MustGet("org_id").(uuid.UUID)

	if err := h.orgService.DeleteOrg(c.Request.Context(), orgID); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(gin.H{"message": "Organization deleted"}))
}

// --- Members ---

// ListMembers godoc
// @Summary List organization members
// @Tags orgs
// @Security BearerAuth
// @Param orgId path string true "Organization ID"
// @Success 200 {object} errors.Response{data=[]MemberResponse}
// @Router /api/v1/orgs/{orgId}/members [get]
func (h *Handler) ListMembers(c *gin.Context) {
	orgID := c.MustGet("org_id").(uuid.UUID)

	members, err := h.orgService.ListMembers(c.Request.Context(), orgID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(members))
}

// UpdateMemberRole godoc
// @Summary Update a member's role
// @Tags orgs
// @Security BearerAuth
// @Param orgId path string true "Organization ID"
// @Param memberId path string true "Membership ID"
// @Param request body UpdateMemberRoleRequest true "Update role"
// @Success 200 {object} errors.Response{data=MemberResponse}
// @Router /api/v1/orgs/{orgId}/members/{memberId} [put]
func (h *Handler) UpdateMemberRole(c *gin.Context) {
	memberIDStr := c.Param("memberId")
	memberID, err := uuid.Parse(memberIDStr)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid member ID"))
		return
	}

	var req UpdateMemberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apiErrors.FromGinValidation(err))
		return
	}

	member, err := h.orgService.UpdateMemberRole(c.Request.Context(), memberID, req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(member))
}

// RemoveMember godoc
// @Summary Remove a member from the organization
// @Tags orgs
// @Security BearerAuth
// @Param orgId path string true "Organization ID"
// @Param memberId path string true "Membership ID"
// @Success 200 {object} errors.Response
// @Router /api/v1/orgs/{orgId}/members/{memberId} [delete]
func (h *Handler) RemoveMember(c *gin.Context) {
	memberIDStr := c.Param("memberId")
	memberID, err := uuid.Parse(memberIDStr)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid member ID"))
		return
	}

	if err := h.orgService.RemoveMember(c.Request.Context(), memberID); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(gin.H{"message": "Member removed"}))
}

// --- Invites ---

// InviteMember godoc
// @Summary Invite a member to the organization
// @Tags orgs
// @Security BearerAuth
// @Param orgId path string true "Organization ID"
// @Param request body InviteMemberRequest true "Invite request"
// @Success 201 {object} errors.Response{data=InviteResponse}
// @Router /api/v1/orgs/{orgId}/invites [post]
func (h *Handler) InviteMember(c *gin.Context) {
	orgID := c.MustGet("org_id").(uuid.UUID)
	claims := c.MustGet("claims").(*authPkg.Claims)

	var req InviteMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apiErrors.FromGinValidation(err))
		return
	}

	invite, err := h.orgService.InviteMember(c.Request.Context(), orgID, claims.UserID, req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, apiErrors.Success(invite))
}

// AcceptInvite godoc
// @Summary Accept an organization invite
// @Tags orgs
// @Security BearerAuth
// @Param token path string true "Invite token"
// @Success 200 {object} errors.Response{data=MemberResponse}
// @Router /api/v1/invites/{token}/accept [post]
func (h *Handler) AcceptInvite(c *gin.Context) {
	token := c.Param("token")
	claims := c.MustGet("claims").(*authPkg.Claims)

	member, err := h.orgService.AcceptInvite(c.Request.Context(), token, claims.UserID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(member))
}

// ListInvites godoc
// @Summary List pending invites for an organization
// @Tags orgs
// @Security BearerAuth
// @Param orgId path string true "Organization ID"
// @Success 200 {object} errors.Response{data=[]InviteResponse}
// @Router /api/v1/orgs/{orgId}/invites [get]
func (h *Handler) ListInvites(c *gin.Context) {
	orgID := c.MustGet("org_id").(uuid.UUID)

	invites, err := h.orgService.ListInvites(c.Request.Context(), orgID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(invites))
}

// RevokeInvite godoc
// @Summary Revoke a pending invite
// @Tags orgs
// @Security BearerAuth
// @Param orgId path string true "Organization ID"
// @Param inviteId path string true "Invite ID"
// @Success 200 {object} errors.Response
// @Router /api/v1/orgs/{orgId}/invites/{inviteId} [delete]
func (h *Handler) RevokeInvite(c *gin.Context) {
	inviteIDStr := c.Param("inviteId")
	inviteID, err := uuid.Parse(inviteIDStr)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid invite ID"))
		return
	}

	if err := h.orgService.RevokeInvite(c.Request.Context(), inviteID); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(gin.H{"message": "Invite revoked"}))
}
