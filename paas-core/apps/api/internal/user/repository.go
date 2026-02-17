package user

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"paas-core/apps/api/internal/model"
)

type txKey struct{}

// Repository defines the user repository interface.
type Repository interface {
	Create(ctx context.Context, user *model.User) error
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListAllUsers(ctx context.Context, filters FilterParams, page, perPage int) ([]model.User, int64, error)
	AssignRole(ctx context.Context, userID uuid.UUID, roleName string) error
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]string, error)
	Transaction(ctx context.Context, fn func(context.Context) error) error
}

type repository struct {
	db *gorm.DB
}

// NewRepository creates a new user repository.
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx
	}
	return r.db
}

func (r *repository) Create(ctx context.Context, user *model.User) error {
	return r.getDB(ctx).WithContext(ctx).Create(user).Error
}

func (r *repository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := r.getDB(ctx).WithContext(ctx).Preload("Roles").Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *repository) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var user model.User
	err := r.getDB(ctx).WithContext(ctx).Preload("Roles").First(&user, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *repository) Update(ctx context.Context, user *model.User) error {
	return r.getDB(ctx).WithContext(ctx).
		Select("name", "email", "avatar_url", "updated_at").
		Save(user).Error
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.getDB(ctx).WithContext(ctx).Delete(&model.User{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *repository) ListAllUsers(ctx context.Context, filters FilterParams, page, perPage int) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	query := r.getDB(ctx).WithContext(ctx).Model(&model.User{}).Preload("Roles")

	if filters.Search != "" {
		searchPattern := "%" + filters.Search + "%"
		query = query.Where("name ILIKE ? OR email ILIKE ?", searchPattern, searchPattern)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage

	validSorts := map[string]bool{"name": true, "email": true, "created_at": true}
	sortField := "created_at"
	if validSorts[filters.Sort] {
		sortField = filters.Sort
	}
	sortOrder := "desc"
	if filters.Order == "asc" {
		sortOrder = "asc"
	}

	orderColumn := clause.OrderByColumn{
		Column: clause.Column{Name: sortField},
		Desc:   sortOrder == "desc",
	}
	query = query.Order(orderColumn).Offset(offset).Limit(perPage)

	if err := query.Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *repository) AssignRole(ctx context.Context, userID uuid.UUID, roleName string) error {
	var role model.Role
	if err := r.getDB(ctx).WithContext(ctx).Where("name = ?", roleName).First(&role).Error; err != nil {
		return err
	}
	userRole := model.UserRole{UserID: userID, RoleID: role.ID}
	return r.getDB(ctx).WithContext(ctx).Create(&userRole).Error
}

func (r *repository) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]string, error) {
	var roleNames []string
	err := r.getDB(ctx).WithContext(ctx).
		Table("roles").
		Select("roles.name").
		Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ?", userID).
		Find(&roleNames).Error
	return roleNames, err
}

func (r *repository) Transaction(ctx context.Context, fn func(context.Context) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, txKey{}, tx)
		return fn(txCtx)
	})
}
