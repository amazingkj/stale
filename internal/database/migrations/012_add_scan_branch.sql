-- Add scan_branch column to sources table
ALTER TABLE sources ADD COLUMN scan_branch TEXT DEFAULT '';
