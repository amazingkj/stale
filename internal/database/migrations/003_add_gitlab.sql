-- Add URL column for self-hosted GitLab instances
ALTER TABLE sources ADD COLUMN url TEXT DEFAULT '';
