package project

import (
	"time"

	"github.com/google/uuid"
)

// CreateProjectRequest is the DTO for creating a project.
type CreateProjectRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=100"`
	Description string `json:"description" binding:"omitempty,max=500"`
	RepoURL     string `json:"repo_url" binding:"omitempty,url"`
}

// UpdateProjectRequest is the DTO for updating a project.
type UpdateProjectRequest struct {
	Name        string `json:"name" binding:"omitempty,min=2,max=100"`
	Description string `json:"description" binding:"omitempty,max=500"`
	RepoURL     string `json:"repo_url" binding:"omitempty,url"`
}

// SetEnvVarRequest is the DTO for creating/updating an env var.
type SetEnvVarRequest struct {
	Key      string `json:"key" binding:"required,min=1,max=255"`
	Value    string `json:"value" binding:"required"`
	IsSecret bool   `json:"is_secret"`
}

// ProjectResponse is the public project representation.
type ProjectResponse struct {
	ID          uuid.UUID `json:"id"`
	OrgID       uuid.UUID `json:"org_id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	RepoURL     string    `json:"repo_url,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// DeploymentResponse is the public deployment representation.
type DeploymentResponse struct {
	ID         uuid.UUID  `json:"id"`
	ProjectID  uuid.UUID  `json:"project_id"`
	Version    string     `json:"version"`
	Status     string     `json:"status"`
	CommitSHA  string     `json:"commit_sha,omitempty"`
	StartedAt  *time.Time `json:"started_at,omitempty"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// EnvVarResponse is the public env var representation (value redacted for secrets).
type EnvVarResponse struct {
	ID        uuid.UUID `json:"id"`
	ProjectID uuid.UUID `json:"project_id"`
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	IsSecret  bool      `json:"is_secret"`
	CreatedAt time.Time `json:"created_at"`
}
