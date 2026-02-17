package billing

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"paas-core/apps/api/internal/model"
)

// Repository defines the billing data access interface.
type Repository interface {
	// Plans
	ListActivePlans(ctx context.Context) ([]model.BillingPlan, error)
	FindPlanByID(ctx context.Context, id uuid.UUID) (*model.BillingPlan, error)
	FindPlanBySlug(ctx context.Context, slug string) (*model.BillingPlan, error)

	// Subscriptions
	CreateSubscription(ctx context.Context, s *model.Subscription) error
	FindActiveSubscription(ctx context.Context, orgID uuid.UUID) (*model.Subscription, error)
	UpdateSubscription(ctx context.Context, s *model.Subscription) error

	// Invoices
	CreateInvoice(ctx context.Context, inv *model.Invoice) error
	FindInvoiceByID(ctx context.Context, id uuid.UUID) (*model.Invoice, error)
	FindInvoiceByXenditID(ctx context.Context, xenditID string) (*model.Invoice, error)
	UpdateInvoice(ctx context.Context, inv *model.Invoice) error
	ListInvoicesByOrg(ctx context.Context, orgID uuid.UUID) ([]model.Invoice, error)

	// Counts for usage
	CountProjectsByOrg(ctx context.Context, orgID uuid.UUID) (int, error)
	CountDeploymentsByOrg(ctx context.Context, orgID uuid.UUID) (int, error)
	CountMembersByOrg(ctx context.Context, orgID uuid.UUID) (int, error)
}

type repository struct {
	db *gorm.DB
}

// NewRepository creates a new billing repository.
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// --- Plans ---

func (r *repository) ListActivePlans(ctx context.Context) ([]model.BillingPlan, error) {
	var plans []model.BillingPlan
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("price_monthly ASC").
		Find(&plans).Error
	return plans, err
}

func (r *repository) FindPlanByID(ctx context.Context, id uuid.UUID) (*model.BillingPlan, error) {
	var plan model.BillingPlan
	err := r.db.WithContext(ctx).First(&plan, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &plan, err
}

func (r *repository) FindPlanBySlug(ctx context.Context, slug string) (*model.BillingPlan, error) {
	var plan model.BillingPlan
	err := r.db.WithContext(ctx).First(&plan, "slug = ?", slug).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &plan, err
}

// --- Subscriptions ---

func (r *repository) CreateSubscription(ctx context.Context, s *model.Subscription) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *repository) FindActiveSubscription(ctx context.Context, orgID uuid.UUID) (*model.Subscription, error) {
	var sub model.Subscription
	err := r.db.WithContext(ctx).
		Preload("Plan").
		Where("org_id = ? AND status IN ?", orgID, []string{"active", "trialing"}).
		Order("created_at DESC").
		First(&sub).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &sub, err
}

func (r *repository) UpdateSubscription(ctx context.Context, s *model.Subscription) error {
	return r.db.WithContext(ctx).Save(s).Error
}

// --- Invoices ---

func (r *repository) CreateInvoice(ctx context.Context, inv *model.Invoice) error {
	return r.db.WithContext(ctx).Create(inv).Error
}

func (r *repository) FindInvoiceByID(ctx context.Context, id uuid.UUID) (*model.Invoice, error) {
	var inv model.Invoice
	err := r.db.WithContext(ctx).First(&inv, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &inv, err
}

func (r *repository) FindInvoiceByXenditID(ctx context.Context, xenditID string) (*model.Invoice, error) {
	var inv model.Invoice
	err := r.db.WithContext(ctx).
		First(&inv, "xendit_invoice_id = ?", xenditID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &inv, err
}

func (r *repository) UpdateInvoice(ctx context.Context, inv *model.Invoice) error {
	return r.db.WithContext(ctx).Save(inv).Error
}

func (r *repository) ListInvoicesByOrg(ctx context.Context, orgID uuid.UUID) ([]model.Invoice, error) {
	var invoices []model.Invoice
	err := r.db.WithContext(ctx).
		Where("org_id = ?", orgID).
		Order("created_at DESC").
		Find(&invoices).Error
	return invoices, err
}

// --- Usage counts ---

func (r *repository) CountProjectsByOrg(ctx context.Context, orgID uuid.UUID) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Project{}).
		Where("org_id = ?", orgID).
		Count(&count).Error
	return int(count), err
}

func (r *repository) CountDeploymentsByOrg(ctx context.Context, orgID uuid.UUID) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Deployment{}).
		Joins("JOIN projects ON projects.id = deployments.project_id").
		Where("projects.org_id = ? AND deployments.status = ?", orgID, "running").
		Count(&count).Error
	return int(count), err
}

func (r *repository) CountMembersByOrg(ctx context.Context, orgID uuid.UUID) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Membership{}).
		Where("org_id = ?", orgID).
		Count(&count).Error
	return int(count), err
}
