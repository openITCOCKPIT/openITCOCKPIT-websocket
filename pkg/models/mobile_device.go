package models

import "github.com/uptrace/bun"

type MobileDevice struct {
	bun.BaseModel `bun:"table:mobile_devices"`

	ID       int    `bun:"id,pk,autoincrement"`
	UserID   int    `bun:"user_id,notnull"`
	DeviceID string `bun:"device_id,notnull"`
}
