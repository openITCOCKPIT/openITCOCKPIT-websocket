package models

import "github.com/uptrace/bun"

type PushNotificationsRelay struct {
	bun.BaseModel `bun:"table:push_notifications_relay"`

	ID      int    `bun:"id,pk,autoincrement"`
	Address string `bun:"address,notnull"`
	Port    int    `bun:"port,notnull"`
	AuthKey string `bun:"auth_key,notnull"`
	Enabled bool   `bun:"enabled,notnull,default:1"`
}
