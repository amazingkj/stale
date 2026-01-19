package domain

import "time"

type IgnoredDependency struct {
	ID        int64     `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	Ecosystem string    `db:"ecosystem" json:"ecosystem,omitempty"`
	Reason    string    `db:"reason" json:"reason,omitempty"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type IgnoredDependencyInput struct {
	Name      string `json:"name"`
	Ecosystem string `json:"ecosystem,omitempty"`
	Reason    string `json:"reason,omitempty"`
}
