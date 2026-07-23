-- Migration v4: Add metadata_failed column to pipeline_states
--
-- Without this column, when stageSaveMetadata fails (e.g. network partition
-- during Supabase write), the pipeline retries up to maxPipelineRetries times
-- and then abandons the recording — deleting both the pipeline state and the
-- local file — even though the uploads to all video hosts succeeded.
--
-- The MetadataFailed flag lets the pipeline recognise "uploads OK, metadata
-- write failed" as a distinct state and keep the file/state for later retry.

ALTER TABLE pipeline_states ADD COLUMN IF NOT EXISTS metadata_failed BOOLEAN NOT NULL DEFAULT FALSE;
