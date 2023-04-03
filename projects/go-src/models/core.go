package models

import "time"

type CoreModel struct {
	BucketKey    string
	TemplateName string
	IsWrapper    bool
	Parent       *string
	Child        *string
	Subject      *string
	Variables    []string
	Plain        string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
