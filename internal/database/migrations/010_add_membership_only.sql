-- Add membership_only column for GitLab to filter projects by membership
ALTER TABLE sources ADD COLUMN membership_only BOOLEAN DEFAULT FALSE;
