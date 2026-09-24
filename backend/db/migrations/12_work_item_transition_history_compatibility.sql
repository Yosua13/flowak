-- Existing deployments may have created work_item_status_history before
-- transition auditing stored the source status.  Keep the migration additive so
-- an already-populated workspace can move cards without data loss.
ALTER TABLE work_item_status_history
    ADD COLUMN IF NOT EXISTS from_status VARCHAR(20) NULL;

