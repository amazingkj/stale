-- Add has_go_mod column to repositories table for Go module support
ALTER TABLE repositories ADD COLUMN has_go_mod BOOLEAN DEFAULT FALSE;
