package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// --- Base ---

// BaseModel provides common fields for all domain models.
type BaseModel struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// --- Auth / RBAC ---

// User represents an authenticated user.
type User struct {
	BaseModel
	Name          string       `gorm:"size:255;not null" json:"name"`
	Email         string       `gorm:"size:255;uniqueIndex;not null" json:"email"`
	PasswordHash  string       `gorm:"size:255" json:"-"` // empty for OAuth-only users
	AvatarURL     string       `gorm:"size:512" json:"avatar_url,omitempty"`
	EmailVerified bool         `gorm:"default:false" json:"email_verified"`
	Roles         []Role       `gorm:"many2many:user_roles;" json:"roles,omitempty"`
	Memberships   []Membership `gorm:"foreignKey:UserID" json:"-"`
}

// Role represents an RBAC role (e.g. admin, user).
type Role struct {
	ID   uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name string `gorm:"size:50;uniqueIndex;not null" json:"name"`
}

// UserRole is the join table for Users ↔ Roles.
type UserRole struct {
	UserID uuid.UUID `gorm:"type:uuid;primaryKey"`
	RoleID uint      `gorm:"primaryKey"`
}

// RefreshToken stores refresh tokens with family-based rotation.
type RefreshToken struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID      uuid.UUID  `gorm:"type:uuid;not null;index"`
	TokenHash   string     `gorm:"size:255;uniqueIndex;not null"`
	Family      uuid.UUID  `gorm:"type:uuid;not null;index"`
	Revoked     bool       `gorm:"default:false"`
	ExpiresAt   time.Time  `gorm:"not null"`
	CreatedAt   time.Time  `gorm:"autoCreateTime"`
	RevokedAt   *time.Time `gorm:""`
}

// EmailVerificationToken stores tokens for email verification.
type EmailVerificationToken struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index"`
	TokenHash string    `gorm:"size:255;uniqueIndex;not null"`
	ExpiresAt time.Time `gorm:"not null"`
	UsedAt    *time.Time
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

// PasswordResetToken stores tokens for password reset via magic link.
type PasswordResetToken struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index"`
	TokenHash string    `gorm:"size:255;uniqueIndex;not null"`
	ExpiresAt time.Time `gorm:"not null"`
	UsedAt    *time.Time
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

// OAuthAccount links a user to an external OAuth provider.
type OAuthAccount struct {
	BaseModel
	UserID       uuid.UUID `gorm:"type:uuid;not null;index" json:"-"`
	Provider     string    `gorm:"size:50;not null;uniqueIndex:idx_oauth_provider_id" json:"provider"`
	ProviderID   string    `gorm:"size:255;not null;uniqueIndex:idx_oauth_provider_id" json:"provider_id"`
	Email        string    `gorm:"size:255" json:"email"`
	AvatarURL    string    `gorm:"size:512" json:"avatar_url,omitempty"`
	AccessToken  string    `gorm:"size:2048" json:"-"`
	RefreshToken string    `gorm:"size:2048" json:"-"`
	User         User      `gorm:"foreignKey:UserID" json:"-"`
}

// FileUpload tracks uploaded files (avatars, attachments, etc.).
type FileUpload struct {
	BaseModel
	OwnerID     uuid.UUID `gorm:"type:uuid;not null;index:idx_file_owner" json:"owner_id"`
	OwnerType   string    `gorm:"size:50;not null;index:idx_file_owner" json:"owner_type"` // "user", "org"
	Key         string    `gorm:"size:512;uniqueIndex;not null" json:"key"`                // storage key
	Filename    string    `gorm:"size:255;not null" json:"filename"`                       // original filename
	ContentType string    `gorm:"size:100;not null" json:"content_type"`
	Size        int64     `gorm:"not null" json:"size"`          // bytes
	Category    string    `gorm:"size:50;not null" json:"category"` // "avatar", "attachment"
}

// --- Multi-Tenant ---

