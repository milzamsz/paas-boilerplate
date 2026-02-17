package org

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"paas-core/apps/api/internal/model"
)

// Repository defines the org data access interface.
type Repository interface {
	Create(ctx context.Context, org *model.Org) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.Org, error)
	FindBySlug(ctx context.Context, slug string) (*model.Org, error)
	Update(ctx context.Context, org *model.Org) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByUser(ctx context.Context, userID uuid.UUID) ([]model.Org, error)

	// Memberships
	CreateMembership(ctx context.Context, m *model.Membership) error
	FindMembership(ctx context.Context, orgID, userID uuid.UUID) (*model.Membership, error)
	UpdateMembership(ctx context.Context, m *model.Membership) error
	DeleteMembership(ctx context.Context, id uuid.UUID) error
	ListMembers(ctx context.Context, orgID uuid.UUID) ([]model.Membership, error)
	CountMembers(ctx context.Context, orgID uuid.UUID) (int64, error)

	// Invites
	CreateInvite(ctx context.Context, inv *model.OrgInvite) error
	FindInviteByToken(ctx context.Context, token string) (*model.OrgInvite, error)
	FindInviteByEmail(ctx context.Context, orgID uuid.UUID, email string) (*model.OrgInvite, error)
	ListInvites(ctx context.Context, orgID uuid.UUID) ([]model.OrgInvite, error)
	DeleteInvite(ctx context.Context, id uuid.UUID) error
	UpdateInvite(ctx context.Context, inv *model.OrgInvite) error

	Transaction(ctx context.Context, fn func(context.Context) error) error
}

type txKey struct{}

type repository struct {
	db *gorm.DB
}

// NewRepository creates a new org repository.
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx
	}
	return r.db
}

// --- Org CRUD ---

func (r *repository) Create(ctx context.Context, org *model.Org) error {
	return r.getDB(ctx).WithContext(ctx).Create(org).Error
}

func (r *repository) FindByID(ctx context.Context, id uuid.UUID) (*model.Org, error) {
	var org model.Org
	err := r.getDB(ctx).WithContext(ctx).First(&org, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &org, err
}

func (r *repository) FindBySlug(ctx context.Context, slug string) (*model.Org, error) {
	var org model.Org
	err := r.getDB(ctx).WithContext(ctx).Where("slug = ?", slug).First(&org).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &org, err
}

func (r *repository) Update(ctx context.Context, org *model.Org) error {
	return r.getDB(ctx).WithContext(ctx).Save(org).Error
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.getDB(ctx).WithContext(ctx).Delete(&model.Org{}, "id = ?", id).Error
}

func (r *repository) ListByUser(ctx context.Context, userID uuid.UUID) ([]model.Org, error) {
	var orgs []model.Org
	err := r.getDB(ctx).WithContext(ctx).
		Joins("JOIN memberships ON memberships.org_id = orgs.id").
		Where("memberships.user_id = ?", userID).
		Find(&orgs).Error
	return orgs, err
}

// --- Memberships ---

func (r *repository) CreateMembership(ctx context.Context, m *model.Membership) error {
	return r.getDB(ctx).WithContext(ctx).Create(m).Error
}

func (r *repository) FindMembership(ctx context.Context, orgID, userID uuid.UUID) (*model.Membership, error) {
	var m model.Membership
	err := r.getDB(ctx).WithContext(ctx).
		Preload("User").
		Where("org_id = ? AND user_id = ?", orgID, userID).
		First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &m, err
}

func (r *repository) UpdateMembership(ctx context.Context, m *model.Membership) error {
	return r.getDB(ctx).WithContext(ctx).Save(m).Error
}

func (r *repository) DeleteMembership(ctx context.Context, id uuid.UUID) error {
	return r.getDB(ctx).WithContext(ctx).Delete(&model.Membership{}, "id = ?", id).Error
}

func (r *repository) ListMembers(ctx context.Context, orgID uuid.UUID) ([]model.Membership, error) {
	var members []model.Membership
	err := r.getDB(ctx).WithContext(ctx).
		Preload("User").
		Where("org_id = ?", orgID).
		Order("joined_at ASC").
		Find(&members).Error
	return members, err
}

func (r *repository) CountMembers(ctx context.Context, orgID uuid.UUID) (int64, error) {
	var count int64
	err := r.getDB(ctx).WithContext(ctx).
		Model(&model.Membership{}).
		Where("org_id = ?", orgID).
		Count(&count).Error
	return count, err
}

// --- Invites ---

func (r *repository) CreateInvite(ctx context.Context, inv *model.OrgInvite) error {
	return r.getDB(ctx).WithContext(ctx).Create(inv).Error
}

func (r *repository) FindInviteByToken(ctx context.Context, token string) (*model.OrgInvite, error) {
	var inv model.OrgInvite
	err := r.getDB(ctx).WithContext(ctx).Where("token = ?", token).First(&inv).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &inv, err
}

func (r *repository) FindInviteByEmail(ctx context.Context, orgID uuid.UUID, email string) (*model.OrgInvite, error) {
	var inv model.OrgInvite
	err := r.getDB(ctx).WithContext(ctx).
		Where("org_id = ? AND email = ? AND accepted_at IS NULL", orgID, email).
		First(&inv).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &inv, err
}

func (r *repository) ListInvites(ctx context.Context, orgID uuid.UUID) ([]model.OrgInvite, error) {
	var invites []model.OrgInvite
	err := r.getDB(ctx).WithContext(ctx).
		Where("org_id = ? AND accepted_at IS NULL", orgID).
		Order("created_at DESC").
		Find(&invites).Error
	return invites, err
}

func (r *repository) DeleteInvite(ctx context.Context, id uuid.UUID) error {
	return r.getDB(ctx).WithContext(ctx).Delete(&model.OrgInvite{}, "id = ?", id).Error
}

func (r *repository) UpdateInvite(ctx context.Context, inv *model.OrgInvite) error {
	return r.getDB(ctx).WithContext(ctx).Save(inv).Error
}

// --- TX ---

func (r *repository) Transaction(ctx context.Context, fn func(context.Context) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, txKey{}, tx)
		return fn(txCtx)
	})
}
