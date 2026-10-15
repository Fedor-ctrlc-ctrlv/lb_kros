
BEGIN;

CREATE TABLE app_users (
    id BIGSERIAL PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'USER'
        CHECK (role IN ('USER', 'ADMIN')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE security_policies (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE policy_rules (
    id BIGSERIAL PRIMARY KEY,
    policy_id BIGINT NOT NULL
        REFERENCES security_policies(id) ON DELETE CASCADE,
    rule_type TEXT NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    severity TEXT NOT NULL DEFAULT 'MEDIUM'
        CHECK (severity IN ('LOW', 'MEDIUM', 'HIGH', 'CRITICAL')),
    parameters JSONB NOT NULL DEFAULT '{}'::jsonb,
    UNIQUE (policy_id, rule_type),
    CHECK (jsonb_typeof(parameters) = 'object')
);

CREATE TABLE hosts (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    hostname TEXT,
    description TEXT NOT NULL DEFAULT '',
    os TEXT,
    architecture TEXT,
    kernel TEXT,
    policy_id BIGINT
        REFERENCES security_policies(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE audits (
    id BIGSERIAL PRIMARY KEY,
    host_id BIGINT NOT NULL
        REFERENCES hosts(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'PENDING'
        CHECK (status IN ('PENDING', 'COMPLETED', 'FAILED')),
    snapshot_version TEXT,
    snapshot_created_at TIMESTAMPTZ,
    raw_snapshot JSONB,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finished_at TIMESTAMPTZ,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (id, host_id)
);

ALTER TABLE hosts
ADD COLUMN baseline_audit_id BIGINT;

ALTER TABLE hosts
ADD CONSTRAINT hosts_baseline_audit_fk
FOREIGN KEY (baseline_audit_id, id)
REFERENCES audits(id, host_id)
DEFERRABLE INITIALLY IMMEDIATE;

CREATE TABLE fs_nodes (
    id BIGSERIAL PRIMARY KEY,
    audit_id BIGINT NOT NULL
        REFERENCES audits(id) ON DELETE CASCADE,
    path TEXT NOT NULL,
    node_type TEXT NOT NULL
        CHECK (node_type IN ('file', 'directory', 'symlink')),
    link_target TEXT,
    sha256 TEXT,
    hash_status TEXT,
    permissions TEXT,
    mode BIGINT,
    uid BIGINT,
    gid BIGINT,
    size BIGINT,
    modified_at TIMESTAMPTZ,
    suid BOOLEAN NOT NULL DEFAULT FALSE,
    sgid BOOLEAN NOT NULL DEFAULT FALSE,
    sticky BOOLEAN NOT NULL DEFAULT FALSE,
    UNIQUE (audit_id, path)
);

CREATE TABLE linux_users (
    id BIGSERIAL PRIMARY KEY,
    audit_id BIGINT NOT NULL
        REFERENCES audits(id) ON DELETE CASCADE,
    username TEXT NOT NULL,
    uid BIGINT NOT NULL,
    gid BIGINT NOT NULL,
    home TEXT,
    shell TEXT,
    raw_data JSONB NOT NULL DEFAULT '{}'::jsonb,
    UNIQUE (audit_id, username)
);

CREATE TABLE linux_groups (
    id BIGSERIAL PRIMARY KEY,
    audit_id BIGINT NOT NULL
        REFERENCES audits(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    gid BIGINT NOT NULL,
    members TEXT[],
    UNIQUE (audit_id, name)
);

CREATE TABLE services (
    id BIGSERIAL PRIMARY KEY,
    audit_id BIGINT NOT NULL
        REFERENCES audits(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    load_state TEXT,
    active_state TEXT,
    sub_state TEXT,
    UNIQUE (audit_id, name)
);

CREATE TABLE cron_jobs (
    id BIGSERIAL PRIMARY KEY,
    audit_id BIGINT NOT NULL
        REFERENCES audits(id) ON DELETE CASCADE,
    source TEXT NOT NULL,
    content TEXT NOT NULL DEFAULT '',
    UNIQUE (audit_id, source)
);

CREATE TABLE violations (
    id BIGSERIAL PRIMARY KEY,
    audit_id BIGINT NOT NULL
        REFERENCES audits(id) ON DELETE CASCADE,
    rule_id BIGINT
        REFERENCES policy_rules(id) ON DELETE SET NULL,
    severity TEXT NOT NULL
        CHECK (severity IN ('LOW', 'MEDIUM', 'HIGH', 'CRITICAL')),
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    entity_type TEXT,
    entity_key TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audits_host_created
    ON audits(host_id, created_at DESC);

CREATE INDEX idx_fs_nodes_audit
    ON fs_nodes(audit_id);

CREATE INDEX idx_linux_users_audit
    ON linux_users(audit_id);

CREATE INDEX idx_linux_groups_audit
    ON linux_groups(audit_id);

CREATE INDEX idx_services_audit
    ON services(audit_id);

CREATE INDEX idx_cron_jobs_audit
    ON cron_jobs(audit_id);

CREATE INDEX idx_violations_audit
    ON violations(audit_id);

CREATE INDEX idx_violations_severity
    ON violations(severity);

CREATE INDEX idx_policy_rules_policy
    ON policy_rules(policy_id);

COMMIT;
