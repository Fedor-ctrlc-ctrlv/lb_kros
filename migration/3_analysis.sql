BEGIN;

ALTER TABLE security_policies
    ADD COLUMN engine_policy JSONB
    CHECK (
        engine_policy IS NULL
        OR jsonb_typeof(engine_policy) = 'object'
    );

CREATE TABLE analysis_runs (
    id BIGSERIAL PRIMARY KEY,

    current_audit_id BIGINT NOT NULL
        REFERENCES audits(id) ON DELETE RESTRICT,

    baseline_audit_id BIGINT NOT NULL
        REFERENCES audits(id) ON DELETE RESTRICT,

    policy_id BIGINT
        REFERENCES security_policies(id) ON DELETE SET NULL,

    policy_snapshot JSONB NOT NULL,
    engine_version TEXT NOT NULL,
    summary JSONB NOT NULL,
    result JSONB NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX analysis_runs_current_idx
    ON analysis_runs(current_audit_id, id DESC);

COMMIT;