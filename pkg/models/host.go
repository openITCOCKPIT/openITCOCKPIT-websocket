package models

import (
	"time"

	"github.com/uptrace/bun"
)

type Host struct {
	bun.BaseModel `bun:"table:hosts"`

	ID                         int       `bun:"id,pk,autoincrement"`
	UUID                       string    `bun:"uuid,unique,notnull"`
	ContainerID                int       `bun:"container_id,notnull"`
	Name                       string    `bun:"name,notnull"`
	Description                *string   `bun:"description,nullzero"`
	HosttemplateID             int       `bun:"hosttemplate_id,notnull"`
	Address                    string    `bun:"address,notnull"`
	CommandID                  *int      `bun:"command_id,nullzero"`
	EventhandlerCommandID      *int      `bun:"eventhandler_command_id,nullzero"`
	TimeperiodID               *int      `bun:"timeperiod_id,nullzero"`
	CheckInterval              *int      `bun:"check_interval,nullzero"`
	RetryInterval              *int      `bun:"retry_interval,nullzero"`
	MaxCheckAttempts           *int      `bun:"max_check_attempts,nullzero"`
	FirstNotificationDelay     *float64  `bun:"first_notification_delay,nullzero"`
	NotificationInterval       *float64  `bun:"notification_interval,nullzero"`
	NotifyOnDown               *int      `bun:"notify_on_down,nullzero"`
	NotifyOnUnreachable        *int      `bun:"notify_on_unreachable,nullzero"`
	NotifyOnRecovery           *int      `bun:"notify_on_recovery,nullzero"`
	NotifyOnFlapping           *int      `bun:"notify_on_flapping,nullzero"`
	NotifyOnDowntime           *int      `bun:"notify_on_downtime,nullzero"`
	FlapDetectionEnabled       *int      `bun:"flap_detection_enabled,nullzero"`
	FlapDetectionOnUp          *int      `bun:"flap_detection_on_up,nullzero"`
	FlapDetectionOnDown        *int      `bun:"flap_detection_on_down,nullzero"`
	FlapDetectionOnUnreachable *int      `bun:"flap_detection_on_unreachable,nullzero"`
	LowFlapThreshold           *float64  `bun:"low_flap_threshold,nullzero"`
	HighFlapThreshold          *float64  `bun:"high_flap_threshold,nullzero"`
	ProcessPerformanceData     *int      `bun:"process_performance_data,nullzero"`
	FreshnessChecksEnabled     *int      `bun:"freshness_checks_enabled,nullzero"`
	FreshnessThreshold         *int      `bun:"freshness_threshold,nullzero"`
	PassiveChecksEnabled       *int      `bun:"passive_checks_enabled,nullzero"`
	EventHandlerEnabled        *int      `bun:"event_handler_enabled,nullzero"`
	ActiveChecksEnabled        *int      `bun:"active_checks_enabled,nullzero"`
	RetainStatusInformation    *int      `bun:"retain_status_information,nullzero"`
	RetainNonstatusInformation *int      `bun:"retain_nonstatus_information,nullzero"`
	NotificationsEnabled       *int      `bun:"notifications_enabled,nullzero"`
	Notes                      *string   `bun:"notes,nullzero"`
	Priority                   *int      `bun:"priority,nullzero"`
	CheckPeriodID              *int      `bun:"check_period_id,nullzero"`
	NotifyPeriodID             *int      `bun:"notify_period_id,nullzero"`
	Tags                       *string   `bun:"tags,nullzero"`
	OwnContacts                int       `bun:"own_contacts,notnull,default:0"`
	OwnContactgroups           int       `bun:"own_contactgroups,notnull,default:0"`
	OwnCustomvariables         int       `bun:"own_customvariables,notnull,default:0"`
	HostURL                    *string   `bun:"host_url,nullzero"`
	SlaID                      *int      `bun:"sla_id,nullzero"`
	SatelliteID                int       `bun:"satellite_id,default:0"`
	HostType                   int       `bun:"host_type,notnull,default:1"`
	Disabled                   int       `bun:"disabled,default:0"`
	UsageFlag                  int       `bun:"usage_flag,notnull"`
	Created                    time.Time `bun:"created,notnull"`
	Modified                   time.Time `bun:"modified,notnull"`
}