// Org represents a tenant organization.
type Org struct {
	BaseModel
	Name        string       `gorm:"size:255;not null" json:"name"`
	Slug        string       `gorm:"size:100;uniqueIndex;not null" json:"slug"`
	LogoURL     string       `gorm:"size:512" json:"logo_url,omitempty"`
	Memberships []Membership `gorm:"foreignKey:OrgID" json:"memberships,omitempty"`
	Projects    []Project    `gorm:"foreignKey:OrgID" json:"projects,omitempty"`
}

// Membership connects a User to an Org with a specific role.
type Membership struct {
	ID     uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_user_org" json:"user_id"`
	OrgID  uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_user_org" json:"org_id"`
	Role   string    `gorm:"size:50;not null;default:'viewer'" json:"role"` // owner, admin, developer, viewer
	User   User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Org    Org       `gorm:"foreignKey:OrgID" json:"org,omitempty"`
	JoinedAt time.Time `gorm:"autoCreateTime" json:"joined_at"`
}

// OrgInvite represents a pending invitation to join an org.
type OrgInvite struct {
	BaseModel
	OrgID     uuid.UUID  `gorm:"type:uuid;not null;index" json:"org_id"`
	Email     string     `gorm:"size:255;not null" json:"email"`
	Role      string     `gorm:"size:50;not null;default:'viewer'" json:"role"`
	Token     string     `gorm:"size:255;uniqueIndex;not null" json:"-"`
	ExpiresAt time.Time  `gorm:"not null" json:"expires_at"`
	AcceptedAt *time.Time `gorm:"" json:"accepted_at,omitempty"`
	InvitedBy uuid.UUID  `gorm:"type:uuid;not null" json:"invited_by"`
}

// --- Projects & Deployments ---

// Project represents a deployable application within an org.
type Project struct {
	BaseModel
	OrgID       uuid.UUID    `gorm:"type:uuid;not null;index" json:"org_id"`
	Name        string       `gorm:"size:255;not null" json:"name"`
	Description string       `gorm:"type:text" json:"description,omitempty"`
	RepoURL     string       `gorm:"size:512" json:"repo_url,omitempty"`
	Deployments []Deployment `gorm:"foreignKey:ProjectID" json:"deployments,omitempty"`
	EnvVars     []EnvVar     `gorm:"foreignKey:ProjectID" json:"-"`
}

// Deployment represents a single deployment event.
type Deployment struct {
	BaseModel
	ProjectID uuid.UUID `gorm:"type:uuid;not null;index" json:"project_id"`
	Version   string    `gorm:"size:100" json:"version"`
	Status    string    `gorm:"size:50;not null;default:'pending'" json:"status"` // pending, building, running, failed, stopped
	CommitSHA string    `gorm:"size:64" json:"commit_sha,omitempty"`
	Logs      string    `gorm:"type:text" json:"-"`
	StartedAt *time.Time `gorm:"" json:"started_at,omitempty"`
	FinishedAt *time.Time `gorm:"" json:"finished_at,omitempty"`
}

// EnvVar stores environment variables for a project.
type EnvVar struct {
	BaseModel
	ProjectID uuid.UUID `gorm:"type:uuid;not null;index" json:"project_id"`
	Key       string    `gorm:"size:255;not null" json:"key"`
	Value     string    `gorm:"type:text;not null" json:"value"` // [TODO] encrypt at rest
	IsSecret  bool      `gorm:"default:false" json:"is_secret"`
}

