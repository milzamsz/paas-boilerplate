package project

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"paas-core/apps/api/internal/model"
)

// Repository defines the project data access interface.
type Repository interface {
	Create(ctx context.Context, p *model.Project) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.Project, error)
	Update(ctx context.Context, p *model.Project) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByOrg(ctx context.Context, orgID uuid.UUID) ([]model.Project, error)
	CountByOrg(ctx context.Context, orgID uuid.UUID) (int64, error)

	// Deployments
	CreateDeployment(ctx context.Context, d *model.Deployment) error
	FindDeploymentByID(ctx context.Context, id uuid.UUID) (*model.Deployment, error)
	UpdateDeployment(ctx context.Context, d *model.Deployment) error
	ListDeployments(ctx context.Context, projectID uuid.UUID, limit int) ([]model.Deployment, error)

	// Env Vars
	SetEnvVar(ctx context.Context, ev *model.EnvVar) error
	ListEnvVars(ctx context.Context, projectID uuid.UUID) ([]model.EnvVar, error)
	DeleteEnvVar(ctx context.Context, id uuid.UUID) error
}

type repository struct {
	db *gorm.DB
}

// NewRepository creates a new project repository.
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// --- Project CRUD ---

func (r *repository) Create(ctx context.Context, p *model.Project) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *repository) FindByID(ctx context.Context, id uuid.UUID) (*model.Project, error) {
	var p model.Project
	err := r.db.WithContext(ctx).First(&p, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &p, err
}

func (r *repository) Update(ctx context.Context, p *model.Project) error {
	return r.db.WithContext(ctx).Save(p).Error
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Project{}, "id = ?", id).Error
}

func (r *repository) ListByOrg(ctx context.Context, orgID uuid.UUID) ([]model.Project, error) {
	var projects []model.Project
	err := r.db.WithContext(ctx).
		Where("org_id = ?", orgID).
		Order("created_at DESC").
		Find(&projects).Error
	return projects, err
}

func (r *repository) CountByOrg(ctx context.Context, orgID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Project{}).
		Where("org_id = ?", orgID).
		Count(&count).Error
	return count, err
}

// --- Deployments ---

func (r *repository) CreateDeployment(ctx context.Context, d *model.Deployment) error {
	return r.db.WithContext(ctx).Create(d).Error
}

func (r *repository) FindDeploymentByID(ctx context.Context, id uuid.UUID) (*model.Deployment, error) {
	var d model.Deployment
	err := r.db.WithContext(ctx).First(&d, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &d, err
}

func (r *repository) UpdateDeployment(ctx context.Context, d *model.Deployment) error {
	return r.db.WithContext(ctx).Save(d).Error
}

func (r *repository) ListDeployments(ctx context.Context, projectID uuid.UUID, limit int) ([]model.Deployment, error) {
	var deployments []model.Deployment
	q := r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Order("created_at DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	err := q.Find(&deployments).Error
	return deployments, err
}

// --- Env Vars ---

func (r *repository) SetEnvVar(ctx context.Context, ev *model.EnvVar) error {
	// Upsert: if key exists for this project, update it
	var existing model.EnvVar
	err := r.db.WithContext(ctx).
		Where("project_id = ? AND key = ?", ev.ProjectID, ev.Key).
		First(&existing).Error
	if err == nil {
		existing.Value = ev.Value
		existing.IsSecret = ev.IsSecret
		return r.db.WithContext(ctx).Save(&existing).Error
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return r.db.WithContext(ctx).Create(ev).Error
	}
	return err
}

func (r *repository) ListEnvVars(ctx context.Context, projectID uuid.UUID) ([]model.EnvVar, error) {
	var envVars []model.EnvVar
	err := r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Order("key ASC").
		Find(&envVars).Error
	return envVars, err
}

func (r *repository) DeleteEnvVar(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.EnvVar{}, "id = ?", id).Error
}
