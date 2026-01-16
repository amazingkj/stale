package domain

type Settings struct {
	// Schedule settings
	ScheduleEnabled bool   `json:"schedule_enabled"`
	ScheduleCron    string `json:"schedule_cron"`

	// Email settings
	EmailEnabled          bool   `json:"email_enabled"`
	EmailSMTPHost         string `json:"email_smtp_host"`
	EmailSMTPPort         int    `json:"email_smtp_port"`
	EmailSMTPUser         string `json:"email_smtp_user"`
	EmailSMTPPass         string `json:"email_smtp_pass,omitempty"`
	EmailFrom             string `json:"email_from"`
	EmailTo               string `json:"email_to"`
	EmailNotifyNewOutdated bool  `json:"email_notify_new_outdated"`
}

type SettingsInput struct {
	// Schedule settings
	ScheduleEnabled *bool   `json:"schedule_enabled,omitempty"`
	ScheduleCron    *string `json:"schedule_cron,omitempty"`

	// Email settings
	EmailEnabled          *bool   `json:"email_enabled,omitempty"`
	EmailSMTPHost         *string `json:"email_smtp_host,omitempty"`
	EmailSMTPPort         *int    `json:"email_smtp_port,omitempty"`
	EmailSMTPUser         *string `json:"email_smtp_user,omitempty"`
	EmailSMTPPass         *string `json:"email_smtp_pass,omitempty"`
	EmailFrom             *string `json:"email_from,omitempty"`
	EmailTo               *string `json:"email_to,omitempty"`
	EmailNotifyNewOutdated *bool  `json:"email_notify_new_outdated,omitempty"`
}

type NewOutdatedReport struct {
	ScanID       int64                `json:"scan_id"`
	NewOutdated  []DependencyWithRepo `json:"new_outdated"`
	TotalScanned int                  `json:"total_scanned"`
}
