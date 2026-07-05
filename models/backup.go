package models

import "time"

// Stucts for backup data representation.
// These structs will be used to serialize and deserialize the backup data in JSON.

type BackupData struct {
	ExportedAt time.Time       `json:"exported_at"`
	Projects   []BackupProject `json:"projects"`
}

type BackupProject struct {
	Name        string             `json:"name"`
	ErrorGroups []BackupErrorGroup `json:"error_groups"`
}

type BackupErrorGroup struct {
	Title       string        `json:"title"`
	Status      string        `json:"status"`
	Fingerprint string        `json:"fingerprint"`
	Count       int           `json:"count"`
	LastSeenAt  time.Time     `json:"last_seen_at"`
	Errors      []BackupError `json:"errors"`
}

type BackupError struct {
	Log        string    `json:"log"`
	StackTrace string    `json:"stack_trace"`
	Severity   string    `json:"severity"`
	OccurredAt time.Time `json:"occurred_at"`
}
