-- 000001_init_schema.down.sql
-- Rollback initial schema

DROP TABLE IF EXISTS invoices;
DROP TABLE IF EXISTS subscriptions;
DROP TABLE IF EXISTS billing_plans;
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS env_vars;
DROP TABLE IF EXISTS deployments;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS org_invites;
DROP TABLE IF EXISTS memberships;
DROP TABLE IF EXISTS orgs;
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS roles;
