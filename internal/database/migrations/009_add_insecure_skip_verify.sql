-- Add insecure_skip_verify column for self-hosted GitLab instances with invalid certs
ALTER TABLE sources ADD COLUMN insecure_skip_verify BOOLEAN DEFAULT FALSE;