// AuditLog records important actions within an org.
type AuditLog struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrgID     uuid.UUID `gorm:"type:uuid;not null;index" json:"org_id"`
	ActorID   uuid.UUID `gorm:"type:uuid;not null" json:"actor_id"`
	Action    string    `gorm:"size:100;not null" json:"action"`
	Resource  string    `gorm:"size:100" json:"resource"`
	Details   string    `gorm:"type:jsonb" json:"details,omitempty"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// --- Billing / Xendit ---

// BillingPlan defines a subscription tier.
type BillingPlan struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name          string    `gorm:"size:100;not null" json:"name"`
	Slug          string    `gorm:"size:50;uniqueIndex;not null" json:"slug"`
	PriceMonthly  int64     `gorm:"not null;default:0" json:"price_monthly"`  // in smallest currency unit (e.g. IDR)
	PriceYearly   int64     `gorm:"not null;default:0" json:"price_yearly"`
	Currency      string    `gorm:"size:3;not null;default:'IDR'" json:"currency"`
	MaxProjects   int       `gorm:"not null;default:1" json:"max_projects"`
	MaxDeployments int      `gorm:"not null;default:10" json:"max_deployments"`
	MaxMembers    int       `gorm:"not null;default:1" json:"max_members"`
	Features      string    `gorm:"type:jsonb" json:"features,omitempty"`
	IsActive      bool      `gorm:"default:true" json:"is_active"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// Subscription links an org to a billing plan.
type Subscription struct {
	BaseModel
	OrgID              uuid.UUID  `gorm:"type:uuid;not null;index" json:"org_id"`
	PlanID             uuid.UUID  `gorm:"type:uuid;not null" json:"plan_id"`
	Plan               BillingPlan `gorm:"foreignKey:PlanID" json:"plan,omitempty"`
	Status             string     `gorm:"size:50;not null;default:'active'" json:"status"` // active, cancelled, past_due, trialing
	BillingCycle       string     `gorm:"size:20;not null;default:'monthly'" json:"billing_cycle"`
	CurrentPeriodStart time.Time  `gorm:"not null" json:"current_period_start"`
	CurrentPeriodEnd   time.Time  `gorm:"not null" json:"current_period_end"`
	CancelledAt        *time.Time `gorm:"" json:"cancelled_at,omitempty"`
	// Xendit fields
	XenditSubscriptionID string `gorm:"size:255" json:"xendit_subscription_id,omitempty"`
	XenditPlanID         string `gorm:"size:255" json:"xendit_plan_id,omitempty"`
}

// Invoice represents a billing invoice.
type Invoice struct {
	BaseModel
	OrgID          uuid.UUID  `gorm:"type:uuid;not null;index" json:"org_id"`
	SubscriptionID uuid.UUID  `gorm:"type:uuid;not null" json:"subscription_id"`
	Amount         int64      `gorm:"not null" json:"amount"`
	Currency       string     `gorm:"size:3;not null;default:'IDR'" json:"currency"`
	Status         string     `gorm:"size:50;not null;default:'pending'" json:"status"` // pending, paid, failed, refunded
	DueDate        time.Time  `gorm:"not null" json:"due_date"`
	PaidAt         *time.Time `gorm:"" json:"paid_at,omitempty"`
	// Xendit fields
	XenditInvoiceID  string `gorm:"size:255" json:"xendit_invoice_id,omitempty"`
	XenditPaymentURL string `gorm:"size:512" json:"xendit_payment_url,omitempty"`
	XenditExternalID string `gorm:"size:255" json:"xendit_external_id,omitempty"`
}

// --- RBAC Constants ---

const (
	RoleOwner     = "owner"
	RoleAdmin     = "admin"
	RoleDeveloper = "developer"
	RoleViewer    = "viewer"
	RoleUser      = "user" // system-level role
	RoleSuperAdmin = "super_admin"
)

// RoleHierarchy defines the power level of each role (higher = more permissions).
var RoleHierarchy = map[string]int{
	RoleViewer:    1,
	RoleDeveloper: 2,
	RoleAdmin:     3,
	RoleOwner:     4,
}

// HasPermission checks if roleA has at least the power of requiredRole.
func HasPermission(userRole, requiredRole string) bool {
	return RoleHierarchy[userRole] >= RoleHierarchy[requiredRole]
}
