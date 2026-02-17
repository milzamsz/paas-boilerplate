-- MyPaaS Database Schema
-- Multi-tenant PaaS platform with billing (Xendit)

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================
-- Users
-- ============================================
CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email           VARCHAR(255) UNIQUE NOT NULL,
    name            VARCHAR(255) NOT NULL,
    password_hash   TEXT NOT NULL,
    avatar_url      TEXT,
    active_org_id   UUID,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);

-- ============================================
-- Organizations (Tenants)
-- ============================================
CREATE TABLE orgs (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name            VARCHAR(255) NOT NULL,
    slug            VARCHAR(255) UNIQUE NOT NULL,
    logo_url        TEXT,
    plan_id         UUID,
    trial_ends_at   TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_orgs_slug ON orgs(slug);

-- ============================================
-- Organization Memberships
-- ============================================
CREATE TABLE org_memberships (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id          UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role            VARCHAR(20) NOT NULL DEFAULT 'viewer'
                    CHECK (role IN ('owner', 'admin', 'developer', 'viewer')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(org_id, user_id)
);

CREATE INDEX idx_org_memberships_org ON org_memberships(org_id);
CREATE INDEX idx_org_memberships_user ON org_memberships(user_id);

-- ============================================
-- Organization Invites
-- ============================================
CREATE TABLE org_invites (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id          UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    email           VARCHAR(255) NOT NULL,
    role            VARCHAR(20) NOT NULL DEFAULT 'viewer',
    token           VARCHAR(255) UNIQUE NOT NULL,
    expires_at      TIMESTAMPTZ NOT NULL,
    accepted_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_org_invites_token ON org_invites(token);

-- ============================================
-- Projects
-- ============================================
CREATE TABLE projects (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id          UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    name            VARCHAR(255) NOT NULL,
    description     TEXT,
    status          VARCHAR(20) NOT NULL DEFAULT 'active'
                    CHECK (status IN ('active', 'archived')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_projects_org ON projects(org_id);

-- ============================================
-- Deployments
-- ============================================
CREATE TABLE deployments (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id          UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    version         VARCHAR(100) NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending', 'building', 'running', 'failed', 'stopped')),
    commit_sha      VARCHAR(40),
    logs            TEXT,
    deployed_by     UUID NOT NULL REFERENCES users(id),
    started_at      TIMESTAMPTZ,
    finished_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_deployments_project ON deployments(project_id);
CREATE INDEX idx_deployments_org ON deployments(org_id);

-- ============================================
-- Environment Variables
-- ============================================
CREATE TABLE env_vars (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id          UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    key             VARCHAR(255) NOT NULL,
    value           TEXT NOT NULL,
    is_secret       BOOLEAN NOT NULL DEFAULT false,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(project_id, key)
);

CREATE INDEX idx_env_vars_project ON env_vars(project_id);

-- ============================================
-- Billing Plans
-- ============================================
CREATE TABLE billing_plans (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name            VARCHAR(100) NOT NULL,
    slug            VARCHAR(100) UNIQUE NOT NULL,
    description     TEXT,
    price_monthly   BIGINT NOT NULL DEFAULT 0,
    price_yearly    BIGINT NOT NULL DEFAULT 0,
    currency        VARCHAR(3) NOT NULL DEFAULT 'IDR',
    max_projects    INT NOT NULL DEFAULT 3,
    max_deployments INT NOT NULL DEFAULT 100,
    max_members     INT NOT NULL DEFAULT 5,
    storage_limit_mb INT NOT NULL DEFAULT 1024,
    features        JSONB DEFAULT '[]',
    is_active       BOOLEAN NOT NULL DEFAULT true,
    sort_order      INT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed default plans
INSERT INTO billing_plans (id, name, slug, description, price_monthly, price_yearly, currency, max_projects, max_deployments, max_members, storage_limit_mb, features, sort_order) VALUES
    (uuid_generate_v4(), 'Free', 'free', 'Get started for free', 0, 0, 'IDR', 2, 50, 3, 512, '["Community support", "Shared resources"]', 1),
    (uuid_generate_v4(), 'Starter', 'starter', 'For small teams', 199000, 1990000, 'IDR', 5, 200, 5, 2048, '["Email support", "Custom domains", "SSL certificates"]', 2),
    (uuid_generate_v4(), 'Pro', 'pro', 'For growing businesses', 599000, 5990000, 'IDR', 20, 1000, 20, 10240, '["Priority support", "Auto-scaling", "Audit logs", "SSO"]', 3),
    (uuid_generate_v4(), 'Enterprise', 'enterprise', 'For large organizations', 1999000, 19990000, 'IDR', -1, -1, -1, -1, '["24/7 support", "SLA", "Dedicated resources", "Custom integrations"]', 4);

-- ============================================
-- Subscriptions
-- ============================================
CREATE TABLE subscriptions (
    id                   UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id               UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    plan_id              UUID NOT NULL REFERENCES billing_plans(id),
    status               VARCHAR(20) NOT NULL DEFAULT 'active'
                         CHECK (status IN ('active', 'trialing', 'past_due', 'canceled', 'expired')),
    billing_cycle        VARCHAR(10) NOT NULL DEFAULT 'monthly'
                         CHECK (billing_cycle IN ('monthly', 'yearly')),
    current_period_start TIMESTAMPTZ NOT NULL,
    current_period_end   TIMESTAMPTZ NOT NULL,
    canceled_at          TIMESTAMPTZ,
    xendit_plan_id       TEXT,
    xendit_sub_id        TEXT,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_subscriptions_org ON subscriptions(org_id);

-- ============================================
-- Invoices
-- ============================================
CREATE TABLE invoices (
    id               UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id           UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    subscription_id  UUID REFERENCES subscriptions(id),
    xendit_invoice_id TEXT,
    invoice_number   VARCHAR(100) UNIQUE NOT NULL,
    amount           BIGINT NOT NULL,
    currency         VARCHAR(3) NOT NULL DEFAULT 'IDR',
    status           VARCHAR(20) NOT NULL DEFAULT 'pending'
                     CHECK (status IN ('pending', 'paid', 'expired', 'failed', 'refunded')),
    description      TEXT,
    invoice_url      TEXT,
    paid_at          TIMESTAMPTZ,
    due_date         TIMESTAMPTZ NOT NULL,
    period_start     TIMESTAMPTZ NOT NULL,
    period_end       TIMESTAMPTZ NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_invoices_org ON invoices(org_id);
CREATE INDEX idx_invoices_number ON invoices(invoice_number);

-- ============================================
-- Audit Logs
-- ============================================
CREATE TABLE audit_logs (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id          UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    user_id         UUID REFERENCES users(id) ON DELETE SET NULL,
    action          VARCHAR(100) NOT NULL,
    resource        VARCHAR(100) NOT NULL,
    resource_id     UUID,
    metadata        JSONB,
    ip_address      INET,
    user_agent      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_org ON audit_logs(org_id);
CREATE INDEX idx_audit_logs_created ON audit_logs(created_at);

-- Add FK from orgs.plan_id to billing_plans
ALTER TABLE orgs ADD CONSTRAINT fk_orgs_plan FOREIGN KEY (plan_id) REFERENCES billing_plans(id);
-- Add FK from users.active_org_id to orgs
ALTER TABLE users ADD CONSTRAINT fk_users_active_org FOREIGN KEY (active_org_id) REFERENCES orgs(id) ON DELETE SET NULL;
