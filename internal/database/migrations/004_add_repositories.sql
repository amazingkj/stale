-- Add repositories column to sources table for filtering specific repos
ALTER TABLE sources ADD COLUMN repositories TEXT DEFAULT '';
