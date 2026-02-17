package project

import (
	"context"

	"github.com/google/uuid"

	apiErrors "paas-core/apps/api/internal/errors"
	"paas-core/apps/api/internal/model"
)

// Service defines the project service interface.
type Service interface {
	CreateProject(ctx context.Context, orgID uuid.UUID, req CreateProjectRequest) (*ProjectResponse, error)
	GetProject(ctx context.Context, projectID uuid.UUID) (*ProjectResponse, error)
	UpdateProject(ctx context.Context, projectID uuid.UUID, req UpdateProjectRequest) (*ProjectResponse, error)
	DeleteProject(ctx context.Context, projectID uuid.UUID) error
	ListProjects(ctx context.Context, orgID uuid.UUID) ([]ProjectResponse, error)

	// Deployments
	CreateDeployment(ctx context.Context, projectID uuid.UUID, version, commitSHA string) (*DeploymentResponse, error)
	ListDeployments(ctx context.Context, projectID uuid.UUID, limit int) ([]DeploymentResponse, error)

	// Env Vars
	SetEnvVar(ctx context.Context, projectID uuid.UUID, req SetEnvVarRequest) (*EnvVarResponse, error)
	ListEnvVars(ctx context.Context, projectID uuid.UUID) ([]EnvVarResponse, error)
	DeleteEnvVar(ctx context.Context, envVarID uuid.UUID) error
}

type service struct {
	repo Repository
}

// NewService creates a new project service.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// --- Project CRUD ---

func (s *service) CreateProject(ctx context.Context, orgID uuid.UUID, req CreateProjectRequest) (*ProjectResponse, error) {
	p := &model.Project{
		OrgID:       orgID,
		Name:        req.Name,
		Description: req.Description,
		RepoURL:     req.RepoURL,
	}

	if err := s.repo.Create(ctx, p); err != nil {
		return nil, apiErrors.InternalServerError(err)
	}

	return toProjectResponse(p), nil
}

func (s *service) GetProject(ctx context.Context, projectID uuid.UUID) (*ProjectResponse, error) {
	p, err := s.repo.FindByID(ctx, projectID)
	if err != nil {
		return nil, apiErrors.InternalServerError(err)
	}
	if p == nil {
		return nil, apiErrors.NotFound("Project not found")
	}
	return toProjectResponse(p), nil
}

func (s *service) UpdateProject(ctx context.Context, projectID uuid.UUID, req UpdateProjectRequest) (*ProjectResponse, error) {
	p, err := s.repo.FindByID(ctx, projectID)
	if err != nil {
		return nil, apiErrors.InternalServerError(err)
	}
	if p == nil {
		return nil, apiErrors.NotFound("Project not found")
	}

	if req.Name != "" {
		p.Name = req.Name
	}
	if req.Description != "" {
		p.Description = req.Description
	}
	if req.RepoURL != "" {
		p.RepoURL = req.RepoURL
	}

	if err := s.repo.Update(ctx, p); err != nil {
		return nil, apiErrors.InternalServerError(err)
	}
	return toProjectResponse(p), nil
}

func (s *service) DeleteProject(ctx context.Context, projectID uuid.UUID) error {
	if err := s.repo.Delete(ctx, projectID); err != nil {
		return apiErrors.InternalServerError(err)
	}
	return nil
}

func (s *service) ListProjects(ctx context.Context, orgID uuid.UUID) ([]ProjectResponse, error) {
	projects, err := s.repo.ListByOrg(ctx, orgID)
	if err != nil {
		return nil, apiErrors.InternalServerError(err)
	}
	var responses []ProjectResponse
	for _, p := range projects {
		responses = append(responses, *toProjectResponse(&p))
	}
	return responses, nil
}

// --- Deployments ---

func (s *service) CreateDeployment(ctx context.Context, projectID uuid.UUID, version, commitSHA string) (*DeploymentResponse, error) {
	d := &model.Deployment{
		ProjectID: projectID,
		Version:   version,
		CommitSHA: commitSHA,
		Status:    "pending",
	}

	if err := s.repo.CreateDeployment(ctx, d); err != nil {
		return nil, apiErrors.InternalServerError(err)
	}
	return toDeploymentResponse(d), nil
}

func (s *service) ListDeployments(ctx context.Context, projectID uuid.UUID, limit int) ([]DeploymentResponse, error) {
	deployments, err := s.repo.ListDeployments(ctx, projectID, limit)
	if err != nil {
		return nil, apiErrors.InternalServerError(err)
	}
	var responses []DeploymentResponse
	for _, d := range deployments {
		responses = append(responses, *toDeploymentResponse(&d))
	}
	return responses, nil
}

// --- Env Vars ---

func (s *service) SetEnvVar(ctx context.Context, projectID uuid.UUID, req SetEnvVarRequest) (*EnvVarResponse, error) {
	ev := &model.EnvVar{
		ProjectID: projectID,
		Key:       req.Key,
		Value:     req.Value,
		IsSecret:  req.IsSecret,
	}

	if err := s.repo.SetEnvVar(ctx, ev); err != nil {
		return nil, apiErrors.InternalServerError(err)
	}
	return toEnvVarResponse(ev), nil
}

func (s *service) ListEnvVars(ctx context.Context, projectID uuid.UUID) ([]EnvVarResponse, error) {
	envVars, err := s.repo.ListEnvVars(ctx, projectID)
	if err != nil {
		return nil, apiErrors.InternalServerError(err)
	}
	var responses []EnvVarResponse
	for _, ev := range envVars {
		responses = append(responses, *toEnvVarResponse(&ev))
	}
	return responses, nil
}

func (s *service) DeleteEnvVar(ctx context.Context, envVarID uuid.UUID) error {
	if err := s.repo.DeleteEnvVar(ctx, envVarID); err != nil {
		return apiErrors.InternalServerError(err)
	}
	return nil
}

// --- Helpers ---

func toProjectResponse(p *model.Project) *ProjectResponse {
	return &ProjectResponse{
		ID:          p.ID,
		OrgID:       p.OrgID,
		Name:        p.Name,
		Description: p.Description,
		RepoURL:     p.RepoURL,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func toDeploymentResponse(d *model.Deployment) *DeploymentResponse {
	return &DeploymentResponse{
		ID:         d.ID,
		ProjectID:  d.ProjectID,
		Version:    d.Version,
		Status:     d.Status,
		CommitSHA:  d.CommitSHA,
		StartedAt:  d.StartedAt,
		FinishedAt: d.FinishedAt,
		CreatedAt:  d.CreatedAt,
	}
}

func toEnvVarResponse(ev *model.EnvVar) *EnvVarResponse {
	value := ev.Value
	if ev.IsSecret {
		value = "********"
	}
	return &EnvVarResponse{
		ID:        ev.ID,
		ProjectID: ev.ProjectID,
		Key:       ev.Key,
		Value:     value,
		IsSecret:  ev.IsSecret,
		CreatedAt: ev.CreatedAt,
	}
}
