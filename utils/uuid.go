package utils

import "github.com/gofrs/uuid/v5"

// NewUUIDv7 returns a time-sortable UUIDv7 string. The library is already a direct dependency,
// so no new module is introduced for ids.
func NewUUIDv7() string {
	if u, err := uuid.NewV7(); err == nil {
		return u.String()
	}
	return uuid.Must(uuid.NewV4()).String()
}
