BEGIN;

ALTER TABLE analysis_runs
    ALTER COLUMN policy_snapshot DROP NOT NULL;

COMMIT;