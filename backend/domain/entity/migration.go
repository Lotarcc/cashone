package entity

import "time"

// Migration represents a database migration record
type Migration struct {
	Version   string    `gorm:"primaryKey"`
	AppliedAt time.Time `gorm:"autoCreateTime"`
}
