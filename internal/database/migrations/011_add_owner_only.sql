-- Add owner_only column for GitHub to filter repos by ownership
ALTER TABLE sources ADD COLUMN owner_only BOOLEAN DEFAULT FALSE;
