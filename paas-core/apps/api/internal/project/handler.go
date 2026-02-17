package project

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	apiErrors "paas-core/apps/api/internal/errors"
)

// Handler handles project-related HTTP requests.
type Handler struct {
	projectService Service
}

// NewHandler creates a new project handler.
func NewHandler(projectService Service) *Handler {
	return &Handler{projectService: projectService}
}

// --- Project CRUD ---

// CreateProject godoc
// @Summary Create a new project
// @Tags projects
// @Security BearerAuth
// @Param orgId path string true "Organization ID"
// @Param request body CreateProjectRequest true "Create project"
// @Success 201 {object} errors.Response{data=ProjectResponse}
// @Router /api/v1/orgs/{orgId}/projects [post]
func (h *Handler) CreateProject(c *gin.Context) {
	orgID := c.MustGet("org_id").(uuid.UUID)

	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apiErrors.FromGinValidation(err))
		return
	}

	project, err := h.projectService.CreateProject(c.Request.Context(), orgID, req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, apiErrors.Success(project))
}

// ListProjects godoc
// @Summary List projects in an organization
// @Tags projects
// @Security BearerAuth
// @Param orgId path string true "Organization ID"
// @Success 200 {object} errors.Response{data=[]ProjectResponse}
// @Router /api/v1/orgs/{orgId}/projects [get]
func (h *Handler) ListProjects(c *gin.Context) {
	orgID := c.MustGet("org_id").(uuid.UUID)

	projects, err := h.projectService.ListProjects(c.Request.Context(), orgID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(projects))
}

// GetProject godoc
// @Summary Get project details
// @Tags projects
// @Security BearerAuth
// @Param orgId path string true "Organization ID"
// @Param projectId path string true "Project ID"
// @Success 200 {object} errors.Response{data=ProjectResponse}
// @Router /api/v1/orgs/{orgId}/projects/{projectId} [get]
func (h *Handler) GetProject(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("projectId"))
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid project ID"))
		return
	}

	project, err := h.projectService.GetProject(c.Request.Context(), projectID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(project))
}

// UpdateProject godoc
// @Summary Update a project
// @Tags projects
// @Security BearerAuth
// @Param orgId path string true "Organization ID"
// @Param projectId path string true "Project ID"
// @Param request body UpdateProjectRequest true "Update project"
// @Success 200 {object} errors.Response{data=ProjectResponse}
// @Router /api/v1/orgs/{orgId}/projects/{projectId} [put]
func (h *Handler) UpdateProject(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("projectId"))
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid project ID"))
		return
	}

	var req UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apiErrors.FromGinValidation(err))
		return
	}

	project, err := h.projectService.UpdateProject(c.Request.Context(), projectID, req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(project))
}

// DeleteProject godoc
// @Summary Delete a project
// @Tags projects
// @Security BearerAuth
// @Param orgId path string true "Organization ID"
// @Param projectId path string true "Project ID"
// @Success 200 {object} errors.Response
// @Router /api/v1/orgs/{orgId}/projects/{projectId} [delete]
func (h *Handler) DeleteProject(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("projectId"))
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid project ID"))
		return
	}

	if err := h.projectService.DeleteProject(c.Request.Context(), projectID); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(gin.H{"message": "Project deleted"}))
}

// --- Deployments ---

// CreateDeployment godoc
// @Summary Trigger a deployment
// @Tags projects
// @Security BearerAuth
// @Param orgId path string true "Organization ID"
// @Param projectId path string true "Project ID"
// @Success 201 {object} errors.Response{data=DeploymentResponse}
// @Router /api/v1/orgs/{orgId}/projects/{projectId}/deployments [post]
func (h *Handler) CreateDeployment(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("projectId"))
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid project ID"))
		return
	}

	type deployReq struct {
		Version   string `json:"version"`
		CommitSHA string `json:"commit_sha"`
	}
	var req deployReq
	_ = c.ShouldBindJSON(&req) // optional body

	deployment, err := h.projectService.CreateDeployment(c.Request.Context(), projectID, req.Version, req.CommitSHA)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, apiErrors.Success(deployment))
}

// ListDeployments godoc
// @Summary List deployments for a project
// @Tags projects
// @Security BearerAuth
// @Param orgId path string true "Organization ID"
// @Param projectId path string true "Project ID"
// @Param limit query int false "Limit" default(20)
// @Success 200 {object} errors.Response{data=[]DeploymentResponse}
// @Router /api/v1/orgs/{orgId}/projects/{projectId}/deployments [get]
func (h *Handler) ListDeployments(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("projectId"))
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid project ID"))
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	deployments, err := h.projectService.ListDeployments(c.Request.Context(), projectID, limit)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(deployments))
}

// --- Env Vars ---

// SetEnvVar godoc
// @Summary Set an environment variable (upsert by key)
// @Tags projects
// @Security BearerAuth
// @Param orgId path string true "Organization ID"
// @Param projectId path string true "Project ID"
// @Param request body SetEnvVarRequest true "Env var"
// @Success 200 {object} errors.Response{data=EnvVarResponse}
// @Router /api/v1/orgs/{orgId}/projects/{projectId}/env [post]
func (h *Handler) SetEnvVar(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("projectId"))
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid project ID"))
		return
	}

	var req SetEnvVarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apiErrors.FromGinValidation(err))
		return
	}

	envVar, err := h.projectService.SetEnvVar(c.Request.Context(), projectID, req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(envVar))
}

// ListEnvVars godoc
// @Summary List environment variables for a project
// @Tags projects
// @Security BearerAuth
// @Param orgId path string true "Organization ID"
// @Param projectId path string true "Project ID"
// @Success 200 {object} errors.Response{data=[]EnvVarResponse}
// @Router /api/v1/orgs/{orgId}/projects/{projectId}/env [get]
func (h *Handler) ListEnvVars(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("projectId"))
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid project ID"))
		return
	}

	envVars, err := h.projectService.ListEnvVars(c.Request.Context(), projectID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(envVars))
}

// DeleteEnvVar godoc
// @Summary Delete an environment variable
// @Tags projects
// @Security BearerAuth
// @Param orgId path string true "Organization ID"
// @Param projectId path string true "Project ID"
// @Param envVarId path string true "Env Var ID"
// @Success 200 {object} errors.Response
// @Router /api/v1/orgs/{orgId}/projects/{projectId}/env/{envVarId} [delete]
func (h *Handler) DeleteEnvVar(c *gin.Context) {
	envVarID, err := uuid.Parse(c.Param("envVarId"))
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid env var ID"))
		return
	}

	if err := h.projectService.DeleteEnvVar(c.Request.Context(), envVarID); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(gin.H{"message": "Environment variable deleted"}))
}
