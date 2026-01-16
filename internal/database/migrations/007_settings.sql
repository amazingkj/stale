-- Settings table for schedule and notification configuration
CREATE TABLE IF NOT EXISTS settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Default settings
INSERT OR IGNORE INTO settings (key, value) VALUES ('schedule_enabled', 'false');
INSERT OR IGNORE INTO settings (key, value) VALUES ('schedule_cron', '0 9 * * *');
INSERT OR IGNORE INTO settings (key, value) VALUES ('email_enabled', 'false');
INSERT OR IGNORE INTO settings (key, value) VALUES ('email_smtp_host', '');
INSERT OR IGNORE INTO settings (key, value) VALUES ('email_smtp_port', '587');
INSERT OR IGNORE INTO settings (key, value) VALUES ('email_smtp_user', '');
INSERT OR IGNORE INTO settings (key, value) VALUES ('email_smtp_pass', '');
INSERT OR IGNORE INTO settings (key, value) VALUES ('email_from', '');
INSERT OR IGNORE INTO settings (key, value) VALUES ('email_to', '');
INSERT OR IGNORE INTO settings (key, value) VALUES ('email_notify_new_outdated', 'true');

-- Track outdated status changes
ALTER TABLE dependencies ADD COLUMN previously_outdated BOOLEAN DEFAULT 0;
